package retry

import (
	"fmt"
	"strconv"
	"testing"

	"context"
	"time"
)

func TestAppendErr(t *testing.T) {
	var e error
	e = appendErr(e, fmt.Errorf("error1"))
	if e.Error() != "saved errors:\n#1: error1" {
		t.Log(e.Error())
		t.Error("actual and expected err didn't match")
	}
}

func TestSave(t *testing.T) {
	t.Run("negative test: bad enum panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("successfully paniced on bad enum: %v", r)
			} else {
				t.Errorf("failed to panic for bad enum: %v", r)
			}
		}()
		Save(testErrors)
	})

	absOutput := []error{
		Error{1, fmt.Errorf("error 1; return 1")},
		Error{2, nil},
		Error{3, fmt.Errorf("error 2; return 3")},
		Error{4, nil},
	}
	var absOutputHist History = absOutput
	fmt.Println(absOutputHist)
	var allErrorsOutput []error
	for _, err := range absOutput {
		if err.(Error).Unwrap() != nil {
			allErrorsOutput = append(allErrorsOutput, err)
		}
	}
	lastError := allErrorsOutput[len(allErrorsOutput)-1:]

	table := []struct {
		enum   errEnum
		output []error
	}{
		{AllStates, absOutput},
		{LastState, absOutput[len(absOutput)-1:]},
		{NoStates, []error{nil}},
		{NoErrors, []error{nil}},
		{AllErrors, allErrorsOutput},
		{LastError, lastError},
	}
	for _, test := range table {
		t.Run(fmt.Sprintf("Save enum: %s", test.enum),
			func(t *testing.T) {
				i := 0
				actual, _ := Do(
					context.TODO(),
					func(context.Context) error {
						r := absOutput[i].(Error).Unwrap()
						i++
						return r
					},
					Always(),
					Save(test.enum),
					StopMaxAttempts(uint(len(absOutput))),
				)

				if merr, ok := actual.(History); ok {
					for i, err := range merr {
						if err != test.output[i] {
							t.Errorf("for enum '%s' expected '%v' at position %d got '%v'", test.enum, test.output[i], i, err)
						}
					}
					if t.Failed() {
						t.Logf("full expected '%v', full actual '%v'", test.output, actual)
					}
				} else {
					if test.output[0] != actual {
						t.Errorf("for enum '%s' expected '%v' got '%v'", test.enum, test.output[0], actual)
					}
				}
			},
		)
	}
}

func TestErrors_Error(t *testing.T) {
	_, err := Do(
		context.TODO(),
		func(ctx context.Context) error {
			return nil
		}, StopMaxAttempts(5),
	)
	if err != nil {
		fmt.Println(err)
		t.Errorf("No errors shouldn't have failed")
	}
}

func TestConfig(t *testing.T) {
	c := Config()
	nilState := State{}
	errState := State{Err: fmt.Errorf("dogs are pretty great")}
	states := []State{nilState, errState}
	for _, s := range states {
		if c.Stop(s) != false {
			t.Errorf("received unexpected stop condition on config with err: %s", s.Err)
		}
		if c.Wait(s) != 0 {
			t.Errorf("received non-zero wait time for config with err: %s", s.Err)
		}
	}
}

var maxAttemptTable = []struct {
	attempts uint
	output   string
}{
	{0, "1"},
	{1, "1"},
	{2, "12"},
	{5, "12345"},
}

func maxAttemptsStr(opts ...Configurer) string {
	i := 1
	out := ""
	opts = append(opts, Always())
	Do(
		context.TODO(),
		func(ctx context.Context) error {
			out = fmt.Sprintf("%s%s", out, strconv.Itoa(i))
			i++
			return nil
		},
		opts...,
	)
	return out
}

func TestMaxAttempts(t *testing.T) {
	for _, test := range maxAttemptTable {
		out := maxAttemptsStr(StopMaxAttempts(test.attempts))
		if out != test.output {
			t.Errorf("expected '%s' got '%s'", test.output, out)
		}
	}
}

func TestOr(t *testing.T) {
	for _, test := range maxAttemptTable {
		out := maxAttemptsStr(StopOr(StopMaxAttempts(test.attempts+5), StopMaxAttempts(test.attempts)))
		if out != test.output {
			t.Errorf("expected '%s' got '%s'", test.output, out)
		}
	}
}

func TestAnd(t *testing.T) {
	for _, test := range maxAttemptTable {
		t.Run(fmt.Sprintf("running StopAnd tests for %d attempts", test.attempts), func(t *testing.T) {
			var delta uint = 1
			if test.attempts <= 0 {
				delta = 0
			}
			out := maxAttemptsStr(StopAnd(StopMaxAttempts(test.attempts), StopMaxAttempts(test.attempts-delta)))
			if out != test.output {
				t.Errorf("expected '%s' got '%s'", test.output, out)
			}
		})
	}
}

func TestWaitFixed(t *testing.T) {
	c := &config{}
	input := time.Minute
	WaitFixed(input).Configure(c)
	if c.Wait(State{}) != time.Minute {
		t.Errorf("wait expected '%s' actual: '%s'", time.Minute, c.Wait(State{}))
	}
}

var antiCompiler error

func benchmarkDo(b *testing.B, attempts uint) {
	var r error
	opts := StopMaxAttempts(attempts)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, r = Do(
			context.TODO(),
			func(ctx context.Context) error {
				return nil
			}, opts)
	}
	antiCompiler = r
}

func benchmarkDoWithConfigurer(b *testing.B, attempts uint) {
	opt := StopMaxAttempts(attempts)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config := Config()
		opt.Configure(config)
		sequentialLoop(
			context.TODO(),
			func(ctx context.Context) error {
				return nil
			}, config)
	}
}

func BenchmarkDo1(b *testing.B) {
	benchmarkDo(b, 1)
}

func BenchmarkDoWithConfigurer1(b *testing.B) {
	benchmarkDoWithConfigurer(b, 1)
}

func BenchmarkDo2(b *testing.B) {
	benchmarkDo(b, 2)
}

func BenchmarkDoWithConfigurer2(b *testing.B) {
	benchmarkDoWithConfigurer(b, 2)
}

func BenchmarkDo10(b *testing.B) {
	benchmarkDo(b, 10)
}

func BenchmarkDoWithConfigurer10(b *testing.B) {
	benchmarkDoWithConfigurer(b, 10)
}

func BenchmarkDo100(b *testing.B) {
	benchmarkDo(b, 100)
}

func BenchmarkDoWithConfigurer100(b *testing.B) {
	benchmarkDoWithConfigurer(b, 100)
}

func BenchmarkDo10000(b *testing.B) {
	benchmarkDo(b, 10000)
}

func BenchmarkDoWithConfigurer10000(b *testing.B) {
	benchmarkDoWithConfigurer(b, 10000)
}
