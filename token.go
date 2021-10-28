package curly

import (
	"fmt"
)

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
	Section
	Define
	Exec
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

type Position struct {
	Line   int
	Column int
}

type Token struct {
	Position
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
	case Define:
		return "<define>"
	case Exec:
		return "<exec>"
	case Section:
		return "<section>"
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
