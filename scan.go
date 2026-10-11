package curly

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"unicode"
)

type Scanner struct {
	input io.RuneScanner
	char  rune

	Position
	old Position

	eof bool
	err error
	str bytes.Buffer
}

func Scan(r io.Reader) *Scanner {
	scan := Scanner{
		input: bufio.NewReader(r),
	}
	scan.Line++
	scan.read()
	return &scan
}

func (s *Scanner) Err() error {
	return s.err
}

func (s *Scanner) Scan() Token {
	var tok Token
	if s.err != nil {
		tok.Type = Invalid
		return tok
	}
	defer s.str.Reset()
	s.skipBlank()

	if s.done() {
		tok.Type = EOF
		return tok
	}
	switch {
	case IsComment(s.char, s.peek()):
		s.scanComment(&tok)
	case IsLetter(s.char) || s.char == '_':
		s.scanIdent(&tok)
	case IsQuote(s.char) || IsApos(s.char):
		s.scanString(&tok)
	case IsNumber(s.char) || s.char == '-' || s.char == '+' || s.char == '.':
		s.scanNumber(&tok)
	case IsDelim(s.char):
		s.scanDelimiter(&tok)
	default:
		tok.Type = Invalid
	}
	return tok
}

func (s *Scanner) scanComment(tok *Token) {
	comment := s.char
	s.read()
	multiline := s.char == '*'
	s.read()
	s.skipBlank()

	for !s.done() {
		if !multiline && IsNL(s.char) {
			break
		} else if multiline && s.char == '*' && s.peek() == '/' {
			s.read()
			s.read()
			break
		}
		s.write()
		s.read()
	}
	tok.Literal = s.str.String()
	tok.Type = CommentLine
	if comment == '#' {
		tok.Type = CommentHash
	}
	if multiline {
		tok.Type = CommentMulti
	}
}

func (s *Scanner) scanIdent(tok *Token) {
	for !s.done() && IsAlpha(s.char) {
		s.write()
		s.read()
	}
	tok.Type = Ident
	tok.Literal = s.str.String()
	switch tok.Literal {
	case "true", "false":
		tok.Type = Boolean
	case "null":
		tok.Type = Null
	default:
		tok.Type = Invalid
	}
}

func (s *Scanner) scanString(tok *Token) {
	quote := s.char
	s.read()
	for !s.done() && s.char != quote {
		if s.char == '\\' {
			s.read()
			if IsNL(s.char) {
				s.write()
				s.read()
				continue
			}
			if ok := s.scanEscape(quote); !ok {
				tok.Type = Invalid
				return
			}
		}
		if s.char < 0x20 {
			tok.Type = Invalid
			break
		}
		s.write()
		s.read()
	}
	tok.Literal = s.str.String()
	tok.Type = String
	if s.char == '\'' {
		tok.Type = StringSingleQuoted
	}
	if s.char != quote {
		tok.Type = Invalid
	} else {
		s.read()
	}
}

func (s *Scanner) scanEscape(quote rune) bool {
	switch s.char {
	case quote:
		s.char = quote
	case '\\':
		s.char = '\\'
	case '/':
		s.char = '/'
	case 'b':
		s.char = '\b'
	case 'f':
		s.char = '\f'
	case 'n':
		s.char = '\n'
	case 'r':
		s.char = '\r'
	case 't':
		s.char = '\t'
	case 'u':
		s.read()
		buf := make([]rune, 4)
		for i := 1; i <= 4; i++ {
			if !IsHex(s.char) {
				return false
			}
			buf[i-1] = s.char
			if i < 4 {
				s.read()
			}
		}
		char, _ := strconv.ParseInt(string(buf), 16, 32)
		s.char = rune(char)
	default:
		return false
	}
	return true
}

func (s *Scanner) scanHexa(tok *Token) {
	s.read()
	s.read()
	s.writeRune('0')
	s.writeRune('x')

	if !IsHex(s.char) {
		tok.Type = Invalid
		return
	}

	for !s.done() && IsHex(s.char) {
		s.write()
		s.read()
	}
	tok.Type = NumberHex
	tok.Literal = s.str.String()
}

func (s *Scanner) scanNumber(tok *Token) {
	tok.Type = Number
	if s.char == '0' && s.peek() == 'x' {
		s.scanHexa(tok)
		return
	}
	if s.char == '.' {
		tok.Type = NumberLeadingDot
		s.writeRune('0')
		s.writeRune('.')
		s.read()
	}
	if s.char == '-' || s.char == '+' {
		if s.char == '-' {
			s.write()
		}
		if s.char == '+' {
			tok.Type = NumberLeadingPlus
		}
		s.read()
	}
	if s.char == '0' && IsNumber(s.peek()) {
		tok.Type = Invalid
		return
	}
	for !s.done() && IsNumber(s.char) {
		s.write()
		s.read()
	}
	tok.Literal = s.str.String()
	if s.char == '.' {
		s.write()
		s.read()
		if !IsNumber(s.char) {
			tok.Type = NumberTrailingDot
			return
		}
		for !s.done() && IsNumber(s.char) {
			s.write()
			s.read()
		}
		tok.Literal = s.str.String()
	}
	if s.char == 'e' || s.char == 'E' {
		s.write()
		s.read()
		if s.char == '-' || s.char == '+' {
			s.write()
			s.read()
		}
		if !IsNumber(s.char) {
			tok.Type = Invalid
			return
		}
		for !s.done() && IsNumber(s.char) {
			s.write()
			s.read()
		}
		tok.Literal = s.str.String()
	}
}

func (s *Scanner) scanDelimiter(tok *Token) {
	switch s.char {
	case '[':
		tok.Type = BegArr
	case ']':
		tok.Type = EndArr
	case '{':
		tok.Type = BegObj
	case '}':
		tok.Type = EndObj
	case ',':
		tok.Type = Comma
	case ';':
		tok.Type = Semicolon
	case ':':
		tok.Type = Colon
	case '=':
		tok.Type = Equal
	default:
		tok.Type = Invalid
	}
	if tok.Type != Invalid {
		s.read()
	}
}

func (s *Scanner) writeRune(c rune) {
	s.str.WriteRune(c)
}

func (s *Scanner) write() {
	s.writeRune(s.char)
}

func (s *Scanner) read() {
	if s.eof || s.err != nil {
		return
	}
	s.old = s.Position
	if s.char == '\n' {
		s.Line++
		s.Column = 0
	}
	s.Column++

	char, _, err := s.input.ReadRune()
	if err != nil {
		if errors.Is(err, io.EOF) {
			s.eof = true
			s.char = 0
		} else {
			s.err = err
		}
		return
	}
	s.char = char
}

func (s *Scanner) peek() rune {
	r, _, err := s.input.ReadRune()
	if err != nil {
		return 0
	}
	s.input.UnreadRune()
	return r
}

func (s *Scanner) done() bool {
	return s.eof || s.err != nil
}

func (s *Scanner) skipBlank() {
	for !s.done() && unicode.IsSpace(s.char) {
		s.read()
	}
}

type Position struct {
	Line   int
	Column int
}

type Token struct {
	Literal string
	Type    rune
	Position
}

func (t Token) String() string {
	var prefix string
	switch t.Type {
	case Colon:
		return "<colon>"
	case Equal:
		return "<equal>"
	case EOF:
		return "<eof>"
	case BegArr:
		return "<beg-arr>"
	case EndArr:
		return "<end-arr>"
	case BegObj:
		return "<beg-obj>"
	case EndObj:
		return "<end-obj>"
	case Comma:
		return "<comma>"
	case Semicolon:
		return "<semicolon>"
	case Boolean:
		prefix = "boolean"
	case Null:
		return "<null>"
	case String, StringSingleQuoted:
		prefix = "string"
	case Number, NumberHex, NumberLeadingPlus, NumberLeadingDot, NumberTrailingDot:
		prefix = "number"
	case Ident:
		prefix = "identifier"
	case CommentHash, CommentLine, CommentMulti:
		prefix = "comment"
	case Invalid:
		prefix = "invalid"
	}
	return fmt.Sprintf("%s(%s)", prefix, t.Literal)
}

const (
	EOF = -(1 + iota)
	BegArr
	EndArr
	BegObj
	EndObj
	Comma
	Semicolon
	Colon
	Equal
	Boolean
	Null
	String
	StringSingleQuoted
	Number
	NumberHex
	NumberLeadingPlus
	NumberLeadingDot
	NumberTrailingDot
	Ident
	CommentLine
	CommentHash
	CommentMulti
	Invalid
)

func IsComment(c, k rune) bool {
	if c == '#' {
		return true
	}
	return c == '/' && (c == k || c == '*')
}

func IsHex(c rune) bool {
	return IsNumber(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func IsNumber(c rune) bool {
	return c >= '0' && c <= '9'
}

func IsLower(c rune) bool {
	return c >= 'a' && c <= 'z'
}

func IsUpper(c rune) bool {
	return c >= 'A' && c <= 'Z'
}

func IsLetter(c rune) bool {
	return IsLower(c) || IsUpper(c)
}

func IsAlpha(c rune) bool {
	return IsLetter(c) || IsNumber(c) || c == '_'
}

func IsApos(c rune) bool {
	return c == '\''
}

func IsQuote(c rune) bool {
	return c == '"'
}

func IsBackQuote(c rune) bool {
	return c == '`'
}

func IsDelim(c rune) bool {
	return c == '{' || c == '}' || c == '[' || c == ']' || c == ',' || c == ';' || c == ':' || c == '='
}

func IsNL(c rune) bool {
	return c == '\n' || c == '\r'
}
