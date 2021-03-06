package state

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	Invalid  reflect.Value
	ErrFound = errors.New("not found")
)

const (
	KeyLoop     = "loop"
	KeyLoop0    = "loop0"
	KeyRevLoop  = "revloop"
	KeyRevLoop0 = "revloop0"
	KeyLength   = "length"
	KeyContext  = "ctx"
)

type FuncMap map[string]interface{}

type State interface {
	Lookup(name string) (reflect.Value, error)
	Define(name string, value reflect.Value) error
	Resolve(name string) (reflect.Value, error)
}

type loopState struct {
	State
	loop   int
	length int
}

func Loop(i, n int, parent State) State {
	return &loopState{
		State:  parent,
		loop:   i,
		length: n,
	}
}

func (s *loopState) Resolve(name string) (reflect.Value, error) {
	var value reflect.Value
	switch name {
	case KeyLength:
		value = reflect.ValueOf(s.length)
	case KeyLoop:
		value = reflect.ValueOf(s.loop + 1)
	case KeyLoop0:
		value = reflect.ValueOf(s.loop)
	case KeyRevLoop:
		value = reflect.ValueOf(s.length - s.loop)
	case KeyRevLoop0:
		value = reflect.ValueOf((s.length - s.loop) - 1)
	default:
		return s.State.Resolve(name)
	}
	return value, nil
}

func (s *loopState) Define(name string, value reflect.Value) error {
	switch name {
	case KeyLength:
	case KeyLoop:
	case KeyLoop0:
	case KeyRevLoop:
	case KeyRevLoop0:
		return s.State.Define(name, value)
	}
	return fmt.Errorf("%s can not be defined", name)
}

type stdState struct {
	parent  State
	current reflect.Value
	filters map[string]interface{}
	locals  map[string]reflect.Value
}

func EmptyState(data interface{}, filters FuncMap) State {
	return EnclosedState(data, nil, filters)
}

func EnclosedState(data interface{}, parent State, filters FuncMap) State {
	return &stdState{
		current: valueOf(data),
		parent:  parent,
		filters: filters,
		locals:  make(map[string]reflect.Value),
	}
}

func (s *stdState) Lookup(name string) (reflect.Value, error) {
	if s.filters == nil && s.parent != nil {
		return s.parent.Lookup(name)
	}
	fn := reflect.ValueOf(s.filters[name])
	if !fn.IsValid() || fn.Kind() != reflect.Func {
		if s.parent != nil {
			return s.parent.Lookup(name)
		}
		return Invalid, ErrFound
	}
	return fn, nil
}

func (s *stdState) Define(key string, value reflect.Value) error {
	if key == KeyContext {
		return fmt.Errorf("%s can not be defined", key)
	}
	s.locals[key] = value
	return nil
}

func (s *stdState) Resolve(key string) (reflect.Value, error) {
	if key == KeyContext {
		return s.current, nil
	}
	v, err := s.find(key)
	if err != nil {
		if r, ok := s.locals[key]; ok {
			return r, nil
		}
	}
	if err != nil && s.parent != nil {
		v, err = s.parent.Resolve(key)
	}
	return v, err
}

func (s *stdState) find(key string) (reflect.Value, error) {
	return s.findWith(key, s.current)
}

func (s *stdState) findWith(key string, value reflect.Value) (reflect.Value, error) {
	switch value.Kind() {
	case reflect.Struct:
		return s.lookupStruct(key, value)
	case reflect.Map:
		return s.lookupMap(key, value)
	case reflect.Ptr:
		return s.findWith(key, value.Elem())
	case reflect.Interface:
		return s.findWith(key, reflect.ValueOf(value.Interface()))
	}
	return Invalid, ErrFound
}

func (s *stdState) lookupStruct(key string, value reflect.Value) (reflect.Value, error) {
	t := value.Type()
	for i := 0; i < value.NumField(); i++ {
		sf := t.Field(i)
		if sf.Name == key || sf.Tag.Get("curly") == key {
			return value.Field(i), nil
		}
	}
	return Invalid, ErrFound
}

func (s *stdState) lookupMap(key string, value reflect.Value) (reflect.Value, error) {
	t := value.Type().Key()
	if !t.AssignableTo(reflect.TypeOf(key)) {
		return Invalid, ErrFound
	}
	val := value.MapIndex(reflect.ValueOf(key))
	if val.IsZero() {
		return Invalid, ErrFound
	}
	return val, nil
}

func valueOf(v interface{}) reflect.Value {
	if v, ok := v.(reflect.Value); ok {
		return v
	}
	return reflect.ValueOf(v)
}
