package jq

import (
	"fmt"
	"math"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

func TestJQProgram(t *testing.T) {
	jq, err := NewJQ(".")
	ok(t, err)
	equals(t, ".", jq.program)
}

func TestTransform(t *testing.T) {
	jq, err := NewJQ(".")
	ok(t, err)

	jq.HandleJson("1")
	equals(t, true, jq.Next())
	equals(t, 1, jq.Value())
	equals(t, false, jq.Next())
}

func TestTransformArray(t *testing.T) {
	jq, err := NewJQ(".[]")
	ok(t, err)

	jq.Handle([]int{1, 2, 3})

	equals(t, true, jq.Next())
	equals(t, 1, jq.Value())

	equals(t, true, jq.Next())
	equals(t, 2, jq.Value())

	equals(t, true, jq.Next())
	equals(t, 3, jq.Value())

	equals(t, false, jq.Next())
}

func TestTransformArrayJson(t *testing.T) {
	jq, err := NewJQ(".[]")
	ok(t, err)

	jq.HandleJson("[1, 2, 3]")

	equals(t, true, jq.Next())
	equals(t, 1, jq.Value())

	equals(t, true, jq.Next())
	equals(t, 2, jq.Value())

	equals(t, true, jq.Next())
	equals(t, 3, jq.Value())

	equals(t, false, jq.Next())
}

// TODO KIND_INVALID

func TestJVNull(t *testing.T) {
	result := NewJV("null").ToGo()
	equals(t, nil, result)
}

func TestJVTrue(t *testing.T) {
	result := NewJV("true").ToGo()
	equals(t, true, result)
}

func TestJVFalse(t *testing.T) {
	result := NewJV("false").ToGo()
	equals(t, false, result)
}

func TestJVInt(t *testing.T) {
	result := NewJV("42").ToGo()
	equals(t, 42, result)
}

func TestJVFloat(t *testing.T) {
	result := NewJV("38.6").ToGo()
	equals(t, 38.6, result)
}

func TestJVString(t *testing.T) {
	result := NewJV("\"foo\"").ToGo()
	equals(t, "foo", result)
}

func TestJVArray(t *testing.T) {
	result := NewJV("[1, 2, 3]").ToGo()
	expected := []interface{}{1, 2, 3}
	equals(t, expected, result)
}

func TestJVObject(t *testing.T) {
	result := NewJV("{\"x\": 1, \"y\": \"two\"}").ToGo()
	expected := map[string]interface{}{
		"x": 1,
		"y": "two",
	}
	equals(t, expected, result)
}

func TestJVFromGoNil(t *testing.T) {
	jv := NewJVFromGo(nil)
	equals(t, nil, jv.ToGo())
}

func TestJVFromGoTrue(t *testing.T) {
	expected := true
	jv := NewJVFromGo(expected)
	equals(t, expected, jv.ToGo())
}

func TestJVFromGoFalse(t *testing.T) {
	expected := false
	jv := NewJVFromGo(expected)
	equals(t, expected, jv.ToGo())
}

// Ints

func TestJVFromGoInt(t *testing.T) {
	expected := 1
	jv := NewJVFromGo(expected)
	equals(t, expected, jv.ToGo())
}

func TestJVFromGoInt8(t *testing.T) {
	expected := 1
	jv := NewJVFromGo(int8(expected))
	equals(t, expected, jv.ToGo())
}

func TestJVFromGoInt16(t *testing.T) {
	expected := 1
	jv := NewJVFromGo(int16(expected))
	equals(t, expected, jv.ToGo())
}

func TestJVFromGoInt32(t *testing.T) {
	expected := 1
	jv := NewJVFromGo(int32(expected))
	equals(t, expected, jv.ToGo())
}

func TestJVFromGoInt64(t *testing.T) {
	expected := 1
	jv := NewJVFromGo(int64(expected))
	equals(t, expected, jv.ToGo())
}

// Uints

func TestJVFromGoUInt(t *testing.T) {
	expected := 1
	jv := NewJVFromGo(uint(expected))
	equals(t, expected, jv.ToGo())
}

func TestJVFromGoUInt8(t *testing.T) {
	expected := 1
	jv := NewJVFromGo(uint8(expected))
	equals(t, expected, jv.ToGo())
}

func TestJVFromGoUInt16(t *testing.T) {
	expected := 1
	jv := NewJVFromGo(uint16(expected))
	equals(t, expected, jv.ToGo())
}

func TestJVFromGoUInt32(t *testing.T) {
	expected := 1
	jv := NewJVFromGo(uint32(expected))
	equals(t, expected, jv.ToGo())
}

func TestJVFromGoUInt64(t *testing.T) {
	expected := 1
	jv := NewJVFromGo(uint64(expected))
	equals(t, expected, jv.ToGo())
}

// Floats

func TestJVFromGoFloat32(t *testing.T) {
	expected := 1.2
	jv := NewJVFromGo(float32(expected))
	actual := jv.ToGo().(float64)

	// allow for some error due to precision conversion
	if math.Abs(expected-actual) > 0.001 {
		equals(t, expected, actual)
	}
}

func TestJVFromGoFloat64(t *testing.T) {
	expected := 1.1
	jv := NewJVFromGo(float64(expected))
	equals(t, expected, jv.ToGo())
}

func TestJVFromGoString(t *testing.T) {
	expected := "foobar"
	jv := NewJVFromGo(expected)
	equals(t, expected, jv.ToGo())
}

// Arrays & Slices

func TestJVFromGoArray(t *testing.T) {
	expected := []interface{}{1, 2, 3}
	asArray := [3]interface{}{1, 2, 3}

	jv := NewJVFromGo(&asArray)
	equals(t, expected, jv.ToGo())
}

func TestJVFromGoIntArray(t *testing.T) {
	expected := []interface{}{1, 2, 3}
	asArray := [3]int{1, 2, 3}

	jv := NewJVFromGo(&asArray)
	equals(t, expected, jv.ToGo())
}

func TestJVFromGoSlice(t *testing.T) {
	expected := []interface{}{1, 2, 3}
	jv := NewJVFromGo(expected)
	equals(t, expected, jv.ToGo())
}

func TestJVFromGoIntSlice(t *testing.T) {
	expected := []interface{}{1, 2, 3}
	asInts := []int{1, 2, 3}

	jv := NewJVFromGo(asInts)
	equals(t, expected, jv.ToGo())
}

// Objects

func TestJVFromGoObject(t *testing.T) {
	expected := map[string]interface{}{
		"x": 1,
		"y": "two",
	}
	jv := NewJVFromGo(expected)
	equals(t, expected, jv.ToGo())
}

// Pointers

func TestJVFromGoIntPointer(t *testing.T) {
	expected := 1
	jv := NewJVFromGo(&expected)
	equals(t, expected, jv.ToGo())
}
