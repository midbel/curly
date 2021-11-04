package state

import (
	"errors"
	"reflect"
)

var (
	Invalid  reflect.Value
	ErrFound = errors.New("not found")
)

type FuncMap map[string]interface{}

type State struct {
	parent  *State
	current reflect.Value
	filters map[string]interface{}
}

func EmptyState(data interface{}, filters FuncMap) *State {
	return EnclosedState(data, nil, filters)
}

func EnclosedState(data interface{}, parent *State, filters FuncMap) *State {
	return &State{
		current: valueOf(data),
		parent:  parent,
		filters: filters,
	}
}

func (s *State) Lookup(name string) (reflect.Value, error) {
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
