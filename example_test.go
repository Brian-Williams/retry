package retry_test

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"log"

	"bufio"

	"bytes"
	"context"
	"github.com/Brian-Williams/retry"
	"os/exec"
)

// ExampleDo_Get adds retrying to the ExampleGet in net/http/example_test.go
func ExampleDo_get() {
	_, err := retry.Do(
		context.TODO(),
		func(ctx context.Context) error {
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
		retry.StopMaxAttempts(5),
	)
	if err != nil {
		log.Fatal(err)
	}
	// Output:
	// User-agent: *
}

// WaitFixed does not adjust for total time, so this is only counting the delta time and not the total time
// If this was written to calculate total time the time would likely round over the durationUnit
func ExampleDoWithHistory_waitFixed() {
	durationUnit := time.Millisecond * 20
	pit := time.Now()
	fmt.Printf("The first execution will be immediate than it will pause 10 * %s inbetween executions\n", durationUnit)

	h, _ := retry.Do(
		context.TODO(),
		func(ctx context.Context) error {
			now := time.Now()
			elapsed := now.Sub(pit)
			fmt.Printf("ran after ~%s\n", elapsed.Round(durationUnit))
			pit = now
			return fmt.Errorf("Always fails")
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
	// saved errors:
	// #1: attempt 1: Always fails
	// #2: attempt 2: Always fails
	// #3: attempt 3: Always fails
}

func ExampleDo_maxAttempts() {
	i := 1
	_, err := retry.Do(
		context.TODO(),
		func(ctx context.Context) error {
			fmt.Println(i)
			i++
			return fmt.Errorf("failed on attempt %d", i-1)
		},
		retry.StopMaxAttempts(3),
	)
	if err != nil {
		// handle the error
		fmt.Println(err)
	}
	// Output:
	// 1
	// 2
	// 3
	// failed on attempt 3
}

func ExampleDo_goRoutine() {
	i := 1
	var wg sync.WaitGroup
	retry.Do(
		context.TODO(),
		func(ctx context.Context) error {
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
}

func ExampleDo_cmd() {
	args := []string{"false", "false", "true"}

	i := 0
	retry.Do(
		context.TODO(),
		func(ctx context.Context) error {
			c := exec.Command(args[i])
			var out bytes.Buffer
			c.Stdout = &out
			err := c.Run()
			fmt.Println(args[i])
			i++
			return err
		},
		retry.StopMaxAttempts(5),
	)
	// Output:
	// false
	// false
	// true
}

func ExampleDo_cmdUntilFailure() {
	args := []string{"true", "true", "false"}

	i := 0
	retry.Do(
		context.TODO(),
		func(ctx context.Context) error {
			c := exec.Command(args[i])
			var out bytes.Buffer
			c.Stdout = &out
			err := c.Run()
			fmt.Println(args[i])
			i++
			return err
		},
		retry.Always(),
		retry.StopOr(retry.StopIfError(), retry.StopMaxAttempts(5)),
	)
	// Output:
	// true
	// true
	// false
}

// This custom StopFunc shows how we can stop if an error occurs, but there is no history saved an the following
// execution would bury the current error.
//
// Without this StopFunc the error would've caused an additional retry, however since we didn't use
// `retry.Save(AllStates)` and no History is being saved it exits
func ExampleStopFunc_custom() {
	i := 0
	_, err := retry.Do(
		context.TODO(),
		func(ctx context.Context) error {
			i++
			fmt.Println(i)
			return fmt.Errorf("failed on attempt %d", i)
		},
		retry.StopFunc(
			func(s retry.State) bool {
				if s.Hist != nil {
					return false
				}
				fmt.Println("will not continue because no history is being saved")
				return true
			},
		),
	)
	fmt.Println(err)
	// Output:
	// 1
	// will not continue because no history is being saved
	// failed on attempt 1
}
