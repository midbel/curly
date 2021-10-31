package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/midbel/curly/internal/parser"
	"github.com/midbel/curly/internal/scanner"
	"github.com/midbel/curly/internal/state"
	"github.com/midbel/curly/internal/token"
	"github.com/midbel/toml"
)

func main() {
	var (
		scan = flag.Bool("s", false, "scan")
		exec = flag.Bool("e", false, "exec")
		demo = flag.Bool("d", false, "demo")
	)
	flag.Parse()

	if *demo {
		err := execDemo()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		return
	}

	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer r.Close()

	if *scan {
		scanTemplate(r)
	} else if *exec {
		err = execTemplate(r, flag.Arg(1))
	} else {
		err = debugTemplate(r)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

const demo = `my repositories:

{{# Repo }}
* {{Name}} (version: {{Version}})
{{/ Repo }}
contacts: {{ Email }}
`

type Repo struct {
	Name    string
	Version string
}

func execDemo() error {
	data := struct {
		Repo  []Repo
		Email string
	}{
		Repo: []Repo{
			{
				Name:    "curly",
				Version: "0.0.1",
			},
			{
				Name:    "toml",
				Version: "0.0.1",
			},
		},
		Email: "midbel@foobar.org",
	}
	t, err := parser.Parse(strings.NewReader(demo))
	if err != nil {
		return err
	}
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	return t.Execute(w, state.EmptyState(data))
}

func execTemplate(r io.Reader, file string) error {
	data := make(map[string]interface{})
	if err := toml.DecodeFile(file, &data); err != nil && file != "" {
		return err
	}
	t, err := parser.Parse(r)
	if err != nil {
		return err
	}
	return t.Execute(os.Stdout, state.EmptyState(data))
}

func debugTemplate(r io.Reader) error {
	return parser.Debug(r, os.Stdout)
}

func scanTemplate(r io.Reader) {
	s, _ := scanner.Scan(r)
	for {
		tok := s.Scan()
		if tok.Type == token.EOF || tok.Type == token.Invalid {
			break
		}
		fmt.Println(tok)
	}
}
