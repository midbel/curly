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

}

func testCmp(t *testing.T) {

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

func getIntValue(v reflect.Value) int {
	return int(v.Int())
}

func getStringValue(v reflect.Value) string {
	return v.String()
}
