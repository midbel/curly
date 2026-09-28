package curly

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Writer struct {
	ws *bufio.Writer

	Indent  string
	Compact bool

	level int
	err   error
}

func Compact(w io.Writer) *Writer {
	ws := NewWriter(w)
	ws.Compact = true
	return ws
}

func NewWriter(w io.Writer) *Writer {
	ws := Writer{
		ws:     bufio.NewWriter(w),
		Indent: "  ",
	}
	return &ws
}

func (w *Writer) Write(value any) error {
	defer w.reset()
	w.writeValue(value)
	if w.err != nil {
		return w.err
	}
	return w.flush()
}

func (w *Writer) writeValue(value any) {
	switch v := value.(type) {
	case map[string]any:
		w.writeObject(v)
	case []any:
		w.writeArray(v)
	default:
		w.writeLiteral(value)
	}
}

func (w *Writer) writeObject(value map[string]any) {
	w.writeRune('{')
	if len(value) == 0 {
		w.writeRune('}')
		return
	}
	w.enter()
	w.writeNL()
	var i int
	for k, v := range value {
		if i > 0 {
			w.writeRune(',')
			w.writeNL()
		}
		w.writePrefix()
		w.writeKey(k)
		w.writeValue(v)
		if w.err != nil {
			break
		}
		i++
	}
	w.leave()
	w.writeNL()
	w.writePrefix()
	w.writeRune('}')
}

func (w *Writer) writeArray(value []any) {
	w.writeRune('[')
	if len(value) == 0 {
		w.writeRune(']')
		return
	}

	w.enter()
	w.writeNL()
	for i := range value {
		if i > 0 {
			w.writeRune(',')
			w.writeNL()
		}
		w.writePrefix()
		w.writeValue(value[i])
		if w.err != nil {
			return
		}
	}
	w.leave()
	w.writeNL()
	w.writePrefix()
	w.writeRune(']')
}

func (w *Writer) writeLiteral(value any) {
	if value == nil {
		w.writeString("null")
		return
	}
	switch v := value.(type) {
	case bool:
		if v {
			w.writeString("true")
		} else {
			w.writeString("false")
		}
	case float64:
		w.writeString(strconv.FormatFloat(v, 'f', -1, 64))
	case int64:
		w.writeString(strconv.FormatInt(v, 10))
	case int:
		w.writeString(strconv.FormatInt(int64(v), 10))
	case string:
		w.writeQuote(v)
	default:
		w.err = fmt.Errorf("unsupported json type %T", value)
	}
}

func (w *Writer) writeKey(key string) {
	w.writeLiteral(key)
	w.writeRune(':')
	if !w.Compact {
		w.writeRune(' ')
	}
}

func (w *Writer) writeQuote(value string) {
	w.writeRune('"')
	for _, r := range value {
		switch r {
		case '"', '\\':
			w.writeRune('\\')
			w.writeRune('"')
		case '\n':
			w.writeString("\n")
		case '\r':
			w.writeString("\r")
		case '\t':
			w.writeString("\t")
		case '\b':
			w.writeString("\b")
		case '\f':
			w.writeString("\f")
		default:
			w.writeRune(r)
		}
	}
	w.writeRune('"')
}

func (w *Writer) writePrefix() {
	if w.Compact || w.level == 0 {
		return
	}
	space := strings.Repeat(w.Indent, w.level)
	w.writeString(space)
}

func (w *Writer) writeNL() {
	if w.Compact {
		return
	}
	w.writeRune('\n')
}

func (w *Writer) writeRune(char rune) {
	if w.err != nil {
		return
	}
	_, w.err = w.ws.WriteRune(char)
}

func (w *Writer) writeString(str string) {
	if w.err != nil {
		return
	}
	_, w.err = w.ws.WriteString(str)
}

func (w *Writer) enter() {
	w.level++
}

func (w *Writer) leave() {
	w.level--
}

func (w *Writer) reset() {
	w.level = 0
	w.err = nil
}

func (w *Writer) flush() error {
	return w.ws.Flush()
}

func escapeChar(r rune) bool {
	switch r {
	case '"':
	case '\\':
	case '\n':
	case '\r':
	case '\t':
	case '\b':
	case '\f':
	default:
		return false
	}
	return true
}
