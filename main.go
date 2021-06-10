package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"html"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

const template1 = `
hello {{& World }} - {{#truthy}}{{&name}}{{/truthy}}
{{^Falsy}}nothing will be rendered {{& falsy}}{{/Falsy}}
>>{{! comment will not be rendered }}<<
`

const template2 = `
my repositories:
{{# repo }}
  {{Name}} (version: {{Version}})
{{/ repo }}
contacts: {{ Email }}
`

const template3 = `
* {{default_tags}}
{{=<% %>=}}
* <% erb_style_tags %>
<%={{ }}=%>
* {{ default_tags_again }}
`

const template4 = `hello {{ array | cut 3 | join "," false }} - {{=<% %>=}} - <% var %>`

const template5 = `
hello {{world | lower}} - {{world | upper | len}}

hello {{ text | split "_" | reverse | join "/" }}
`

type Name struct {
	Name    string
	Version int
}

func main() {
	c := struct {
		World string `tag:"world"`
		Text  string `tag:"text"`
	}{
		World: "World",
		Text:  "under_score_text_with_extra_words",
	}
	exec(template5, c)
}

func scan(template string) {
	s, _ := Scan(strings.NewReader(template))
	for {
		tok := s.Scan()
		if tok.Type == EOF {
			break
		}
		fmt.Println(tok)
	}
}

func exec(template string, data interface{}) {
	r := strings.NewReader(strings.TrimSpace(template))

	t, err := Parse(r)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	debug(t)

	if data == nil {
		return
	}
	err = t.Execute(os.Stdout, data)
	fmt.Println()
	fmt.Println(err)
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

type Node interface {
	Execute(io.Writer, interface{}) error
	execute(io.StringWriter, *state) error
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

type Parser struct {
	scan *Scanner
	curr Token
	peek Token

	parsers map[rune]func() (Node, error)
}

func Parse(r io.Reader) (*Template, error) {
	p, err := NewParser(r)
	if err != nil {
		return nil, err
	}
	return p.Parse()
}

func NewParser(r io.Reader) (*Parser, error) {
	s, err := Scan(r)
	if err != nil {
		return nil, err
	}

	var p Parser
	p.scan = s
	p.parsers = map[rune]func() (Node, error){
		Block:       p.parseBlock,
		Inverted:    p.parseBlock,
		EscapeVar:   p.parseVariable,
		UnescapeVar: p.parseVariable,
		Comment:     p.parseComment,
		Partial:     p.parsePartial,
		Delim:       p.parseDelim,
	}

	p.next()
	p.next()

	return &p, nil
}

func (p *Parser) Parse() (*Template, error) {
	var t Template
	for !p.done() {
		var (
			node Node
			err  error
		)
		switch p.curr.Type {
		case Literal:
			node, err = p.parseLiteral()
		case Open:
			p.next()
			node, err = p.parseNode()
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unexpected token: %s", p.curr)
		}
		if err != nil {
			return nil, err
		}
		if node != nil {
			t.nodes = append(t.nodes, node)
		}
	}
	return &t, nil
}

func (p *Parser) parsePartial() (Node, error) {
	p.next()

	r, err := os.Open(p.curr.Literal)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	x, err := NewParser(r)
	if err != nil {
		return nil, err
	}
	t, err := x.Parse()
	if err != nil {
		return nil, err
	}
	return t, p.ensureClose()
}

func (p *Parser) parseDelim() (Node, error) {
	p.next()
	if p.curr.Type != Literal {
		return nil, p.unexpectedToken()
	}
	left := p.curr.Literal
	p.next()
	if p.curr.Type != Literal {
		return nil, p.unexpectedToken()
	}
	right := p.curr.Literal

	p.scan.SetDelimiter(left, right)
	return nil, p.ensureClose()
}

func (p *Parser) parseBlock() (Node, error) {
	b := BlockNode{
		inverted: p.curr.Type == Inverted,
	}
	p.next()
	key, err := p.parseKey()
	if err != nil {
		return nil, err
	}
	b.key = key
	if err := p.ensureClose(); err != nil {
		return nil, err
	}
	for !p.done() {
		if p.curr.Type == Literal {
			node, err := p.parseLiteral()
			if err != nil {
				return nil, err
			}
			b.nodes = append(b.nodes, node)
			continue
		}
		if p.curr.Type != Open {
			return nil, p.unexpectedToken()
		}
		p.next()
		if p.curr.Type == End {
			break
		}
		node, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if node != nil {
			b.nodes = append(b.nodes, node)
		}
	}
	if p.curr.Type != End {
		return nil, p.unexpectedToken()
	}
	p.next()
	if p.curr.Type != Ident && p.curr.Literal != b.key.name {
		return nil, p.unexpectedToken()
	}
	return &b, p.ensureClose()
}

func (p *Parser) parseNode() (Node, error) {
	parse, ok := p.parsers[p.curr.Type]
	if !ok {
		return nil, fmt.Errorf("parsing block: unexpected token %s", p.curr)
	}
	return parse()
}

func (p *Parser) parseLiteral() (Node, error) {
	defer p.next()
	return &LiteralNode{str: p.curr.Literal}, nil
}

func (p *Parser) parseComment() (Node, error) {
	p.next()
	n := CommentNode{str: p.curr.Literal}
	return &n, p.ensureClose()
}

func (p *Parser) parseVariable() (Node, error) {
	n := VariableNode{
		unescap: p.curr.unescape(),
	}
	p.next()
	key, err := p.parseKey()
	if err != nil {
		return nil, err
	}
	n.key = key
	return &n, nil
}

func (p *Parser) parseKey() (Key, error) {
	var k Key
	if p.curr.Type != Ident {
		return k, p.unexpectedToken()
	}
	k.name = p.curr.Literal
	for {
		if p.peek.Type != Pipe {
			break
		}
		p.next()
		p.next()
		f, err := p.parseFilter()
		if err != nil {
			return k, err
		}
		k.filters = append(k.filters, f)
	}
	return k, p.ensureClose()
}

func (p *Parser) parseFilter() (Filter, error) {
	var f Filter
	if p.curr.Type != Ident {
		return f, p.unexpectedToken()
	}
	f.name = p.curr.Literal
	for {
		if !p.peek.isValue() {
			break
		}
		p.next()
		a := Argument{
			literal: p.curr.Literal,
			kind:    p.curr.Type,
		}
		f.args = append(f.args, a)
	}
	return f, nil
}

func (p *Parser) ensureClose() error {
	p.next()
	if p.curr.Type != Close {
		return p.unexpectedToken()
	}
	p.next()
	return nil
}

var ErrUnexpected = errors.New("unexpected token")

func (p *Parser) unexpectedToken() error {
	return fmt.Errorf("%w: %s", ErrUnexpected, p.curr)
}

func (p *Parser) done() bool {
	return p.curr.Type == EOF
}

func (p *Parser) next() {
	p.curr = p.peek
	p.peek = p.scan.Scan()
}

const (
	EOF rune = -(iota + 1)
	Literal
	Ident
	Integer
	Float
	Bool
	Open
	Close
	Delim
	Block
	Inverted
	EscapeVar
	UnescapeVar
	Pipe
	Rev
	And
	Or
	Not
	Partial
	Comment
	End
	Invalid
)

const (
	amper      = '&'
	pound      = '#'
	caret      = '^'
	slash      = '/'
	lbrace     = '{'
	rbrace     = '}'
	bang       = '!'
	angle      = '>'
	equal      = '='
	space      = ' '
	tab        = '\t'
	cr         = '\r'
	nl         = '\n'
	underscore = '_'
	pipe       = '|'
	squote     = '\''
	dquote     = '"'
	dot        = '.'
	dash       = '-'
)

type Token struct {
	Literal string
	Unescap bool
	Type    rune
}

func (t Token) isValue() bool {
	switch t.Type {
	case Literal, Integer, Float, Bool, Ident:
		return true
	default:
		return false
	}
}

func (t *Token) unescape() bool {
	return t.Type == UnescapeVar
}

func (t Token) String() string {
	var prefix string
	switch t.Type {
	case EOF:
		return "<eof>"
	case And:
		return "<and>"
	case Or:
		return "<or>"
	case Pipe:
		return "<pipe>"
	case Open:
		return "<open>"
	case Close:
		return "<close>"
	case Block:
		return "<block>"
	case Inverted:
		return "<inverted>"
	case Comment:
		return "<comment>"
	case Partial:
		return "<partial>"
	case EscapeVar, UnescapeVar:
		return "<variable>"
	case Delim:
		return "<delimiter>"
	case End:
		return "<end>"
	case Not:
		return "<not>"
	case Rev:
		return "<reverse>"
	case Ident:
		prefix = "identifier"
	case Integer:
		prefix = "integer"
	case Float:
		prefix = "float"
	case Bool:
		prefix = "boolean"
	case Literal:
		prefix = "literal"
	case Invalid:
		prefix = "invalid"
	default:
		prefix = "unknown"
	}
	return fmt.Sprintf("%s(%s)", prefix, t.Literal)
}

type Scanner struct {
	input []byte
	curr  int
	next  int
	char  rune

	line   int
	column int
	seen   int

	scan    func(*Token)
	between bool

	left  []rune
	right []rune
}

func Scan(r io.Reader) (*Scanner, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	s := Scanner{
		input: bytes.ReplaceAll(buf, []byte{cr, nl}, []byte{nl}),
		line:  1,
		left:  []rune{lbrace, lbrace},
		right: []rune{rbrace, rbrace},
	}
	s.read()
	return &s, nil
}

func (s *Scanner) Scan() (t Token) {
	if s.isEOF() {
		t.Type = EOF
		return t
	}
	if s.scan != nil {
		s.scan(&t)
		return t
	}
	switch {
	case s.isOpen():
		s.skipOpen()
		t.Type = Open
		s.scan, s.between = s.scanType, true
	case s.isClose():
		s.skipClose()
		t.Type = Close
		s.scan, s.between = nil, false
	case isOperator(s.char):
		s.scanOperator(&t)
		s.scan = s.scanIdent
	case isQuote(s.char):
		s.scanString(&t)
	case isDigit(s.char):
		s.scanNumber(&t)
	default:
		if s.between {
			s.scanIdent(&t)
		} else {
			s.scanLiteral(&t)
		}
	}
	return t
}

func (s *Scanner) SetDelimiter(left, right string) {
	if left != "" {
		s.left = []rune(left)
	}
	if right != "" {
		s.right = []rune(right)
	}
}

func (s *Scanner) scanComment(t *Token) {
	s.scan = nil

	pos := s.curr
	for !s.isEOF() && !s.isClose() {
		s.read()
	}
	t.Type = Literal
	if !s.isClose() {
		t.Type = Invalid
	}
	t.Literal = strings.TrimSpace(string(s.input[pos:s.curr]))
}

func (s *Scanner) scanLiteral(t *Token) {
	pos := s.curr
	for !s.isEOF() && !s.isOpen() {
		s.read()
	}
	t.Type = Literal
	t.Literal = string(s.input[pos:s.curr])
}

func (s *Scanner) scanIdent(t *Token) {
	s.scan = nil
	s.skipBlank()
	defer s.skipBlank()
	if !isLetter(s.char) {
		t.Type = Invalid
		return
	}
	pos := s.curr
	for isIdent(s.char) {
		s.read()
	}
	t.Literal = string(s.input[pos:s.curr])
	switch t.Literal {
	case "true", "false":
		t.Type = Bool
	default:
		t.Type = Ident
	}
}

func (s *Scanner) scanOperator(t *Token) {
	t.Type = Invalid
	switch k := s.peek(); s.char {
	case pipe:
		t.Type = Pipe
		if k == pipe {
			t.Type = Or
		}
	case amper:
		if k == amper {
			t.Type = And
		}
	case bang:
		t.Type = Not
	case dash:
		t.Type = Rev
	default:
	}
	if t.Type == And || t.Type == Or {
		s.read()
	}
	s.read()
	s.skipBlank()
}

func (s *Scanner) scanString(t *Token) {
	var (
		quote = s.char
		pos   = s.curr
	)
	s.read()
	pos = s.curr
	for !s.isEOF() && s.char != quote {
		s.read()
	}
	t.Type = Literal
	t.Literal = string(s.input[pos:s.curr])
	if s.char != quote {
		t.Type = Invalid
	}
	s.read()
	s.skipBlank()
}

func (s *Scanner) scanNumber(t *Token) {
	pos := s.curr
	for isDigit(s.char) {
		s.read()
	}
	t.Type = Integer
	if s.char == dot {
		for isDigit(s.char) {
			s.read()
		}
		t.Type = Float
	}
	t.Literal = string(s.input[pos:s.curr])
	s.skipBlank()
}

func (s *Scanner) scanOpenDelimiter(t *Token) {
	// {{=<punct><blank>
	s.scan = s.scanCloseDelimiter
	s.scanDelim(t, func() bool { return !isBlank(s.char) })
	if !isBlank(s.char) {
		t.Type = Invalid
	} else {
		s.skipBlank()
	}
}

func (s *Scanner) scanCloseDelimiter(t *Token) {
	// <blank><punct>=}}
	s.scan = nil
	s.scanDelim(t, func() bool { return s.char != equal })
	if s.char != equal {
		t.Type = Invalid
	} else {
		s.read()
	}
}

func (s *Scanner) scanDelim(t *Token, accept func() bool) {
	pos := s.curr
	for !s.isEOF() && accept() {
		if isLetter(s.char) || isDigit(s.char) {
			t.Type = Invalid
		}
		s.read()
	}
	if t.Type == 0 {
		t.Type = Literal
	}
	t.Literal = string(s.input[pos:s.curr])
}

func (s *Scanner) scanType(t *Token) {
	s.scan = s.scanIdent
	switch s.char {
	case pound:
		t.Type = Block
	case caret:
		t.Type = Inverted
	case bang:
		t.Type = Comment
		s.scan = s.scanComment
	case equal:
		t.Type = Delim
		s.scan = s.scanOpenDelimiter
	case angle:
		t.Type = Partial
	case slash:
		t.Type = End
	case amper:
		t.Type = UnescapeVar
	default:
		t.Type = EscapeVar
	}
	if t.Type != EscapeVar {
		s.read()
	}
	s.skipBlank()
}

func (s *Scanner) read() {
	if s.curr >= len(s.input) {
		s.char = 0
		return
	}
	r, n := utf8.DecodeRune(s.input[s.next:])
	if r == utf8.RuneError {
		s.char = 0
		s.next = len(s.input)
	}
	last := s.char
	s.char, s.curr, s.next = r, s.next, s.next+n

	if last == nl {
		s.line++
		s.seen, s.column = s.column, 1
	} else {
		s.column++
	}
}

func (s *Scanner) isEOF() bool {
	return s.char == 0 || s.char == utf8.RuneError
}

func (s *Scanner) isOpen() bool {
	return s.isTag(s.left)
}

func (s *Scanner) isClose() bool {
	return s.isTag(s.right)
}

func (s *Scanner) isTag(set []rune) bool {
	var (
		siz int
		ok  bool
	)
	for i := 0; i < len(set); i++ {
		char, n := utf8.DecodeRune(s.input[s.curr+siz:])
		if ok = set[i] == char; !ok {
			return false
		}
		siz += n
	}
	return ok
}

func (s *Scanner) skipOpen() {
	s.skipN(len(s.left))
}

func (s *Scanner) skipClose() {
	s.skipN(len(s.right))
}

func (s *Scanner) skipN(n int) {
	for i := 0; i < n; i++ {
		s.read()
	}
}

func (s *Scanner) skipNL() {
	for s.char == nl {
		s.read()
	}
}

func (s *Scanner) skipBlank() {
	for isBlank(s.char) {
		s.read()
	}
}

func (s *Scanner) peek() rune {
	r, _ := utf8.DecodeRune(s.input[s.next:])
	return r
}

func isTag(r rune) bool {
	switch r {
	case pound, caret, angle, slash, amper, pipe, equal:
	default:
		return false
	}
	return true
}

func isOperator(r rune) bool {
	return r == pipe || r == amper || r == bang || r == dash
}

func isQuote(r rune) bool {
	return r == squote || r == dquote
}

func isIdent(r rune) bool {
	return isLetter(r) || isDigit(r) || r == underscore
}

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isBlank(r rune) bool {
	return r == space || r == tab
}

func debug(n Node) {
	debugWithLevel(n, 0)
}

func debugWithLevel(n Node, level int) {
	prefix := strings.Repeat(" ", level)
	fmt.Print(prefix)
	switch n := n.(type) {
	case *Template:
		fmt.Print("template [")
		fmt.Println()
		for i := range n.nodes {
			debugWithLevel(n.nodes[i], level+2)
		}
		fmt.Print(prefix)
		fmt.Println("]")
	case *BlockNode:
		fmt.Print("block(key: ")
		fmt.Print(n.key.name)
		fmt.Print(") [")
		fmt.Println()
		for i := range n.nodes {
			debugWithLevel(n.nodes[i], level+2)
		}
		fmt.Print(prefix)
		fmt.Println("]")
	case *VariableNode:
		fmt.Print("variable(key: ")
		fmt.Print(n.key.name)
		fmt.Print(", unescape: ")
		fmt.Print(n.unescap)
		fmt.Print(")")
		fmt.Println()
	case *CommentNode:
		fmt.Print("comment(")
		fmt.Print(n.str)
		fmt.Print(")")
		fmt.Println()
	case *LiteralNode:
		fmt.Print("literal(str: ")
		for j, i := range strings.Split(n.str, "\n") {
			if j > 0 {
				fmt.Println()
				fmt.Print(prefix)
			}
			fmt.Print(i)
		}
		fmt.Print(")")
		fmt.Println()
	}
}
