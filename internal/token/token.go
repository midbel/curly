package token

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
	OpenTrim
	Close
	CloseTrim
	Delim
	Block
	Inverted
	EscapeVar
	UnescapeVar
	EnvVar
	SpecialVar
	Pipe
	Assignment
	Partial
	Section
	Define
	Exec
	Comment
	End
	BegGrp
	EndGrp
	Invalid
)

type Position struct {
	Line   int
	Column int
}

func (p Position) String() string {
	return fmt.Sprintf("%d,%d", p.Line, p.Column)
}

type Token struct {
	Position
	Literal string
	Unescap bool
	Type    rune
}

func CreateToken(literal string, kind rune) Token {
	return Token{
		Literal: literal,
		Type:    kind,
	}
}

func (t Token) TrimLeft() bool {
	return t.Type == CloseTrim
}

func (t Token) TrimRight() bool {
	return t.Type == OpenTrim
}

func (t Token) Equal(other Token) bool {
	return t.Literal == other.Literal && t.Type == other.Type
}

func (t Token) IsEOF() bool {
	return t.Type == EOF
}

func (t Token) IsStructural() bool {
	switch t.Type {
	case Open, Close, Delim, Block, Inverted, Pipe, Partial, Section, Define, Exec, End:
		return true
	default:
		return false
	}
}

func (t Token) IsValue() bool {
	switch t.Type {
	case Literal, Integer, Float, Bool, Ident:
		return true
	default:
		return false
	}
}

func (t *Token) Unescape() bool {
	return t.Type == UnescapeVar
}

func (t *Token) Special() bool {
	return t.Type == SpecialVar
}

func (t Token) String() string {
	var prefix string
	switch t.Type {
	case EOF:
		return "<eof>"
	case Pipe:
		return "<pipe>"
	case Open:
		return "<open>"
	case OpenTrim:
		return "<open-trim>"
	case Close:
		return "<close>"
	case CloseTrim:
		return "<close-trim>"
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
	case Assignment:
		return "<assignment>"
	case Partial:
		return "<partial>"
	case EscapeVar, UnescapeVar:
		return "<variable>"
	case SpecialVar:
		return "<special-variable>"
	case EnvVar:
		return "<env-variable>"
	case Delim:
		return "<delimiter>"
	case End:
		return "<end>"
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
	case BegGrp:
		return "begin-group"
	case EndGrp:
		return "end-group"
	case Invalid:
		prefix = "invalid"
	default:
		prefix = "unknown"
	}
	return fmt.Sprintf("%s(%s)", prefix, t.Literal)
}
