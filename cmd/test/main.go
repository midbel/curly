package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/midbel/curly"
	"github.com/midbel/toml"
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
	t, err := curly.Parse(r)
	if err != nil {
		return err
	}
	return t.Execute(os.Stdout, data)
}

func debugTemplate(r io.Reader) error {
	return curly.Debug(r, os.Stdout)
}

func scanTemplate(r io.Reader) {
	s, _ := curly.Scan(r)
	for {
		tok := s.Scan()
		if tok.Type == curly.EOF || tok.Type == curly.Invalid {
			break
		}
		fmt.Println(tok)
	}
}
