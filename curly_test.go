package curly_test

import (
	"fmt"
	"os"
	"path/filepath"
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

func ExampleTemplate_ParseFiles() {
	type Change struct {
		Date string
		Desc string
	}
	type Repo struct {
		Name    string
		Version string
		Changes []Change
	}
	type Dev struct {
		Name  string
		Repos []Repo
	}
	const demo = `
{{# devs -}}
{{Name}}
{{@ repository Repos -}}
{{/ devs }}
	`
	c := struct {
		Dev []Dev `curly:"devs"`
	}{
		Dev: []Dev{
			{
				Name: "rustine",
				Repos: []Repo{
					{
						Name:    "data",
						Version: "1.0.0",
						Changes: []Change{
							{
								Date: "2021-11-06 19:22:00",
								Desc: "first major release",
							},
						},
					},
				},
			},
			{
				Name: "midbel",
				Repos: []Repo{
					{
						Name:    "fig",
						Version: "0.0.4",
						Changes: []Change{
							{
								Date: "2021-11-06 19:22:00",
								Desc: "restart from scratch",
							},
							{
								Date: "2021-06-06 15:45:00",
								Desc: "initial commit",
							},
						},
					},
					{
						Name:    "curly",
						Version: "0.1.0",
						Changes: []Change{
							{
								Date: "2021-07-18 11:00:00",
								Desc: "initial commit",
							},
							{
								Date: "2021-11-07 15:45:00",
								Desc: "test parse files",
							},
						},
					},
				},
			},
		},
	}
	files := []struct {
		File    string
		Content string
	}{
		{
			File: "repo.txt",
			Content: `
{{< repository -}}
{{#Repos -}}
- {{Name}} ({{Version}})
{{@ changes Changes -}}
{{/Repos -}}
{{/ repository -}}`,
		},
		{
			File: "change.txt",
			Content: `
{{< changes -}}
{{# Changes -}}
  - {{Date}}: {{Desc}}
{{/ Changes -}}
{{/ changes -}}`,
		},
	}
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		fmt.Println("mkdir tmp:", err)
		return
	}
	defer os.RemoveAll(dir)

	var list []string
	for _, f := range files {
		f.File = filepath.Join(dir, f.File)
		buf := strings.TrimSpace(f.Content)
		if err := os.WriteFile(f.File, []byte(buf), 0644); err != nil {
			fmt.Println("writing file error:", err)
			return
		}
		list = append(list, f.File)
	}
	t, err := curly.New("demo").Funcs(curly.Filters).Parse(strings.NewReader(demo))
	if err != nil {
		fmt.Println("error parsing template (1):", err)
		return
	}
	if t, err = t.ParseFiles(list...); err != nil {
		fmt.Println("error parsing template (2):", err)
		return
	}
	t.Execute(os.Stdout, c)
	// Output:
	// rustine
	// - data (1.0.0)
	//   - 2021-11-06 19:22:00: first major release
	// midbel
	// - fig (0.0.4)
	//   - 2021-11-06 19:22:00: restart from scratch
	//   - 2021-06-06 15:45:00: initial commit
	// - curly (0.1.0)
	//   - 2021-07-18 11:00:00: initial commit
	//   - 2021-11-07 15:45:00: test parse files
}
