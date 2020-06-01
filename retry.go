package retry

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"context"
)

// Func is a function that can be retried
type Func func(ctx context.Context) error

// Do is a general retry function.
//
// Do is the entry point for retrying. It will always run a function at least once.
//
// If combining multiple configurable options use one of the combination helpers like "StopOr".
// If multiple configurations are passed in for the same configurable option the *last* configuration will be the one
// that takes effect.
func Do(ctx context.Context, f Func, opts ...Configurer) error {
	config := Config()

	for _, opt := range opts {
		opt.Configure(config)
	}

	_, err := doWithConfigurer(ctx, f, config)
	return err
}

func DoWithHistory(ctx context.Context, f Func, opts ...Configurer) (history, err error) {
	config := Config()

	for _, opt := range opts {
		opt.Configure(config)
	}

	return doWithConfigurer(ctx, f, config)
}

func doWithConfigurer(ctx context.Context, f Func, config *config) (history error, err error) {
	var attempt uint = 1
	var errs error

	for {
		e := State{Attempt: attempt, Err: nil, Hist: errs}
		e.Err = f(ctx)
		e.Hist = config.save(e)
		if !config.retry(e.Err) {
			return e.Hist, e.Err
		}
		if config.stop(e) {
			return e.Hist, e.Err
		}
		time.Sleep(config.wait(e))
		errs = e.Hist
		attempt++
	}
}

// CONFIGURATION

type Configurer interface {
	Configure(config *config)
}

// RetryFunc is a simplified independent stop condition that only has access to the error
// It is technically not required, but simplifies the API for the most common use case of attempt until no error/success
type RetryFunc func(err error) bool

// StopFunc is a function that stops on true and continues retrying on false
type StopFunc func(e State) bool

// WaitFunc is a function that tells us how long to wait inbetween retry attempts
//
// It is important to note that this is not an absolute time between retry attempt *starts* it is the time inbetween one
// retry attempt ending and the next one starting.
type WaitFunc func(e State) time.Duration

// SaveFunc is a function that handles saving/storage of errors
type SaveFunc func(e State) error

// Configure satisfies the Configurer interface
func (f RetryFunc) Configure(config *config) {
	config.retry = f
}

// Configure satisfies the Configurer interface
func (f StopFunc) Configure(config *config) {
	config.stop = f
}

// Configure satisfies the Configurer interface
func (f WaitFunc) Configure(config *config) {
	config.wait = f
}

// Configure satisfies the Configurer interface
func (f SaveFunc) Configure(config *config) {
	config.save = f
}

// Retry options

func retryIfErrorFunc(err error) bool {
	if err != nil {
		return true
	}
	return false
}

func IfError() RetryFunc {
	return retryIfErrorFunc
}

func IfNoError() RetryFunc {
	return func(err error) bool {
		if err != nil {
			return false
		}
		return true
	}
}

func Always() RetryFunc {
	return func(error) bool {
		return true
	}
}

// Stop options

func StopMaxAttempts(n uint) StopFunc {
	return func(s State) bool {
		if s.Attempt >= n {
			return true
		}
		return false
	}
}

func StopIfError() StopFunc {
	return func(s State) bool {
		if s.Err != nil {
			return true
		}
		return false
	}
}

func StopIfNoError() StopFunc {
	return func(s State) bool {
		if s.Err != nil {
			return false
		}
		return true
	}
}

// stopOnError compares the most recent error to a desired error and stops only when those errors match
func stopOnError(err error) StopFunc {
	return func(s State) bool {
		if s.Err == err {
			return true
		}
		return false
	}
}

// Wait options

func WaitFixed(delay time.Duration) WaitFunc {
	return func(_ State) time.Duration {
		return delay
	}
}

// Error options

type Error struct {
	Attempt uint
	Err     error
}

func (e Error) Unwrap() error { return e.Err }

func (e Error) Error() string {
	return fmt.Sprintf("attempt %d: %v", e.Attempt, e.Err)
}

type History []error

// Error method return string representation of Error
// It is an implementation of error interface
func (e History) Error() string {
	errLog := make([]string, len(e))
	for i, l := range e {
		var errmsg string = "<nil>"
		if l != nil {
			errmsg = l.Error()
		}
		errLog[i] = fmt.Sprintf("#%d: %s", i+1, errmsg)
	}

	return fmt.Sprintf("saved errors:\n%s", strings.Join(errLog, "\n"))
}

func appendErr(e error, err error) History {
	rErr, ok := e.(History)
	if !ok {
		rErr = History{err}
		return rErr
	}
	rErr = append(rErr, err)
	return rErr
}

type errEnum int

// errEnums for how to store errors during retries
// If an errEnum is added please update Save
//
//go:generate stringer -type=errEnum
const (
	// AllStates will save both non-nil and nil errors
	AllStates errEnum = iota
	LastState
	NoStates

	// AllErrors should only be used when doing simple comparisons against the returned error slice.
	// It loses the connection between the retry number and the state, so in general AllStates is a better option.
	AllErrors
	LastError

	NoErrors = NoStates

	// testErrors is for negative testing
	testErrors errEnum = 999
)

func noStatesFunc(_ State) error {
	return nil
}

// Save decides what parts of the error history to save and return to the user.
//
// Save can be used with `DoWithHistory`.
// Save can also be used if any Configure function wants to inspect previous executions
func Save(errProperty errEnum) SaveFunc {
	switch errProperty {
	//default:
	//	fallthrough
	case AllStates:
		return func(s State) error {
			return appendErr(s.Hist, Error{s.Attempt, s.Err})
		}
	case LastState:
		return func(s State) error {
			return Error{s.Attempt, s.Err}
		}
	case NoStates:
		return noStatesFunc
	case AllErrors:
		return func(s State) error {
			if s.Err != nil {
				return appendErr(s.Hist, Error{s.Attempt, s.Err})
			}
			return s.Hist
		}
	case LastError:
		return func(s State) error {
			if s.Err != nil {
				return Error{s.Attempt, s.Err}
			}
			return s.Hist
		}
	}
	panic(fmt.Sprintf("unrecognized enum for error configuration: '%+v'", errProperty))
}

// Config provides a config for retry's Do functions
//
// The returned config retries on non-nil errors, has no stop condition, 0 wait time, and will not save any history.
func Config() *config {
	return &config{
		retry: retryIfErrorFunc,
		stop: func(_ State) bool {
			return false
		},
		wait: func(_ State) time.Duration {
			return 0
		},
		save: noStatesFunc,
	}
}

// config configures which functions to use for retrying logic
type config struct {
	_     struct{}
	retry RetryFunc
	stop  StopFunc
	wait  WaitFunc
	save  SaveFunc
}

// State provides the execution info for a single retry attempt
type State struct {
	_ struct{}
	// Attempt is the current number of attempts
	Attempt uint
	// Err is the current iterations error
	Err error
	// Hist is the History of the execution
	Hist error
}

// COMBINATION UTILITIES

// StopOr takes two or more StopFunc and returns a StopFunc that stops if any is true
func StopOr(fs ...StopFunc) StopFunc {
	return func(s State) bool {
		var b bool
		for _, sf := range fs {
			b = b || sf(s)
		}
		return b
	}
}

// StopAnd takes two or more StopFunc and returns a StopFunc that stops if all are true
func StopAnd(fs ...StopFunc) StopFunc {
	return func(s State) bool {
		var b = true
		for _, sf := range fs {
			b = b && sf(s)
		}
		return b
	}
}

// WRAPPING UTILITIES

// UnwrappedError
var NotWrappedError = errors.New("error doesn't have an 'Unwrap' method")

// Unwrap is like the standard libraries Unwrap function except it returns a `NotWrappedError` in the case of being
// handed an error without an Unwrap method instead of returning nil
func Unwrap(err error) error {
	u, ok := err.(interface {
		Unwrap() error
	})
	if !ok {
		return NotWrappedError
	}
	return u.Unwrap()
}
