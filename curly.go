package curly

import (
	"bufio"
	"io"
	"os"
	"path/filepath"

	"github.com/midbel/curly/internal/parser"
	"github.com/midbel/curly/internal/state"
)

type FuncMap map[string]interface{}

type Template struct {
	name  string
	funcs FuncMap
	root  parser.Node
}

func New(name string) *Template {
	return &Template{
		name:  name,
		funcs: make(FuncMap),
	}
}

func ParseFile(file string) (*Template, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return New(filepath.Base(file)).Parse(r)
}

func Parse(r io.Reader) (*Template, error) {
	return New("").Parse(r)
}

func (t *Template) Parse(r io.Reader) (*Template, error) {
	node, err := parser.Parse(r)
	if err != nil {
		return nil, err
	}
	t.root = node
	return t, nil
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
	var (
		set   parser.Nodeset
		r, ok = t.root.(*parser.RootNode)
	)
	if ok {
		set = r.Named
	}
	return r.Execute(wr, set, state.EmptyState(data))
}
