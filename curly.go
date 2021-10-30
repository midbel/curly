package curly

import (
	"bufio"
	"io"

	"github.com/midbel/curly/internal/parser"
	"github.com/midbel/curly/internal/state"
)

type FuncMap map[string]interface{}

type Template struct {
	name  string
	funcs FuncMap
	set   map[string]*Template
	root  parser.Node
}

func New(name string) *Template {
	return &Template{
		name:  name,
		funcs: make(FuncMap),
	}
}

func (t *Template) Funcs(fm FuncMap) *Template {
	for k, f := range fm {
		t.funcs[k] = f
	}
	return t
}

func (t *Template) Execute(w io.Writer, data interface{}) error {
	wr := bufio.NewWriter(w)
	defer wr.Flush()
	return t.root.Execute(wr, state.EmptyState(data))
}
