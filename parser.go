package curly

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var errSyntax = errors.New("syntax error")

func Json(r io.Reader) (any, error) {
	p := newStdParser(r)
	return p.Parse()
}

func Ucl(r io.Reader) (any, error) {
	p := newUclParser(r)
	return p.Parse()
}

func Five(r io.Reader) (any, error) {
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
		return p.parseBool()
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
	i, err := strconv.ParseInt(p.currentLiteral(), 10, 64)
	if err == nil {
		return i, nil
	}
	n, err := strconv.ParseFloat(p.currentLiteral(), 64)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (p *stdParser) parseBool() (any, error) {
	defer p.next()
	if p.currentLiteral() == "true" {
		return true, nil
	}
	if p.currentLiteral() == "false" {
		return false, nil
	}
	return false, syntaxError("invalid boolean value")
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
		return p.parseBool()
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
	var (
		base  = 10
		value = p.currentLiteral()
	)
	if p.is(NumberHex) {
		base = 16
		value = strings.TrimPrefix(value, "0x")
	}
	defer p.next()
	i, err := strconv.ParseInt(value, base, 64)
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

func (p *fiveParser) parseBool() (any, error) {
	defer p.next()
	if p.currentLiteral() == "true" {
		return true, nil
	}
	if p.currentLiteral() == "false" {
		return false, nil
	}
	return false, syntaxError("invalid boolean value")
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
	for !p.done() {
		_, err := p.parseMember()
		if err != nil {
			return nil, err
		}
		// p.skipComments()
		// _, err := p.parsePath()
		// if err != nil {
		// 	return nil, err
		// }
		// switch {
		// case p.is(Equal) || p.is(Colon):
		// 	p.next()
		// 	_, _, err = p.parseValue()
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// case p.is(BegObj):
		// 	_, err = p.parseObject()
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// default:
		// 	return nil, syntaxError("unexpected token after key")
		// }
	}
	return nil, nil
}

func (p *uclParser) parseValue() (any, bool, error) {
	var (
		ret any
		comp bool
		err error
	)
	switch p.curr.Type {
	case BegObj:
		ret, err = p.parseObject()
		comp = true
	case BegArr:
		ret, err = p.parseArray()
		comp = true
	case Null:
		ret = p.parseNull()
	case Boolean:
		ret, err = p.parseBool()
	case Number, NumberHex, NumberLeadingPlus, NumberLeadingDot, NumberTrailingDot:
		ret, err = p.parseNumber()
	case String, StringSingleQuoted, Ident:
		ret = p.parseString()
	default:
		err = syntaxError("invalid value")
	}
	if err != nil {
		return nil, false, err
	}
	return ret, comp, nil
}

func (p *uclParser) parsePath() (any, error) {
	for i := 0; !p.done(); i++ {
		if p.is(Equal) || p.is(Colon) || p.is(BegObj) {
			if i == 0 {
				return nil, syntaxError("missing key")
			}
			break
		}
		_, err := p.parseKey()
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (p *uclParser) parseKey() (any, error) {
	switch {
	case p.is(String) || p.is(StringSingleQuoted) || p.is(Ident):
		fmt.Println("> key", p.curr)
		p.next()
	default:
		return nil, syntaxError("unexpected token in key")
	}	
	return nil, nil
}

func (p *uclParser) parseMember() (any, error) {
	p.skipComments()
	_, err := p.parsePath()
	if err != nil {
		return nil, err
	}
	switch {
	case p.is(Equal) || p.is(Colon):
		p.next()
	case p.is(BegObj):
	default:
		return nil, syntaxError("missing separator after key")
	}
	ret, composite, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	fmt.Println(">> object value", ret)
	switch {
	case !composite && (p.is(Semicolon) || p.is(Comma)):
		p.next()
	case p.is(EndObj):
	default:
		if !composite {
			return nil, syntaxError("expected ',' or ';' or '}'")
		}
	}
	return ret, nil
}

func (p *uclParser) parseObject() (any, error) {
	p.next()
	for !p.done() && !p.is(EndObj) {
		_, err := p.parseMember()
		if err != nil {
			return nil, err
		}
	}
	if !p.is(EndObj) {
		return nil, syntaxError("missing '}' at end of object")
	}
	p.next()
	return nil, nil
}

func (p *uclParser) parseArray() (any, error) {
	p.next()
	for !p.done() && !p.is(EndArr) {
		ret, _, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		fmt.Println(">> array value", ret)
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
	return nil, nil
}

func (p *uclParser) parseNumber() (any, error) {
	var (
		base  = 10
		value = p.currentLiteral()
	)
	if p.is(NumberHex) {
		base = 16
		value = strings.TrimPrefix(value, "0x")
	}
	defer p.next()
	i, err := strconv.ParseInt(value, base, 64)
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

func (p *uclParser) parseBool() (any, error) {
	defer p.next()
	if p.currentLiteral() == "true" {
		return true, nil
	}
	if p.currentLiteral() == "false" {
		return false, nil
	}
	return false, syntaxError("invalid boolean value")
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
