package parser_test

import (
	"testing"

	"github.com/midbel/curly/internal/parser"
)

type parseCase struct {
	Name  string
	Input string
	Ok    bool
}

var testcases = []parseCase{
	{
		Name: "empty",
		Ok:   true,
	},
	{
		Name:  "comment",
		Input: "{{!comment}}",
		Ok:    true,
	},
	{
		Name:  "variable",
		Input: "{{variable}} - {{&variable}}",
		Ok:    true,
	},
	{
		Name:  "block",
		Input: "{{# block}}echo {{&variable}}{{/ block}}",
		Ok:    true,
	},
	{
		Name:  "block-comment",
		Input: "{{# block}}echo {{&variable}}{{!comment}}{{/ block}}",
		Ok:    true,
	},
	{
		Name:  "nexted-block",
		Input: "{{# block}}echo {{&variable}}{{#nest}}another block{{/nest}}{{/ block}}",
		Ok:    true,
	},
	{
		Name:  "block-with-filters",
		Input: "{{# block | take 0}}echo {{&variable}}{{/ block}}",
		Ok:    true,
	},
	{
		Name:  "inverted-block",
		Input: "{{^ block}}echo {{&variable}}{{/ block}}",
		Ok:    true,
	},
	{
		Name:  "delimiters",
		Input: "{{= <% %> =}}<% variable %> <%={{ }}=%>",
		Ok:    true,
	},
	{
		Name:  "call",
		Input: "{{@ template}}",
		Ok:    true,
	},
	{
		Name:  "call-with-filters",
		Input: "{{@ template ctx | lower}}",
		Ok:    true,
	},
	{
		Name:  "define",
		Input: "{{< define}}block{{/define}}",
		Ok:    true,
	},
	{
		Name:  "section",
		Input: "{{%section}}section{{/section}}",
		Ok:    true,
	},
	// errors
	{
		Name:  "block-error",
		Input: "{{#errb}}error{{/error}}",
	},
	{
		Name:  "section-error",
		Input: "{{#errs}}error{{/error}}",
	},
	{
		Name:  "define-error",
		Input: "{{#errd}}error{{/error}}",
	},
}

func TestParser(t *testing.T) {
	for _, c := range testcases {
		_, err := parser.ParseString(c.Input)
		switch {
		default:
		case c.Ok && err != nil:
			t.Errorf("%s: unexpected error %s", c.Name, err)
		case !c.Ok && err == nil:
			t.Errorf("%s: expected error but got none (%s)", c.Name, c.Input)
		}
	}
}
