package ppstern

import (
	"bytes"
	"encoding/json"
	"sort"
	"strconv"

	"github.com/fatih/color"
	"golang.org/x/exp/maps"
)

type JsonFormatter struct {
	keyColor *color.Color
}

func NewJsonFormatter() *JsonFormatter {
	return &JsonFormatter{
		keyColor: color.New(color.FgBlue),
	}
}

func (f *JsonFormatter) Format(v any) []byte {
	var buf bytes.Buffer
	f.pretty(v, &buf)
	return buf.Bytes()
}

func (f *JsonFormatter) pretty(v any, buf *bytes.Buffer) {
	switch v := v.(type) {
	case map[string]any:
		f.prettyMap(v, buf)
	case []any:
		f.prettyArray(v, buf)
	case string:
		f.prettyString(v, buf)
	case float64:
		buf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
	case json.Number:
		buf.WriteString(v.String())
	case bool:
		buf.WriteString(strconv.FormatBool(v))
	case nil:
		buf.WriteString("null")
	}
}

func (f *JsonFormatter) prettyMap(m map[string]any, buf *bytes.Buffer) {
	keys := maps.Keys(m)
	sort.Strings(keys)

	buf.WriteString("{")
	for i, key := range keys {
		buf.WriteString(f.keyColor.Sprintf(`"%s":`, key))
		f.pretty(m[key], buf)
		if i < len(m)-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString("}")
}

func (f *JsonFormatter) prettyArray(a []any, buf *bytes.Buffer) {
	buf.WriteString("[")
	for i, v := range a {
		f.pretty(v, buf)
		if i < len(a)-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString("]")
}

func (f *JsonFormatter) prettyString(s string, buf *bytes.Buffer) {
	b, _ := json.Marshal(s)
	buf.Write(b)
}
