package curly

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
	debug(bufio.NewWriter(w), n)
	return nil
}

func debug(w io.Writer, n Node) {
	debugWithLevel(w, n, 0)
}

func debugWithLevel(w io.Writer, n Node, level int) {
	prefix := strings.Repeat(" ", level)
	fmt.Fprint(w, prefix)
	switch n := n.(type) {
	case *Template:
		fmt.Fprint(w, "template [")
		fmt.Fprintln(w)
		for i := range n.nodes {
			debugWithLevel(w, n.nodes[i], level+2)
		}
		fmt.Fprint(w, prefix)
		fmt.Fprintln(w, "]")
	case *BlockNode:
		fmt.Fprint(w, "block(key: ")
		fmt.Fprint(w, n.key.name)
		fmt.Fprint(w, ") [")
		fmt.Fprintln(w)
		for i := range n.nodes {
			debugWithLevel(w, n.nodes[i], level+2)
		}
		fmt.Fprint(w, prefix)
		fmt.Fprintln(w, "]")
	case *VariableNode:
		fmt.Fprint(w, "variable(key: ")
		fmt.Fprint(w, n.key.name)
		fmt.Fprint(w, ", unescape: ")
		fmt.Fprint(w, n.unescap)
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
