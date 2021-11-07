package filters_test

import (
	"reflect"
	"testing"

	"github.com/midbel/curly/internal/filters"
)

func TestFilters(t *testing.T) {
	t.Run("len", testLen)
	t.Run("math", testMath)
	t.Run("cmp", testCmp)
	t.Run("array", testArray)
}

func testLen(t *testing.T) {
	data := []struct {
		Input interface{}
		Want  int
	}{
		{
			Input: "hello",
			Want:  5,
		},
		{
			Input: []string{"hello", "world"},
			Want:  2,
		},
	}
	for _, d := range data {
		v, err := filters.Len(getValue(d.Input))
		if err != nil {
			t.Errorf("unexpected error! got %s", err)
			continue
		}
		if got := getIntValue(v); got != d.Want {
			t.Errorf("result mismatched! want %d, got %d", d.Want, got)
		}
	}
}

func testMath(t *testing.T) {
	var x, y int
	ret, err := filters.Increment(getValue(x))
	checkInt(t, ret, err, 1)

	ret, err = filters.Decrement(ret)
	checkInt(t, ret, err, x)

	ret, err = filters.Mul(getValue(x), getValue(x))
	checkInt(t, ret, err, x)

	x, y = 10, 2
	ret, err = filters.Div(getValue(x), getValue(y))
	checkInt(t, ret, err, 5)
	ret, err = filters.Mod(getValue(x), getValue(y))
	checkInt(t, ret, err, 0)

	x = 2
	ret, err = filters.Pow(getValue(x), getValue(y))
	checkInt(t, ret, err, 4)

	arr := []float64{1, 7, 9, -9, 10}
	ret, err = filters.Min(toArrayValues(arr)...)
	checkFloat(t, ret, err, -9)
	ret, err = filters.Max(toArrayValues(arr)...)
	checkFloat(t, ret, err, 10)
}

func testCmp(t *testing.T) {
	ret, err := filters.Equal(getValue("foobar"), getValue("foobar"))
	checkBool(t, ret, err, true)
	ret, err = filters.Equal(getValue(true), getValue(true))
	checkBool(t, ret, err, true)
	ret, err = filters.Equal(getValue(true), getValue(false))
	checkBool(t, ret, err, false)
	ret, err = filters.NotEqual(getValue(true), getValue(false))
	checkBool(t, ret, err, true)
	ret, err = filters.Equal(getValue("foobar"), getValue("foo"))
	checkBool(t, ret, err, false)
	ret, err = filters.Equal(getValue(100), getValue(100))
	checkBool(t, ret, err, true)
	ret, err = filters.Equal(getValue(100.0), getValue(1.0))
	checkBool(t, ret, err, false)
	ret, err = filters.NotEqual(getValue("foobar"), getValue("foo"))
	checkBool(t, ret, err, true)
	ret, err = filters.Equal(getValue(100), getValue(100))
	checkBool(t, ret, err, true)
	ret, err = filters.Lesser(getValue(1), getValue(100))
	checkBool(t, ret, err, true)
	ret, err = filters.LesserEqual(getValue(1), getValue(100))
	checkBool(t, ret, err, true)
	ret, err = filters.Equal(getValue(100.0), getValue(1.0))
	checkBool(t, ret, err, false)
	ret, err = filters.Greater(getValue(100.0), getValue(1.0))
	checkBool(t, ret, err, true)
	ret, err = filters.GreaterEqual(getValue(100.0), getValue(1.0))
	checkBool(t, ret, err, true)
	ret, err = filters.GreaterEqual(getValue(100.0), getValue(100.0))
	checkBool(t, ret, err, true)

	ret, err = filters.And(getValue(true), getValue(true))
	checkBool(t, ret, err, true)
	ret, err = filters.And(getValue(true), getValue(false))
	checkBool(t, ret, err, false)
	ret, err = filters.Or(getValue(true), getValue(false))
	checkBool(t, ret, err, true)
	ret, err = filters.Or(getValue(false), getValue(false))
	checkBool(t, ret, err, false)

	ret, err = filters.Not(getValue(true))
	checkBool(t, ret, err, false)
	ret, err = filters.Not(getValue(false))
	checkBool(t, ret, err, true)
}

func testArray(t *testing.T) {
	var (
		arr = []string{"hello", "foo", "bar", "world"}
		ret reflect.Value
		err error
	)
	ret, err = filters.First(getValue(arr))
	checkString(t, ret, err, "hello")
	ret, err = filters.First(getValue([]string{}))
	checkString(t, ret, err, "")
	ret, err = filters.Last(getValue(arr))
	checkString(t, ret, err, "world")
	ret, err = filters.Last(getValue([]string{}))
	checkString(t, ret, err, "")

	ret, err = filters.FirstN(getValue(arr), 3)
	checkStringArray(t, ret, err, arr[:3])

	ret, err = filters.LastN(getValue(arr), 2)
	checkStringArray(t, ret, err, arr[2:])

	ret, err = filters.Reverse(getValue(arr))
	checkStringArray(t, ret, err, []string{"world", "bar", "foo", "hello"})

	ret, err = filters.Concat(getValue(arr[:2]), getValue(arr[2:]))
	checkStringArray(t, ret, err, arr)

	ret, err = filters.Append(getValue(arr[:3]), getValue(arr[3]))
	checkStringArray(t, ret, err, arr)
}

func checkStringArray(t *testing.T, val reflect.Value, err error, want []string) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error! got %s", err)
		return
	}
	if len(want) != val.Len() {
		t.Errorf("length mismatched! want %d, got %d", len(want), val.Len())
		return
	}
	for i := 0; i < val.Len(); i++ {
		checkString(t, val.Index(i), nil, want[i])
	}
}

func checkBool(t *testing.T, val reflect.Value, err error, want bool) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error! got %s", err)
		return
	}
	if got := getBoolValue(val); got != want {
		t.Errorf("result mismatched! want %t, got %t", want, got)
	}
}

func checkInt(t *testing.T, val reflect.Value, err error, want int) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error! got %s", err)
		return
	}
	if got := getIntValue(val); got != want {
		t.Errorf("result mismatched! want %d, got %d", want, got)
	}
}

func checkFloat(t *testing.T, val reflect.Value, err error, want float64) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error! got %s", err)
		return
	}
	if got := getFloatValue(val); got != want {
		t.Errorf("result mismatched! want %f, got %f", want, got)
	}
}

func checkString(t *testing.T, val reflect.Value, err error, want string) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error! got %s", err)
		return
	}
	if got := getStringValue(val); got != want {
		t.Errorf("result mismatched! want %s, got %s", want, got)
	}
}

func getValue(v interface{}) reflect.Value {
	return reflect.ValueOf(v)
}

func toArrayValues(arr []float64) []reflect.Value {
	var vs []reflect.Value
	for i := range arr {
		vs = append(vs, getValue(arr[i]))
	}
	return vs
}

func getIntValue(v reflect.Value) int {
	return int(v.Int())
}

func getFloatValue(v reflect.Value) float64 {
	return v.Float()
}

func getBoolValue(v reflect.Value) bool {
	return v.Bool()
}

func getStringValue(v reflect.Value) string {
	return v.String()
}
