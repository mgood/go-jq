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
	defer jq.Close()

	equals(t, ".", jq.program)
}

func TestTransform(t *testing.T) {
	jq, err := NewJQ(".")
	ok(t, err)
	defer jq.Close()

	jq.HandleJson("1")
	equals(t, true, jq.Next())
	equals(t, 1, jq.Value())
	equals(t, false, jq.Next())
}

func TestTransformArray(t *testing.T) {
	jq, err := NewJQ(".[]")
	ok(t, err)
	defer jq.Close()

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
	defer jq.Close()

	jq.HandleJson("[1, 2, 3]")

	equals(t, true, jq.Next())
	equals(t, 1, jq.Value())

	equals(t, true, jq.Next())
	equals(t, 2, jq.Value())

	equals(t, true, jq.Next())
	equals(t, 3, jq.Value())

	equals(t, false, jq.Next())
}

func TestTransformArrayJsonString(t *testing.T) {
	jq, err := NewJQ(".[]")
	ok(t, err)
	defer jq.Close()

	jq.HandleJson("[[1], [2], [3]]")

	equals(t, true, jq.Next())
	equals(t, "[1]", jq.ValueJson())

	equals(t, true, jq.Next())
	equals(t, "[2]", jq.ValueJson())

	equals(t, true, jq.Next())
	equals(t, "[3]", jq.ValueJson())

	equals(t, false, jq.Next())
}

// TODO KIND_INVALID

func assertJsonParsed(t *testing.T, expected interface{}, json string) {
	jv := parseJson(json)
	result := jvToGo(jv)
	freeJv(jv)
	equals(t, expected, result)
}

func TestJVNull(t *testing.T) {
	assertJsonParsed(t, nil, "null")
}

func TestJVTrue(t *testing.T) {
	assertJsonParsed(t, true, "true")
}

func TestJVFalse(t *testing.T) {
	assertJsonParsed(t, false, "false")
}

func TestJVInt(t *testing.T) {
	assertJsonParsed(t, 42, "42")
}

func TestJVFloat(t *testing.T) {
	assertJsonParsed(t, 38.6, "38.6")
}

func TestJVString(t *testing.T) {
	assertJsonParsed(t, "foo", "\"foo\"")
}

func TestJVArray(t *testing.T) {
	expected := []interface{}{1, 2, 3}
	assertJsonParsed(t, expected, "[1, 2, 3]")
}

func TestJVObject(t *testing.T) {
	json := "{\"x\": 1, \"y\": \"two\"}"
	expected := map[string]interface{}{
		"x": 1,
		"y": "two",
	}
	assertJsonParsed(t, expected, json)
}

func assertGoJvConversion(t *testing.T, expected interface{}, value interface{}) {
	jv := goToJv(value)
	actual := jvToGo(jv)
	freeJv(jv)
	equals(t, expected, actual)
}

func TestJVFromGoNil(t *testing.T) {
	assertGoJvConversion(t, nil, nil)
}

func TestJVFromGoTrue(t *testing.T) {
	assertGoJvConversion(t, true, true)
}

func TestJVFromGoFalse(t *testing.T) {
	assertGoJvConversion(t, false, false)
}

// Ints

func TestJVFromGoInt(t *testing.T) {
	assertGoJvConversion(t, 1, 1)
}

func TestJVFromGoInt8(t *testing.T) {
	assertGoJvConversion(t, 1, int8(1))
}

func TestJVFromGoInt16(t *testing.T) {
	assertGoJvConversion(t, 1, int16(1))
}

func TestJVFromGoInt32(t *testing.T) {
	assertGoJvConversion(t, 1, int32(1))
}

func TestJVFromGoInt64(t *testing.T) {
	assertGoJvConversion(t, 1, int64(1))
}

// Uints

func TestJVFromGoUInt(t *testing.T) {
	assertGoJvConversion(t, 1, uint(1))
}

func TestJVFromGoUInt8(t *testing.T) {
	assertGoJvConversion(t, 1, uint8(1))
}

func TestJVFromGoUInt16(t *testing.T) {
	assertGoJvConversion(t, 1, uint16(1))
}

func TestJVFromGoUInt32(t *testing.T) {
	assertGoJvConversion(t, 1, uint32(1))
}

func TestJVFromGoUInt64(t *testing.T) {
	assertGoJvConversion(t, 1, uint64(1))
}

// Floats

func TestJVFromGoFloat32(t *testing.T) {
	expected := 1.2
	jv := goToJv(float32(expected))
	actual := jvToGo(jv).(float64)

	// allow for some error due to precision conversion
	if math.Abs(expected-actual) > 0.001 {
		equals(t, expected, actual)
	}
}

func TestJVFromGoFloat64(t *testing.T) {
	assertGoJvConversion(t, 1.1, float64(1.1))
}

func TestJVFromGoString(t *testing.T) {
	assertGoJvConversion(t, "foobar", "foobar")
}

// Arrays & Slices

func TestJVFromGoArray(t *testing.T) {
	expected := []interface{}{1, 2, 3}
	asArray := [3]interface{}{1, 2, 3}
	assertGoJvConversion(t, expected, asArray)
}

func TestJVFromGoIntArray(t *testing.T) {
	expected := []interface{}{1, 2, 3}
	asArray := [3]int{1, 2, 3}
	assertGoJvConversion(t, expected, asArray)
}

func TestJVFromGoSlice(t *testing.T) {
	expected := []interface{}{1, 2, 3}
	assertGoJvConversion(t, expected, expected)
}

func TestJVFromGoIntSlice(t *testing.T) {
	expected := []interface{}{1, 2, 3}
	asInts := []int{1, 2, 3}
	assertGoJvConversion(t, expected, asInts)
}

// Objects

func TestJVFromGoObject(t *testing.T) {
	expected := map[string]interface{}{
		"x": 1,
		"y": "two",
	}
	assertGoJvConversion(t, expected, expected)
}

// Pointers

func TestJVFromGoIntPointer(t *testing.T) {
	expected := 1
	assertGoJvConversion(t, expected, &expected)
}

// JSON

func TestDumpJSONRefCount(t *testing.T) {
	text := "{\"foo\":1}"
	jv := parseJson(text)

	// check that dumpJson keeps the same refcount
	// and that repeated use on the same value doesn't crash
	equals(t, 1, refcount(jv))
	equals(t, text, dumpJson(jv))
	equals(t, text, dumpJson(jv))
	equals(t, 1, refcount(jv))

	// afterward, we should be able to free it and the refcount
	// decreases to 0
	freeJv(jv)
	equals(t, 0, refcount(jv))
}
