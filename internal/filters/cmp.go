package filters

import (
	"reflect"
)

func And(fst, snd reflect.Value) (reflect.Value, error) {
	if err := isBool(fst); err != nil {
		return zero, err
	}
	if err := isBool(snd); err != nil {
		return zero, err
	}
	cmp := fst.Bool() && snd.Bool()
	return reflect.ValueOf(cmp), nil
}

func Or(fst, snd reflect.Value) (reflect.Value, error) {
	if err := isBool(fst); err != nil {
		return zero, err
	}
	if err := isBool(snd); err != nil {
		return zero, err
	}
	cmp := fst.Bool() || snd.Bool()
	return reflect.ValueOf(cmp), nil
}

func Not(fst reflect.Value) (reflect.Value, error) {
	if err := isBool(fst); err != nil {
		return zero, err
	}
	cmp := !fst.Bool()
	return reflect.ValueOf(cmp), nil
}

func Equal(fst, snd reflect.Value) (reflect.Value, error) {
	accept := func(err error) bool {
		return err == nil
	}
	var cmp bool
	switch {
	case accept(isNumeric(fst)) && accept(isNumeric(snd)):
		v1, _ := toFloat(fst)
		v2, _ := toFloat(snd)
		cmp = v1 == v2
	case accept(isString(fst)) && accept(isString(snd)):
		cmp = fst.String() == snd.String()
	case accept(isBool(fst)) && accept(isBool(snd)):
		cmp = fst.Bool() == snd.Bool()
	default:
		return zero, ErrIncompatible
	}
	return reflect.ValueOf(cmp), nil
}

func NotEqual(fst, snd reflect.Value) (reflect.Value, error) {
	cmp, err := Equal(fst, snd)
	if err != nil {
		return zero, err
	}
	return reflect.ValueOf(!cmp.Bool()), nil
}

func Greater(fst, snd reflect.Value) (reflect.Value, error) {
	accept := func(err error) bool {
		return err == nil
	}
	var cmp bool
	switch {
	case accept(isNumeric(fst)) && accept(isNumeric(snd)):
		v1, _ := toFloat(fst)
		v2, _ := toFloat(snd)
		cmp = v1 > v2
	case accept(isString(fst)) && accept(isString(snd)):
		cmp = fst.String() > snd.String()
	default:
		return zero, ErrIncompatible
	}
	return reflect.ValueOf(cmp), nil
}

func GreaterEqual(fst, snd reflect.Value) (reflect.Value, error) {
	var (
		cmp1, err1 = Greater(fst, snd)
		cmp2, err2 = Equal(fst, snd)
	)
	if err1 != nil || err2 != nil {
		return zero, ErrIncompatible
	}
	cmp := cmp1.Bool() || cmp2.Bool()
	return reflect.ValueOf(cmp), nil
}

func Lesser(fst, snd reflect.Value) (reflect.Value, error) {
	accept := func(err error) bool {
		return err == nil
	}
	var cmp bool
	switch {
	case accept(isNumeric(fst)) && accept(isNumeric(snd)):
		v1, _ := toFloat(fst)
		v2, _ := toFloat(snd)
		cmp = v1 < v2
	case accept(isString(fst)) && accept(isString(snd)):
		cmp = fst.String() < snd.String()
	default:
		return zero, ErrIncompatible
	}
	return reflect.ValueOf(cmp), nil
}

func LesserEqual(fst, snd reflect.Value) (reflect.Value, error) {
	var (
		cmp1, err1 = Lesser(fst, snd)
		cmp2, err2 = Equal(fst, snd)
	)
	if err1 != nil || err2 != nil {
		return zero, ErrIncompatible
	}
	cmp := cmp1.Bool() || cmp2.Bool()
	return reflect.ValueOf(cmp), nil
}
