package curly

import (
	"bytes"
	"io"
	"strings"
	"unicode/utf8"
)

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
	case rangle:
		t.Type = Partial
	case langle:
		t.Type = Define
	case arobase:
		t.Type = Exec
	case percent:
		t.Type = Section
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
	case pound, caret, rangle, langle, percent, arobase, slash, amper, pipe, equal:
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
