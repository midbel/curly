package curly_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/midbel/curly"
)

func ExampleTemplate_Define() {
	const demo = `
{{< character}}
{{# character}}
  - {{name}}{{#role}}: {{role}}{{/role}}
{{/character}}
{{/character}}

{{#movies}}
  [[{{title}}]]
  {{@ character character }}
{{/movies}}
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
						Name: "Luke Skywalker",
						Role: "hero",
					},
					{
						Name: "Leia Organa",
						Role: "hero",
					},
					{
						Name: "Anakin Skywalker",
						Role: "",
					},
				},
			},
			{
				Title: "star wars: the empire strikes back",
			},
		},
	}
	t, err := curly.Parse(strings.NewReader(demo))
	if err != nil {
		fmt.Println(err)
		return
	}
	t.Execute(os.Stdout, data)

	// Output:
	// [star wars: a new hope]
	//   - Luke Skywalker: hero
	//   - Leia Organa: hero
	//   - Anakin Skywalker
	// [star wars: the empire strikes back]
}

func ExampleTemplate() {
	const demo = `
repositories:
{{# repo}}
- {{Name}} (version: {{Version}})
{{/ repo}}

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
		fmt.Println(err)
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
