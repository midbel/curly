package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/midbel/toml"
	"github.com/midbel/curly/internal/parser"
	"github.com/midbel/curly/internal/state"
	"github.com/midbel/curly/internal/token"
	"github.com/midbel/curly/internal/scanner"
)

func main() {
	var (
		scan = flag.Bool("s", false, "scan")
		exec = flag.Bool("e", false, "exec")
	)
	flag.Parse()

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
