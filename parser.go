package curly

import (
	"errors"
	"fmt"
	"io"
	"strconv"
)

var errSyntax = errors.New("syntax error")

func Decode(r io.Reader) (any, error) {
	p := newStdParser(r)
	return p.Parse()
}

func Decode5(r io.Reader) (any, error) {
	p := newFiveParser(r)
	return p.Parse()
}

type stdParser struct {
	scan *Scanner
	curr Token
	peek Token
}

func newStdParser(r io.Reader) *stdParser {
	p := &stdParser{
		scan: Scan(r),
	}
	p.next()
	p.next()
	return p
}

func (p *stdParser) Parse() (any, error) {
	val, err := p.parse()
	if err != nil {
		return nil, err
	}
	if !p.done() {
		return nil, syntaxError("unexpected token after value")
	}
	return val, p.scan.Err()
}

func (p *stdParser) parse() (any, error) {
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
	default:
		return nil, syntaxError("invalid token")
	}
}

func (p *stdParser) parseKey() (string, error) {
	if !p.is(String) {
		return "", syntaxError("object key should be string")
	}
	key := p.currentLiteral()
	p.next()
	if !p.is(Colon) {
		return "", syntaxError("missing colon after key")
	}
	p.next()
	return key, nil
}

func (p *stdParser) parseObject() (any, error) {
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
			if p.is(EndObj) {
				return nil, syntaxError("trailing comma not allowed")
			}
		case p.is(EndObj):
		default:
			return nil, syntaxError("expected ',' or '}'")
		}
	}
	if !p.is(EndObj) {
		return nil, syntaxError("missing '}' at end of object")
	}
	p.next()
	return obj, nil
}

func (p *stdParser) parseArray() (any, error) {
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
			if p.is(EndArr) {
				return nil, syntaxError("trailing comma not allowed")
			}
		case p.is(EndArr):
		default:
			return nil, syntaxError("expected ',' or ']'")
		}
	}
	if !p.is(EndArr) {
		return nil, syntaxError("missing ']' at end of array")
	}
	p.next()
	return arr, nil
}

func (p *stdParser) parseNumber() (any, error) {
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

func (p *stdParser) parseBool() any {
	defer p.next()
	if p.currentLiteral() == "true" {
		return true
	}
	return false
}

func (p *stdParser) parseString() any {
	defer p.next()
	return p.currentLiteral()
}

func (p *stdParser) parseNull() any {
	defer p.next()
	return nil
}

func (p *stdParser) done() bool {
	return p.is(EOF)
}

func (p *stdParser) is(kind rune) bool {
	return p.curr.Type == kind
}

func (p *stdParser) next() {
	p.curr = p.peek
	p.peek = p.scan.Scan()
}

func (p *stdParser) currentLiteral() string {
	return p.curr.Literal
}

type fiveParser struct {
	scan *Scanner
	curr Token
	peek Token
}

func newFiveParser(r io.Reader) *fiveParser {
	p := &fiveParser{
		scan: Scan(r),
	}
	p.next()
	p.next()
	return p
}

func (p *fiveParser) Parse() (any, error) {
	val, err := p.parse()
	if err != nil {
		return nil, err
	}
	if !p.done() {
		return nil, syntaxError("unexpected token after value")
	}
	return val, p.scan.Err()
}

func (p *fiveParser) parse() (any, error) {
	if err := p.scan.Err(); err != nil {
		return nil, err
	}
	switch p.curr.Type {
	case BegArr:
		return p.parseArray()
	case BegObj:
		return p.parseObject()
	case String, StringSingleQuoted:
		return p.parseString(), nil
	case Number, NumberHex, NumberTrailingDot, NumberLeadingDot, NumberLeadingPlus:
		return p.parseNumber()
	case Boolean:
		return p.parseBool(), nil
	case Null:
		return p.parseNull(), nil
	case CommentLine, CommentMulti:
		p.skipComments()
		return p.parse()
	default:
		return nil, syntaxError("invalid token")
	}
}

func (p *fiveParser) parseKey() (string, error) {
	switch p.curr.Type {
	case String, StringSingleQuoted, Ident:
	default:
		return "", syntaxError("object key should be string")
	}
	key := p.currentLiteral()
	p.next()
	if !p.is(Colon) {
		return "", syntaxError("missing colon after key")
	}
	p.next()
	return key, nil
}

func (p *fiveParser) parseObject() (any, error) {
	p.next()
	obj := make(map[string]any)
	for !p.done() {
		p.skipComments()
		if p.is(EndObj) {
			break
		}
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
		case p.is(EndObj):
		default:
			return nil, syntaxError("expected ',' or '}'")
		}
	}
	if !p.is(EndObj) {
		return nil, syntaxError("missing '}' at end of object")
	}
	p.next()
	return obj, nil
}

func (p *fiveParser) parseArray() (any, error) {
	p.next()
	var arr []any
	for !p.done() {
		p.skipComments()
		if p.is(EndArr) {
			break
		}
		a, err := p.parse()
		if err != nil {
			return nil, err
		}
		arr = append(arr, a)
		switch {
		case p.is(Comma):
			p.next()
		case p.is(EndArr):
		default:
			return nil, syntaxError("expected ',' or ']'")
		}
	}
	if !p.is(EndArr) {
		return nil, syntaxError("missing ']' at end of array")
	}
	p.next()
	return arr, nil
}

func (p *fiveParser) parseNumber() (any, error) {
	base := 10
	if p.is(NumberHex) {
		base = 16
	}
	defer p.next()
	i, err := strconv.ParseInt(p.currentLiteral(), base, 64)
	if err == nil {
		return i, nil
	}
	if err != nil && p.is(NumberHex) {
		return nil, syntaxError("invalid hex number")
	}
	n, err := strconv.ParseFloat(p.currentLiteral(), 64)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (p *fiveParser) parseBool() any {
	defer p.next()
	if p.currentLiteral() == "true" {
		return true
	}
	return false
}

func (p *fiveParser) parseString() any {
	defer p.next()
	return p.currentLiteral()
}

func (p *fiveParser) parseNull() any {
	defer p.next()
	return nil
}

func (p *fiveParser) skipComments() {
	for p.is(CommentLine) || p.is(CommentMulti) {
		p.next()
	}
}

func (p *fiveParser) done() bool {
	return p.is(EOF)
}

func (p *fiveParser) is(kind rune) bool {
	return p.curr.Type == kind
}

func (p *fiveParser) next() {
	p.curr = p.peek
	p.peek = p.scan.Scan()
}

func (p *fiveParser) currentLiteral() string {
	return p.curr.Literal
}

type uclParser struct {
	scan *Scanner
	curr Token
	peek Token
}

func newUclParser(r io.Reader) *uclParser {
	p := &uclParser{
		scan: Scan(r),
	}
	p.next()
	p.next()
	return p
}

func (p *uclParser) Parse() (any, error) {
	val, err := p.parse()
	if err != nil {
		return nil, err
	}
	if !p.done() {
		return nil, syntaxError("unexpected token after value")
	}
	return val, p.scan.Err()
}

func (p *uclParser) parse() (any, error) {
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
	case Ident:
		return nil, nil
	case Number:
		return p.parseNumber()
	case Boolean:
		return p.parseBool(), nil
	case Null:
		return p.parseNull(), nil
	case CommentHash, CommentMulti:
		p.skipComments()
		return p.parse()
	default:
		return nil, syntaxError("invalid token")
	}
}

func (p *uclParser) parseObject() (any, error) {
	return nil, nil
}

func (p *uclParser) parseArray() (any, error) {
	return nil, nil
}

func (p *uclParser) parseNumber() (any, error) {
	base := 10
	if p.is(NumberHex) {
		base = 16
	}
	defer p.next()
	i, err := strconv.ParseInt(p.currentLiteral(), base, 64)
	if err == nil {
		return i, nil
	}
	if err != nil && p.is(NumberHex) {
		return nil, syntaxError("invalid hex number")
	}
	n, err := strconv.ParseFloat(p.currentLiteral(), 64)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (p *uclParser) parseBool() any {
	defer p.next()
	if p.currentLiteral() == "true" {
		return true
	}
	return false
}

func (p *uclParser) parseString() any {
	defer p.next()
	return p.currentLiteral()
}

func (p *uclParser) parseNull() any {
	defer p.next()
	return nil
}

func (p *uclParser) skipComments() {
	for p.is(CommentHash) || p.is(CommentMulti) {
		p.next()
	}
}

func (p *uclParser) done() bool {
	return p.is(EOF)
}

func (p *uclParser) is(kind rune) bool {
	return p.curr.Type == kind
}

func (p *uclParser) next() {
	p.curr = p.peek
	p.peek = p.scan.Scan()
}

func (p *uclParser) currentLiteral() string {
	return p.curr.Literal
}

func syntaxError(msg string) error {
	return fmt.Errorf("%w: %s", errSyntax, msg)
}
