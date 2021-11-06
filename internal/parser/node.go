package parser

import (
	"fmt"
	"html"
	"io"
	"reflect"
	"strconv"

	"github.com/midbel/curly/internal/state"
	"github.com/midbel/curly/internal/token"
)

type Key interface {
	Ident() string
	resolve(state.State) (reflect.Value, error)
	isZero() bool
}

type Node interface {
	Execute(io.StringWriter, Nodeset, state.State) error
}

type NodeList []Node

func (n NodeList) Execute(w io.StringWriter, ns Nodeset, s state.State) error {
	for i := range n {
		if err := n[i].Execute(w, ns, s); err != nil {
			return err
		}
	}
	return nil
}

type Nodeset map[string]Node

func (ns Nodeset) Resolve(name string) Node {
	var n Node
	if ns != nil {
		n = ns[name]
	}
	return n
}

func (ns Nodeset) Merge(other Nodeset) Nodeset {
	for k, v := range other {
		if _, ok := ns[k]; ok {
			continue
		}
		ns[k] = v
	}
	return ns
}

type RootNode struct {
	Nodes NodeList
	Named Nodeset
}

func (r *RootNode) Execute(w io.StringWriter, ns Nodeset, data state.State) error {
	return r.Nodes.Execute(w, ns, data)
}

func (c *RootNode) Register(name string, node Node) {
	if c.Named == nil {
		c.Named = make(Nodeset)
	}
	c.Named[name] = node
}

type CommentNode struct {
	str string
}

func (c *CommentNode) Execute(_ io.StringWriter, _ Nodeset, _ state.State) error {
	return nil
}

type DefineNode struct {
	name  string
	nodes NodeList
}

func (d *DefineNode) Execute(w io.StringWriter, ns Nodeset, s state.State) error {
	return d.nodes.Execute(w, ns, s)
}

type PartialNode struct {
	file string
}

func (p *PartialNode) Execute(w io.StringWriter, ns Nodeset, data state.State) error {
	n, err := ParseFile(p.file)
	if err != nil {
		return err
	}
	if root, ok := n.(*RootNode); ok {
		ns = root.Named.Merge(ns)
	}
	return n.Execute(w, ns, data)
}

type ExecNode struct {
	name string
	key  Key
}

func (e *ExecNode) Execute(w io.StringWriter, ns Nodeset, data state.State) error {
	n := ns.Resolve(e.name)
	if n == nil {
		return fmt.Errorf("node %s not found", e.name)
	}
	if !e.key.isZero() {
		val, err := e.key.resolve(data)
		if err != nil {
			return err
		}
		data = state.EnclosedState(val, data, nil)
	}
	return n.Execute(w, ns, data)
}

type SectionNode struct {
	name  string
	nodes NodeList
}

func (s *SectionNode) Execute(w io.StringWriter, ns Nodeset, data state.State) error {
	n := ns.Resolve(s.name)
	if n != nil {
		return n.Execute(w, ns, data)
	}
	return s.nodes.Execute(w, ns, data)
}

type LiteralNode struct {
	str string
}

func (i *LiteralNode) Execute(w io.StringWriter, _ Nodeset, _ state.State) error {
	w.WriteString(i.str)
	return nil
}

type BlockNode struct {
	inverted bool
	key      Key
	nodes    NodeList
}

func (b *BlockNode) Execute(w io.StringWriter, ns Nodeset, data state.State) error {
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
		err = b.nodes.Execute(w, ns, state.EnclosedState(val, data, nil))
	case reflect.Array, reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			err = b.nodes.Execute(w, ns, state.EnclosedState(val.Index(i), data, nil))
			if err != nil {
				return nil
			}
		}
	default:
		err = b.nodes.Execute(w, ns, data)
	}
	return err
}

type AssignmentNode struct {
	ident string
	key   Key
}

func (a *AssignmentNode) Execute(_ io.StringWriter, _ Nodeset, data state.State) error {
	val, err := a.key.resolve(data)
	if err != nil {
		return err
	}
	data.Define(a.ident, val)
	return nil
}

type VariableNode struct {
	key     Key
	unescap bool
}

func (v *VariableNode) Execute(w io.StringWriter, _ Nodeset, data state.State) error {
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

func (a Argument) get(data state.State) reflect.Value {
	var val reflect.Value
	switch a.kind {
	case token.Integer:
		if arg, err := strconv.ParseInt(a.literal, 0, 64); err == nil {
			val = reflect.ValueOf(arg)
		}
	case token.Float:
		if arg, err := strconv.ParseFloat(a.literal, 64); err == nil {
			val = reflect.ValueOf(arg)
		}
	case token.Literal:
		val = reflect.ValueOf(a.literal)
	case token.Bool:
		if arg, err := strconv.ParseBool(a.literal); err == nil {
			val = reflect.ValueOf(arg)
		}
	case token.Ident:
		val, _ = data.Resolve(a.literal)
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

func (f Filter) apply(data state.State, value reflect.Value) (reflect.Value, error) {
	fn, err := data.Lookup(f.name)
	if err != nil {
		return fn, err
	}
	var (
		typ  = fn.Type()
		nin  = typ.NumIn()
		nout = typ.NumOut()
		args = append([]reflect.Value{value}, f.arguments(data)...)
	)
	if nin == 0 || nout == 0 || nout > 2 || len(args) != nin {
		return state.Invalid, nil
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
			return state.Invalid, nil
		}
		args[i] = args[i].Convert(argtyp)
	}
	rs := fn.Call(args)
	if len(rs) == 2 && rs[1].Type() == errorType {
		err, _ = rs[1].Interface().(error)
	}
	if rs[0].Type() == reflectValueType {
		rs[0] = rs[0].Interface().(reflect.Value)
	}
	return rs[0], err
}

func (f Filter) arguments(data state.State) []reflect.Value {
	as := make([]reflect.Value, len(f.args))
	for i := range f.args {
		as[i] = f.args[i].get(data)
	}
	return as
}

type ValueKey struct {
	literal string
	kind    rune
	filters []Filter
}

func (k ValueKey) Ident() string {
	return k.literal
}

func (k ValueKey) isZero() bool {
	return k.literal == ""
}

func (k ValueKey) resolve(data state.State) (reflect.Value, error) {
	var (
		raw interface{}
		err error
	)
	switch k.kind {
	case token.Integer:
		raw, err = strconv.ParseInt(k.literal, 0, 64)
	case token.Float:
		raw, err = strconv.ParseFloat(k.literal, 64)
	case token.Literal:
		raw = k.literal
	case token.Bool:
		raw, err = strconv.ParseBool(k.literal)
	default:
		err = fmt.Errorf("unsupported type")
	}
	if err != nil {
		return state.Invalid, err
	}
	value := reflect.ValueOf(raw)
	for i := range k.filters {
		value, err = k.filters[i].apply(data, value)
		if err != nil {
			value = state.Invalid
			break
		}
	}
	return value, err
}

type IdentKey struct {
	name    string
	filters []Filter
}

func (k IdentKey) Ident() string {
	return k.name
}

func (k IdentKey) isZero() bool {
	return k.name == ""
}

func (k IdentKey) resolve(data state.State) (reflect.Value, error) {
	value, err := data.Resolve(k.name)
	if err != nil {
		return state.Invalid, err
	}
	for i := range k.filters {
		value, err = k.filters[i].apply(data, value)
		if err != nil {
			value = state.Invalid
			break
		}
	}
	return value, err
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
