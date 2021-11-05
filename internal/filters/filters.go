package filters

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	zero            reflect.Value
	ErrIncompatible = errors.New("incompatible type")
)

func Len(value reflect.Value) (reflect.Value, error) {
	var n int
	switch value.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
		n = value.Len()
	default:
	}
	return reflect.ValueOf(n), nil
}

func isArray(value reflect.Value) error {
	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		return nil
	default:
		return fmt.Errorf("%s can not be used as an array", value)
	}
}

func isString(value reflect.Value) error {
	switch value.Kind() {
	case reflect.String:
		return nil
	default:
		return fmt.Errorf("%s can not be used as a string", value)
	}
}

func isNumeric(value reflect.Value) error {
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	case reflect.Float32, reflect.Float64:
	default:
		return fmt.Errorf("%s can not be used as a number", value)
	}
	return nil
}
