# Retry
Retry is a flexible retrying library for all your repeated needs.

As usual with retry examples in go, here is how to implement a maximum of 5
retries to the `ExampleGet` test from `net/http`.
```go
// ExampleDo_Get adds retrying to the ExampleGet in net/http/example_test.go
func ExampleDo_Get() {
	err := retry.Do(
		func() error {
			res, err := http.Get("http://www.google.com/robots.txt")
			if err != nil {
				return err
			}
			r := bufio.NewReader(res.Body)
      robots, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				return err
			}
			fmt.Printf("%s", robots)
		},
		retry.StopMaxAttempts(5),
	)
	if err != nil {
		log.Fatal(err)
	}
}
```

### Benchmarks
Benchmark history can be viewed [here](https://brian-williams.github.io/retry/dev/bench/)
