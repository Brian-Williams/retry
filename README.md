# retry
[![GoDoc][godoc image]][godoc]


Retry is a flexible retrying library for all your repeated needs.

As usual with retry examples in go, here is how to implement a maximum of 5
retries to the `ExampleGet` test from `net/http`.
```go
// ExampleDo_Get adds retrying to the ExampleGet in net/http/example_test.go
func ExampleDo_Get() {
	err := retry.Do(
		func() error {
			// copied from https://golang.org/pkg/net/http/#example_Get
			res, err := http.Get("http://www.google.com/robots.txt")
			if err != nil {
				return err
			}
            robots, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				return err
			}
			fmt.Printf("%s", robots)
			// by default retry will halt if a nil error is returned
			return nil
		},
		// stop retrying after 5 attempts
		retry.StopMaxAttempts(5),
	)
	if err != nil {
		log.Fatal(err)
	}
}
```
Godoc has more [examples].

### Contributing
I'm happy to accept PRs for your use case.

If you are making a large change or extension consider opening an issue
first, so that we can discuss the best way to add your change!

### Benchmarks
Features and usability will always take priority.

[Benchmarks] are kept primarily to confirm the library is fit for purpose
and not degrading.

[godoc]: https://godoc.org/github.com/Brian-Williams/retry
[godoc image]: https://godoc.org/github.com/Brian-Williams/retry?status.png
[examples]: https://godoc.org/github.com/Brian-Williams/retry#pkg-examples
[Benchmarks]: https://brian-williams.github.io/retry/dev/bench/
