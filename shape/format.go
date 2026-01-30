package shape

import (
	"encoding/json"
	"strings"
)

// Format renders a shape in .shape-like form.
func Format(s *Shape) string {
	if s == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(s.Name)
	b.WriteString(" : ")
	b.WriteString(typeLabel(s.Type))
	b.WriteString("\n")
	if obj := ObjectType(s.Type); obj != nil {
		writeFields(&b, obj, 4)
	}
	return b.String()
}

// FormatFile renders all shapes in a .shape file.
func FormatFile(f *File) string {
	if f == nil || len(f.Shapes) == 0 {
		return ""
	}
	var b strings.Builder
	for i, sh := range f.Shapes {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(Format(sh))
	}
	return b.String()
}

// FormatJSON renders the shape as JSON.
func FormatJSON(s *Shape) (string, error) {
	out := shapeJSON(s)
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// FormatJSONFile renders all shapes in a .shape file as JSON.
func FormatJSONFile(f *File) (string, error) {
	if f == nil {
		return "", nil
	}
	shapes := make([]interface{}, 0, len(f.Shapes))
	for _, sh := range f.Shapes {
		shapes = append(shapes, shapeJSON(sh))
	}
	out := map[string]interface{}{
		"shapes": shapes,
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func shapeJSON(s *Shape) map[string]interface{} {
	return map[string]interface{}{
		"name": s.Name,
		"type": toJSONType(s.Type),
	}
}

func typeLabel(t *Type) string {
	if t == nil {
		return ""
	}
	if t.Kind == KindArray {
		return typeLabel(t.Elem) + "[]"
	}
	switch t.Kind {
	case KindString:
		return "string"
	case KindInt:
		return "int"
	case KindFloat:
		return "float"
	case KindBool:
		return "bool"
	case KindNull:
		return "null"
	case KindAny:
		return "any"
	case KindObject:
		return "object"
	case KindRef:
		if t.RefName != "" {
			return t.RefName
		}
		return "ref"
	case KindUnion:
		return "union"
	default:
		return ""
	}
}

func writeFields(b *strings.Builder, obj *Type, indent int) {
	if obj == nil {
		return
	}
	pad := strings.Repeat(" ", indent)
	for _, f := range obj.Fields {
		prefix := "+"
		if !f.Required {
			prefix = "-"
		}
		b.WriteString(pad)
		b.WriteString(prefix)
		b.WriteString(" ")
		b.WriteString(f.Name)
		if f.Alias != "" {
			b.WriteString("(")
			b.WriteString(f.Alias)
			b.WriteString(")")
		}
		b.WriteString(" : ")
		b.WriteString(typeLabel(f.Type))
		b.WriteString("\n")
		if nested := ObjectType(f.Type); nested != nil {
			writeFields(b, nested, indent+4)
		}
	}
}

func toJSONType(t *Type) interface{} {
	if t == nil {
		return nil
	}
	out := map[string]interface{}{
		"kind": t.Kind.String(),
	}
	switch t.Kind {
	case KindArray:
		out["elem"] = toJSONType(t.Elem)
	case KindObject:
		fields := make([]interface{}, 0, len(t.Fields))
		for _, f := range t.Fields {
			fields = append(fields, map[string]interface{}{
				"name":     f.Name,
				"alias":    f.Alias,
				"required": f.Required,
				"type":     toJSONType(f.Type),
			})
		}
		out["fields"] = fields
	case KindUnion:
		opts := make([]interface{}, 0, len(t.Options))
		for _, opt := range t.Options {
			opts = append(opts, toJSONType(opt))
		}
		out["options"] = opts
	case KindRef:
		out["ref"] = t.RefName
	}
	return out
}
