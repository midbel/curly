package curly

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var ErrUnexpected = errors.New("unexpected token")

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

func parseFile(f string) (*Template, error) {
	r, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	return Parse(r)
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
		Section:     p.parseSection,
		Define:      p.parseDefine,
		Exec:        p.parseExec,
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
	t := New("")
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
			return nil, p.unexpectedToken()
		}
		if err != nil {
			return nil, err
		}
		if node != nil {
			t.nodes = append(t.nodes, node)
		}
	}
	return t, nil
}

func (p *Parser) parseSection() (Node, error) {
	p.next()
	if p.curr.Type != Ident {
		return nil, p.unexpectedToken()
	}
	s := SectionNode{
		name: p.curr.Literal,
	}
	if err := p.ensureClose(); err != nil {
		return nil, err
	}
	ns, err := p.parseBody(s.name)
	if err != nil {
		return nil, err
	}
	s.nodes = ns
	return &s, nil
}

func (p *Parser) parseDefine() (Node, error) {
	p.next()
	if p.curr.Type != Ident {
		return nil, p.unexpectedToken()
	}
	d := DefineNode{
		name: p.curr.Literal,
	}
	if err := p.ensureClose(); err != nil {
		return nil, err
	}
	ns, err := p.parseBody(d.name)
	if err != nil {
		return nil, err
	}
	d.nodes = ns
	return &d, nil
}

func (p *Parser) parseExec() (Node, error) {
	p.next()
	if p.curr.Type != Ident {
		return nil, p.unexpectedToken()
	}
	e := ExecNode{
		name: p.curr.Literal,
	}
	if p.peek.Type != Ident {
		return &e, p.ensureClose()
	}
	p.next()
	key, err := p.parseKey()
	if err != nil {
		return nil, err
	}
	e.key = key
	return &e, nil
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
	ns, err := p.parseBody(b.key.name)
	if err != nil {
		return nil, err
	}
	b.nodes = ns
	return &b, nil
}

func (p *Parser) parseBody(name string) ([]Node, error) {
	var ns []Node
	for !p.done() {
		if p.curr.Type == Literal {
			node, err := p.parseLiteral()
			if err != nil {
				return nil, err
			}
			ns = append(ns, node)
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
			ns = append(ns, node)
		}
	}
	if p.curr.Type != End {
		return nil, p.unexpectedToken()
	}
	p.next()
	if p.curr.Type != Ident && p.curr.Literal != name {
		return nil, p.unexpectedToken()
	}
	return ns, p.ensureClose()
}

func (p *Parser) parseNode() (Node, error) {
	parse, ok := p.parsers[p.curr.Type]
	if !ok {
		return nil, p.unexpectedToken()
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

func (p *Parser) unexpectedToken() error {
	return fmt.Errorf("<%d:%d> %w: %s", p.curr.Line, p.curr.Column, ErrUnexpected, p.curr)
}

func (p *Parser) done() bool {
	return p.curr.Type == EOF
}

func (p *Parser) next() {
	p.curr = p.peek
	p.peek = p.scan.Scan()
}
