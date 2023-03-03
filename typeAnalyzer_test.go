package gexelizer

import (
	"fmt"
	"reflect"
	"testing"
)

func TestTypeAnalyzer_OneField(t *testing.T) {
	type oneField struct {
		One string
	}
	expected := typeInfo{
		isPtr:           false,
		primaryKeyIndex: -1,
		fields: []fieldInfo{
			{
				name:       "One",
				kind:       kindPrimitive,
				order:      0,
				fieldIndex: 0,
			},
		},
	}
	info, err := analyzeType(reflect.TypeOf(oneField{}))
	if err != nil {
		t.Fatal(err)
	}
	if err := typeInfosEqual(expected, info); err != nil {
		t.Error(err)
	}
}

func TestTypeAnalyzer_TwoFields(t *testing.T) {
	type twoFields struct {
		One string
		Two string
	}
	expected := typeInfo{
		isPtr:           false,
		primaryKeyIndex: -1,
		fields: []fieldInfo{
			{
				name:       "One",
				kind:       kindPrimitive,
				order:      0,
				fieldIndex: 0,
			},
			{
				name:       "Two",
				kind:       kindPrimitive,
				order:      1,
				fieldIndex: 1,
			},
		},
	}
	info, err := analyzeType(reflect.TypeOf(twoFields{}))
	if err != nil {
		t.Fatal(err)
	}
	if err := typeInfosEqual(expected, info); err != nil {
		t.Fatal(err)
	}
}

func TestTypeAnalyzer_UnexportedFields(t *testing.T) {
	type unexportedFields struct {
		one   string
		two   string
		Three string
	}
	expected := typeInfo{
		isPtr:           false,
		primaryKeyIndex: -1,
		fields: []fieldInfo{
			{
				name:       "Three",
				kind:       kindPrimitive,
				order:      0,
				fieldIndex: 2,
			},
		},
	}
	info, err := analyzeType(reflect.TypeOf(unexportedFields{}))
	if err != nil {
		t.Fatal(err)
	}
	if err := typeInfosEqual(expected, info); err != nil {
		t.Fatal(err)
	}
}

func TestTypeAnalyzer_PrimitiveFields(t *testing.T) {
	// Test all primitive types, all uints, ints, all floats, bool, string
	type primitiveFields struct {
		Bool    bool
		String  string
		Int     int
		Int8    int8
		Int16   int16
		Int32   int32
		Int64   int64
		Uint    uint
		Uint8   uint8
		Uint16  uint16
		Uint32  uint32
		Uint64  uint64
		Float32 float32
		Float64 float64
	}
	expected := typeInfo{
		isPtr:           false,
		primaryKeyIndex: -1,
		fields: []fieldInfo{
			{
				name:       "Bool",
				kind:       kindPrimitive,
				order:      0,
				fieldIndex: 0,
			},
			{
				name:       "String",
				kind:       kindPrimitive,
				order:      1,
				fieldIndex: 1,
			},
			{
				name:       "Int",
				kind:       kindPrimitive,
				order:      2,
				fieldIndex: 2,
			},
			{
				name:       "Int8",
				kind:       kindPrimitive,
				order:      3,
				fieldIndex: 3,
			},
			{
				name:       "Int16",
				kind:       kindPrimitive,
				order:      4,
				fieldIndex: 4,
			},
			{
				name:       "Int32",
				kind:       kindPrimitive,
				order:      5,
				fieldIndex: 5,
			},
			{
				name:       "Int64",
				kind:       kindPrimitive,
				order:      6,
				fieldIndex: 6,
			},
			{
				name:       "Uint",
				kind:       kindPrimitive,
				order:      7,
				fieldIndex: 7,
			},
			{
				name:       "Uint8",
				kind:       kindPrimitive,
				order:      8,
				fieldIndex: 8,
			},
			{
				name:       "Uint16",
				kind:       kindPrimitive,
				order:      9,
				fieldIndex: 9,
			},
			{
				name:       "Uint32",
				kind:       kindPrimitive,
				order:      10,
				fieldIndex: 10,
			},
			{
				name:       "Uint64",
				kind:       kindPrimitive,
				order:      11,
				fieldIndex: 11,
			},
			{
				name:       "Float32",
				kind:       kindPrimitive,
				order:      12,
				fieldIndex: 12,
			},
			{
				name:       "Float64",
				kind:       kindPrimitive,
				order:      13,
				fieldIndex: 13,
			},
		},
	}
	info, err := analyzeType(reflect.TypeOf(primitiveFields{}))
	if err != nil {
		t.Fatal(err)
	}
	if err := typeInfosEqual(expected, info); err != nil {
		t.Fatal(err)
	}
}

func TestTypeAnalyzer_TagName(t *testing.T) {
	type tagName struct {
		Two string `gex:"one"`
		One string `gex:"two,"`
	}
	expected := typeInfo{
		isPtr:           false,
		primaryKeyIndex: -1,
		fields: []fieldInfo{
			{
				name:       "one",
				kind:       kindPrimitive,
				order:      0,
				fieldIndex: 0,
			},
			{
				name:       "two",
				kind:       kindPrimitive,
				order:      1,
				fieldIndex: 1,
			},
		},
	}
	info, err := analyzeType(reflect.TypeOf(tagName{}))
	if err != nil {
		t.Fatal(err)
	}
	if err := typeInfosEqual(expected, info); err != nil {
		t.Fatal(err)
	}
}

func TestTypeAnalyzer_TagOrder(t *testing.T) {
	type tagOrder struct {
		Two string `gex:"two,order:1"`
		One string `gex:"one,order:0"`
	}
	expected := typeInfo{
		isPtr:           false,
		primaryKeyIndex: -1,
		fields: []fieldInfo{
			{
				name:       "one",
				kind:       kindPrimitive,
				order:      0,
				fieldIndex: 1,
			},
			{
				name:       "two",
				kind:       kindPrimitive,
				order:      1,
				fieldIndex: 0,
			},
		},
	}
	info, err := analyzeType(reflect.TypeOf(tagOrder{}))
	if err != nil {
		t.Fatal(err)
	}
	if err := typeInfosEqual(expected, info); err != nil {
		t.Fatal(err)
	}
}

func TestTypeAnalyzer_TagPrimaryKey(t *testing.T) {
	type tagPrimaryKey struct {
		Two string `gex:"two,primary"`
		One string `gex:"one"`
	}
	expected := typeInfo{
		isPtr:           false,
		primaryKeyIndex: 0,
		fields: []fieldInfo{
			{
				name:         "two",
				isPrimaryKey: true,
				kind:         kindPrimitive,
				order:        0,
				fieldIndex:   0,
			},
			{
				name:       "one",
				kind:       kindPrimitive,
				order:      1,
				fieldIndex: 1,
			},
		},
	}
	info, err := analyzeType(reflect.TypeOf(tagPrimaryKey{}))
	if err != nil {
		t.Fatal(err)
	}
	if err := typeInfosEqual(expected, info); err != nil {
		t.Fatal(err)
	}
}

func TestTypeAnalyzer_TagPrimaryKeyMultiple(t *testing.T) {
	type tagPrimaryKeyMultiple struct {
		Two string `gex:"two,primary"`
		One string `gex:"one,primary"`
	}
	_, err := analyzeType(reflect.TypeOf(tagPrimaryKeyMultiple{}))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTypeAnalyzer_TagIgnore(t *testing.T) {
	type tagIgnore struct {
		Two string `gex:"-"`
		One string `gex:"one"`
	}
	expected := typeInfo{
		isPtr:           false,
		primaryKeyIndex: -1,
		fields: []fieldInfo{
			{
				name:       "one",
				kind:       kindPrimitive,
				order:      0,
				fieldIndex: 1,
			},
		},
	}
	info, err := analyzeType(reflect.TypeOf(tagIgnore{}))
	if err != nil {
		t.Fatal(err)
	}
	if err := typeInfosEqual(expected, info); err != nil {
		t.Fatal(err)
	}
}

func TestTypeAnalyzer_EmbeddedStruct(t *testing.T) {
	//TODO issues with embedded structs and structs in general i suspect
	type es struct {
		One string `gex:"one"`
		Two string `gex:"two"`
	}
	type embeddedStruct struct {
		es
		Three string `gex:"three"`
		Four  string `gex:"four"`
	}
	expected := typeInfo{
		isPtr:           false,
		primaryKeyIndex: -1,
		fields: []fieldInfo{
			{
				name:       "one",
				kind:       kindPrimitive,
				order:      0,
				fieldIndex: 0,
			},
			{
				name:       "two",
				kind:       kindPrimitive,
				order:      1,
				fieldIndex: 1,
			},
			{
				name:       "three",
				kind:       kindPrimitive,
				order:      2,
				fieldIndex: 2,
			},
			{
				name:       "four",
				kind:       kindPrimitive,
				order:      3,
				fieldIndex: 3,
			},
		},
	}
	info, err := analyzeType(reflect.TypeOf(embeddedStruct{}))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", info)
	if err := typeInfosEqual(expected, info); err != nil {
		t.Fatal(err)
	}
}

// typeInfoEqual compares two typeInfo structs and returns nil if they are equal, error otherwise, with expected and got values
func typeInfosEqual(a, b typeInfo) error {
	if a.isPtr != b.isPtr {
		return fmt.Errorf("isPtr: expected %t, got %t", a.isPtr, b.isPtr)
	}
	if len(a.fields) != len(b.fields) {
		return fmt.Errorf("fields length: expected %d, got %d", len(a.fields), len(b.fields))
	}
	if a.primaryKeyIndex != b.primaryKeyIndex {
		return fmt.Errorf("primaryKeyIndex: expected %d, got %d", a.primaryKeyIndex, b.primaryKeyIndex)
	}
	for i := range a.fields {
		if a.fields[i].name != b.fields[i].name {
			return fmt.Errorf("fields[%d].name: expected %s, got %s", i, a.fields[i].name, b.fields[i].name)
		}
		if a.fields[i].kind != b.fields[i].kind {
			return fmt.Errorf("fields[%d].kind: expected %d, got %d", i, a.fields[i].kind, b.fields[i].kind)
		}
		if a.fields[i].isPrimaryKey != b.fields[i].isPrimaryKey {
			return fmt.Errorf("fields[%d].isPrimaryKey: expected %t, got %t", i, a.fields[i].isPrimaryKey, b.fields[i].isPrimaryKey)
		}
		if a.fields[i].order != b.fields[i].order {
			return fmt.Errorf("fields[%d].order: expected %d, got %d", i, a.fields[i].order, b.fields[i].order)
		}
		if a.fields[i].fieldIndex != b.fields[i].fieldIndex {
			return fmt.Errorf("fields[%d].fieldIndex: expected %d, got %d", i, a.fields[i].fieldIndex, b.fields[i].fieldIndex)
		}
		if a.fields[i].structInfo != nil && b.fields[i].structInfo != nil {
			if err := typeInfosEqual(*a.fields[i].structInfo, *b.fields[i].structInfo); err != nil {
				return fmt.Errorf("fields[%d].structInfo: %s", i, err)
			}
		}
		if a.fields[i].structInfo == nil && b.fields[i].structInfo != nil {
			return fmt.Errorf("fields[%d].structInfo: expected nil, got %+v", i, b.fields[i].structInfo)
		}
		if a.fields[i].structInfo != nil && b.fields[i].structInfo == nil {
			return fmt.Errorf("fields[%d].structInfo: expected %+v, got nil", i, a.fields[i].structInfo)
		}
	}
	return nil
}
