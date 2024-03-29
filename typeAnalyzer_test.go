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
		t:              reflect.TypeOf(oneField{}),
		primaryKeyName: "",
		orderedColumns: []string{"one"},
		nameToField:    map[string]fieldInfo{"one": {name: "One", order: 0, isPrimaryKey: false, index: []int{0}, kind: kindPrimitive}},
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
		t:              reflect.TypeOf(twoFields{}),
		primaryKeyName: "",
		orderedColumns: []string{"one", "two"},
		nameToField: map[string]fieldInfo{
			"one": {name: "One", order: 0, isPrimaryKey: false, index: []int{0}, kind: kindPrimitive},
			"two": {name: "Two", order: 1, isPrimaryKey: false, index: []int{1}, kind: kindPrimitive},
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
		t:              reflect.TypeOf(unexportedFields{}),
		primaryKeyName: "",
		orderedColumns: []string{"three"},
		nameToField:    map[string]fieldInfo{"three": {name: "Three", order: 0, isPrimaryKey: false, index: []int{2}, kind: kindPrimitive}},
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
		t:              reflect.TypeOf(primitiveFields{}),
		primaryKeyName: "",
		orderedColumns: []string{"bool", "string", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64"},
		nameToField: map[string]fieldInfo{
			"bool":    {name: "Bool", order: 0, isPrimaryKey: false, index: []int{0}, kind: kindPrimitive},
			"string":  {name: "String", order: 1, isPrimaryKey: false, index: []int{1}, kind: kindPrimitive},
			"int":     {name: "Int", order: 2, isPrimaryKey: false, index: []int{2}, kind: kindPrimitive},
			"int8":    {name: "Int8", order: 3, isPrimaryKey: false, index: []int{3}, kind: kindPrimitive},
			"int16":   {name: "Int16", order: 4, isPrimaryKey: false, index: []int{4}, kind: kindPrimitive},
			"int32":   {name: "Int32", order: 5, isPrimaryKey: false, index: []int{5}, kind: kindPrimitive},
			"int64":   {name: "Int64", order: 6, isPrimaryKey: false, index: []int{6}, kind: kindPrimitive},
			"uint":    {name: "Uint", order: 7, isPrimaryKey: false, index: []int{7}, kind: kindPrimitive},
			"uint8":   {name: "Uint8", order: 8, isPrimaryKey: false, index: []int{8}, kind: kindPrimitive},
			"uint16":  {name: "Uint16", order: 9, isPrimaryKey: false, index: []int{9}, kind: kindPrimitive},
			"uint32":  {name: "Uint32", order: 10, isPrimaryKey: false, index: []int{10}, kind: kindPrimitive},
			"uint64":  {name: "Uint64", order: 11, isPrimaryKey: false, index: []int{11}, kind: kindPrimitive},
			"float32": {name: "Float32", order: 12, isPrimaryKey: false, index: []int{12}, kind: kindPrimitive},
			"float64": {name: "Float64", order: 13, isPrimaryKey: false, index: []int{13}, kind: kindPrimitive},
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
		Two string `gex:"column:one"`
		One string `gex:"column:two,"`
	}
	expected := typeInfo{
		t:              reflect.TypeOf(tagName{}),
		primaryKeyName: "",
		orderedColumns: []string{"one", "two"},
		nameToField: map[string]fieldInfo{
			"one": {name: "one", order: 0, isPrimaryKey: false, index: []int{0}, kind: kindPrimitive},
			"two": {name: "two", order: 1, isPrimaryKey: false, index: []int{1}, kind: kindPrimitive},
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
		Two string `gex:"column:two,order:1"`
		One string `gex:"column:one,order:0"`
	}
	expected := typeInfo{
		t:              reflect.TypeOf(tagOrder{}),
		primaryKeyName: "",
		orderedColumns: []string{"one", "two"},
		nameToField: map[string]fieldInfo{
			"one": {name: "one", order: 0, isPrimaryKey: false, index: []int{1}, kind: kindPrimitive},
			"two": {name: "two", order: 1, isPrimaryKey: false, index: []int{0}, kind: kindPrimitive},
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
		Two string `gex:"column:two,primary"`
		One string `gex:"column:one"`
	}
	expected := typeInfo{
		t:              reflect.TypeOf(tagPrimaryKey{}),
		primaryKeyName: "two",
		orderedColumns: []string{"two", "one"},
		nameToField: map[string]fieldInfo{
			"two": {name: "two", order: 0, isPrimaryKey: true, index: []int{0}, kind: kindPrimitive},
			"one": {name: "one", order: 1, isPrimaryKey: false, index: []int{1}, kind: kindPrimitive},
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
		Two string `gex:"column:two,primary"`
		One string `gex:"column:one,primary"`
	}
	_, err := analyzeType(reflect.TypeOf(tagPrimaryKeyMultiple{}))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTypeAnalyzer_TagIgnore(t *testing.T) {
	type tagIgnore struct {
		Two string `gex:"-"`
		One string `gex:"column:one"`
	}
	expected := typeInfo{
		t:              reflect.TypeOf(tagIgnore{}),
		primaryKeyName: "",
		orderedColumns: []string{"one"},
		nameToField: map[string]fieldInfo{
			"one": {name: "one", order: 0, isPrimaryKey: false, index: []int{1}, kind: kindPrimitive},
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
	type es struct {
		One string `gex:"column:one"`
		Two string `gex:"column:two"`
	}
	type embeddedStruct struct {
		es
		Three string `gex:"column:three"`
		Four  string `gex:"column:four"`
	}
	expected := typeInfo{
		t:              reflect.TypeOf(embeddedStruct{}),
		primaryKeyName: "",
		orderedColumns: []string{"one", "two", "three", "four"},
		nameToField: map[string]fieldInfo{
			"one":   {name: "one", order: 0, isPrimaryKey: false, index: []int{0, 0}, kind: kindPrimitive},
			"two":   {name: "two", order: 1, isPrimaryKey: false, index: []int{0, 1}, kind: kindPrimitive},
			"three": {name: "three", order: 2, isPrimaryKey: false, index: []int{1}, kind: kindPrimitive},
			"four":  {name: "four", order: 3, isPrimaryKey: false, index: []int{2}, kind: kindPrimitive},
		},
	}
	info, err := analyzeType(reflect.TypeOf(embeddedStruct{}))
	if err != nil {
		t.Fatal(err)
	}
	if err := typeInfosEqual(expected, info); err != nil {
		t.Fatal(err)
	}
}
func TestTypeAnalyzer_PrefixedEmbeddedStruct(t *testing.T) {
	type es struct {
		One string `gex:"column:one"`
		Two string `gex:"column:two"`
	}
	type embeddedStruct struct {
		es    `gex:"prefix:es_"`
		Three string `gex:"column:three"`
		Four  string `gex:"column:four"`
	}
	expected := typeInfo{
		t:              reflect.TypeOf(embeddedStruct{}),
		primaryKeyName: "",
		orderedColumns: []string{"es_one", "es_two", "three", "four"},
		nameToField: map[string]fieldInfo{
			"es_one": {name: "es_one", order: 0, isPrimaryKey: false, index: []int{0, 0}, kind: kindPrimitive},
			"es_two": {name: "es_two", order: 1, isPrimaryKey: false, index: []int{0, 1}, kind: kindPrimitive},
			"three":  {name: "three", order: 2, isPrimaryKey: false, index: []int{1}, kind: kindPrimitive},
			"four":   {name: "four", order: 3, isPrimaryKey: false, index: []int{2}, kind: kindPrimitive},
		},
	}
	info, err := analyzeType(reflect.TypeOf(embeddedStruct{}))
	if err != nil {
		t.Fatal(err)
	}
	if err := typeInfosEqual(expected, info); err != nil {
		t.Fatal(err)
	}
}

func TestTypeAnalyzer_StructField(t *testing.T) {
	type sf struct {
		One string `gex:"column:one"`
		Two string `gex:"column:two"`
	}
	type structField struct {
		Sf    sf
		Three string `gex:"column:three"`
		Four  string `gex:"column:four"`
	}
	expected := typeInfo{
		t:              reflect.TypeOf(structField{}),
		primaryKeyName: "",
		orderedColumns: []string{"sf.one", "sf.two", "three", "four"},
		nameToField: map[string]fieldInfo{
			"sf.one": {name: "Sf.one", order: 0, isPrimaryKey: false, index: []int{0, 0}, kind: kindPrimitive},
			"sf.two": {name: "Sf.two", order: 1, isPrimaryKey: false, index: []int{0, 1}, kind: kindPrimitive},
			"three":  {name: "three", order: 2, isPrimaryKey: false, index: []int{1}, kind: kindPrimitive},
			"four":   {name: "four", order: 3, isPrimaryKey: false, index: []int{2}, kind: kindPrimitive},
		},
	}
	info, err := analyzeType(reflect.TypeOf(structField{}))
	if err != nil {
		t.Fatal(err)
	}
	if err := typeInfosEqual(expected, info); err != nil {
		t.Fatal(err)
	}
}
func TestTypeAnalyzer_NamedStructField(t *testing.T) {
	type sf struct {
		One string `gex:"column:one"`
		Two string `gex:"column:two"`
	}
	type structField struct {
		Sf    sf     `gex:"column:f"`
		Three string `gex:"column:three"`
		Four  string `gex:"column:four"`
	}
	expected := typeInfo{
		t:              reflect.TypeOf(structField{}),
		primaryKeyName: "",
		orderedColumns: []string{"f.one", "f.two", "three", "four"},
		nameToField: map[string]fieldInfo{
			"f.one": {name: "f.one", order: 0, isPrimaryKey: false, index: []int{0, 0}, kind: kindPrimitive},
			"f.two": {name: "f.two", order: 1, isPrimaryKey: false, index: []int{0, 1}, kind: kindPrimitive},
			"three": {name: "three", order: 2, isPrimaryKey: false, index: []int{1}, kind: kindPrimitive},
			"four":  {name: "four", order: 3, isPrimaryKey: false, index: []int{2}, kind: kindPrimitive},
		},
	}
	info, err := analyzeType(reflect.TypeOf(structField{}))
	if err != nil {
		t.Fatal(err)
	}
	if err := typeInfosEqual(expected, info); err != nil {
		t.Fatal(err)
	}
}

func TestTypeAnalyzer_StructFieldPrefixes(t *testing.T) {
	type sf struct {
		One string `gex:"column:one"`
		Two string `gex:"column:two"`
	}
	type structField struct {
		Sf    sf     `gex:"column:f,prefix:foo."`
		Sf1   sf     `gex:"column:sf,noprefix"`
		Three string `gex:"column:three"`
		Four  string `gex:"column:four"`
	}
	expected := typeInfo{
		t:              reflect.TypeOf(structField{}),
		primaryKeyName: "",
		orderedColumns: []string{"foo.one", "foo.two", "one", "two", "three", "four"},
		nameToField: map[string]fieldInfo{
			"foo.one": {name: "foo.one", order: 0, isPrimaryKey: false, index: []int{0, 0}, kind: kindPrimitive},
			"foo.two": {name: "foo.two", order: 1, isPrimaryKey: false, index: []int{0, 1}, kind: kindPrimitive},
			"one":     {name: "one", order: 2, isPrimaryKey: false, index: []int{1, 0}, kind: kindPrimitive},
			"two":     {name: "two", order: 3, isPrimaryKey: false, index: []int{1, 1}, kind: kindPrimitive},
			"three":   {name: "three", order: 4, isPrimaryKey: false, index: []int{2}, kind: kindPrimitive},
			"four":    {name: "four", order: 5, isPrimaryKey: false, index: []int{3}, kind: kindPrimitive},
		},
	}
	info, err := analyzeType(reflect.TypeOf(structField{}))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", info.orderedColumns)
	if err := typeInfosEqual(expected, info); err != nil {
		t.Fatal(err)
	}
}

func TestTypeAnalyzer_StructFieldPtr(t *testing.T) {
	type sf struct {
		One string `gex:"column:one,primary"`
		Two string `gex:"column:two"`
	}
	type structFieldPtr struct {
		Sf    *sf
		Three string `gex:"column:three"`
		Four  string `gex:"column:four"`
	}
	expected := typeInfo{
		t:              reflect.TypeOf(structFieldPtr{}),
		primaryKeyName: "sf.one",
		orderedColumns: []string{"sf.one", "sf.two", "three", "four"},
		nameToField: map[string]fieldInfo{
			"sf.one": {name: "Sf.one", order: 0, isPrimaryKey: true, index: []int{0, 0}, kind: kindPrimitive},
			"sf.two": {name: "Sf.two", order: 1, isPrimaryKey: false, index: []int{0, 1}, kind: kindPrimitive},
			"three":  {name: "three", order: 2, isPrimaryKey: false, index: []int{1}, kind: kindPrimitive},
			"four":   {name: "four", order: 3, isPrimaryKey: false, index: []int{2}, kind: kindPrimitive},
		},
	}
	info, err := analyzeType(reflect.TypeOf(structFieldPtr{}))
	if err != nil {
		t.Fatal(err)
	}
	if err := typeInfosEqual(expected, info); err != nil {
		t.Fatal(err)
	}
}

func TestTypeAnalyzer_StructFieldOrderTag(t *testing.T) {
	type sf struct {
		Two string `gex:"column:two,order:1"`
		One string `gex:"column:one,order:0"`
	}
	type structFieldOrderTag struct {
		Sf    sf     `gex:"column:sf,order:0"`
		Three string `gex:"column:three"`
		Four  string `gex:"column:four"`
	}
	expected := typeInfo{
		t:              reflect.TypeOf(structFieldOrderTag{}),
		primaryKeyName: "",
		orderedColumns: []string{"sf.one", "sf.two", "three", "four"},
		nameToField: map[string]fieldInfo{
			"sf.one": {name: "sf.one", order: 0, isPrimaryKey: false, index: []int{0, 1}, kind: kindPrimitive},
			"sf.two": {name: "sf.two", order: 1, isPrimaryKey: false, index: []int{0, 0}, kind: kindPrimitive},
			"three":  {name: "three", order: 2, isPrimaryKey: false, index: []int{1}, kind: kindPrimitive},
			"four":   {name: "four", order: 3, isPrimaryKey: false, index: []int{2}, kind: kindPrimitive},
		},
	}
	info, err := analyzeType(reflect.TypeOf(structFieldOrderTag{}))
	if err != nil {
		t.Fatal(err)
	}
	if err := typeInfosEqual(expected, info); err != nil {
		t.Fatal(err)
	}
}

func TestTypeAnalyzer_PrimitiveSlice(t *testing.T) {
	type primitiveSlice struct {
		Slice []string `gex:""`
	}
	_, err := analyzeType(reflect.TypeOf(primitiveSlice{}))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTypeAnalyzer_SliceStruct(t *testing.T) {
	type sliceStruct struct {
		One string `gex:"column:one"`
	}
	type sliceStructSlice struct {
		ID    int64         `gex:"column:id,primary"`
		Slice []sliceStruct `gex:""`
	}
	expected := typeInfo{
		t:              reflect.TypeOf(sliceStructSlice{}),
		primaryKeyName: "id",
		orderedColumns: []string{"id", "slice.one"},
		nameToField: map[string]fieldInfo{
			"id":        {name: "id", order: 0, isPrimaryKey: true, index: []int{0}, kind: kindPrimitive},
			"slice.one": {name: "Slice.one", order: 1, isPrimaryKey: false, index: []int{1, 0}, kind: kindPrimitive},
		},
	}
	info, err := analyzeType(reflect.TypeOf(sliceStructSlice{}))
	if err != nil {
		t.Fatal(err)
	}
	if err := typeInfosEqual(expected, info); err != nil {
		t.Fatal(err)
	}
}

// typeInfoEqual compares two typeInfo structs and returns nil if they are equal, error otherwise, with expected and got values
func typeInfosEqual(a, b typeInfo) error {
	if len(a.nameToField) != len(b.nameToField) {
		return fmt.Errorf("fields length: expected %d, got %d", len(a.nameToField), len(b.nameToField))
	}
	if a.primaryKeyName != b.primaryKeyName {
		return fmt.Errorf("primaryKeyIndex: expected %s, got %s", a.primaryKeyName, b.primaryKeyName)
	}
	for i, c := range a.orderedColumns {
		if b.orderedColumns[i] != c {
			return fmt.Errorf("orderedColumns: expected %s, got %s", c, b.orderedColumns[i])
		}
	}
	for k, v := range a.nameToField {
		if !b.nameToField[k].equal(v) {
			return fmt.Errorf("nameToField: expected %+v, got %+v", v, b.nameToField[k])
		}
	}
	return nil
}
