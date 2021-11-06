package filters

import (
	"fmt"
	"math"
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
	return doMath(fst, lst, func(v1, v2 float64) float64 {
		return v1 + v2
	})
}

func Sub(fst, lst reflect.Value) (reflect.Value, error) {
	if err := isNumeric(fst); err != nil {
		return zero, err
	}
	if err := isNumeric(lst); err != nil {
		return zero, err
	}
	return doMath(fst, lst, func(v1, v2 float64) float64 {
		return v1 - v2
	})
}

func Mul(fst, lst reflect.Value) (reflect.Value, error) {
	if err := isNumeric(fst); err != nil {
		return zero, err
	}
	if err := isNumeric(lst); err != nil {
		return zero, err
	}
	return doMath(fst, lst, func(v1, v2 float64) float64 {
		return v1 * v2
	})
}

func Div(fst, lst reflect.Value) (reflect.Value, error) {
	if err := isNumeric(fst); err != nil {
		return zero, err
	}
	if err := isNumeric(lst); err != nil {
		return zero, err
	}
	return doMath(fst, lst, func(v1, v2 float64) float64 {
		if v2 == 0 {
			return math.NaN()
		}
		return v1 / v2
	})
}

func Mod(fst, lst reflect.Value) (reflect.Value, error) {
	if err := isNumeric(fst); err != nil {
		return zero, err
	}
	if err := isNumeric(lst); err != nil {
		return zero, err
	}
	return doMath(fst, lst, func(v1, v2 float64) float64 {
		if v2 == 0 {
			return math.NaN()
		}
		return math.Mod(v1, v2)
	})
}

func Pow(fst, lst reflect.Value) (reflect.Value, error) {
	if err := isNumeric(fst); err != nil {
		return zero, err
	}
	if err := isNumeric(lst); err != nil {
		return zero, err
	}
	return doMath(fst, lst, func(v1, v2 float64) float64 {
		return math.Pow(v1, v2)
	})
}

func Min(list ...reflect.Value) (reflect.Value, error) {
	var (
		min  float64
		curr float64
		kind reflect.Kind
	)
	for i := range list {
		curr, kind = toFloat(list[i])
		if kind == reflect.Invalid {
			return zero, fmt.Errorf("%s: invalid value", list[i])
		}
		if i == 0 || curr < min {
			min, kind = curr, list[i].Kind()
			continue
		}
	}
	return toValue(min, kind, kind), nil
}

func Max(list ...reflect.Value) (reflect.Value, error) {
	var (
		max  float64
		curr float64
		kind reflect.Kind
	)
	for i := range list {
		curr, kind = toFloat(list[i])
		if kind == reflect.Invalid {
			return zero, fmt.Errorf("%s: invalid value", list[i])
		}
		if i == 0 || curr > max {
			max, kind = curr, list[i].Kind()
			continue
		}
	}
	return toValue(max, kind, kind), nil
}

func Increment(fst reflect.Value) (reflect.Value, error) {
	snd := reflect.ValueOf(1)
	return Add(fst, snd)
}

func Decrement(fst reflect.Value) (reflect.Value, error) {
	snd := reflect.ValueOf(1)
	return Sub(fst, snd)
}

func toFloat(v reflect.Value) (float64, reflect.Kind) {
	var (
		val  float64
		kind = reflect.Invalid
	)
	switch k := v.Kind(); {
	default:
	case isFloat(k):
		val, kind = v.Float(), reflect.Float64
	case isInt(k):
		val, kind = float64(v.Int()), reflect.Int
	case isUint(k):
		val, kind = float64(v.Uint()), reflect.Uint
	}
	return val, kind
}

func doMath(fst, lst reflect.Value, do func(float64, float64) float64) (reflect.Value, error) {
	var (
		v1, k1 = toFloat(fst)
		v2, k2 = toFloat(lst)
		ret    = do(v1, v2)
	)
	if math.IsNaN(ret) {
		return zero, fmt.Errorf("invalid operation between %f and %f", v1, v2)
	}
	return toValue(ret, k1, k2), nil
}

func toValue(v float64, k1, k2 reflect.Kind) reflect.Value {
	if k1 == reflect.Float64 || k2 == reflect.Float64 {
		return reflect.ValueOf(v)
	}
	if k1 == reflect.Uint || k2 == reflect.Uint {
		return reflect.ValueOf(uint(v))
	}
	return reflect.ValueOf(int(v))
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
