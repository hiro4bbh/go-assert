package goassert

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestTestingTB(t *testing.T) {
	// Test whether TestingTB is subset of testing.TB.
	var tb testing.TB
	tb = t
	New(tb)
}

func TestHookedTestingTBNormalUse(t *testing.T) {
	// Only normal uses
	tb := NewHookedTestingTB("test")
	if expected, got := "test", tb.Name(); expected != got {
		t.Fatalf("test: expected %q, but got %q returned by Name()", expected, got)
	}
	tb.Log("hello", "world")
	tb.Logf("%s %s!", "hello", "world")
	tb.Helper()
	if failed := tb.Failed(); failed {
		t.Fatalf("test: unexpected Failed() == true")
	}
	if !reflect.DeepEqual(tb.Messages, []string{"helloworld", "hello world!"}) {
		t.Fatalf("test: unexpected Messages: %#v", tb.Messages)
	}
	if !(len(tb.Helpers) == 1 && strings.HasSuffix(tb.Helpers[0], "assert_test.go:25")) {
		t.Fatalf("test: unexpected Helpers: %#v", tb.Helpers)
	}
}

func TestHookedTestingTBContinueAfterError(t *testing.T) {
	// Error should not prevent the following execution
	tb := NewHookedTestingTB("test")
	tb.Error("hell", "world")
	if failed := tb.Failed(); !failed {
		t.Fatalf("test: unexpected Failed() == false")
	}
	tb.Logf("%s %s!!", "hell", "world")
	if failed := tb.Failed(); !failed {
		t.Fatalf("test: unexpected Failed() == false")
	}
	if !reflect.DeepEqual(tb.Messages, []string{"ERROR: hellworld", "hell world!!"}) {
		t.Fatalf("test: unexpected Messages: %#v", tb.Messages)
	}
}

func TestHookedTestingTBContinueAfterErrorf(t *testing.T) {
	// Error and Errorf should not prevent the following execution
	tb := NewHookedTestingTB("test")
	tb.Errorf("%s %s!!", "hell", "world")
	if failed := tb.Failed(); !failed {
		t.Fatalf("test: unexpected Failed() == false")
	}
	tb.Logf("%s %s!!", "hell", "world")
	if !reflect.DeepEqual(tb.Messages, []string{"ERROR: hell world!!", "hell world!!"}) {
		t.Fatalf("test: unexpected Messages: %#v", tb.Messages)
	}
}

func TestHookedTestingTBFatalPreventsExecution(t *testing.T) {
	// Fatal should prevent the following execution
	tb := NewHookedTestingTB("test")
	var panicObj interface{}
	func() {
		defer func() {
			panicObj = recover()
		}()
		tb.Fatal("HELL", "WORLD")
	}()
	if panicStr, ok := panicObj.(string); !(ok && panicStr == "HookedTestingTB(\"test\"): FAIL NOW") {
		t.Fatalf("test: unexpected panic message: %s", panicObj)
	}
	if failed := tb.Failed(); !failed {
		t.Fatalf("test: unexpected Failed() == false")
	}
	if !reflect.DeepEqual(tb.Messages, []string{"FATAL: HELLWORLD"}) {
		t.Fatalf("test: unexpected Messages: %#v", tb.Messages)
	}
}

func TestHookedTestingTBFatalfPreventsExecution(t *testing.T) {
	// Fatalf should prevent the following execution
	tb := NewHookedTestingTB("test")
	var panicObj interface{}
	func() {
		defer func() {
			panicObj = recover()
		}()
		tb.Fatalf("%s %s!!!", "HELL", "WORLD")
	}()
	if panicStr, ok := panicObj.(string); !(ok && panicStr == "HookedTestingTB(\"test\"): FAIL NOW") {
		t.Fatalf("test: unexpected panic message: %s", panicObj)
	}
	if failed := tb.Failed(); !failed {
		t.Fatalf("test: unexpected Failed() == false")
	}
	if !reflect.DeepEqual(tb.Messages, []string{"FATAL: HELL WORLD!!!"}) {
		t.Fatalf("test: unexpected Messages: %#v", tb.Messages)
	}
}

func TestAssertEqual(t *testing.T) {
	// test1: Test helper registration
	tb1 := NewHookedTestingTB("test1")
	New(tb1, "hello").Equal("hello")
	// NOTICE: We cannot identify the location of any helper exactly, because of go cover tool inserts some code into source code files.
	if len(tb1.Helpers) != 1 {
		t.Fatalf("test1: unexpected Helpers: %#v", tb1.Helpers)
	}
	// test2: only normal uses
	tb2 := NewHookedTestingTB("test2")
	New(tb2).Equal()
	New(tb2, "hello").Equal("hello")
	New(tb2, "hello", "world").Equal("hello", "world")
	New(tb2, "hello", "world").Equal(func() (string, string) {
		return "hello", "world"
	}())
	if tb2.Failed() {
		t.Fatalf("test2: unexpected Failed() == true")
	}
	if len(tb2.Messages) != 0 {
		t.Fatalf("test2: unexpected Messages: %#v", tb2.Messages)
	}
	// test3: error cases followed by one fatal exit cases
	tb3 := NewHookedTestingTB("test3")
	New(tb3, "hello").Equal("hell")
	New(tb3, "hello", "world").Equal("hell", "w0rld")
	func() {
		defer func() {
			if o := recover(); o != nil {
			}
		}()
		New(tb3).Equal("hello")
	}()
	if !tb3.Failed() {
		t.Fatalf("test3: unexpected Failed() == false")
	}
	if !reflect.DeepEqual(tb3.Messages, []string{
		"ERROR: at #0 value, expected \"hello\" (string), but got \"hell\" (string)",
		"ERROR: at #0 value, expected \"hello\" (string), but got \"hell\" (string)\nat #1 value, expected \"world\" (string), but got \"w0rld\" (string)",
		"FATAL: expected 0 value(s), but got 1 value(s)",
	}) {
		t.Fatalf("test3: unexpected Messages: %#v", tb3.Messages)
	}
}

func TestAssertEqualWithoutError(t *testing.T) {
	// test1: Test helper registration
	tb1 := NewHookedTestingTB("test1")
	New(tb1, "hello").EqualWithoutError("hello", error(nil))
	// NOTICE: We cannot identify the location of any helper exactly, because of go cover tool inserts some code into source code files.
	if len(tb1.Helpers) != 2 {
		t.Fatalf("test1: unexpected Helpers: %#v", tb1.Helpers)
	}
	// test2: only normal uses
	tb2 := NewHookedTestingTB("test2")
	New(tb2, "hello").EqualWithoutError("hello", error(nil))
	New(tb2, "hello", "world").EqualWithoutError("hello", "world", error(nil))
	New(tb2, "hello", "world").EqualWithoutError(func() (string, string, error) {
		return "hello", "world", nil
	}())
	if tb2.Failed() {
		t.Fatalf("test2: unexpected Failed() == true")
	}
	if len(tb2.Messages) != 0 {
		t.Fatalf("test2: unexpected Messages: %#v", tb2.Messages)
	}
	// test3: error cases followed by one fatal exit cases
	tb3 := NewHookedTestingTB("test3")
	New(tb3, "hello").EqualWithoutError("hell", error(nil))
	New(tb3, "hello", "world").EqualWithoutError("hell", "w0rld", error(nil))
	func() {
		defer func() {
			if o := recover(); o != nil {
			}
		}()
		New(tb3).EqualWithoutError("hello", fmt.Errorf("w0rld"))
	}()
	if !tb3.Failed() {
		t.Fatalf("test3: unexpected Failed() == false")
	}
	if !reflect.DeepEqual(tb3.Messages, []string{
		"ERROR: at #0 value, expected \"hello\" (string), but got \"hell\" (string)",
		"ERROR: at #0 value, expected \"hello\" (string), but got \"hell\" (string)\nat #1 value, expected \"world\" (string), but got \"w0rld\" (string)",
		"FATAL: unexpected error: w0rld",
	}) {
		t.Fatalf("test3: unexpected Messages: %#v", tb3.Messages)
	}
	// test4: fatal exit at no returning value
	tb4 := NewHookedTestingTB("test4")
	func() {
		defer func() {
			if o := recover(); o != nil {
			}
		}()
		New(tb4).EqualWithoutError(error(nil))
	}()
	if !tb4.Failed() {
		t.Fatalf("test4: unexpected Failed() == false")
	}
	if !reflect.DeepEqual(tb4.Messages, []string{"FATAL: actual_err must be at least two: (actual..., err)"}) {
		t.Fatalf("test4: unexpected Messages: %#v", tb4.Messages)
	}
}

func TestAssertExpectError(t *testing.T) {
	// test1: Test helper registration
	tb1 := NewHookedTestingTB("test1")
	New(tb1, "world").ExpectError("hello", fmt.Errorf("world"))
	// NOTICE: We cannot identify the location of any helper exactly, because of go cover tool inserts some code into source code files.
	if len(tb1.Helpers) != 1 {
		t.Fatalf("test1: unexpected Helpers: %#v", tb1.Helpers)
	}
	// test2: only normal uses
	tb2 := NewHookedTestingTB("test2")
	New(tb2).ExpectError(fmt.Errorf("hello"))
	New(tb2, "world").ExpectError("hello", fmt.Errorf("world"))
	New(tb2, "!").ExpectError(func() (string, string, error) {
		return "hello", "world", fmt.Errorf("!")
	}())
	if tb2.Failed() {
		t.Fatalf("test2: unexpected Failed() == true")
	}
	if len(tb2.Messages) != 0 {
		t.Fatalf("test2: unexpected Messages: %#v", tb2.Messages)
	}
	// test3: fatal exit cases
	tb3_1 := NewHookedTestingTB("test3_1")
	func() {
		defer func() {
			if o := recover(); o != nil {
			}
		}()
		New(tb3_1).ExpectError()
	}()
	if !tb3_1.Failed() {
		t.Fatalf("test3_1: unexpected Failed() == false")
	}
	if !reflect.DeepEqual(tb3_1.Messages, []string{"FATAL: _err must be at least one: (_..., err)"}) {
		t.Fatalf("test3_1: unexpected Messages: %#v", tb3_1.Messages)
	}
	tb3_2 := NewHookedTestingTB("test3_2")
	func() {
		defer func() {
			if o := recover(); o != nil {
			}
		}()
		New(tb3_2).ExpectError("hello", "world")
	}()
	if !tb3_2.Failed() {
		t.Fatalf("test3_2: unexpected Failed() == false")
	}
	if !reflect.DeepEqual(tb3_2.Messages, []string{"FATAL: the last element of _err must be error, but got string"}) {
		t.Fatalf("test3_2: unexpected Messages: %#v", tb3_2.Messages)
	}
	tb3_3 := NewHookedTestingTB("test3_3")
	func() {
		defer func() {
			if o := recover(); o != nil {
			}
		}()
		New(tb3_3).ExpectError(error(nil))
	}()
	if !tb3_3.Failed() {
		t.Fatalf("test3_3: unexpected Failed() == false")
	}
	if !reflect.DeepEqual(tb3_3.Messages, []string{"FATAL: expected an error, but got no error"}) {
		t.Fatalf("test3_3: unexpected Messages: %#v", tb3_3.Messages)
	}
	tb3_4 := NewHookedTestingTB("test3_4")
	func() {
		defer func() {
			if o := recover(); o != nil {
			}
		}()
		New(tb3_4, "hello", "world").ExpectError(fmt.Errorf("hello"))
	}()
	if !tb3_4.Failed() {
		t.Fatalf("test3_4: unexpected Failed() == false")
	}
	if !reflect.DeepEqual(tb3_4.Messages, []string{"FATAL: the number of error pattern must be at most one"}) {
		t.Fatalf("test3_4: unexpected Messages: %#v", tb3_4.Messages)
	}
	tb3_5 := NewHookedTestingTB("test3_5")
	func() {
		defer func() {
			if o := recover(); o != nil {
			}
		}()
		New(tb3_5, 0xdeadbeef).ExpectError(fmt.Errorf("hello"))
	}()
	if !tb3_5.Failed() {
		t.Fatalf("test3_5: unexpected Failed() == false")
	}
	if !reflect.DeepEqual(tb3_5.Messages, []string{"FATAL: error pattern must be string"}) {
		t.Fatalf("test3_5: unexpected Messages: %#v", tb3_5.Messages)
	}
	tb3_6 := NewHookedTestingTB("test3_6")
	func() {
		defer func() {
			if o := recover(); o != nil {
			}
		}()
		New(tb3_6, "hello[").ExpectError(fmt.Errorf("hello"))
	}()
	if !tb3_6.Failed() {
		t.Fatalf("test3_6: unexpected Failed() == false")
	}
	if !reflect.DeepEqual(tb3_6.Messages, []string{"FATAL: malformed expected error pattern: error parsing regexp: missing closing ]: `[`"}) {
		t.Fatalf("test3_6: unexpected Messages: %#v", tb3_6.Messages)
	}
	tb3_7 := NewHookedTestingTB("test3_6")
	func() {
		defer func() {
			if o := recover(); o != nil {
			}
		}()
		New(tb3_7, "hello").ExpectError(fmt.Errorf("hell0"))
	}()
	if !tb3_7.Failed() {
		t.Fatalf("test3_7: unexpected Failed() == false")
	}
	if !reflect.DeepEqual(tb3_7.Messages, []string{"FATAL: expected error pattern \"hello\", but got error \"hell0\""}) {
		t.Fatalf("test3_7: unexpected Messages: %#v", tb3_7.Messages)
	}
}

func TestAssertSucceedNew(t *testing.T) {
	// test1: Test helper registration
	tb1 := NewHookedTestingTB("test1")
	New(tb1).SucceedNew("hello", error(nil))
	// NOTICE: We cannot identify the location of any helper exactly, because of go cover tool inserts some code into source code files.
	if len(tb1.Helpers) != 1 {
		t.Fatalf("test1: unexpected Helpers: %#v", tb1.Helpers)
	}
	// test2: only normal uses
	tb2 := NewHookedTestingTB("test2")
	if obj, ok := New(tb2).SucceedNew("hello", error(nil)).(string); !(ok && obj == "hello") {
		t.Fatalf("test2: unexpected result: %#v (%T)", obj, obj)
	}
	if tb2.Failed() {
		t.Fatalf("test2: unexpected Failed() == true")
	}
	if len(tb2.Messages) != 0 {
		t.Fatalf("test2: unexpected Messages: %#v", tb2.Messages)
	}
	// test3: fatal exit case
	tb3 := NewHookedTestingTB("test3")
	func() {
		defer func() {
			if o := recover(); o != nil {
			}
		}()
		New(tb3).SucceedNew("hello", fmt.Errorf("world"))
	}()
	if !tb3.Failed() {
		t.Fatalf("test3: unexpected Failed() == false")
	}
	if !reflect.DeepEqual(tb3.Messages, []string{"FATAL: unexpected error in New-style function: world"}) {
		t.Fatalf("test3: unexpected Messages: %#v", tb3.Messages)
	}
}

func TestAssertSucceedWithoutError(t *testing.T) {
	// test1: Test helper registration
	tb1 := NewHookedTestingTB("test1")
	New(tb1).SucceedWithoutError(error(nil))
	// NOTICE: We cannot identify the location of any helper exactly, because of go cover tool inserts some code into source code files.
	if len(tb1.Helpers) != 1 {
		t.Fatalf("test1: unexpected Helpers: %#v", tb1.Helpers)
	}
	// test2: only normal uses
	tb2 := NewHookedTestingTB("test2")
	New(tb1).SucceedWithoutError(error(nil))
	if tb2.Failed() {
		t.Fatalf("test2: unexpected Failed() == true")
	}
	if len(tb2.Messages) != 0 {
		t.Fatalf("test2: unexpected Messages: %#v", tb2.Messages)
	}
	// test3: fatal exit case
	tb3 := NewHookedTestingTB("test3")
	func() {
		defer func() {
			if o := recover(); o != nil {
			}
		}()
		New(tb3).SucceedWithoutError(fmt.Errorf("hello"))
	}()
	if !tb3.Failed() {
		t.Fatalf("test3: unexpected Failed() == false")
	}
	if !reflect.DeepEqual(tb3.Messages, []string{"ERROR: unexpected error: hello"}) {
		t.Fatalf("test3: unexpected Messages: %#v", tb3.Messages)
	}
}
