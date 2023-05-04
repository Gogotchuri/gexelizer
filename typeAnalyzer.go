package gexelizer

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type fieldInfo struct {
	name         string
	aliases      []string
	order        int
	nextPrefix   string
	isPrimaryKey bool
	index        []int
	kind         kind
	required     bool
	omitEmpty    bool
	defaultValue string
}

func (i fieldInfo) isChildOf(b fieldInfo) bool {
	if len(i.index) <= len(b.index) {
		return false
	}
	for k := 0; k < len(b.index); k++ {
		if i.index[k] != b.index[k] {
			return false
		}
	}
	return true
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
	sliceFieldInfo *fieldInfo
}

func (info typeInfo) containsSlice() bool {
	return info.sliceFieldInfo != nil
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
		fi := info.nameToField[strings.ToLower(name)]
		fi.order = i
		fi.nextPrefix = ""
		info.nameToField[strings.ToLower(name)] = fi
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
				info.primaryKeyName = strings.ToLower(fi.name)
			}

			if fi.kind == kindSlice {
				if encounteredSlice {
					return typeInfo{}, fmt.Errorf("only one slice is allowed")
				}
				encounteredSlice = true
				info.sliceFieldInfo = &fi
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
				continue
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
			lowerName := strings.TrimSpace(strings.ToLower(fi.name))
			if existingFI, ok := info.nameToField[lowerName]; ok {
				//If the new field is shallower in the struct, need to add it
				if len(existingFI.index) > len(fi.index) {
					info.nameToField[lowerName] = fi
					info.orderedColumns = append(info.orderedColumns, lowerName) //To be sorted later
				}
				//If the new field is at the same level, we error out
				if len(existingFI.index) == len(fi.index) {
					return typeInfo{}, fmt.Errorf("duplicate field name: %s", fi.name)
				}
				//If the new field is in the deeper levels of the struct, we just ignore it
			} else {
				info.nameToField[lowerName] = fi
				info.orderedColumns = append(info.orderedColumns, lowerName) //To be sorted later
			}

			//Go over and assign for aliases
			for _, alias := range fi.aliases {
				lowerName := strings.TrimSpace(strings.ToLower(alias))
				if existingFI, ok := info.nameToField[lowerName]; ok {
					//If the new field is shallower in the struct, we need to add it
					if len(existingFI.index) > len(fi.index) {
						info.nameToField[lowerName] = fi
					}
					//If the new field is at the same level, we error out
					if len(existingFI.index) == len(fi.index) {
						return typeInfo{}, fmt.Errorf("duplicate field name through alias: %s", lowerName)
					}
					//If the new field is in the upper levels of the struct, we just overwrite it
				} else {
					info.nameToField[lowerName] = fi
				}
			}
		}
	}
	if info.primaryKeyName == "" && encounteredSlice {
		return typeInfo{}, fmt.Errorf("primary key is required when a slice is present")
	}
	info.sortColumns()
	return info, nil
}

func analyzeField(field reflect.StructField, currentNode toTraverse, i int) (fieldInfo, error) {
	index := make([]int, 0, len(currentNode.indexPrefix)+len(field.Index))
	index = append(index, currentNode.indexPrefix...)
	index = append(index, field.Index...)
	tagOpts := parseTagOptions(field, i)
	// Get field kind
	typeKind, err := getKind(field.Type)
	if err != nil {
		return fieldInfo{}, err
	}
	if typeKind == kindSlice {
		// For slices, we only allow slices of structs
		if field.Type.Elem().Kind() != reflect.Struct {
			return fieldInfo{}, fmt.Errorf("unsupported slice type: %s", field.Type.Elem().Kind())
		}
	}
	// Get field prefix
	prefix := getNextFieldPrefix(field, tagOpts.column, currentNode.columnPrefix, typeKind)
	for i, alias := range tagOpts.aliases {
		tagOpts.aliases[i] = currentNode.columnPrefix + alias
	}
	return fieldInfo{
		isPrimaryKey: tagOpts.primaryKey,
		order:        tagOpts.order,
		omitEmpty:    tagOpts.omitEmpty,
		aliases:      tagOpts.aliases,
		name:         currentNode.columnPrefix + tagOpts.column,
		kind:         typeKind,
		index:        index,
		nextPrefix:   prefix, //For nested structs
		required:     tagOpts.required,
		defaultValue: tagOpts.defaultValue,
	}, nil
}

type tagOptions struct {
	column       string
	defaultValue string
	order        int
	primaryKey   bool
	required     bool
	omitEmpty    bool
	aliases      []string
}

func parseTagOptions(field reflect.StructField, i int) tagOptions {
	tag := field.Tag.Get(mainTag)
	segments := strings.Split(tag, mainSeparator)
	options := tagOptions{}
	for _, o := range segments {
		//Column aliases
		if strings.HasPrefix(o, aliasesTag) {
			alias := strings.TrimPrefix(o, aliasesTag)
			if alias != "" {
				aliasesList := strings.Split(alias, listSeparator)
				options.aliases = append(options.aliases, aliasesList...)
			}
			continue
		}

		//Column primary name
		if strings.HasPrefix(o, columnTag) {
			col := strings.TrimPrefix(o, columnTag)
			if col != "" {
				options.column = col
			}
			continue
		}
		//Order
		if strings.HasPrefix(o, orderTag) {
			order, err := strconv.Atoi(strings.TrimPrefix(o, orderTag))
			if err != nil {
				options.order = i
			}
			options.order = order
			continue
		}
		//Default
		if strings.HasPrefix(o, defaultTag) {
			options.defaultValue = strings.TrimPrefix(o, defaultTag)
			continue
		}
		//Required
		if strings.TrimSpace(o) == requiredTag {
			options.required = true
			continue
		}
		//OmitEmpty
		if strings.TrimSpace(o) == omitEmptyTag {
			options.omitEmpty = true
			continue
		}
		//Primary Key
		if strings.TrimSpace(o) == primaryKeyTag {
			options.primaryKey = true
			continue
		}
	}
	if options.column == "" {
		options.column = field.Name
	}
	return options
}

func getNextFieldPrefix(field reflect.StructField, name, prevPrefix string, k kind) string {
	prefix := ""
	segments := strings.Split(field.Tag.Get(mainTag), mainSeparator)
	noPrefix := false
	for _, segment := range segments {
		if strings.TrimSpace(segment) == noprefixTag {
			noPrefix = true
		}
		if strings.HasPrefix(strings.TrimSpace(segment), prefixTag) {
			prefix = prevPrefix + strings.TrimPrefix(segment, prefixTag)
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

func isFieldIgnored(field reflect.StructField) bool {
	return field.Tag.Get(mainTag) == ignoreTag
}

func isUnexported(field reflect.StructField) bool {
	return field.PkgPath != ""
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
