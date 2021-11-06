package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func Debug(r io.Reader, w io.Writer) error {
	n, err := Parse(r)
	if err != nil {
		return err
	}
	ws := bufio.NewWriter(w)
	defer ws.Flush()

	debug(ws, n)
	return nil
}

func debug(w io.Writer, n Node) {
	debugWithLevel(w, n, 0)
}

func debugWithLevel(w io.Writer, n Node, level int) {
	prefix := strings.Repeat(" ", level)
	fmt.Fprint(w, prefix)
	switch n := n.(type) {
	case *RootNode:
		fmt.Fprint(w, "template [")
		fmt.Fprintln(w)
		for i := range n.Named {
			debugWithLevel(w, n.Named[i], level+2)
		}
		for i := range n.Nodes {
			debugWithLevel(w, n.Nodes[i], level+2)
		}
		fmt.Fprint(w, prefix)
		fmt.Fprintln(w, "]")
	case *AssignmentNode:
		fmt.Fprint(w, "assignment(name: ")
		fmt.Fprint(w, n.ident)
		if key, filters := getKeyFields(n.key); key != "" {
			fmt.Fprint(w, ", key: ")
			fmt.Fprint(w, key)
			printFilters(w, filters)
		}
		fmt.Fprintln(w, ")")
	case *ExecNode:
		fmt.Fprint(w, "exec(name: ")
		fmt.Fprint(w, n.name)
		if key, filters := getKeyFields(n.key); key != "" {
			fmt.Fprint(w, ", key: ")
			fmt.Fprint(w, key)
			printFilters(w, filters)
		}
		fmt.Fprintln(w, ")")
	case *DefineNode:
		fmt.Fprint(w, "define(name: ")
		fmt.Fprint(w, n.name)
		fmt.Fprintln(w, ") [")
		for i := range n.nodes {
			debugWithLevel(w, n.nodes[i], level+2)
		}
		fmt.Fprintln(w, "]")
	case *SectionNode:
		fmt.Fprint(w, "section(name: ")
		fmt.Fprint(w, n.name)
		fmt.Fprintln(w, ")")
	case *BlockNode:
		key, filters := getKeyFields(n.key)
		fmt.Fprint(w, "block(key: ")
		fmt.Fprint(w, key)
		printFilters(w, filters)
		fmt.Fprint(w, ") [")
		fmt.Fprintln(w)
		for i := range n.nodes {
			debugWithLevel(w, n.nodes[i], level+2)
		}
		fmt.Fprint(w, prefix)
		fmt.Fprintln(w, "]")
	case *VariableNode:
		key, filters := getKeyFields(n.key)
		fmt.Fprint(w, "variable(key: ")
		fmt.Fprint(w, key)
		fmt.Fprint(w, ", unescape: ")
		fmt.Fprint(w, n.unescap)
		printFilters(w, filters)
		fmt.Fprint(w, ")")
		fmt.Fprintln(w)
	case *CommentNode:
		fmt.Fprint(w, "comment(")
		fmt.Fprint(w, n.str)
		fmt.Fprint(w, ")")
		fmt.Fprintln(w)
	case *LiteralNode:
		fmt.Fprint(w, "literal(str: ")
		for j, i := range strings.Split(n.str, "\n") {
			if j > 0 {
				fmt.Fprintln(w)
				fmt.Fprint(w, prefix)
			}
			fmt.Fprint(w, i)
		}
		fmt.Fprint(w, ")")
		fmt.Fprintln(w)
	}
}

func getKeyFields(k Key) (string, []Filter) {
	switch k := k.(type) {
	case IdentKey:
		return k.name, k.filters
	case ValueKey:
		return k.literal, k.filters
	default:
		return "", nil
	}
}

func printFilters(w io.Writer, filters []Filter) {
	if len(filters) == 0 {
		return
	}
	fmt.Fprint(w, ", filter: ")
	for i, f := range filters {
		if i > 0 {
			fmt.Fprint(w, " | ")
		}
		fmt.Fprint(w, f.name)
	}
}
