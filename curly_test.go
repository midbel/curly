package curly_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/midbel/curly"
)

func ExampleTemplate_Define() {
	const demo = `
{{< list }}
{{- # character -}}
  {{loop}}/{{length}}: {{ name | title }}{{#role}}: {{role}}{{/role}}
{{/ character -}}
{{/ list }}


{{#movies -}}
[[{{title}}]]
{{@ list character -}}
{{/movies -}}
	`

	type Character struct {
		Name string `curly:"name"`
		Role string `curly:"role"`
	}

	type Movie struct {
		Title     string      `curly:"title"`
		Character []Character `curly:"character"`
	}

	data := struct {
		Movies []Movie `curly:"movies"`
	}{
		Movies: []Movie{
			{
				Title: "star wars: a new hope",
				Character: []Character{
					{
						Name: "luke skywalker",
						Role: "hero",
					},
					{
						Name: "leia organa",
						Role: "hero",
					},
					{
						Name: "anakin skywalker",
						Role: "",
					},
				},
			},
			{
				Title: "star wars: the empire strikes back",
			},
		},
	}
	t, err := curly.New("demo").Funcs(curly.Filters).Parse(strings.NewReader(demo))
	if err != nil {
		fmt.Println(err)
		return
	}
	t.Execute(os.Stdout, data)

	// Output:
	// [[star wars: a new hope]]
	//   1/3: Luke Skywalker: hero
	//   2/3: Leia Organa: hero
	//   3/3: Anakin Skywalker
	// [[star wars: the empire strikes back]]
}

func ExampleTemplate_Block() {
	const demo = `
{{! comment are not rendered }}
{{< contact -}}
contact: {{email | lower -}}
{{/contact -}}
repositories:
{{# repo -}}
- {{Name}} (version: {{Version}})
{{/ repo -}}

{{% contact }}
contact: noreply@foobar.org
{{/ contact }}
{{% licence -}}
licence: MIT
{{/ licence}}
	`
	type Repo struct {
		Name    string
		Version string
	}

	data := struct {
		Email string `curly:"email"`
		Repos []Repo `curly:"repo"`
	}{
		Email: "midbel@foobar.org",
		Repos: []Repo{
			{
				Name:    "curly",
				Version: "0.0.1",
			},
			{
				Name:    "toml",
				Version: "0.1.1",
			},
		},
	}
	t, err := curly.New("demo").Funcs(curly.Filters).Parse(strings.NewReader(demo))
	if err != nil {
		fmt.Println("error parsing template:", err)
		return
	}
	t.Execute(os.Stdout, data)

	// Output:
	// repositories:
	// - curly (version: 0.0.1)
	// - toml (version: 0.1.1)
	//
	// contact: midbel@foobar.org
	// licence: MIT
}

func ExampleTemplate_Filters() {
	const demo = `
repositories:
{{# repo | reverse -}}
- {{Name}} (version: {{Version}})
{{/ repo -}}

contact: {{email}}
  `

	type Repo struct {
		Name    string
		Version string
	}

	data := struct {
		Email string `curly:"email"`
		Repos []Repo `curly:"repo"`
	}{
		Email: "midbel@foobar.org",
		Repos: []Repo{
			{
				Name:    "curly",
				Version: "0.0.1",
			},
			{
				Name:    "toml",
				Version: "0.1.1",
			},
		},
	}
	t, err := curly.New("demo").Funcs(curly.Filters).Parse(strings.NewReader(demo))
	if err != nil {
		fmt.Println("error parsing template:", err)
		return
	}
	t.Execute(os.Stdout, data)

	// Output:
	// repositories:
	// - toml (version: 0.1.1)
	// - curly (version: 0.0.1)
	//
	// contact: midbel@foobar.org
}

func ExampleTemplate() {
	const demo = `
repositories:
{{# repo -}}
- {{Name}} (version: {{Version}})
{{/ repo -}}

contact: {{email}}
  `

	type Repo struct {
		Name    string
		Version string
	}

	data := struct {
		Email string `curly:"email"`
		Repos []Repo `curly:"repo"`
	}{
		Email: "midbel@foobar.org",
		Repos: []Repo{
			{
				Name:    "curly",
				Version: "0.0.1",
			},
			{
				Name:    "toml",
				Version: "0.1.1",
			},
		},
	}
	t, err := curly.Parse(strings.NewReader(demo))
	if err != nil {
		fmt.Println("error parsing template:", err)
		return
	}
	t.Execute(os.Stdout, data)

	// Output:
	// repositories:
	// - curly (version: 0.0.1)
	// - toml (version: 0.1.1)
	//
	// contact: midbel@foobar.org
}
