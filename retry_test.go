package retry

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/pkg/errors"
	"time"
)

func TestAppendErr(t *testing.T) {
	var e error
	e = appendErr(e, errors.New("error1"))
	if e.Error() != "Saved errors:\n#1: error1" {
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

	absOutput := []error{errors.New("error 1; return 1"), nil, errors.New("error 2; return 3"), nil}
	var allErrorsOutput []error
	for _, err := range absOutput {
		if err != nil {
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
				actual, _ := DoWithHistory(
					func() error {
						r := absOutput[i]
						i++
						return r
					},
					RetryAlways(),
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
	err := Do(
		func() error {
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
	errState := State{err: errors.New("dogs are pretty great")}
	states := []State{nilState, errState}
	for _, s := range states {
		if c.stop(s) != false {
			t.Errorf("received unexpected stop condition on config with err: %s", s.err)
		}
		if c.wait(s) != 0 {
			t.Errorf("received non-zero wait time for config with err: %s", s.err)
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
	opts = append(opts, RetryAlways())
	Do(
		func() error {
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
			var delta uint = 0
			//if test.attempts != 0 {
			//	delta = 1
			//}
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
	if c.wait(State{}) != time.Minute {
		t.Errorf("wait expected '%s' actual: '%s'", time.Minute, c.wait(State{}))
	}
}

var antiCompiler error

func benchmarkDo(b *testing.B, attempts uint) {
	var r error
	opts := StopMaxAttempts(attempts)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r = Do(func() error {
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
		doWithConfigurer(func() error {
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
