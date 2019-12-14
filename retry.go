package retry

import (
	"fmt"
	"strings"
	"time"
)

// Func is a function that can be retried
type Func func() error

// Do is a general retry tool. It will retry forever, but can be modified to your needs.
//
// Do is the entry point for retrying. It will always run a function at least once.
//
// If you want to combine multiple configurable options use one of the combination helpers like "StopOr".
// If you pass in multiple configurations for the same configurable option the *last* configuration will be the one that
// takes effect. There is not a valid use case for doing that and may panic in the future.
func Do(f Func, opts ...Configurer) error {
	config := Config()

	for _, opt := range opts {
		opt.Configure(config)
	}

	_, err :=  doWithConfigurer(f, config)
	return err
}

func DoWithHistory(f Func, opts ...Configurer) (history, err error) {
	config := Config()

	for _, opt := range opts {
		opt.Configure(config)
	}

	return doWithConfigurer(f, config)
}

func doWithConfigurer(f Func, config *config) (history error, err error) {
	// TODO: is it valid that we should give a stop option for the first run
	// check if we should start any runs
	//if config.stop(State{attempt: 0, err: nil}) {
	//	return config.error
	//}

	var attempt uint = 1
	var errs error

	for {
		e := State{attempt: attempt, err: nil, errs: errs}
		e.err = f()
		e.errs = config.save(e)
		if config.stop(e) {
			return e.errs, e.err
		}
		time.Sleep(config.wait(e))
		errs = e.errs
		attempt++
	}
}

// CONFIGURATION

type Configurer interface {
	Configure(config *config)
}

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

// Stop options

func StopMaxAttempts(n uint) StopFunc {
	return func(s State) bool {
		if s.attempt >= n {
			return true
		}
		return false
	}

}

func StopIfError() StopFunc {
	return func(s State) bool {
		if s.err != nil {
			return true
		}
		return false
	}

}

func StopIfNoError() StopFunc {
	return func(s State) bool {
		if s.err != nil {
			return false
		}
		return true
	}
}

// stopOnError compares the most recent error to a desired error and stops only when those errors match
func stopOnError(err error) StopFunc {
	return func(s State) bool {
		if s.err == err {
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

type History []error

// Error method return string representation of Error
// It is an implementation of error interface
func (e History) Error() string {
	errLog := make([]string, len(e))
	for i, l := range e {
		errmsg := "<nil>"
		if l != nil {
			errmsg = l.Error()
		}
		errLog[i] = fmt.Sprintf("#%d: %s", i+1, errmsg)
	}

	return fmt.Sprintf("Saved errors:\n%s", strings.Join(errLog, "\n"))
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

	// AllErrors should only be used when you want to do simple comparisons against the returned error slice.
	// It loses the connection between the retry number and the state, so in general AllStates is a better option.
	AllErrors
	LastError

	NoErrors = NoStates

	// testErrors is for negative testing
	testErrors errEnum = 999
)

// TODO: it's slightly strange to have the error configurer use an enum when none of the other Configurerers do, this
// might be okay simply to stop users from configuring errors more than once.
// It's possible this is just a naming problem though and the SaveFunc should have something more meaninfully named verb
// like 'save'. 'SaveAllStates' makes a lot more sense than 'ErrorAllStates' as an input
func Save(errProperty errEnum) SaveFunc {
	switch errProperty {
	//default:
	//	fallthrough
	case AllStates:
		return func(s State) error {
			//s.errs = multierror.Append(s.errs, s.err)
			return appendErr(s.errs, s.err)
		}
	case LastState:
		return func(e State) error {
			return e.err
		}
	case NoStates:
		return func(_ State) error {
			return nil
		}
	case AllErrors:
		return func(e State) error {
			if e.err != nil {
				//e.errs = multierror.Append(e.errs, e.err)
				return appendErr(e.errs, e.err)
			}
			return e.errs
		}
	case LastError:
		return func(e State) error {
			if e.err != nil {
				return e.err
			}
			return e.errs
		}
	}
	panic(fmt.Sprintf("unrecognized enum for error configuration: '%+v'", errProperty))
}

func Config() *config {
	c := &config{
		stop: func(_ State) bool {
			return false
		},
		wait: func(s State) time.Duration {
			return 0
		},
	}
	Save(NoStates).Configure(c)
	return c
}

// config configures which functions to use for retrying logic
type config struct {
	_    struct{}
	stop StopFunc
	wait WaitFunc
	save SaveFunc
}

// State provides the execution info for a single retry attempt
type State struct {
	_ struct{}
	// attempt is the current number of attempts
	// attempt is effectively 1 indexed as it will only be 0 before the `Func` to be retried is called
	attempt uint
	// err is the current iterations error
	err error
	// errs
	errs error
}

// COMBINATION UTILITIES

func StopOr(fs ...StopFunc) StopFunc {
	return func(s State) bool {
		var b bool
		for _, sf := range fs {
			b = b || sf(s)
		}
		return b
	}
}

func StopAnd(fs ...StopFunc) StopFunc {
	return func(s State) bool {
		var b = true
		for _, sf := range fs {
			b = b && sf(s)
		}
		return b
	}
}
