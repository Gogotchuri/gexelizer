package gexelizer

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

const DefaultTag = "gex"

type kind int

const (
	kindPrimitive kind = iota
	kindSlice
	kindStruct
	kindPrimitivePtr
	kindStructPtr
)

type typeAnalyzer struct {
	// expandedFieldIndex current expanded field index
	currentExpandedFieldIndex int
}

type fieldInfo struct {
	isPrimaryKey       bool
	order              int
	fieldIndex         int
	expandedFieldIndex int
	name               string
	kind               kind
	structInfo         *typeInfo
}

type typeInfo struct {
	primaryKeyIndex int
	isPtr           bool
	fields          []fieldInfo
	namesToIndex    map[string]int
}

func analyzeType(t reflect.Type) (typeInfo, error) {
	ta := typeAnalyzer{}
	return ta.analyzeType(t)
}

func (ta typeAnalyzer) analyzeType(t reflect.Type) (typeInfo, error) {
	isStruct := t.Kind() == reflect.Struct
	isStructPtr := t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
	if !isStruct && !isStructPtr {
		return typeInfo{}, fmt.Errorf("unsupported type: %s", t.Kind())
	}
	info, err := ta.analyzeStruct(t, 0)
	if err != nil {
		return typeInfo{}, err
	}
	return info, nil
}

func (ta typeAnalyzer) analyzeStruct(t reflect.Type, depth int) (typeInfo, error) {
	if err := isAllowed(t, depth); err != nil {
		return typeInfo{}, err
	}
	info := typeInfo{
		primaryKeyIndex: -1,
		isPtr:           t.Kind() == reflect.Ptr,
		namesToIndex:    make(map[string]int),
	}
	if t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	exportedIndex := 0
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		//Expand embedded anonymous struct and compositions
		if field.Anonymous {
			//TODO issues with embedded structs, should add later
			//embeddedInfo, err := analyzeStruct(field.Type, depth) //we don't need to increase depth here, because we embed the struct fields
			//if err != nil {
			//	return typeInfo{}, err
			//}
			//info.fields = append(info.fields, embeddedInfo.fields...)
			//if info.primaryKeyIndex == -1 {
			//	info.primaryKeyIndex = embeddedInfo.primaryKeyIndex
			//}
			////Merge namesToIndex
			//for k, v := range embeddedInfo.namesToIndex {
			//	info.namesToIndex[k] = v
			//}
			//exportedIndex += len(embeddedInfo.fields)
			continue
		}
		//Skip unexported field
		if isUnexported(field) {
			continue
		}
		//Skip ignored field
		if isFieldIgnored(field) {
			continue
		}
		fi, err := ta.analyzeField(field, i, exportedIndex, depth)
		if err != nil {
			return typeInfo{}, err
		}
		info.fields = append(info.fields, fi)
		info.namesToIndex[fi.name] = fi.fieldIndex
		if fi.isPrimaryKey {
			if info.primaryKeyIndex != -1 {
				return typeInfo{}, fmt.Errorf("multiple primary keys are not allowed")
			}
			info.primaryKeyIndex = fi.fieldIndex
		}
		exportedIndex++
	}
	sort.Slice(info.fields, func(i, j int) bool {
		return info.fields[i].order < info.fields[j].order
	})
	return info, nil
}

func (ta typeAnalyzer) analyzeField(field reflect.StructField, fieldIndex, exportedIndex, depth int) (fieldInfo, error) {
	// Get field order
	order := getFieldOrder(field, exportedIndex)
	// Get field name
	name := getFieldName(field)
	// Get field kind
	kind, err := getKind(field.Type)
	if err != nil {
		return fieldInfo{}, err
	}
	isPrimaryKey := isPrimaryKey(field)
	// if kind is struct, analyze it
	var structInfo *typeInfo
	if kind == kindStruct || kind == kindStructPtr || kind == kindSlice {
		if kind == kindSlice {
			// For slices, we only allow slices of structs
			if field.Type.Elem().Kind() != reflect.Struct {
				return fieldInfo{}, fmt.Errorf("unsupported slice type: %s", field.Type.Elem().Kind())
			}
		}
		s, err := ta.analyzeStruct(field.Type, depth+1)
		if err != nil {
			return fieldInfo{}, err
		}
		structInfo = &s
	}
	// Increment currentExpandedFieldIndex
	ta.currentExpandedFieldIndex++
	return fieldInfo{
		isPrimaryKey:       isPrimaryKey,
		order:              order,
		expandedFieldIndex: ta.currentExpandedFieldIndex,
		fieldIndex:         fieldIndex,
		name:               name,
		kind:               kind,
		structInfo:         structInfo,
	}, nil
}

func isPrimaryKey(field reflect.StructField) bool {
	segments := strings.Split(field.Tag.Get(DefaultTag), ",")
	for _, segment := range segments {
		if strings.TrimSpace(segment) == "primary" {
			return true
		}
	}
	return false
}

func getKind(t reflect.Type) (kind, error) {
	switch t.Kind() {
	case reflect.Struct:
		return kindStruct, nil
	case reflect.Slice:
		return kindSlice, nil
	case reflect.Ptr:
		switch t.Elem().Kind() {
		case reflect.Ptr, reflect.Slice:
			return -1, fmt.Errorf("pointer to pointer is not supported")
		default:
			return getPointedKind(t.Elem().Kind())
		}
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32,
		reflect.Float64, reflect.String:
		return kindPrimitive, nil
	default:
		return -1, fmt.Errorf("unsupported type: %s", t.Kind())
	}
}

func getPointedKind(v reflect.Kind) (kind, error) {
	switch v {
	case reflect.Struct:
		return kindStructPtr, nil
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32,
		reflect.Float64, reflect.String:
		return kindPrimitivePtr, nil
	default:
		return -1, fmt.Errorf("unsupported type: %s", v)
	}
}

func getFieldOrder(field reflect.StructField, i int) int {
	tag := field.Tag.Get(DefaultTag)
	if tag == "" {
		return i
	}
	segments := strings.Split(tag, ",")
	for _, o := range segments {
		if !strings.HasPrefix(o, "order:") {
			continue
		}
		order, err := strconv.Atoi(o[6:])
		if err != nil {
			return i
		}
		return order
	}
	return i
}

func getFieldName(field reflect.StructField) string {
	// Tag has values separated by comma, first value is always the name, so we split it
	tagValue := field.Tag.Get(DefaultTag)
	if tagValue == "" {
		return field.Name
	}
	name := strings.Split(tagValue, ",")[0]
	if name == "" {
		name = field.Name
	}
	return name
}

func isFieldIgnored(field reflect.StructField) bool {
	return field.Tag.Get(DefaultTag) == "-"
}

func isUnexported(field reflect.StructField) bool {
	return field.PkgPath != ""
}

func isPrimitive(k reflect.Kind) bool {
	return k == reflect.Bool || k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 || k == reflect.Int32 || k == reflect.Int64 ||
		k == reflect.Uint || k == reflect.Uint8 || k == reflect.Uint16 || k == reflect.Uint32 || k == reflect.Uint64 || k == reflect.Float32 ||
		k == reflect.Float64 || k == reflect.String
}

func isAllowed(t reflect.Type, depth int) error {
	// Max depth is 2, because we don't support deep struct composition, which cant be represented in excel
	if depth > 1 {
		return fmt.Errorf("unsupported struct composition, depth: %d", depth)
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		// We don't support pointer to slice or pointer to pointer
		if t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
			return fmt.Errorf("pointer to the type is not allowed: %s", t.Kind())
		}
	}
	// For depth 0, we only allow structs or pointers to structs
	if depth == 0 && t.Kind() != reflect.Struct {
		return fmt.Errorf("unsupported type: %s", t.Kind())
	}
	// For depth 1, we allow structs, pointers to structs + slices and primitives
	if depth == 1 && t.Kind() != reflect.Struct && t.Kind() != reflect.Slice && isPrimitive(t.Kind()) {
		return fmt.Errorf("unsupported type: %s", t.Kind())
	}
	return nil
}
