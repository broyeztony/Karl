package shape

// TypeKind represents the kind of a shape type.
type TypeKind int

const (
	KindString TypeKind = iota
	KindInt
	KindFloat
	KindBool
	KindNull
	KindAny
	KindObject
	KindArray
	KindRef
	KindUnion // reserved for future union-based types
)

// Type describes a shape type.
type Type struct {
	Kind    TypeKind
	Elem    *Type
	Fields  []Field
	Options []*Type // reserved for future union-based types
	RefName string
}

// Field is a named entry in an object shape.
type Field struct {
	Name     string
	Alias    string
	Required bool
	Type     *Type
}

// Shape is the top-level declaration in a .shape file.
type Shape struct {
	Name string
	Type *Type
}

// File is the parsed representation of a .shape file.
// It may contain multiple top-level shapes.
type File struct {
	Shapes []*Shape
	ByName map[string]*Shape
}

// ObjectType returns the underlying object type used for fields,
// whether the current type is object or array-of-object.
func ObjectType(t *Type) *Type {
	if t == nil {
		return nil
	}
	if t.Kind == KindObject {
		return t
	}
	if t.Kind == KindArray && t.Elem != nil && t.Elem.Kind == KindObject {
		return t.Elem
	}
	return nil
}

func (k TypeKind) String() string {
	switch k {
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
	case KindArray:
		return "array"
	case KindRef:
		return "ref"
	case KindUnion:
		return "union"
	default:
		return "unknown"
	}
}

// AliasFor returns the alias for a given field name, if any.
func AliasFor(t *Type, field string) (string, bool) {
	if t == nil || t.Kind != KindObject {
		return "", false
	}
	for _, f := range t.Fields {
		if f.Name == field {
			if f.Alias == "" {
				return "", false
			}
			return f.Alias, true
		}
	}
	return "", false
}
