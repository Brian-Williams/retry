// retry is an opinionated, but unassuming retrying library
//
// By default it will retry the provided function forever until the retried function returns `nil` with no pause
// in-between executions. It always runs the function at least once.
//
// Retry has 3 metrics for retrying logic: retry condition, stop conditions, wait time
// Optionally you may save the error history and inspect that post execution.
//
// The retry and stop conditions are functionally inverted, but the same. They are separate only so that stop conditions
// can be changed without needing to do || logic to continue retrying on error. Technically the retry condition isn't
// required and only stop conditions could be used, but that isn't how most people think about retrying and would
// increase the verbosity for simple cases.
// For example the simplest case of stopping after n attempts would change from the actual:
// 	`retry.Do(func() error{ return nil }, retry.StopMaxAttempts(n)`
// To the more verbose:
// 	`retry.Do(func() error{ return nil }, retry.StopOr(retry.StopIfNoError(), retry.StopMaxAttempts(n))`)
// Both of the above work with the actual API, but the second would be required if there wasn't a default retry
// condition. Because a dummmy function that always returns nil is passed in both of the above will always run once. IF
// a real function was run that could return a non-nil error it would run at least once, up to n times.
//
// If you want to instead retry until an error you may pass in `retry.IfNoError()`:
// 	`retry.Do(func() error{ return nil}, retry.IfNoError(), retry.StopMaxAttempts(n)`
//	Because of the always nil return this will run n times when n > 0.
//
// Because the assumed common case of retrying on error logic if you have a complex case you may pass in
// `retry.Always()` and then do all halt conditional logic in the stop combinatorial function.
// 	`retry.Do(func() error{ return nil }, retry.Always(), retry.StopOr(retry.StopIfError(), retry.StopMaxAttempts(5))`
package retry
