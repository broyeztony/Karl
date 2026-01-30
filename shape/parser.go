package shape

import (
	"fmt"
	"strings"
)

// Parse parses a .shape file content into a single Shape.
// It fails if the file declares multiple shapes.
func Parse(content string) (*Shape, error) {
	file, err := ParseFile(content)
	if err != nil {
		return nil, err
	}
	if len(file.Shapes) != 1 {
		return nil, fmt.Errorf("shape parse error: expected exactly one shape")
	}
	return file.Shapes[0], nil
}

// ParseFile parses a .shape file content into a File that can contain
// multiple top-level shapes.
func ParseFile(content string) (*File, error) {
	lines := collectLines(content)
	if len(lines) == 0 {
		return nil, fmt.Errorf("shape parse error: empty file")
	}

	file := &File{Shapes: []*Shape{}, ByName: map[string]*Shape{}}
	var contexts []context
	var current *Shape

	for i := 0; i < len(lines); i++ {
		ln := lines[i]
		if ln.indent == 0 {
			name, typeStr, err := splitDecl(ln.text)
			if err != nil {
				return nil, lineError(ln, err.Error())
			}
			if !isIdent(name) {
				return nil, lineError(ln, "shape name must be a valid identifier")
			}
			if _, exists := file.ByName[name]; exists {
				return nil, lineError(ln, "duplicate shape name: "+name)
			}
			rootType, err := parseType(typeStr)
			if err != nil {
				return nil, lineError(ln, err.Error())
			}
			current = &Shape{Name: name, Type: rootType}
			file.Shapes = append(file.Shapes, current)
			file.ByName[name] = current
			contexts = []context{}
			if obj := ObjectType(rootType); obj != nil {
				contexts = append(contexts, context{indent: ln.indent + 4, target: obj})
			}
			continue
		}
		if current == nil {
			return nil, lineError(ln, "field declared outside of a shape")
		}
		if len(contexts) == 0 {
			return nil, lineError(ln, "field declared outside of an object shape")
		}
		for len(contexts) > 0 && ln.indent < contexts[len(contexts)-1].indent {
			contexts = contexts[:len(contexts)-1]
		}
		if len(contexts) == 0 || ln.indent != contexts[len(contexts)-1].indent {
			return nil, lineError(ln, "unexpected indentation")
		}
		ctx := contexts[len(contexts)-1]
		field, fieldType, err := parseField(ln.text)
		if err != nil {
			return nil, lineError(ln, err.Error())
		}
		field.Type = fieldType
		ctx.target.Fields = append(ctx.target.Fields, field)
		if obj := ObjectType(fieldType); obj != nil {
			contexts = append(contexts, context{indent: ln.indent + 4, target: obj})
		}
	}

	if err := resolveFile(file); err != nil {
		return nil, err
	}
	return file, nil
}

type context struct {
	indent int
	target *Type
}

type parsedLine struct {
	num    int
	indent int
	text   string
}

func collectLines(content string) []parsedLine {
	lines := []parsedLine{}
	raw := strings.Split(content, "\n")
	for i, line := range raw {
		line = strings.TrimRight(line, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		indent := countIndent(line)
		text := strings.TrimSpace(line)
		lines = append(lines, parsedLine{num: i + 1, indent: indent, text: text})
	}
	return lines
}

func countIndent(line string) int {
	count := 0
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case ' ':
			count++
		case '\t':
			count += 4
		default:
			return count
		}
	}
	return count
}

func splitDecl(text string) (string, string, error) {
	parts := strings.SplitN(text, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected ':' in declaration")
	}
	name := strings.TrimSpace(parts[0])
	typeStr := strings.TrimSpace(parts[1])
	if name == "" || typeStr == "" {
		return "", "", fmt.Errorf("invalid declaration")
	}
	return name, typeStr, nil
}

func parseField(text string) (Field, *Type, error) {
	left, typeStr, err := splitDecl(text)
	if err != nil {
		return Field{}, nil, err
	}
	required := true
	if strings.HasPrefix(left, "+") {
		left = strings.TrimSpace(left[1:])
		required = true
	} else if strings.HasPrefix(left, "-") {
		left = strings.TrimSpace(left[1:])
		required = false
	}

	name := left
	alias := ""
	if idx := strings.Index(left, "("); idx >= 0 {
		if !strings.HasSuffix(left, ")") {
			return Field{}, nil, fmt.Errorf("alias must end with ')' ")
		}
		name = strings.TrimSpace(left[:idx])
		alias = strings.TrimSpace(left[idx+1 : len(left)-1])
	}
	if name == "" || !isIdent(name) {
		return Field{}, nil, fmt.Errorf("field name must be a valid identifier")
	}
	fieldType, err := parseType(typeStr)
	if err != nil {
		return Field{}, nil, err
	}
	return Field{Name: name, Alias: alias, Required: required, Type: fieldType}, fieldType, nil
}

func parseType(text string) (*Type, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("missing type")
	}
	if strings.HasSuffix(text, "[]") {
		base := strings.TrimSpace(strings.TrimSuffix(text, "[]"))
		elem, err := parseType(base)
		if err != nil {
			return nil, err
		}
		return &Type{Kind: KindArray, Elem: elem}, nil
	}
	if strings.Contains(text, "|") {
		return nil, fmt.Errorf("union types are not supported yet")
	}
	switch text {
	case "string":
		return &Type{Kind: KindString}, nil
	case "int":
		return &Type{Kind: KindInt}, nil
	case "float":
		return &Type{Kind: KindFloat}, nil
	case "bool":
		return &Type{Kind: KindBool}, nil
	case "null":
		return &Type{Kind: KindNull}, nil
	case "any":
		return &Type{Kind: KindAny}, nil
	case "object":
		return &Type{Kind: KindObject, Fields: []Field{}}, nil
	default:
		if isIdent(text) {
			return &Type{Kind: KindRef, RefName: text}, nil
		}
		return nil, fmt.Errorf("unknown type: %s", text)
	}
}

func isIdent(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		ch := name[i]
		if i == 0 {
			if !(ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')) {
				return false
			}
			continue
		}
		if !(ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
			return false
		}
	}
	return true
}

func lineError(line parsedLine, msg string) error {
	return fmt.Errorf("shape parse error at line %d: %s", line.num, msg)
}
