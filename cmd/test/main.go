package main

import (
	"fmt"
	"strings"

	"github.com/midbel/curly"
)

const template1 = `
hello {{& World }} - {{#truthy}}{{&name}}{{/truthy}}
{{^Falsy}}nothing will be rendered {{& falsy}}{{/Falsy}}
>>{{! comment will not be rendered }}<<
`

const template2 = `
my repositories:
{{# repo }}
  {{Name}} (version: {{Version}})
{{/ repo }}
contacts: {{ Email }}
`

const template3 = `
* {{default_tags}}
{{=<% %>=}}
* <% erb_style_tags %>
<%={{ }}=%>
* {{ default_tags_again }}
`

const template4 = `hello {{ array | cut 3 | join "," false }} - {{=<% %>=}} - <% var %>`

const template5 = `
hello {{world | lower}} - {{world | upper | len}}

hello {{ text | split "_" | reverse | join "/" }}
`

type Name struct {
	Name    string
	Version int
}

func main() {
	c := struct {
		World string `tag:"world"`
		Text  string `tag:"text"`
	}{
		World: "World",
		Text:  "under_score_text_with_extra_words",
	}
	exec(template5, c)
}

func scan(template string) {
	s, _ := curly.Scan(strings.NewReader(template))
	for {
		tok := s.Scan()
		if tok.Type == curly.EOF {
			break
		}
		fmt.Println(tok)
	}
}

func exec(template string, data interface{}) {
	r := strings.NewReader(strings.TrimSpace(template))
	curly.Debug(r)
}
