package interpreter

import (
	"bytes"
	"sort"
	"strings"
)

// PrettyPrinter is an interface for values that can be pretty-printed
type PrettyPrinter interface {
	Pretty(indent int) string
}

func (a *Array) Pretty(indent int) string {
	if len(a.Elements) == 0 {
		return "[]"
	}
	
	var out bytes.Buffer
	out.WriteString("[\n")
	
	indentStr := strings.Repeat("  ", indent+1)
	
	for i, el := range a.Elements {
		out.WriteString(indentStr)
		if pp, ok := el.(PrettyPrinter); ok {
			out.WriteString(pp.Pretty(indent + 1))
		} else {
			out.WriteString(el.Inspect())
		}
		
		if i < len(a.Elements)-1 {
			out.WriteString(",")
		}
		out.WriteString("\n")
	}
	
	out.WriteString(strings.Repeat("  ", indent))
	out.WriteString("]")
	
	return out.String()
}

func (o *Object) Pretty(indent int) string {
	if len(o.Pairs) == 0 {
		return "{}"
	}
	
	var out bytes.Buffer
	out.WriteString("{\n")
	
	indentStr := strings.Repeat("  ", indent+1)
	
	// Sort keys for deterministic output
	keys := make([]string, 0, len(o.Pairs))
	for k := range o.Pairs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	for i, k := range keys {
		v := o.Pairs[k]
		out.WriteString(indentStr)
		out.WriteString(k)
		out.WriteString(": ")
		
		if pp, ok := v.(PrettyPrinter); ok {
			out.WriteString(pp.Pretty(indent + 1))
		} else {
			out.WriteString(v.Inspect())
		}
		
		if i < len(keys)-1 {
			out.WriteString(",")
		}
		out.WriteString("\n")
	}
	
	out.WriteString(strings.Repeat("  ", indent))
	out.WriteString("}")
	
	return out.String()
}

func (m *Map) Pretty(indent int) string {
	if len(m.Pairs) == 0 {
		return "map{}"
	}
	
	var out bytes.Buffer
	out.WriteString("map{\n")
	
	indentStr := strings.Repeat("  ", indent+1)
	
	// Sort keys for deterministic output is hard for mixed types, but we can try roughly by string representation
	type kv struct {
		k MapKey
		v Value
		s string
	}
	pairs := make([]kv, 0, len(m.Pairs))
	for k, v := range m.Pairs {
		pairs = append(pairs, kv{k, v, formatMapKey(k)})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].s < pairs[j].s
	})
	
	for i, pair := range pairs {
		out.WriteString(indentStr)
		out.WriteString(pair.s)
		out.WriteString(": ")
		
		if pp, ok := pair.v.(PrettyPrinter); ok {
			out.WriteString(pp.Pretty(indent + 1))
		} else {
			out.WriteString(pair.v.Inspect())
		}
		
		if i < len(pairs)-1 {
			out.WriteString(",")
		}
		out.WriteString("\n")
	}
	
	out.WriteString(strings.Repeat("  ", indent))
	out.WriteString("}")
	
	return out.String()
}

func (s *Set) Pretty(indent int) string {
	if len(s.Elements) == 0 {
		return "set{}"
	}
	
	var out bytes.Buffer
	out.WriteString("set{\n")
	
	indentStr := strings.Repeat("  ", indent+1)
	
	keys := make([]string, 0, len(s.Elements))
	for k := range s.Elements {
		keys = append(keys, formatMapKey(k))
	}
	sort.Strings(keys)
	
	for i, k := range keys {
		out.WriteString(indentStr)
		out.WriteString(k)
		
		if i < len(keys)-1 {
			out.WriteString(",")
		}
		out.WriteString("\n")
	}
	
	out.WriteString(strings.Repeat("  ", indent))
	out.WriteString("}")
	
	return out.String()
}
