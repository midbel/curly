package filters

import (
	"math/rand"
	"reflect"
)

func Rand() (reflect.Value, error) {
	n := rand.Int()
	return reflect.ValueOf(n), nil
}

func Add(fst, lst reflect.Value) (reflect.Value, error) {
	if err := isNumeric(fst); err != nil {
		return zero, err
	}
	if err := isNumeric(lst); err != nil {
		return zero, err
	}
	return zero, nil
}

func Sub(fst, lst reflect.Value) (reflect.Value, error) {
	if err := isNumeric(fst); err != nil {
		return zero, err
	}
	if err := isNumeric(lst); err != nil {
		return zero, err
	}
	return zero, nil
}

func Mul(fst, lst reflect.Value) (reflect.Value, error) {
	if err := isNumeric(fst); err != nil {
		return zero, err
	}
	if err := isNumeric(lst); err != nil {
		return zero, err
	}
	return zero, nil
}

func Div(fst, lst reflect.Value) (reflect.Value, error) {
	if err := isNumeric(fst); err != nil {
		return zero, err
	}
	if err := isNumeric(lst); err != nil {
		return zero, err
	}
	return zero, nil
}

func Mod(fst, lst reflect.Value) (reflect.Value, error) {
	if err := isNumeric(fst); err != nil {
		return zero, err
	}
	if err := isNumeric(lst); err != nil {
		return zero, err
	}
	return zero, nil
}

func Pow(fst, lst reflect.Value) (reflect.Value, error) {
	if err := isNumeric(fst); err != nil {
		return zero, err
	}
	if err := isNumeric(lst); err != nil {
		return zero, err
	}
	return zero, nil
}

func Min(list ...reflect.Value) (reflect.Value, error) {
	return zero, nil
}

func Max(list ...reflect.Value) (reflect.Value, error) {
	return zero, nil
}

func Increment(fst reflect.Value) (reflect.Value, error) {
	snd := reflect.ValueOf(1)
	return Add(fst, snd)
}

func Decrement(fst reflect.Value) (reflect.Value, error) {
	snd := reflect.ValueOf(1)
	return Sub(fst, snd)
}
