package parser_test

import (
	"strings"
	"testing"

	"github.com/midbel/curly/internal/filters"
	"github.com/midbel/curly/internal/parser"
	"github.com/midbel/curly/internal/state"
)

type nodeCase struct {
	Name  string
	Input string
	Want  string
	Ok    bool
}

var nodecases = []nodeCase{
	{
		Name:  "simple-assignment",
		Input: `{{:var "foobar"}}{{var}}`,
		Want:  "foobar",
		Ok:    true,
	},
	{
		Name:  "simple-assignment-with-filters",
		Input: `{{:var "foobar" | upper}}{{var}}`,
		Want:  "FOOBAR",
		Ok:    true,
	},
	{
		Name:  "assignment-from-context-with-filters",
		Input: `{{:var name | upper}}{{var}}`,
		Want:  "FOOBAR",
		Ok:    true,
	},
	{
		Name:  "loop-index",
		Input: `{{# list }}{{# loop0}} | {{/loop0}}{{loop}} - {{loop0}} - {{revloop}}{{/ list}}`,
		Want:  "1 - 0 - 3 | 2 - 1 - 2 | 3 - 2 - 1",
		Ok:    true,
	},
	{
		Name:  "variable-and-int-math",
		Input: `{{: var 4}}{{var | add 1}}`,
		Want:  "5",
		Ok:    true,
	},
	{
		Name:  "variable-and-float-math",
		Input: `{{: var 4.1}}{{var | add 1}}`,
		Want:  "5.1",
		Ok:    true,
	},
}

func TestNode(t *testing.T) {
	var (
		filters = state.FuncMap{
			"lower": strings.ToLower,
			"upper": strings.ToUpper,
			"add":   filters.Add,
			"sub":   filters.Sub,
		}
		ctx = struct {
			Name string   `curly:"name"`
			List []string `curly:"list"`
		}{
			Name: "foobar",
			List: []string{"foo", "bar", "foo"},
		}
		state = state.EmptyState(ctx, filters)
	)
	for _, c := range nodecases {
		n, err := parser.ParseString(c.Input)
		if err != nil {
			t.Errorf("%s: expecting no error parsing %s! got %s", c.Name, c.Input, err)
			continue
		}
		var str strings.Builder
		err = n.Execute(&str, nil, state)
		switch {
		default:
		case c.Ok && err != nil:
			t.Errorf("%s: unexpected error %s", c.Name, err)
			c.Ok = false
		case !c.Ok && err == nil:
			t.Errorf("%s: expected error but got none (%s)", c.Name, c.Input)
			c.Ok = false
		}
		if !c.Ok {
			continue
		}
		if c.Want != str.String() {
			t.Errorf("%s: results mismatched! want %s, got %s", c.Name, c.Want, str.String())
		}
	}
}
