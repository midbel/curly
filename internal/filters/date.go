package filters

import (
	"reflect"
	"time"
)

func Now() (reflect.Value, error) {
	t := time.Now()
	return reflect.ValueOf(t), nil
}
