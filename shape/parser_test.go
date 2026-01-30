package shape

import "testing"

func TestParseSimpleObject(t *testing.T) {
	input := `User : object
    + name : string
    - age  : int
`
	shape, err := Parse(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if shape.Name != "User" {
		t.Fatalf("expected shape name User, got %q", shape.Name)
	}
	if shape.Type.Kind != KindObject {
		t.Fatalf("expected object type, got %v", shape.Type.Kind)
	}
	if len(shape.Type.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(shape.Type.Fields))
	}
	if !shape.Type.Fields[0].Required || shape.Type.Fields[1].Required {
		t.Fatalf("required/optional flags not parsed")
	}
}

func TestParseAliasAndObjectArray(t *testing.T) {
	input := `Colors : object[]
    + color : string
    + value : string
    - userAgent(User Agent) : string
`
	shape, err := Parse(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if shape.Type.Kind != KindArray {
		t.Fatalf("expected array type")
	}
	if shape.Type.Elem == nil || shape.Type.Elem.Kind != KindObject {
		t.Fatalf("expected array of object")
	}
	fields := shape.Type.Elem.Fields
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(fields))
	}
	if fields[2].Alias != "User Agent" {
		t.Fatalf("expected alias 'User Agent', got %q", fields[2].Alias)
	}
}

func TestParsePrimitiveArray(t *testing.T) {
	input := `Vehicules : string[]`
	shape, err := Parse(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if shape.Type.Kind != KindArray || shape.Type.Elem.Kind != KindString {
		t.Fatalf("expected string array")
	}
}

func TestIndentationError(t *testing.T) {
	input := `User : object
  + name : string
`
	_, err := Parse(input)
	if err == nil {
		t.Fatalf("expected indentation error")
	}
}

func TestParseFileMultipleShapes(t *testing.T) {
	input := `Color : object
    + name : string
Colors : Color[]`
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(file.Shapes) != 2 {
		t.Fatalf("expected 2 shapes, got %d", len(file.Shapes))
	}
	color := file.ByName["Color"]
	if color == nil || color.Type.Kind != KindObject {
		t.Fatalf("expected Color object shape")
	}
	colors := file.ByName["Colors"]
	if colors == nil || colors.Type.Kind != KindArray {
		t.Fatalf("expected Colors array shape")
	}
	if colors.Type.Elem == nil || colors.Type.Elem.Kind != KindObject {
		t.Fatalf("expected Colors to reference Color object")
	}
	if len(colors.Type.Elem.Fields) != 1 {
		t.Fatalf("expected referenced object fields to be resolved")
	}
}

func TestParseFileUnknownReference(t *testing.T) {
	input := `Colors : Color[]`
	_, err := ParseFile(input)
	if err == nil {
		t.Fatalf("expected unknown type error")
	}
}
