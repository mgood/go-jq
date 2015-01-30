package jq

// #cgo LDFLAGS: -ljq
// #include <jq.h>
// #include <jv.h>
import "C"
import (
	"errors"
	"fmt"
	"reflect"
)

type JQ struct {
	program   string
	state     *C.jq_state
	lastValue *JV
}

func NewJQ(program string) (*JQ, error) {
	state := C.jq_init()
	jq := &JQ{program, state, nil}
	if err := jq.compile(program); err != nil {
		return nil, err
	}
	return jq, nil
}

func (jq *JQ) Handle(value interface{}) {
	jq.start(NewJVFromGo(value))
}

func (jq *JQ) HandleJson(text string) {
	jq.start(NewJV(text))
}

func (jq *JQ) Next() bool {
	// FIXME this raises assertion if called before start()
	jq.lastValue = jq.next()
	return jq.lastValue.isValid()
}

func (jq *JQ) Value() interface{} {
	return jq.lastValue.ToGo()
}

func (jq *JQ) ValueJson() string {
	return jq.lastValue.ToJson()
}

// JQ APIs

func (jq *JQ) compile(program string) error {
	_ = C.jq_compile(jq.state, C.CString(program))
	return nil
}

func (jq *JQ) compileArgs(program string, args *JV) {
	C.jq_compile_args(jq.state, C.CString(program), args.value)
}

func (jq *JQ) start(jv *JV) {
	C.jq_start(jq.state, jv.value, 0)
}

func (jq *JQ) next() *JV {
	jv := JV{C.jq_next(jq.state)}
	return &jv
}

func (jq *JQ) teardown() {
	C.jq_teardown(&jq.state)
}

// JSON values

type JV struct {
	value C.jv
}

func NewJV(value string) *JV {
	parsed := C.jv_parse(C.CString(value))
	return &JV{parsed}
}

func NewJVFromGo(value interface{}) *JV {
	return &JV{goToJv(value)}
}

func (jv *JV) Copy() *JV {
	return &JV{C.jv_copy(jv.value)}
}

func (jv *JV) ToJson() string {
	strJv := C.jv_dump_string(jv.value, 0)
	result := C.jv_string_value(strJv)
	C.jv_free(strJv)
	return C.GoString(result)
}

func (jv *JV) ToGo() interface{} {
	return jvToGo(jv.value)
}

func goToJv(v interface{}) C.jv {
	if v == nil {
		return C.jv_null()
	}

	value := reflect.Indirect(reflect.ValueOf(v))

	switch value.Type().Kind() {
	case reflect.Bool:
		if value.Bool() {
			return C.jv_true()
		} else {
			return C.jv_false()
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return C.jv_number(C.double(value.Int()))
	// TODO reflect.Uintptr?
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return C.jv_number(C.double(value.Uint()))
	case reflect.Float32, reflect.Float64:
		return C.jv_number(C.double(value.Float()))
	case reflect.String:
		return C.jv_string(C.CString(value.String()))
	case reflect.Array, reflect.Slice:
		n := value.Len()
		arr := C.jv_array_sized(C.int(n))
		for i := 0; i < n; i++ {
			item := goToJv(value.Index(i).Interface())
			arr = C.jv_array_set(C.jv_copy(arr), C.int(i), item)
		}
		return arr
	case reflect.Map:
		// TODO assert key is string?
		object := C.jv_object()
		for _, k := range value.MapKeys() {
			key := goToJv(k.Interface())
			mapValue := goToJv(value.MapIndex(k).Interface())
			object = C.jv_object_set(object, key, mapValue)
		}
		return object
	}

	msg := fmt.Sprintf("unknown type for: %v", value.Interface())

	return C.jv_invalid_with_msg(C.jv_string(C.CString(msg)))
}

func jvToGo(value C.jv) interface{} {
	switch C.jv_get_kind(value) {
	case C.JV_KIND_INVALID:
		return errors.New("invalid")
	case C.JV_KIND_NULL:
		return nil
	case C.JV_KIND_FALSE:
		return false
	case C.JV_KIND_TRUE:
		return true
	case C.JV_KIND_NUMBER:
		number := C.jv_number_value(value)
		if C.jv_is_integer(value) == 0 {
			return float64(number)
		} else {
			return int(number)
		}
	case C.JV_KIND_STRING:
		return C.GoString(C.jv_string_value(value))
	case C.JV_KIND_ARRAY:
		length := C.jv_array_length(C.jv_copy(value))
		arr := make([]interface{}, length)
		for i := range arr {
			arr[i] = jvToGo(C.jv_array_get(C.jv_copy(value), C.int(i)))
		}
		return arr
	case C.JV_KIND_OBJECT:
		result := make(map[string]interface{})
		var k, v C.jv
		for jv_i := C.jv_object_iter(value); C.jv_object_iter_valid(value, jv_i) != 0; jv_i = C.jv_object_iter_next(value, jv_i) {
			k = C.jv_object_iter_key(value, jv_i)
			v = C.jv_object_iter_value(value, jv_i)
			result[C.GoString(C.jv_string_value(k))] = jvToGo(v)
		}
		return result
	default:
		return errors.New("unknown type")
	}
}

func (jv *JV) isValid() bool {
	return C.jv_is_valid(jv.value) != 0
}
