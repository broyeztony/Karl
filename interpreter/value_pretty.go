package interpreter

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

// ANSI Color Codes
const (
	ColorReset   = "\033[0m"
	ColorKey     = "\033[1;34m" // Bold Blue
	ColorString  = "\033[32m"   // Green
	ColorNumber  = "\033[33m"   // Yellow
	ColorBool    = "\033[35m"   // Purple
	ColorNull    = "\033[1;30m" // Bold Grey
	ColorSymbol  = "\033[37m"   // White (for brackets, commas)
)

// PrettyPrinter is an interface for values that can be pretty-printed
type PrettyPrinter interface {
	Pretty(indent int) string
}

func colorize(val Value) string {
	s := val.Inspect()
	switch val.Type() {
	case STRING:
		return ColorString + s + ColorReset
	case INTEGER, FLOAT:
		return ColorNumber + s + ColorReset
	case BOOLEAN:
		return ColorBool + s + ColorReset
	case NULL:
		return ColorNull + s + ColorReset
	default:
		return s
	}
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
			out.WriteString(colorize(el))
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
		out.WriteString(ColorKey + fmt.Sprintf("%q", k) + ColorReset)
		out.WriteString(": ")
		
		if pp, ok := v.(PrettyPrinter); ok {
			out.WriteString(pp.Pretty(indent + 1))
		} else {
			out.WriteString(colorize(v))
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
		keyStr := formatMapKey(k)
		// Colorize key based on type if possible, or just treat as key color
		// For maps, keys can be anything. Let's color them as keys for consistency or values?
		// Let's stick to key color for consistency with Objects
		if k.Type == STRING {
			keyStr = ColorKey + keyStr + ColorReset
		} else {
			// For non-string keys, maybe use their value color?
			// But they are keys. Let's use Key color to distinguish.
			keyStr = ColorKey + keyStr + ColorReset
		}
		
		pairs = append(pairs, kv{k, v, keyStr})
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
			out.WriteString(colorize(pair.v))
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
		out.WriteString(ColorString + k + ColorReset) // Treat set elements as values? Strings mostly?
		
		if i < len(keys)-1 {
			out.WriteString(",")
		}
		out.WriteString("\n")
	}
	
	out.WriteString(strings.Repeat("  ", indent))
	out.WriteString("}")
	
	return out.String()
}
