package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/midbel/curly"
)

func main() {
	scan := flag.Bool("x", false, "scan")
	flag.Parse()

	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer r.Close()

	if *scan {
		err = tokenize(r)
	} else {
		err = parse(r)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func tokenize(r io.Reader) error {
	scan := curly.Scan(r)
	scan.KeepNL()
	for {
		tok := scan.Scan()
		fmt.Println(tok)
		if tok.Type == curly.EOF {
			break
		}
	}
	return nil
}

func parse(r io.Reader) error {
	_, err := curly.Ucl(r)
	return err
}
