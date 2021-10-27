package curly

import (
	"fmt"
	"io"
	"strings"
)

func Debug(r io.Reader) {

}

func debug(n Node) {
	debugWithLevel(n, 0)
}

func debugWithLevel(n Node, level int) {
	prefix := strings.Repeat(" ", level)
	fmt.Print(prefix)
	switch n := n.(type) {
	case *Template:
		fmt.Print("template [")
		fmt.Println()
		for i := range n.nodes {
			debugWithLevel(n.nodes[i], level+2)
		}
		fmt.Print(prefix)
		fmt.Println("]")
	case *BlockNode:
		fmt.Print("block(key: ")
		fmt.Print(n.key.name)
		fmt.Print(") [")
		fmt.Println()
		for i := range n.nodes {
			debugWithLevel(n.nodes[i], level+2)
		}
		fmt.Print(prefix)
		fmt.Println("]")
	case *VariableNode:
		fmt.Print("variable(key: ")
		fmt.Print(n.key.name)
		fmt.Print(", unescape: ")
		fmt.Print(n.unescap)
		fmt.Print(")")
		fmt.Println()
	case *CommentNode:
		fmt.Print("comment(")
		fmt.Print(n.str)
		fmt.Print(")")
		fmt.Println()
	case *LiteralNode:
		fmt.Print("literal(str: ")
		for j, i := range strings.Split(n.str, "\n") {
			if j > 0 {
				fmt.Println()
				fmt.Print(prefix)
			}
			fmt.Print(i)
		}
		fmt.Print(")")
		fmt.Println()
	}
}
