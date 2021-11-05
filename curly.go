package curly

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/midbel/curly/internal/filters"
	"github.com/midbel/curly/internal/parser"
	"github.com/midbel/curly/internal/state"
)

type FuncMap map[string]interface{}

var Filters = FuncMap{
	// strings filters
	"split": strings.Split,
	"join":  strings.Join,
	"lower": strings.ToLower,
	"upper": strings.ToUpper,
	"title": strings.Title,
	"trim":  strings.TrimSpace,
	// filename filters
	"basename": filepath.Base,
	"dirname":  filepath.Dir,
	// array/slice filters
	"reverse": filters.Reverse,
	"first":   filters.First,
	"last":    filters.Last,
	"firstn":  filters.FirstN,
	"lastn":   filters.LastN,
	// checksum filters
	"md5sum":    filters.SumMD5,
	"shasum":    filters.SumSHA,
	"sha256sum": filters.SumSHA256,
	"sha512sum": filters.SumSHA512,
	// math filters
	"add":       filters.Add,
	"sub":       filters.Sub,
	"mul":       filters.Mul,
	"div":       filters.Div,
	"mod":       filters.Mod,
	"pow":       filters.Pow,
	"min":       filters.Min,
	"max":       filters.Max,
	"rand":      filters.Rand,
	"increment": filters.Increment,
	// time function
	"now": filters.Now,
}

type Template struct {
	name      string
	filters   FuncMap
	root      parser.Node
	templates map[string]*Template
}

func New(name string) *Template {
	return &Template{
		name:      name,
		filters:   make(FuncMap),
		templates: make(map[string]*Template),
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
		t.filters[k] = f
	}
	return t
}

func (t *Template) ParseFiles(files ...string) (*Template, error) {
	for _, f := range files {
		tpl, err := ParseFile(f)
		if err != nil {
			return nil, err
		}
		_ = tpl
	}
	return t, nil
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
	filters := state.FuncMap(t.filters)
	return r.Execute(wr, set, state.EmptyState(data, filters))
}

func (t *Template) ExecuteTemplate(name string, w io.Writer, data interface{}) error {
	tpl, ok := t.templates[name]
	if !ok {
		return fmt.Errorf("%s: template not defined", name)
	}
	return tpl.Execute(w, data)
}
