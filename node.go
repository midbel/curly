package curly

import (
	"bufio"
	"io"
	"reflect"
	"strconv"
)

type Node interface {
	Execute(io.Writer, interface{}) error
	execute(io.StringWriter, *state) error
}

type CommentNode struct {
	str string
}

func (c *CommentNode) Execute(w io.Writer, _ interface{}) error {
	wr := bufio.NewWriter(w)
	defer wr.Flush()
	return c.execute(wr, nil)
}

func (c *CommentNode) execute(w io.StringWriter, _ *state) error {
	w.WriteString("")
	return nil
}

type LiteralNode struct {
	str string
}

func (i *LiteralNode) Execute(w io.Writer, _ interface{}) error {
	wr := bufio.NewWriter(w)
	defer wr.Flush()
	return i.execute(wr, nil)
}

func (i *LiteralNode) execute(w io.StringWriter, _ *state) error {
	w.WriteString(i.str)
	return nil
}

type BlockNode struct {
	inverted bool
	key      Key
	nodes    []Node
}

func (b *BlockNode) Execute(w io.Writer, data interface{}) error {
	wr := bufio.NewWriter(w)
	defer wr.Flush()
	return b.execute(wr, emptyState(data))
}

func (b *BlockNode) execute(w io.StringWriter, data *state) error {
	val, err := b.key.resolve(data)
	if err != nil {
		return nil
	}
	ok := isTrue(val)
	if b.inverted {
		ok = !ok
	}
	if !ok {
		return nil
	}
	switch k := val.Kind(); k {
	case reflect.Struct, reflect.Map:
		b.executeNodes(w, enclosedState(val, data))
	case reflect.Array, reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			err = b.executeNodes(w, enclosedState(val.Index(i), data))
			if err != nil {
				return nil
			}
		}
	default:
		b.executeNodes(w, data)
	}
	return nil
}

func (b *BlockNode) executeNodes(w io.StringWriter, data *state) error {
	for i := range b.nodes {
		err := b.nodes[i].execute(w, data)
		if err != nil {
			return err
		}
	}
	return nil
}

type VariableNode struct {
	key     Key
	unescap bool
}

func (v *VariableNode) Execute(w io.Writer, data interface{}) error {
	wr := bufio.NewWriter(w)
	defer wr.Flush()
	return v.execute(wr, emptyState(data))
}

func (v *VariableNode) execute(w io.StringWriter, data *state) error {
	val, err := v.key.resolve(data)
	if err != nil {
		return nil
	}
	str, err := stringify(val, !v.unescap)
	if err == nil {
		w.WriteString(str)
	}
	return err
}

type Argument struct {
	literal string
	kind    rune
}

func (a Argument) get(data *state) reflect.Value {
	var val reflect.Value
	switch a.kind {
	case Integer:
		if arg, err := strconv.ParseInt(a.literal, 0, 64); err == nil {
			val = reflect.ValueOf(arg)
		}
	case Float:
		if arg, err := strconv.ParseFloat(a.literal, 64); err == nil {
			val = reflect.ValueOf(arg)
		}
	case Literal:
		val = reflect.ValueOf(a.literal)
	case Bool:
		if arg, err := strconv.ParseBool(a.literal); err == nil {
			val = reflect.ValueOf(arg)
		}
	case Ident:
		val, _ = data.resolve(a.literal)
	default:
	}
	return val
}

type Filter struct {
	name string
	args []Argument
}

var (
	errorType        = reflect.TypeOf((*error)(nil)).Elem()
	reflectValueType = reflect.TypeOf((*reflect.Value)(nil)).Elem()
)

func (f Filter) apply(data *state, value reflect.Value) (reflect.Value, error) {
	fn := reflect.ValueOf(filters[f.name])
	if !fn.IsValid() || fn.Kind() != reflect.Func {
		return value, nil
	}
	var (
		typ  = fn.Type()
		nin  = typ.NumIn()
		nout = typ.NumOut()
		args = append([]reflect.Value{value}, f.arguments(data)...)
	)
	if nin == 0 || nout == 0 || nout > 2 || len(args) != nin {
		return invalid, nil
	}
	for i := 0; i < nin; i++ {
		argtyp := fn.Type().In(i)
		if argtyp == reflectValueType {
			args[i] = reflect.ValueOf(args[i])
			continue
		}
		if !args[i].IsValid() && canBeNil(args[i].Type()) {
			args[i] = reflect.Zero(argtyp)
		}
		if args[i].Type().AssignableTo(argtyp) {
			continue
		}
		if !args[i].Type().ConvertibleTo(argtyp) {
			return invalid, nil
		}
		args[i] = args[i].Convert(argtyp)
	}
	rs := fn.Call(args)
	var err error
	if len(rs) == 2 && rs[1].Type() == errorType {
		err = rs[1].Interface().(error)
	}
	if rs[0].Type() == reflectValueType {
		rs[0] = rs[0].Interface().(reflect.Value)
	}
	return rs[0], err
}

func (f Filter) arguments(data *state) []reflect.Value {
	as := make([]reflect.Value, len(f.args))
	for i := range f.args {
		as[i] = f.args[i].get(data)
	}
	return as
}

type Key struct {
	name    string
	filters []Filter
}

func (k Key) resolve(data *state) (reflect.Value, error) {
	value, err := data.resolve(k.name)
	if err != nil {
		return invalid, err
	}
	for i := range k.filters {
		value, err = k.filters[i].apply(data, value)
		if err != nil {
			value = invalid
			break
		}
	}
	return value, err
}
