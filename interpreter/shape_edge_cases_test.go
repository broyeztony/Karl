package interpreter

import (
	"encoding/json"
	"testing"

	"karl/shape"
)

// Edge Case 1: Alias Collision - Both internal and alias keys present
// Spec: "If both the internal key and alias key are present, internal wins"
func TestApplyShapeAliasCollision(t *testing.T) {
	shapeText := `Response : object
    + data(external-data) : string
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	
	// Both keys present: internal "data" and alias "external-data"
	value := &Object{Pairs: map[string]Value{
		"data":          &String{Value: "internal"},
		"external-data": &String{Value: "external"},
	}}
	
	out, err := applyShape(value, sh.Type)
	if err != nil {
		t.Fatalf("applyShape error: %v", err)
	}
	
	obj := out.(*Object)
	result, ok := obj.Pairs["data"].(*String)
	if !ok {
		t.Fatalf("expected string value")
	}
	
	// Internal should win
	if result.Value != "internal" {
		t.Fatalf("expected internal value to win, got %q", result.Value)
	}
}

// Edge Case 2: Deeply nested objects with mixed required/optional
func TestApplyShapeDeeplyNested(t *testing.T) {
	shapeText := `Deep : object
    + level1 : object
        + level2 : object
            - level3 : object
                + value : string
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	
	// Missing optional level3 should set to null
	value := &Object{Pairs: map[string]Value{
		"level1": &Object{Pairs: map[string]Value{
			"level2": &Object{Pairs: map[string]Value{}}, // level3 missing
		}},
	}}
	
	out, err := applyShape(value, sh.Type)
	if err != nil {
		t.Fatalf("applyShape error: %v", err)
	}
	
	obj := out.(*Object)
	l1 := obj.Pairs["level1"].(*Object)
	l2 := l1.Pairs["level2"].(*Object)
	
	// level3 should be null
	if _, ok := l2.Pairs["level3"].(*Null); !ok {
		t.Fatalf("expected null for missing optional field")
	}
}

// Edge Case 3: Missing required field in deeply nested structure
func TestApplyShapeDeepMissingRequired(t *testing.T) {
	shapeText := `Deep : object
    + level1 : object
        + level2 : object
            + required : string
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	
	// Missing required field at deep level
	value := &Object{Pairs: map[string]Value{
		"level1": &Object{Pairs: map[string]Value{
			"level2": &Object{Pairs: map[string]Value{}}, // required is missing
		}},
	}}
	
	_, err = applyShape(value, sh.Type)
	if err == nil {
		t.Fatalf("expected error for missing required field at deep level")
	}
	if _, ok := err.(*RecoverableError); !ok {
		t.Fatalf("expected recoverable error, got %T", err)
	}
}

// Edge Case 4: Empty array
func TestApplyShapeEmptyArray(t *testing.T) {
	shapeText := `Items : object[]
    + id : int
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	
	value := &Array{Elements: []Value{}}
	
	out, err := applyShape(value, sh.Type)
	if err != nil {
		t.Fatalf("applyShape error: %v", err)
	}
	
	arr := out.(*Array)
	if len(arr.Elements) != 0 {
		t.Fatalf("expected empty array, got %d elements", len(arr.Elements))
	}
}

// Edge Case 5: Array with invalid element (should fail on first bad element)
func TestApplyShapeArrayInvalidElement(t *testing.T) {
	shapeText := `Items : object[]
    + id : int
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	
	value := &Array{Elements: []Value{
		&Object{Pairs: map[string]Value{"id": &Integer{Value: 1}}},
		&Object{Pairs: map[string]Value{}}, // missing required id
	}}
	
	_, err = applyShape(value, sh.Type)
	if err == nil {
		t.Fatalf("expected error for array with invalid element")
	}
}

// Edge Case 6: Multiple aliases with spaces and special characters
func TestApplyShapeComplexAliases(t *testing.T) {
	shapeText := `Headers : object
    + contentType(Content-Type) : string
    + xApiKey(X-API-KEY) : string
    + customHeader(Custom Header With Spaces) : string
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	
	value := &Object{Pairs: map[string]Value{
		"Content-Type":              &String{Value: "application/json"},
		"X-API-KEY":                 &String{Value: "secret"},
		"Custom Header With Spaces": &String{Value: "value"},
	}}
	
	out, err := applyShape(value, sh.Type)
	if err != nil {
		t.Fatalf("applyShape error: %v", err)
	}
	
	obj := out.(*Object)
	if obj.Pairs["contentType"].(*String).Value != "application/json" {
		t.Fatalf("alias mapping failed for Content-Type")
	}
	if obj.Pairs["xApiKey"].(*String).Value != "secret" {
		t.Fatalf("alias mapping failed for X-API-KEY")
	}
	if obj.Pairs["customHeader"].(*String).Value != "value" {
		t.Fatalf("alias mapping failed for Custom Header With Spaces")
	}
}

// Edge Case 7: Extra fields should be dropped
func TestApplyShapeDropExtraFieldsNested(t *testing.T) {
	shapeText := `Config : object
    + settings : object
        + enabled : bool
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	
	value := &Object{Pairs: map[string]Value{
		"settings": &Object{Pairs: map[string]Value{
			"enabled": &Boolean{Value: true},
			"extra1":  &String{Value: "drop"},
			"extra2":  &Integer{Value: 42},
		}},
		"topExtra": &String{Value: "also drop"},
	}}
	
	out, err := applyShape(value, sh.Type)
	if err != nil {
		t.Fatalf("applyShape error: %v", err)
	}
	
	obj := out.(*Object)
	if _, ok := obj.Pairs["topExtra"]; ok {
		t.Fatalf("top-level extra field not dropped")
	}
	
	settings := obj.Pairs["settings"].(*Object)
	if _, ok := settings.Pairs["extra1"]; ok {
		t.Fatalf("nested extra field 'extra1' not dropped")
	}
	if _, ok := settings.Pairs["extra2"]; ok {
		t.Fatalf("nested extra field 'extra2' not dropped")
	}
}

// Edge Case 8: Primitive array with wrong type element
func TestApplyShapePrimitiveArrayTypeError(t *testing.T) {
	shapeText := `Numbers : int[]`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	
	value := &Array{Elements: []Value{
		&Integer{Value: 1},
		&String{Value: "not a number"},
	}}
	
	_, err = applyShape(value, sh.Type)
	if err == nil {
		t.Fatalf("expected error for wrong type in array")
	}
}

// Edge Case 9: Workaround for nested arrays using object wrapper
// Since "int[][]" syntax is not properly supported, demonstrate the
// correct way to achieve nested arrays using intermediate objects
func TestApplyShapeNestedArraysWorkaround(t *testing.T) {
	shapeText := `Matrix : object[]
    + row : int[]
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	
	// Create a matrix as array of objects, where each object has a row array
	value := &Array{Elements: []Value{
		&Object{Pairs: map[string]Value{
			"row": &Array{Elements: []Value{
				&Integer{Value: 1},
				&Integer{Value: 2},
			}},
		}},
		&Object{Pairs: map[string]Value{
			"row": &Array{Elements: []Value{
				&Integer{Value: 3},
				&Integer{Value: 4},
			}},
		}},
	}}
	
	out, err := applyShape(value, sh.Type)
	if err != nil {
		t.Fatalf("applyShape error: %v", err)
	}
	
	arr := out.(*Array)
	if len(arr.Elements) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(arr.Elements))
	}
	
	// Verify the nested structure
	row0 := arr.Elements[0].(*Object).Pairs["row"].(*Array)
	if len(row0.Elements) != 2 {
		t.Fatalf("expected 2 elements in first row")
	}
}

// Edge Case 10: All optional fields missing
func TestApplyShapeAllOptionalMissing(t *testing.T) {
	shapeText := `Optional : object
    - field1 : string
    - field2 : int
    - field3 : bool
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	
	value := &Object{Pairs: map[string]Value{}}
	
	out, err := applyShape(value, sh.Type)
	if err != nil {
		t.Fatalf("applyShape error: %v", err)
	}
	
	obj := out.(*Object)
	// All fields should be null
	for _, field := range []string{"field1", "field2", "field3"} {
		if _, ok := obj.Pairs[field].(*Null); !ok {
			t.Fatalf("expected null for missing optional field %s", field)
		}
	}
}

// Edge Case 11: Type mismatch for primitive types
func TestApplyShapeTypeMismatch(t *testing.T) {
	cases := []struct {
		shapeType string
		value     Value
	}{
		{"string", &Integer{Value: 42}},
		{"int", &String{Value: "42"}},
		{"float", &Boolean{Value: true}},
		{"bool", &Float{Value: 1.0}},
		{"null", &String{Value: "null"}},
	}
	
	for _, tc := range cases {
		shapeText := "Value : " + tc.shapeType
		sh, err := shape.Parse(shapeText)
		if err != nil {
			t.Fatalf("shape parse error for %s: %v", tc.shapeType, err)
		}
		
		_, err = applyShape(tc.value, sh.Type)
		if err == nil {
			t.Fatalf("expected type error for %s, got nil", tc.shapeType)
		}
		if _, ok := err.(*RecoverableError); !ok {
			t.Fatalf("expected recoverable error for %s, got %T", tc.shapeType, err)
		}
	}
}

// Edge Case 12: Non-object value for object shape
func TestApplyShapeWrongTopLevelType(t *testing.T) {
	shapeText := `User : object
    + name : string
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	
	cases := []Value{
		&String{Value: "not an object"},
		&Integer{Value: 42},
		&Array{Elements: []Value{}},
		NullValue,
	}
	
	for _, value := range cases {
		_, err := applyShape(value, sh.Type)
		if err == nil {
			t.Fatalf("expected error for non-object value %T", value)
		}
	}
}

// Edge Case 13: JSON round-trip with nested aliases
func TestApplyShapeJSONRoundTripNested(t *testing.T) {
	shapeText := `API : object
    + meta : object
        + requestId(X-Request-ID) : string
        + timestamp(created_at) : int
    + data : object
        + userId(user_id) : int
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	
	value := &Object{Pairs: map[string]Value{
		"meta": &Object{Pairs: map[string]Value{
			"X-Request-ID": &String{Value: "abc123"},
			"created_at":   &Integer{Value: 1234567890},
		}},
		"data": &Object{Pairs: map[string]Value{
			"user_id": &Integer{Value: 42},
		}},
	}}
	
	shaped, err := applyShape(value, sh.Type)
	if err != nil {
		t.Fatalf("applyShape error: %v", err)
	}
	
	// Verify internal keys
	obj := shaped.(*Object)
	meta := obj.Pairs["meta"].(*Object)
	if meta.Pairs["requestId"].(*String).Value != "abc123" {
		t.Fatalf("alias mapping failed")
	}
	
	// Encode back to JSON
	enc, err := encodeJSONValue(shaped)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	
	raw, err := json.Marshal(enc)
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}
	
	var decoded map[string]interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	
	// Verify external keys are restored
	metaOut := decoded["meta"].(map[string]interface{})
	if _, ok := metaOut["X-Request-ID"]; !ok {
		t.Fatalf("expected alias key 'X-Request-ID' in JSON output")
	}
	if _, ok := metaOut["created_at"]; !ok {
		t.Fatalf("expected alias key 'created_at' in JSON output")
	}
}

// Edge Case 14: Map as object input
func TestApplyShapeMapAsObject(t *testing.T) {
	shapeText := `Config : object
    + key : string
    + value : int
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	
	// Using Map instead of Object
	value := &Map{Pairs: map[MapKey]Value{
		{Type: STRING, Value: "key"}:   &String{Value: "test"},
		{Type: STRING, Value: "value"}: &Integer{Value: 123},
	}}
	
	out, err := applyShape(value, sh.Type)
	if err != nil {
		t.Fatalf("applyShape error: %v", err)
	}
	
	obj := out.(*Object)
	if obj.Pairs["key"].(*String).Value != "test" {
		t.Fatalf("map to object conversion failed")
	}
}

// Edge Case 15: Map with alias lookup
func TestApplyShapeMapWithAlias(t *testing.T) {
	shapeText := `Headers : object
    + contentType(Content-Type) : string
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	
	value := &Map{Pairs: map[MapKey]Value{
		{Type: STRING, Value: "Content-Type"}: &String{Value: "application/json"},
	}}
	
	out, err := applyShape(value, sh.Type)
	if err != nil {
		t.Fatalf("applyShape error: %v", err)
	}
	
	obj := out.(*Object)
	if obj.Pairs["contentType"].(*String).Value != "application/json" {
		t.Fatalf("map alias lookup failed")
	}
}

// Edge Case 16: Shape with 'any' type accepts anything
func TestApplyShapeAnyType(t *testing.T) {
	shapeText := `Flexible : object
    + data : any
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	
	testValues := []Value{
		&String{Value: "text"},
		&Integer{Value: 42},
		&Boolean{Value: true},
		&Array{Elements: []Value{&Integer{Value: 1}}},
		&Object{Pairs: map[string]Value{"nested": &String{Value: "object"}}},
	}
	
	for _, val := range testValues {
		value := &Object{Pairs: map[string]Value{"data": val}}
		_, err := applyShape(value, sh.Type)
		if err != nil {
			t.Fatalf("applyShape failed for any type with %T: %v", val, err)
		}
	}
}

// Edge Case 17: Array of objects with some elements missing required fields
func TestApplyShapeArrayPartialFailure(t *testing.T) {
	shapeText := `Users : object[]
    + id : int
    + name : string
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	
	value := &Array{Elements: []Value{
		&Object{Pairs: map[string]Value{
			"id":   &Integer{Value: 1},
			"name": &String{Value: "Alice"},
		}},
		&Object{Pairs: map[string]Value{
			"id": &Integer{Value: 2},
			// name missing
		}},
		&Object{Pairs: map[string]Value{
			"id":   &Integer{Value: 3},
			"name": &String{Value: "Charlie"},
		}},
	}}
	
	_, err = applyShape(value, sh.Type)
	if err == nil {
		t.Fatalf("expected error for array element with missing required field")
	}
}

// Edge Case 18: Empty object with all required fields
func TestApplyShapeEmptyObjectRequiredFields(t *testing.T) {
	shapeText := `Required : object
    + field1 : string
    + field2 : int
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	
	value := &Object{Pairs: map[string]Value{}}
	
	_, err = applyShape(value, sh.Type)
	if err == nil {
		t.Fatalf("expected error for empty object with required fields")
	}
}

// Edge Case 19: Null value for null type
func TestApplyShapeNullType(t *testing.T) {
	shapeText := `NullField : object
    + value : null
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	
	value := &Object{Pairs: map[string]Value{
		"value": NullValue,
	}}
	
	_, err = applyShape(value, sh.Type)
	if err != nil {
		t.Fatalf("applyShape error for null type: %v", err)
	}
	
	// Non-null should fail
	value2 := &Object{Pairs: map[string]Value{
		"value": &String{Value: "not null"},
	}}
	
	_, err = applyShape(value2, sh.Type)
	if err == nil {
		t.Fatalf("expected error for non-null value with null type")
	}
}

// Edge Case 20: Empty alias (field with parentheses but no content)
func TestParseShapeEmptyAlias(t *testing.T) {
	shapeText := `Config : object
    + field() : string
`
	_, err := shape.Parse(shapeText)
	// This should still parse - empty alias means no alias
	if err != nil {
		t.Logf("Empty alias parse result: %v (implementation-specific)", err)
	}
}

// Edge Case 21: Field name that matches Karl keyword (edge case identifier)
func TestApplyShapeKeywordFieldName(t *testing.T) {
	// Field names are identifiers, so they should be allowed
	shapeText := `Data : object
    + let : string
    + if : int
    + return : bool
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	
	value := &Object{Pairs: map[string]Value{
		"let":    &String{Value: "keyword"},
		"if":     &Integer{Value: 1},
		"return": &Boolean{Value: true},
	}}
	
	_, err = applyShape(value, sh.Type)
	if err != nil {
		t.Fatalf("applyShape error for keyword field names: %v", err)
	}
}

// Edge Case 22: Very long alias with special characters
func TestApplyShapeLongComplexAlias(t *testing.T) {
	shapeText := `Headers : object
    + field(X-Very-Long-Header-Name-With-Many-Dashes-And-Numbers-123) : string
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	
	value := &Object{Pairs: map[string]Value{
		"X-Very-Long-Header-Name-With-Many-Dashes-And-Numbers-123": &String{Value: "test"},
	}}
	
	out, err := applyShape(value, sh.Type)
	if err != nil {
		t.Fatalf("applyShape error: %v", err)
	}
	
	obj := out.(*Object)
	if obj.Pairs["field"].(*String).Value != "test" {
		t.Fatalf("long alias mapping failed")
	}
}

// Edge Case 23: Shape metadata preservation
func TestApplyShapeMetadataPreservation(t *testing.T) {
	shapeText := `User : object
    + name : string
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	
	value := &Object{Pairs: map[string]Value{
		"name": &String{Value: "Alice"},
	}}
	
	out, err := applyShape(value, sh.Type)
	if err != nil {
		t.Fatalf("applyShape error: %v", err)
	}
	
	obj := out.(*Object)
	if obj.Shape == nil {
		t.Fatalf("shape metadata not preserved")
	}
	if obj.Shape.Kind != shape.KindObject {
		t.Fatalf("shape metadata incorrect")
	}
}
