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
	if value.Kind() == reflect.String {
		return nil
	}
	return fmt.Errorf("%s can not be used as a string", value)
}

func isNumeric(value reflect.Value) error {
	k := value.Kind()
	if isInt(k) || isUint(k) || isFloat(k) {
		return nil
	}
	return fmt.Errorf("%s can not be used as a number", value)
}

func isBool(value reflect.Value) error {
	if value.Kind() == reflect.Bool {
		return nil
	}
	return fmt.Errorf("%s can not be used as a boolean", value)
}

func isInt(k reflect.Kind) bool {
	return k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 || k == reflect.Int32 || k == reflect.Int64
}

func isUint(k reflect.Kind) bool {
	return k == reflect.Uint || k == reflect.Uint8 || k == reflect.Uint16 || k == reflect.Uint32 || k == reflect.Uint64
}

func isFloat(k reflect.Kind) bool {
	return k == reflect.Float32 || k == reflect.Float64
}
