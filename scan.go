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
	mode

	Position
	old Position

	eof bool
	err error
	str bytes.Buffer
}

func Scan(r io.Reader, mode mode) *Scanner {
	scan := Scanner{
		input: bufio.NewReader(r),
		mode:  mode,
	}
	scan.Line = 1
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
	case s.mode.isExtended() && IsComment(s.char, s.peek()):
		s.scanComment(&tok)
	case s.mode.isExtended() && IsLetter(s.char):
		s.scanLiteral(&tok)
	case IsLower(s.char):
		s.scanIdent(&tok)
	case IsQuote(s.char) || (s.mode.isExtended() && IsApos(s.char)):
		s.scanString(&tok)
	case IsNumber(s.char) || s.char == '-':
		s.scanNumber(&tok)
	case s.mode.isExtended() && (s.char == '+' || s.char == '.'):
		s.scanNumber(&tok)
	case IsDelim(s.char):
		s.scanDelimiter(&tok)
	default:
		tok.Type = Invalid
	}
	return tok
}

func (s *Scanner) scanLiteral(tok *Token) {
	for !s.done() && IsAlpha(s.char) {
		s.write()
		s.read()
	}
	tok.Literal = s.str.String()
	switch tok.Literal {
	case "true", "false":
		tok.Type = Boolean
	case "null":
		tok.Type = Null
	default:
		tok.Type = Ident
	}
}

func (s *Scanner) scanComment(tok *Token) {
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
	tok.Type = Comment
}

func (s *Scanner) scanIdent(tok *Token) {
	for !s.done() && IsAlpha(s.char) {
		s.write()
		s.read()
	}
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
		s.write()
		s.read()
	}
	tok.Literal = s.str.String()
	tok.Type = String
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
	for !s.done() && IsHex(s.char) {
		s.write()
		s.read()
	}
	tok.Literal = s.str.String()
}

func (s *Scanner) scanNumber(tok *Token) {
	tok.Type = Number
	if s.mode.isExtended() && s.char == '0' && s.peek() == 'x' {
		s.scanHexa(tok)
		return
	}
	if s.mode.isExtended() && s.char == '.' {
		s.writeRune('0')
		s.writeRune('.')
		s.read()
	}
	if s.char == '-' || s.char == '+' {
		if s.char == '-' {
			s.write()
		}
		s.read()
	}
	if s.char == '0' && s.peek() != '.' {
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
			if !s.mode.isExtended() {
				tok.Type = Invalid
			}
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
	case ':':
		tok.Type = Colon
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
	defer s.input.UnreadRune()
	r, _, _ := s.input.ReadRune()
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
	case Transform:
		return "<transform>"
	case Doc:
		return "<document>"
	case Ternary:
		return "<ternary>"
	case Colon:
		return "<colon>"
	case BegGrp:
		return "<beg-grp>"
	case EndGrp:
		return "<end-grp>"
	case And:
		return "<and>"
	case Or:
		return "<or>"
	case In:
		return "<in>"
	case Add:
		return "<add>"
	case Sub:
		return "<subtract>"
	case Mul:
		return "<multiply>"
	case Div:
		return "<divide>"
	case Mod:
		return "<modulo>"
	case Eq:
		return "<equal>"
	case Ne:
		return "<not-equal>"
	case Lt:
		return "<lesser-than>"
	case Le:
		return "<lesser-eq>"
	case Gt:
		return "<greater-than>"
	case Ge:
		return "<greater-eq>"
	case Concat:
		return "<concat>"
	case Map:
		return "<map>"
	case Parent:
		return "<parent>"
	case Wildcard:
		return "<wildcard>"
	case Descent:
		return "<descend>"
	case Range:
		return "<range>"
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
	case Boolean:
		prefix = "boolean"
	case Null:
		return "<null>"
	case String:
		prefix = "string"
	case Number:
		prefix = "number"
	case Ident:
		prefix = "identifier"
	case Func:
		prefix = "function"
	case Comment:
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
	Colon
	Boolean
	Null
	String
	Number
	Ident
	Func
	Comment
	// query token
	Doc
	BegGrp
	EndGrp
	In
	And
	Or
	Add
	Sub
	Mul
	Div
	Mod
	Eq
	Ne
	Lt
	Le
	Gt
	Ge
	Concat
	Ternary
	Map
	Parent
	Wildcard
	Descent
	Range
	Transform
	// common
	Invalid
)
