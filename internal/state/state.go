package state

import (
	"errors"
	"reflect"
)

var (
	Invalid  reflect.Value
	ErrFound = errors.New("not found")
)

type State struct {
	parent  *State
	current reflect.Value
}

func EmptyState(data interface{}) *State {
	return EnclosedState(data, nil)
}

func EnclosedState(data interface{}, parent *State) *State {
	return &State{
		current: valueOf(data),
		parent:  parent,
	}
}

func (s *State) Resolve(key string) (reflect.Value, error) {
	v, err := s.find(key)
	if err != nil && s.parent != nil {
		v, err = s.parent.find(key)
	}
	return v, err
}

func (s *State) find(key string) (reflect.Value, error) {
	return s.findWith(key, s.current)
}

func (s *State) findWith(key string, value reflect.Value) (reflect.Value, error) {
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

func (s *State) lookupStruct(key string, value reflect.Value) (reflect.Value, error) {
	t := value.Type()
	for i := 0; i < value.NumField(); i++ {
		sf := t.Field(i)
		if sf.Name == key || sf.Tag.Get("curly") == key {
			return value.Field(i), nil
		}
	}
	return Invalid, ErrFound
}

func (s *State) lookupMap(key string, value reflect.Value) (reflect.Value, error) {
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
