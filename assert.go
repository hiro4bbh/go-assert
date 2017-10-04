package goassert

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
)

// TestingTB is an interface mimicking testing.TB (except for Skip*) interface which prevents users to implement itself.
// See testing.TB for details.
type TestingTB interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fail()
	FailNow()
	Failed() bool
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	Name() string
	Helper()
}

// HookedTestingTB implements TestingTB.
// This is designed to be used in testing test frameworks.
// See TestingTB for details if methods are not documented.
type HookedTestingTB struct {
	// Messages is the slice of the logged messages.
	Messages []string
	// Helpers is the slice of the registered helper functions.
	// Helper functions are identified by string "file:line".
	Helpers []string
	// name is the name of TestingTB for method Name.
	name string
	// failed indicates whether the current test has failed already or not.
	failed bool
}

// NewHookedTestingTB returns a new HookedTestingTB.
func NewHookedTestingTB(name string) *HookedTestingTB {
	return &HookedTestingTB{
		Messages: []string{},
		Helpers: []string{},
		name: name,
	}
}

func (tb *HookedTestingTB) Error(args ...interface{}) {
	tb.Log(append([]interface{}{"ERROR: "}, args...)...)
	tb.Fail()
}

func (tb *HookedTestingTB) Errorf(format string, args ...interface{}) {
	tb.Logf("ERROR: " + format, args...)
	tb.Fail()
}

func (tb *HookedTestingTB) Fail() {
	tb.failed = true
}

// FailNow panics for stopping the current test execution.
func (tb *HookedTestingTB) FailNow() {
	tb.Fail()
	panic(fmt.Sprintf("HookedTestingTB(%q): FAIL NOW", tb.name))
}

func (tb *HookedTestingTB) Failed() bool {
	return tb.failed
}

func (tb *HookedTestingTB) Fatal(args ...interface{}) {
	tb.Log(append([]interface{}{"FATAL: "}, args...)...)
	tb.FailNow()
}

func (tb *HookedTestingTB) Fatalf(format string, args ...interface{}) {
	tb.Logf("FATAL: " + format, args...)
	tb.FailNow()
}

func (tb *HookedTestingTB) Helper() {
	_, file, line, _ := runtime.Caller(1)
	tb.Helpers = append(tb.Helpers, fmt.Sprintf("%s:%d", file, line))
}

func (tb *HookedTestingTB) Log(args ...interface{}) {
	tb.Messages = append(tb.Messages, fmt.Sprint(args...))
}

func (tb *HookedTestingTB) Logf(format string, args ...interface{}) {
	tb.Messages = append(tb.Messages, fmt.Sprintf(format, args...))
}

func (tb *HookedTestingTB) Name() string {
	return tb.name
}

// Assert is an assertion wrapper.
type Assert struct {
	tb       TestingTB
	expected []interface{}
}

// NewAssert returns a new Assert with the testing context and the expected values.
func New(tb TestingTB, expected ...interface{}) *Assert {
	return &Assert{
		tb:       tb,
		expected: expected,
	}
}

// Equal checks that the given actual values equals the expected values.
func (assert *Assert) Equal(actual ...interface{}) {
	assert.tb.Helper()
	if len(assert.expected) != len(actual) {
		assert.tb.Fatalf("expected %d value(s), but got %d value(s)", len(assert.expected), len(actual))
	} else {
		str := ""
		for i, expected := range assert.expected {
			if !reflect.DeepEqual(expected, actual[i]) {
				if str != "" {
					str += "\n"
				}
				str += fmt.Sprintf("at #%d value, expected %#v (%T), but got %#v (%T)", i, expected, expected, actual[i], actual[i])
			}
		}
		if str != "" {
			assert.tb.Errorf("%s", str)
		}
	}
}

// EqualWithoutError checks that the given actual values equals the expected values without any error.
func (assert *Assert) EqualWithoutError(actual_err ...interface{}) {
	assert.tb.Helper()
	if len(actual_err) < 2 {
		assert.tb.Fatalf("actual_err must be at least two: (actual..., err)")
	}
	err := actual_err[len(actual_err)-1]
	if err != nil {
		assert.tb.Fatalf("unexpected error: %s", err)
	}
	assert.Equal(actual_err[0 : len(actual_err)-1]...)
}

// ExpectError checks that the error is returned expectedly.
// The expected values must be none or one error pattern string.
func (assert *Assert) ExpectError(_err ...interface{}) {
	assert.tb.Helper()
	if len(_err) < 1 {
		assert.tb.Fatalf("_err must be at least one: (_..., err)")
	}
	err, ok := _err[len(_err)-1].(error)
	if !ok && _err[len(_err)-1] != nil {
		assert.tb.Fatalf("the last element of _err must be error, but got %T", _err[len(_err)-1])
	}
	if err == nil {
		assert.tb.Fatalf("expected an error, but got no error")
		return
	}
	if len(assert.expected) > 0 {
		if len(assert.expected) != 1 {
			assert.tb.Fatalf("the number of error pattern must be at most one")
			return
		}
		pattern, ok := assert.expected[0].(string)
		if !ok {
			assert.tb.Fatalf("error pattern must be string")
			return
		}
		if matched, e := regexp.MatchString(pattern, err.Error()); e != nil {
			assert.tb.Fatalf("malformed expected error pattern: %s", e)
			return
		} else if !matched {
			assert.tb.Fatalf("expected error pattern %q, but got error %q", pattern, err)
			return
		}
	}
}

// SuccessNew checks that New-style function succeeds without any error.
func (assert *Assert) SucceedNew(o interface{}, err error) interface{} {
	assert.tb.Helper()
	if err != nil {
		assert.tb.Fatalf("unexpected error in New-style function: %s", err)
	}
	return o
}

// SuccessWithoutError check that the function succeeds without any error.
func (assert *Assert) SucceedWithoutError(err error) {
	assert.tb.Helper()
	if err != nil {
		assert.tb.Errorf("unexpected error: %s", err)
	}
}
