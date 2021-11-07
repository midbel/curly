package filters

import (
	"fmt"
	"reflect"
)

func Get(value reflect.Value, field string) (reflect.Value, error) {
	switch {
	case accept(isMap(value)):
		return getMap(value, field)
	case accept(isArray(value)):
		return getStruct(value, field)
	default:
		return zero, ErrIncompatible
	}
}

func getStruct(value reflect.Value, field string) (reflect.Value, error) {
	return value.GetFieldByName(field), nil
}

func getMap(value reflect.Value, field string) (reflect.Value, error) {
	val := reflect.ValueOf(field)
	if !val.Type().AssignableTo(value.Type().Key()) {
		return zero, ErrIncompatible
	}
	return value.MapIndex(val), nil
}

func Keys(value reflect.Value) (reflect.Value, error) {
	if err := isMap(value); err != nil {
		return zero, err
	}
	var (
		vs = reflect.MakeSlice(value.Type().Key(), 0, value.Len())
		it = value.MapRange()
	)
	for it.Next() {
		vs = reflect.Append(vs, it.Key())
	}
	return vs, nil
}

func Values(value reflect.Value) (reflect.Value, error) {
	if err := isMap(value); err != nil {
		return zero, err
	}
	var (
		vs = reflect.MakeSlice(value.Type().Elem(), 0, value.Len())
		it = value.MapRange()
	)
	for it.Next() {
		vs = reflect.Append(vs, it.Value())
	}
	return vs, nil
}

func Index(value, index reflect.Value) (reflect.Value, error) {
	switch {
	case accept(isMap(value)):
		return indexMap(value, index)
	case accept(isArray(value)):
		return indexArray(value, index)
	default:
		return zero, ErrIncompatible
	}
}

func indexArray(value, index reflect.Value) (reflect.Value, error) {
	if err := isNumeric(index); err != nil {
		return zero, err
	}
	i, _ := toInt(index)
	if i < 0 || i >= value.Len() {
		return zero, fmt.Errorf("index out of range")
	}
	return value.Index(i), nil
}

func indexMap(value, index reflect.Value) (reflect.Value, error) {
	if !index.Type().AssignableTo(value.Type().Key()) {
		return zero, ErrIncompatible
	}
	return value.MapIndex(index), nil
}

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
