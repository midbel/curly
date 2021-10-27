package curly

import (
	"bufio"
	"errors"
	"fmt"
	"html"
	"io"
	"reflect"
	"strconv"
	"strings"
)

type Template struct {
	nodes []Node
}

func (t *Template) Execute(w io.Writer, data interface{}) error {
	wr := bufio.NewWriter(w)
	defer wr.Flush()
	return t.execute(wr, emptyState(data))
}

func (t *Template) execute(w io.StringWriter, data *state) error {
	for i := range t.nodes {
		err := t.nodes[i].execute(w, data)
		if err != nil {
			return err
		}
	}
	return nil
}

var filters = map[string]interface{}{
	"join":    strings.Join,
	"split":   strings.Split,
	"replace": strings.Replace,
	"lower":   strings.ToLower,
	"upper":   strings.ToUpper,
	"count":   strings.Count,
	"trim":    strings.TrimSpace,
	"repeat":  strings.Repeat,
	"len":     length,
	"first":   first,
	"last":    last,
	"firstn":  firstN,
	"lastn":   lastN,
	"reverse": reverse,
}

func reverse(value reflect.Value) reflect.Value {
	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		var (
			size  = value.Len()
			slice = reflect.MakeSlice(value.Type(), 0, size)
		)
		for i := size - 1; i >= 0; i-- {
			slice = reflect.Append(slice, value.Index(i))
		}
		return slice
	default:
		return invalid
	}
}

func first(value reflect.Value) reflect.Value {
	return firstN(value, 1)
}

func last(value reflect.Value) reflect.Value {
	return lastN(value, 1)
}

func firstN(value reflect.Value, n int) reflect.Value {
	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		if value.Len() < n {
			return invalid
		}
		return value.Slice(0, n)
	default:
		return invalid
	}
}

func lastN(value reflect.Value, n int) reflect.Value {
	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		if value.Len() < n {
			return invalid
		}
		return value.Slice(value.Len()-n, value.Len())
	default:
		return invalid
	}
}

func length(value reflect.Value) int {
	switch value.Kind() {
	case reflect.Array, reflect.Slice, reflect.String, reflect.Map:
		return value.Len()
	default:
		return 0
	}
}

var invalid reflect.Value

var errFound = errors.New("not found")

type state struct {
	parent  *state
	current reflect.Value
}

func emptyState(data interface{}) *state {
	return enclosedState(data, nil)
}

func enclosedState(data interface{}, parent *state) *state {
	return &state{
		current: valueOf(data),
		parent:  parent,
	}
}

func (s *state) resolve(key string) (reflect.Value, error) {
	v, err := s.find(key)
	if err != nil && s.parent != nil {
		v, err = s.parent.find(key)
	}
	return v, err
}

func (s *state) find(key string) (reflect.Value, error) {
	return s.findWith(key, s.current)
}

func (s *state) findWith(key string, value reflect.Value) (reflect.Value, error) {
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
	return invalid, errFound
}

func (s *state) lookupStruct(key string, value reflect.Value) (reflect.Value, error) {
	t := value.Type()
	for i := 0; i < value.NumField(); i++ {
		sf := t.Field(i)
		if sf.Name == key || sf.Tag.Get("tag") == key {
			return value.Field(i), nil
		}
	}
	return invalid, errFound
}

func (s *state) lookupMap(key string, value reflect.Value) (reflect.Value, error) {
	t := value.Type().Key()
	if !t.AssignableTo(reflect.TypeOf(key)) {
		return invalid, errFound
	}
	val := value.MapIndex(reflect.ValueOf(key))
	if val.IsZero() {
		return invalid, errFound
	}
	return val, nil
}

func valueOf(v interface{}) reflect.Value {
	if v, ok := v.(reflect.Value); ok {
		return v
	}
	return reflect.ValueOf(v)
}

func isTrue(v reflect.Value) bool {
	if !v.IsValid() {
		return false
	}
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0
	case reflect.Map, reflect.Array, reflect.Slice, reflect.String:
		return v.Len() != 0
	case reflect.Ptr, reflect.Interface:
		return !v.IsNil()
	case reflect.Struct:
		return true
	default:
		return false
	}
}

func canBeNil(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return true
	default:
		return false
	}
}

func stringify(v reflect.Value, escape bool) (string, error) {
	var (
		str string
		err error
	)
	switch v.Kind() {
	case reflect.String:
		str = v.String()
	case reflect.Bool:
		str = strconv.FormatBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		str = strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		str = strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		str = strconv.FormatFloat(v.Float(), 'g', -1, 64)
	default:
		err = fmt.Errorf("%s can not be stringify", v)
	}
	if err == nil && escape {
		str = html.EscapeString(str)
	}
	return str, err
}
