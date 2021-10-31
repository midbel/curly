package state

import (
	"reflect"
	"strings"
)

var Filters = map[string]interface{}{
	"join":    strings.Join,
	"split":   strings.Split,
	"replace": strings.Replace,
	"lower":   strings.ToLower,
	"upper":   strings.ToUpper,
	"count":   strings.Count,
	"trim":    strings.TrimSpace,
	"repeat":  strings.Repeat,
	"len":     length,
	"first":   first,
	"last":    last,
	"firstn":  firstN,
	"lastn":   lastN,
	"reverse": reverse,
}

func Lookup(name string) (reflect.Value, error) {
	fn := reflect.ValueOf(Filters[name])
	if !fn.IsValid() || fn.Kind() != reflect.Func {
		return Invalid, ErrFound
	}
	return fn, nil
}

func reverse(value reflect.Value) reflect.Value {
	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		var (
			size  = value.Len()
			slice = reflect.MakeSlice(value.Type(), 0, size)
		)
		for i := size - 1; i >= 0; i-- {
			slice = reflect.Append(slice, value.Index(i))
		}
		return slice
	default:
		return Invalid
	}
}

func first(value reflect.Value) reflect.Value {
	return firstN(value, 1)
}

func last(value reflect.Value) reflect.Value {
	return lastN(value, 1)
}

func firstN(value reflect.Value, n int) reflect.Value {
	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		if value.Len() < n {
			return Invalid
		}
		return value.Slice(0, n)
	default:
		return Invalid
	}
}

func lastN(value reflect.Value, n int) reflect.Value {
	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		if value.Len() < n {
			return Invalid
		}
		return value.Slice(value.Len()-n, value.Len())
	default:
		return Invalid
	}
}

func length(value reflect.Value) int {
	switch value.Kind() {
	case reflect.Array, reflect.Slice, reflect.String, reflect.Map:
		return value.Len()
	default:
		return 0
	}
}
