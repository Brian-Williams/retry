package retry_test

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"log"

	"bufio"

	"bytes"
	"github.com/Brian-Williams/retry"
	"github.com/pkg/errors"
	"os/exec"
)

// ExampleDo_Get adds retrying to the ExampleGet in net/http/example_test.go
func ExampleDo_Get() {
	errs := retry.Do(
		func() error {
			res, err := http.Get("http://www.google.com/robots.txt")
			if err != nil {
				return err
			}
			r := bufio.NewReader(res.Body)
			robots, err := r.ReadString('\n')
			res.Body.Close()
			if err != nil {
				return err
			}
			fmt.Printf("%s", robots)
			return nil
		},
		retry.StopOr(retry.StopMaxAttempts(5), retry.StopIfNoError()),
	)
	if errs != nil {
		log.Fatal(errs)
	}
	// Output:
	// User-agent: *
}

// WaitFixed does not adjust for total time, so this is only counting the delta time and not the total time
// If this was written to calculate total time the time would likely round over the durationUnit
func ExampleDoWithHistory_WaitFixed() {
	durationUnit := time.Millisecond * 20
	pit := time.Now()
	fmt.Printf("The first execution will be immediate than it will pause 10 * %s inbetween executions\n", durationUnit)

	h, _ := retry.DoWithHistory(
		func() error {
			now := time.Now()
			elapsed := now.Sub(pit)
			fmt.Printf("ran after ~%s\n", elapsed.Round(durationUnit))
			pit = now
			return errors.New("Always fails")
		},
		retry.StopOr(retry.StopMaxAttempts(3), retry.StopIfNoError()),
		retry.WaitFixed(10*durationUnit),
		retry.Save(retry.AllStates),
	)
	if h != nil {
		fmt.Println(h)
	}
	// Output:
	// The first execution will be immediate than it will pause 10 * 20ms inbetween executions
	// ran after ~0s
	// ran after ~200ms
	// ran after ~200ms
	// Saved errors:
	// #1: Always fails
	// #2: Always fails
	// #3: Always fails
}

func ExampleDo_MaxAttempts() {
	i := 1
	errs := retry.Do(
		func() error {
			fmt.Println(i)
			i++
			return nil
		},
		retry.StopMaxAttempts(3),
	)
	if errs != nil {
		//log.Fatal(errs)
	}
	// Output:
	// 1
	// 2
	// 3
}

func ExampleDo_GoRoutine() {
	i := 1
	var wg sync.WaitGroup
	retry.Do(
		func() error {
			wg.Add(1)
			go func(in int) {
				time.Sleep(time.Millisecond * time.Duration(in))
				fmt.Println(in)
				wg.Done()
			}(i)
			i++
			return nil
		},
		retry.StopMaxAttempts(3),
	)
	wg.Wait()
	// Output:
	// 1
	// 2
	// 3
}

func ExampleDo_Cmd() {
	args := []string{"true", "true", "false"}

	i := 0
	retry.Do(
		func() error {
			c := exec.Command(args[i])
			var out bytes.Buffer
			c.Stdout = &out
			err := c.Run()
			fmt.Println(args[i])
			i++
			return err
		},
		retry.StopOr(retry.StopIfError(), retry.StopMaxAttempts(5)),
	)
	// Output:
	// true
	// true
	// false
}
