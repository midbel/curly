package filters

import (
	"fmt"
	"reflect"
)

func First(value reflect.Value) (reflect.Value, error) {
	ret, err := FirstN(value, 1)
	if err != nil {
		return zero, err
	}
	if ret.Len() == 0 {
		return reflect.Zero(value.Type().Elem()), nil
	}
	return ret.Index(0), nil
}

func Last(value reflect.Value) (reflect.Value, error) {
	ret, err := LastN(value, 1)
	if err != nil {
		return zero, err
	}
	if ret.Len() == 0 {
		return reflect.Zero(value.Type().Elem()), nil
	}
	return ret.Index(0), nil
}

func FirstN(value reflect.Value, n int) (reflect.Value, error) {
	if err := isArray(value); err != nil {
		return zero, err
	}
	if value.Len() == 0 {
		return value, nil
	}
	if n >= value.Len() {
		return zero, fmt.Errorf("index is out of range")
	}
	v := reflect.MakeSlice(value.Type(), 0, n)
	for i := 0; i < n; i++ {
		v = reflect.Append(v, value.Index(i))
	}
	return v, nil
}

func LastN(value reflect.Value, n int) (reflect.Value, error) {
	if err := isArray(value); err != nil {
		return zero, err
	}
	if value.Len() == 0 {
		return value, nil
	}
	if n >= value.Len() {
		return zero, fmt.Errorf("index is out of range")
	}
	v := reflect.MakeSlice(value.Type(), 0, n)
	for i := value.Len() - n; i < value.Len(); i++ {
		v = reflect.Append(v, value.Index(i))
	}
	return v, nil
}

func Reverse(value reflect.Value) (reflect.Value, error) {
	if err := isArray(value); err != nil {
		return zero, err
	}
	v := reflect.MakeSlice(value.Type(), 0, value.Len())
	for i := value.Len() - 1; i >= 0; i-- {
		v = reflect.Append(v, value.Index(i))
	}
	return v, nil
}

func Concat(value, other reflect.Value) (reflect.Value, error) {
	if err := isArray(value); err != nil {
		return zero, err
	}
	if err := isArray(other); err != nil {
		return zero, err
	}
	if !value.Type().AssignableTo(other.Type()) {
		return zero, ErrIncompatible
	}
	for i := 0; i < other.Len(); i++ {
		value = reflect.Append(value, other.Index(i))
	}
	return value, nil
}

func Append(value, other reflect.Value) (reflect.Value, error) {
	if err := isArray(value); err != nil {
		return zero, err
	}
	if !value.Type().Elem().AssignableTo(other.Type()) {
		return zero, ErrIncompatible
	}
	return reflect.Append(value, other), nil
}
