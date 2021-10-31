package scanner_test

import (
	"strings"
	"testing"

	"github.com/midbel/curly/internal/scanner"
	"github.com/midbel/curly/internal/token"
)

var tokens = []token.Token{
	token.CreateToken("", token.Open),
	token.CreateToken("", token.Block),
	token.CreateToken("ident", token.Ident),
	token.CreateToken("", token.Close),
	token.CreateToken("\nliteral ", token.Literal),
	token.CreateToken("", token.Open),
	token.CreateToken("", token.EscapeVar),
	token.CreateToken("escape", token.Ident),
	token.CreateToken("", token.Close),
	token.CreateToken(" ", token.Literal),
	token.CreateToken("", token.Open),
	token.CreateToken("", token.UnescapeVar),
	token.CreateToken("unescape", token.Ident),
	token.CreateToken("", token.Close),
	token.CreateToken("\n", token.Literal),
	token.CreateToken("", token.Open),
	token.CreateToken("", token.End),
	token.CreateToken("ident", token.Ident),
	token.CreateToken("", token.Close),
	token.CreateToken("\n", token.Literal),
	token.CreateToken("", token.Open),
	token.CreateToken("", token.Inverted),
	token.CreateToken("falsy", token.Ident),
	token.CreateToken("", token.Close),
	token.CreateToken("empty", token.Literal),
	token.CreateToken("", token.Open),
	token.CreateToken("", token.End),
	token.CreateToken("falsy", token.Ident),
	token.CreateToken("", token.Close),
	token.CreateToken("\n", token.Literal),
	token.CreateToken("", token.Open),
	token.CreateToken("", token.Comment),
	token.CreateToken("comment", token.Literal),
	token.CreateToken("", token.Close),
	token.CreateToken("\n", token.Literal),
	token.CreateToken("", token.Open),
	token.CreateToken("", token.Delim),
	token.CreateToken("<%", token.Literal),
	token.CreateToken("%>", token.Literal),
	token.CreateToken("", token.Close),
	token.CreateToken("\n", token.Literal),
	token.CreateToken("", token.Open),
	token.CreateToken("", token.Partial),
	token.CreateToken("partial", token.Ident),
	token.CreateToken("", token.Close),
	token.CreateToken("\n", token.Literal),
	token.CreateToken("", token.Open),
	token.CreateToken("", token.Define),
	token.CreateToken("define", token.Ident),
	token.CreateToken("", token.Close),
	token.CreateToken("\n", token.Literal),
	token.CreateToken("", token.Open),
	token.CreateToken("", token.Exec),
	token.CreateToken("call", token.Ident),
	token.CreateToken("", token.Close),
	token.CreateToken("\n", token.Literal),
	token.CreateToken("", token.Open),
	token.CreateToken("", token.Section),
	token.CreateToken("section", token.Ident),
	token.CreateToken("", token.Close),
	token.CreateToken("\n", token.Literal),
	token.CreateToken("", token.Open),
	token.CreateToken("", token.EscapeVar),
	token.CreateToken("text", token.Ident),
	token.CreateToken("", token.Pipe),
	token.CreateToken("split", token.Ident),
	token.CreateToken("_", token.Literal),
	token.CreateToken("", token.Pipe),
	token.CreateToken("firstn", token.Ident),
	token.CreateToken("1", token.Integer),
	token.CreateToken("", token.Pipe),
	token.CreateToken("add", token.Ident),
	token.CreateToken("2.3", token.Float),
	token.CreateToken("3.2", token.Float),
	token.CreateToken("", token.Close),
	token.CreateToken("\n", token.Literal),
	token.CreateToken("", token.Open),
	token.CreateToken("", token.EscapeVar),
	token.CreateToken("text", token.Ident),
	token.CreateToken("", token.Pipe),
	token.CreateToken("check", token.Ident),
	token.CreateToken("text", token.Ident),
	token.CreateToken("", token.And),
	token.CreateToken("", token.BegGrp),
	token.CreateToken("", token.Not),
	token.CreateToken("false", token.Bool),
	token.CreateToken("", token.Or),
	token.CreateToken("true", token.Bool),
	token.CreateToken("", token.EndGrp),
	token.CreateToken("", token.Close),
	token.CreateToken("\n", token.Literal),
	token.CreateToken("", token.Open),
	token.CreateToken("", token.Block),
	token.CreateToken("block", token.Ident),
	token.CreateToken("", token.Close),
	token.CreateToken("- value", token.Literal),
	token.CreateToken("", token.Open),
	token.CreateToken("", token.End),
	token.CreateToken("block", token.Ident),
	token.CreateToken("", token.Close),
}

const sample = `
{{#ident}}
literal {{escape}} {{&unescape}}
{{/ident}}
{{^ falsy}}empty{{/ falsy}}
{{! comment }}
{{= <% %> =}}
{{> partial}}
{{< define}}
{{@ call}}
{{% section}}
{{text | split "_" | firstn 1 | add 2.3 3.2 }}
{{text | check text && (!false || true) }}
{{#block}}- value{{/block}}
`

func TestScanner(t *testing.T) {
	s, err := scanner.Scan(strings.NewReader(strings.TrimSpace(sample)))
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	var total int
	for i := 0; ; i++ {
		tok := s.Scan()
		if tok.IsEOF() {
			break
		}
		if tok.Type == token.Invalid {
			t.Errorf("invalid token! want %s", tokens[i])
			t.FailNow()
		}
		if i >= len(tokens) {
			t.Errorf("too many tokens generated! want %d, got %d", len(tokens), i)
			t.FailNow()
		}
		if !tok.Equal(tokens[i]) {
			t.Errorf("tokens mismatched! want %s, got %s", tokens[i], tok)
		}
		total++
	}
	if total != len(tokens) {
		t.Errorf("mismatched tokens generated! want %d, got %d", len(tokens), total)
	}
}
