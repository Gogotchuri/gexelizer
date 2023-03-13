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

type fieldInfo struct {
	name         string
	order        int
	nextPrefix   string
	isPrimaryKey bool
	index        []int
	kind         kind
}

func (i fieldInfo) equal(b fieldInfo) bool {
	if len(i.index) != len(b.index) {
		return false
	}
	for k := 0; k < len(i.index) && k < len(b.index); k++ {
		if i.index[k] != b.index[k] {
			return false
		}
	}
	return i.name == b.name && i.order == b.order && i.nextPrefix == b.nextPrefix && i.isPrimaryKey == b.isPrimaryKey && i.kind == b.kind
}

type typeInfo struct {
	t              reflect.Type
	primaryKeyName string
	orderedColumns []string
	nameToField    map[string]fieldInfo
}

func (info typeInfo) sortColumns() {
	sort.Slice(info.orderedColumns, func(a, b int) bool {
		nameA := info.orderedColumns[a]
		nameB := info.orderedColumns[b]
		infoA := info.nameToField[nameA]
		infoB := info.nameToField[nameB]
		if infoA.isPrimaryKey && !infoB.isPrimaryKey {
			return true
		}
		if !infoA.isPrimaryKey && infoB.isPrimaryKey {
			return false
		}
		if len(infoA.index) > len(infoB.index) {
			return infoA.index[len(infoB.index)-1] < infoB.index[len(infoB.index)-1]
		}
		if len(infoA.index) < len(infoB.index) {
			return infoA.index[len(infoA.index)-1] < infoB.index[len(infoA.index)-1]
		}
		for k := 0; k < len(infoA.index); k++ {
			if k == len(infoA.index)-1 {
				return infoA.order < infoB.order
			}
			if infoA.index[k] != infoB.index[k] {
				return infoA.index[k] < infoB.index[k]
			}
		}
		return infoA.order < infoB.order
	})
	for i, name := range info.orderedColumns {
		fi := info.nameToField[name]
		fi.order = i
		fi.nextPrefix = ""
		info.nameToField[name] = fi
	}
}

func analyzeType(t reflect.Type) (typeInfo, error) {
	isStruct := t.Kind() == reflect.Struct
	isStructPtr := t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
	if !isStruct && !isStructPtr {
		return typeInfo{}, fmt.Errorf("unsupported type: %s", t.Kind())
	}
	if isStructPtr {
		t = t.Elem()
	}
	info, err := analyzeStruct(t)
	if err != nil {
		return typeInfo{}, err
	}
	return info, nil
}

type toTraverse struct {
	t            reflect.Type
	indexPrefix  []int
	columnPrefix string
}

func analyzeStruct(t reflect.Type) (typeInfo, error) {
	info := typeInfo{
		t:           t,
		nameToField: make(map[string]fieldInfo),
	}
	queue := []toTraverse{{t: t}}
	encounteredSlice := false
	for len(queue) > 0 {
		currentNode := queue[0]
		queue = queue[1:]
		currentType := currentNode.t
		for i := 0; i < currentType.NumField(); i++ {
			field := currentType.Field(i)
			//Skip unexported field
			if isUnexported(field) && !field.Anonymous {
				continue
			}
			//Skip ignored field
			if isFieldIgnored(field) {
				continue
			}
			fi, err := analyzeField(field, currentNode, i)
			if err != nil {
				return typeInfo{}, err
			}
			if fi.isPrimaryKey {
				if info.primaryKeyName != "" {
					return typeInfo{}, fmt.Errorf("multiple primary keys are not allowed")
				}
				info.primaryKeyName = fi.name
			}

			if fi.kind == kindSlice {
				if encounteredSlice {
					return typeInfo{}, fmt.Errorf("only one slice is allowed")
				}
				encounteredSlice = true
				t := field.Type.Elem()
				if t.Kind() == reflect.Ptr {
					t = t.Elem()
				}
				if t.Kind() == reflect.Struct {
					queue = append(queue, toTraverse{
						t:            t,
						indexPrefix:  fi.index,
						columnPrefix: fi.nextPrefix,
					})
				}
			} else if fi.kind == kindStruct || fi.kind == kindStructPtr {
				t := field.Type
				if fi.kind == kindStructPtr {
					t = t.Elem()
				}
				queue = append(queue, toTraverse{
					t:            t,
					indexPrefix:  fi.index,
					columnPrefix: fi.nextPrefix,
				})
				continue
			}
			if _, ok := info.nameToField[fi.name]; ok {
				return typeInfo{}, fmt.Errorf("duplicate field name: %s", fi.name)
			}
			info.nameToField[fi.name] = fi
			info.orderedColumns = append(info.orderedColumns, fi.name) //To be sorted later
		}
	}
	info.sortColumns()
	return info, nil
}

func analyzeField(field reflect.StructField, currentNode toTraverse, i int) (fieldInfo, error) {
	index := make([]int, 0, len(currentNode.indexPrefix)+len(field.Index))
	index = append(index, currentNode.indexPrefix...)
	index = append(index, field.Index...)
	// Get field order
	order := getFieldOrder(field, i)
	// Get field name
	name := getFieldName(field)
	// Get field kind
	typeKind, err := getKind(field.Type)
	if err != nil {
		return fieldInfo{}, err
	}
	isPrimaryKey := isPrimaryKey(field)
	if typeKind == kindSlice {
		// For slices, we only allow slices of structs
		if field.Type.Elem().Kind() != reflect.Struct {
			return fieldInfo{}, fmt.Errorf("unsupported slice type: %s", field.Type.Elem().Kind())
		}
	}
	// Get field prefix
	prefix := getNextFieldPrefix(field, name, currentNode.columnPrefix, typeKind)
	return fieldInfo{
		isPrimaryKey: isPrimaryKey,
		order:        order,
		name:         currentNode.columnPrefix + name,
		kind:         typeKind,
		index:        index,
		nextPrefix:   prefix, //For nested structs
	}, nil
}

func getNextFieldPrefix(field reflect.StructField, name, prevPrefix string, k kind) string {
	prefix := ""
	segments := strings.Split(field.Tag.Get(DefaultTag), ",")
	noPrefix := false
	for _, segment := range segments {
		if strings.TrimSpace(segment) == "noprefix" {
			noPrefix = true
		}
		if strings.HasPrefix(strings.TrimSpace(segment), "prefix:") {
			prefix = prevPrefix + strings.TrimSpace(strings.TrimPrefix(segment, "prefix:"))
		}
	}
	if noPrefix || (field.Anonymous && prefix == "") {
		return prevPrefix
	}
	if k != kindPrimitive && k != kindPrimitivePtr {
		if prefix == "" {
			prefix = prevPrefix + name + "."
		}
		return prefix
	}
	return prevPrefix
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
		order, err := strconv.Atoi(strings.TrimPrefix(o, "order:"))
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
