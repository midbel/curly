package scanner

import (
	"bytes"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/midbel/curly/internal/token"
)

const (
	amper      = '&'
	pound      = '#'
	caret      = '^'
	slash      = '/'
	lbrace     = '{'
	rbrace     = '}'
	bang       = '!'
	rangle     = '>'
	langle     = '<'
	arobase    = '@'
	percent    = '%'
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

type Scanner struct {
	input []byte
	curr  int
	next  int
	char  rune

	line   int
	column int
	seen   int

	scan    func(*token.Token)
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

func (s *Scanner) Position() token.Position {
	return token.Position{
		Line:   s.line,
		Column: s.column,
	}
}

func (s *Scanner) Scan() token.Token {
	var t token.Token
	t.Position = s.Position()
	if s.isEOF() {
		t.Type = token.EOF
		return t
	}
	if s.scan != nil {
		s.scan(&t)
		return t
	}
	switch {
	case s.isOpen():
		s.skipOpen()
		t.Type = token.Open
		s.scan, s.between = s.scanType, true
	case s.isClose():
		s.skipClose()
		t.Type = token.Close
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

func (s *Scanner) scanComment(t *token.Token) {
	s.scan = nil

	pos := s.curr
	for !s.isEOF() && !s.isClose() {
		s.read()
	}
	t.Type = token.Literal
	if !s.isClose() {
		t.Type = token.Invalid
	}
	t.Literal = strings.TrimSpace(string(s.input[pos:s.curr]))
}

func (s *Scanner) scanLiteral(t *token.Token) {
	pos := s.curr
	for !s.isEOF() && !s.isOpen() {
		s.read()
	}
	t.Type = token.Literal
	t.Literal = string(s.input[pos:s.curr])
}

func (s *Scanner) scanIdent(t *token.Token) {
	s.scan = nil
	s.skipBlank()
	defer s.skipBlank()
	if !isLetter(s.char) {
		t.Type = token.Invalid
		return
	}
	pos := s.curr
	for isIdent(s.char) {
		s.read()
	}
	t.Literal = string(s.input[pos:s.curr])
	switch t.Literal {
	case "true", "false":
		t.Type = token.Bool
	default:
		t.Type = token.Ident
	}
}

func (s *Scanner) scanOperator(t *token.Token) {
	t.Type = token.Invalid
	switch k := s.peek(); s.char {
	case pipe:
		t.Type = token.Pipe
		if k == pipe {
			t.Type = token.Or
		}
	case amper:
		if k == amper {
			t.Type = token.And
		}
	case bang:
		t.Type = token.Not
	case dash:
		t.Type = token.Rev
	default:
	}
	if t.Type == token.And || t.Type == token.Or {
		s.read()
	}
	s.read()
	s.skipBlank()
}

func (s *Scanner) scanString(t *token.Token) {
	var (
		quote = s.char
		pos   = s.curr
	)
	s.read()
	pos = s.curr
	for !s.isEOF() && s.char != quote {
		s.read()
	}
	t.Type = token.Literal
	t.Literal = string(s.input[pos:s.curr])
	if s.char != quote {
		t.Type = token.Invalid
	}
	s.read()
	s.skipBlank()
}

func (s *Scanner) scanNumber(t *token.Token) {
	pos := s.curr
	for isDigit(s.char) {
		s.read()
	}
	t.Type = token.Integer
	if s.char == dot {
		for isDigit(s.char) {
			s.read()
		}
		t.Type = token.Float
	}
	t.Literal = string(s.input[pos:s.curr])
	s.skipBlank()
}

func (s *Scanner) scanOpenDelimiter(t *token.Token) {
	// {{=<punct><blank>
	s.scan = s.scanCloseDelimiter
	s.scanDelim(t, func() bool { return !isBlank(s.char) })
	if !isBlank(s.char) {
		t.Type = token.Invalid
	} else {
		s.skipBlank()
	}
}

func (s *Scanner) scanCloseDelimiter(t *token.Token) {
	// <blank><punct>=}}
	s.scan = nil
	s.scanDelim(t, func() bool { return s.char != equal })
	if s.char != equal {
		t.Type = token.Invalid
	} else {
		s.read()
	}
}

func (s *Scanner) scanDelim(t *token.Token, accept func() bool) {
	pos := s.curr
	for !s.isEOF() && accept() {
		if isLetter(s.char) || isDigit(s.char) {
			t.Type = token.Invalid
		}
		s.read()
	}
	if t.Type == 0 {
		t.Type = token.Literal
	}
	t.Literal = string(s.input[pos:s.curr])
}

func (s *Scanner) scanType(t *token.Token) {
	s.scan = s.scanIdent
	switch s.char {
	case pound:
		t.Type = token.Block
	case caret:
		t.Type = token.Inverted
	case bang:
		t.Type = token.Comment
		s.scan = s.scanComment
	case equal:
		t.Type = token.Delim
		s.scan = s.scanOpenDelimiter
	case rangle:
		t.Type = token.Partial
	case langle:
		t.Type = token.Define
	case arobase:
		t.Type = token.Exec
	case percent:
		t.Type = token.Section
	case slash:
		t.Type = token.End
	case amper:
		t.Type = token.UnescapeVar
	default:
		t.Type = token.EscapeVar
	}
	if t.Type != token.EscapeVar {
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
