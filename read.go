package curly

import (
	"errors"
	"fmt"
	"io"
	"strconv"
)

var errSyntax = errors.New("syntax error")

func Decode(r io.Reader) (any, error) {
	p := createParser(r, stdMode)
	return p.Parse()
}

func Decode5(r io.Reader) (any, error) {
	p := createParser(r, json5Mode)
	return p.Parse()
}

type Parser struct {
	scan *Scanner
	curr Token
	peek Token

	mode
}

func createParser(r io.Reader, jm mode) *Parser {
	p := &Parser{
		scan: Scan(r, jm),
		mode: jm,
	}
	p.next()
	p.next()
	return p
}

func (p *Parser) Parse() (any, error) {
	val, err := p.parse()
	if err != nil {
		return nil, err
	}
	if !p.done() {
		return nil, p.syntaxError("unexpected token after value")
	}
	return val, p.scan.Err()
}

func (p *Parser) parse() (any, error) {
	if err := p.scan.Err(); err != nil {
		return nil, err
	}
	switch p.curr.Type {
	case BegArr:
		return p.parseArray()
	case BegObj:
		return p.parseObject()
	case String:
		return p.parseString(), nil
	case Number:
		return p.parseNumber()
	case Boolean:
		return p.parseBool(), nil
	case Null:
		return p.parseNull(), nil
	case Comment:
		if p.mode == stdMode {
			return nil, p.syntaxError("comments not supported in standard mode")
		}
		p.skipComment()
		return p.parse()
	default:
		return nil, p.syntaxError("invalid token")
	}
}

func (p *Parser) parseKey() (string, error) {
	switch {
	case p.is(String):
	case p.is(Ident) && p.mode.isExtended():
	default:
		return "", p.syntaxError("object key should be string")
	}
	key := p.currentLiteral()
	p.next()
	if !p.is(Colon) {
		return "", p.syntaxError("missing colon after key")
	}
	p.next()
	return key, nil
}

func (p *Parser) parseObject() (any, error) {
	p.next()
	obj := make(map[string]any)
	for !p.done() && !p.is(EndObj) {
		k, err := p.parseKey()
		if err != nil {
			return nil, err
		}
		a, err := p.parse()
		if err != nil {
			return nil, err
		}

		obj[k] = a
		switch {
		case p.is(Comma):
			p.next()
			if p.is(EndObj) && !p.mode.isExtended() {
				return nil, p.syntaxError("trailing comma not allowed")
			}
		case p.is(EndObj):
		default:
			return nil, p.syntaxError("expected ',' or '}'")
		}
	}
	if !p.is(EndObj) {
		return nil, p.syntaxError("missing '}' at end of object")
	}
	p.next()
	return obj, nil
}

func (p *Parser) parseArray() (any, error) {
	p.next()
	var arr []any
	for !p.done() && !p.is(EndArr) {
		a, err := p.parse()
		if err != nil {
			return nil, err
		}
		arr = append(arr, a)
		switch {
		case p.is(Comma):
			p.next()
			if p.is(EndArr) && !p.mode.isExtended() {
				return nil, p.syntaxError("trailing comma not allowed")
			}
		case p.is(EndArr):
		default:
			return nil, p.syntaxError("expected ',' or ']'")
		}
	}
	if !p.is(EndArr) {
		return nil, p.syntaxError("missing ']' at end of array")
	}
	p.next()
	return arr, nil
}

func (p *Parser) parseNumber() (any, error) {
	defer p.next()
	i, err := strconv.ParseInt(p.currentLiteral(), 0, 64)
	if err == nil {
		return i, nil
	}
	n, err := strconv.ParseFloat(p.currentLiteral(), 64)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (p *Parser) parseBool() any {
	defer p.next()
	if p.currentLiteral() == "true" {
		return true
	}
	return false
}

func (p *Parser) parseString() any {
	defer p.next()
	return p.currentLiteral()
}

func (p *Parser) parseNull() any {
	defer p.next()
	return nil
}

func (p *Parser) skipComment() {
	for _, p.is(Comment) {
		p.next()
	}
}

func (p *Parser) done() bool {
	return p.is(EOF)
}

func (p *Parser) is(kind rune) bool {
	return p.curr.Type == kind
}

func (p *Parser) next() {
	p.curr = p.peek
	p.peek = p.scan.Scan()
}

func (p *Parser) currentLiteral() string {
	return p.curr.Literal
}

func (p *Parser) syntaxError(msg string) error {
	return fmt.Errorf("%w: %s", errSyntax, msg)
}

type mode int8

const (
	stdMode mode = 1 << iota
	json5Mode
)

func (m mode) isStd() bool {
	return m == stdMode
}

func (m mode) isExtended() bool {
	return m == json5Mode
}
