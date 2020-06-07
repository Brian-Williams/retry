package concurrent

import (
	"context"
	"sync"
	"time"
	"github.com/Brian-Williams/retry"
)

// Concurrent is a general retry function.
//
// Concurrent is the entry point for retrying. It will always run a function at least once.
//
// If combining multiple configurable options use one of the combination helpers like "StopOr".
// If multiple configurations are passed in for the same configurable option the *last* configuration will be the one
// that takes effect.
func Concurrent(ctx context.Context, f retry.Func, opts ...retry.Configurer) (history, err error) {
	config := retry.Config()

	for _, opt := range opts {
		opt.Configure(config)
	}

	var attempt uint = 1
	var errs error
	var wg sync.WaitGroup

	for {
		e := retry.State{Attempt: attempt, Err: nil, Hist: errs}

		// e.Err = f(ctx)
		wg.Add(1)
		c := make(chan error, 1)
		go func() {
			defer wg.Done()
			c <- f(ctx)
		}()
		select {
		case <-ctx.Done():
			e.Err = ctx.Err()
			config.Retry = retry.Never()
		case err := <-c:
			e.Err = err
		}

		e.Hist = config.Save(e)
		if !config.Retry(e.Err) {
			return e.Hist, e.Err
		}
		if config.Stop(e) {
			return e.Hist, e.Err
		}

		// time.Sleep(config.wait(e))
		select {
		case <-ctx.Done():
			// This is the only error that doesn't go into Hist
			return e.Hist, ctx.Err()
			// TODO: should the WaitFunc simply return a channel?
			// If WaitFunc is still a Duration, should it short circuit based on ctx.Deadline?
		case <-time.After(config.Wait(e)):
		}

		errs = e.Hist
		attempt++
	}
}
