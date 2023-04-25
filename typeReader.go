package gexelizer

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

type TypeReader[T any] struct {
	file           ExcelFileReader
	typeInfo       typeInfo
	headers        []string
	nextRowToRead  uint
	options        *Options
	headersToIndex map[string]int
	rows           [][]string

	previousPrimaryKey string
}

func ReadExcel[T any](reader io.Reader, opts ...Options) ([]T, error) {
	r := &TypeReader[T]{}
	file, err := readExcel(reader)
	if err != nil {
		return nil, err
	}
	r.file = file
	if len(opts) > 0 {
		r.options = &opts[0]
	} else {
		r.options = DefaultOptions()
	}
	r.options.HeaderRow -= 1
	r.options.DataStartRow -= 1
	r.nextRowToRead = r.options.HeaderRow
	if err := r.analyzeType(); err != nil {
		return nil, err
	}
	return r.Read()
}
func ReadExcelFile[T any](filename string, opts ...Options) ([]T, error) {
	r := &TypeReader[T]{}
	file, err := readExcelFile(filename)
	if err != nil {
		return nil, err
	}
	r.file = file
	if len(opts) > 0 {
		r.options = &opts[0]
	} else {
		r.options = DefaultOptions()
	}
	r.options.HeaderRow -= 1
	r.options.DataStartRow -= 1
	r.nextRowToRead = r.options.HeaderRow
	if err := r.analyzeType(); err != nil {
		return nil, err
	}
	return r.Read()
}

// NewTypeReader creates a new TypeReader[T] instance
func NewTypeReader[T any](reader io.Reader, opts ...Options) (*TypeReader[T], error) {
	if reader == nil {
		return nil, fmt.Errorf("reader cannot be nil")
	}
	r := &TypeReader[T]{}
	file, err := readExcel(reader)
	if err != nil {
		return nil, err
	}
	r.file = file
	if len(opts) > 0 {
		r.options = &opts[0]
	} else {
		r.options = DefaultOptions()
	}
	r.nextRowToRead = r.options.HeaderRow
	if err := r.analyzeType(); err != nil {
		return nil, err
	}
	return r, nil
}

// Read reads the prepared excel file and returns a slice of T objects or an error
func (t *TypeReader[T]) Read() (result []T, err error) {
	//panic recover
	defer func() {
		if r := recover(); r != nil {
			result = nil
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	for i := 0; t.nextRowToRead < uint(len(t.rows)); i++ {
		row := t.rows[t.nextRowToRead]
		t.nextRowToRead++
		var toRead T
		pk, err := t.readSingle(row, &toRead)
		if err != nil {
			return nil, err
		}
		isFirstRow := i == 0
		// if the primary key is different than the previous one, append the object to the result
		if isFirstRow || !t.typeInfo.containsSlice() || t.previousPrimaryKey != pk {
			t.previousPrimaryKey = pk
			result = append(result, toRead)
			continue
		}
		// if the primary key is the same as the previous one, append the slice element to the previous object
		sliceIndexInT := t.typeInfo.sliceFieldInfo.index
		previous := &(result[len(result)-1])
		previousSlice := reflect.ValueOf(previous).Elem().FieldByIndex(sliceIndexInT)
		currentSlice := reflect.ValueOf(toRead).FieldByIndex(sliceIndexInT)
		previousSlice.Set(reflect.AppendSlice(previousSlice, currentSlice))
	}
	return result, nil
}

// ReadSingle reads a single row from the prepared excel file and returns the row parsed into T type object or an error
func (t *TypeReader[T]) readSingle(row []string, toRead *T) (string, error) {
	v := reflect.ValueOf(toRead).Elem()
	if !t.typeInfo.containsSlice() {
		return "", t.readSingleWithoutSlice(row, v)
	}
	// if the type contains a slice, we need to read the slice elements as well
	primaryKey := ""
	sliceFI := *t.typeInfo.sliceFieldInfo
	sliceFV := v.FieldByIndex(sliceFI.index)
	firstElem := reflect.New(sliceFV.Type().Elem()).Elem()
	elementSet := false
	for i := 0; i < len(t.typeInfo.orderedColumns); i++ {
		col := t.typeInfo.orderedColumns[i]
		fi := t.typeInfo.nameToField[col]
		if fi.kind == kindSlice {
			continue
		}
		if fi.isChildOf(sliceFI) {
			fieldToSet, err := firstElem.FieldByIndexErr(fi.index[len(sliceFI.index):])
			if err != nil {
				continue
			}
			isEmpty, err := t.setParsedValue(fieldToSet, col, fi, row)
			if err != nil {
				return "", err
			}
			if !isEmpty {
				elementSet = true
			}
		} else {
			toContinue, err := t.readNonSliceField(v, fi, col, row)
			if toContinue {
				continue
			}
			if err != nil {
				return "", err
			}
		}
		if fi.isPrimaryKey {
			primaryKey = strings.ToLower(strings.TrimSpace(row[t.headersToIndex[col]]))
		}
	}
	if elementSet {
		sliceFV.Set(reflect.Append(sliceFV, firstElem))
	}
	return primaryKey, nil
}

// setParsedValue sets the value of the field based on the column value
// if the field is optional and the value is empty or not present, it returns true
// if the field is required and the value is empty or not present, it returns an error
func (t *TypeReader[T]) setParsedValue(v reflect.Value, col string, info fieldInfo, row []string) (bool, error) {
	headerIndex, columnExists := t.headersToIndex[col]
	if !columnExists && info.defaultValue == "" {
		if !info.required && !info.isPrimaryKey {
			return true, nil
		}
		return true, fmt.Errorf("required column %s is not present", col)
	}
	var rowVal string
	if columnExists {
		rowVal = row[headerIndex]
	}
	if rowVal == "" && info.defaultValue != "" {
		rowVal = info.defaultValue
	}
	// check if the field is optional and the value is empty
	if rowVal == "" {
		if !info.required && !info.isPrimaryKey {
			return true, nil
		}
		//TODO here we have an issue, if struct is not present at all, required shouldn't be taken into consideration
		return true, fmt.Errorf("required column %s is empty", col)
	}
	if parsed, err := parseStringIntoType(rowVal, v.Type()); err != nil {
		return false, fmt.Errorf("error parsing cell value: %v", err)
	} else {
		//TODO wrapper types are not supported
		if v.Type() == reflect.TypeOf(Date("")) {
			parsed = Date(rowVal)
		}
		v.Set(reflect.ValueOf(parsed))
	}
	return false, nil
}

func (t *TypeReader[T]) analyzeType() (err error) {
	//panic recover
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error analyzing type: %v", r)
		}
	}()
	var toRead T
	info, err := analyzeType(reflect.TypeOf(toRead))
	if err != nil {
		return err
	}
	t.typeInfo = info
	t.rows, err = t.file.GetDefaultSheetRows()
	if err != nil {
		return err
	}
	if len(t.rows) < int(t.options.HeaderRow) {
		return fmt.Errorf("header row is out of bounds")
	}
	t.headers = make([]string, len(t.rows[t.options.HeaderRow]))
	for i, header := range t.rows[t.options.HeaderRow] {
		t.headers[i] = strings.TrimSpace(strings.ToLower(header))
	}
	t.headersToIndex = make(map[string]int, len(t.headers))
	for i, header := range t.headers {
		t.headersToIndex[header] = i
		//Add index for the primary column title in case this is an alias
		fi, exists := t.typeInfo.nameToField[header]
		if !exists {
			continue
		}
		lowerName := strings.ToLower(fi.name)
		t.headersToIndex[lowerName] = i
	}
	//Check if all required fields are present
	for _, col := range t.typeInfo.orderedColumns {
		fi := t.typeInfo.nameToField[col]
		if _, exists := t.headersToIndex[col]; fi.required && !exists {
			return fmt.Errorf("required field %s is missing", col)
		}
	}
	t.nextRowToRead = t.options.DataStartRow

	// normalize matrix width
	for i := range t.rows {
		if len(t.rows[i]) < len(t.headers) {
			t.rows[i] = append(t.rows[i], make([]string, len(t.headers)-len(t.rows[i]))...)
		}
	}
	return nil
}

func (t *TypeReader[T]) readSingleWithoutSlice(row []string, v reflect.Value) error {
	for i := 0; i < len(t.typeInfo.orderedColumns); i++ {
		col := t.typeInfo.orderedColumns[i]
		fi := t.typeInfo.nameToField[col]
		fv, err := fieldByIndexInit(v, fi.index)
		if err != nil {
			continue
		}
		if _, err = t.setParsedValue(fv, col, fi, row); err != nil {
			return err
		}
	}
	return nil
}

func (t *TypeReader[T]) readNonSliceField(v reflect.Value, fi fieldInfo, col string, row []string) (toContinue bool, err error) {
	var parent, grandParent reflect.Value
	parent, err = v.FieldByIndexErr(fi.index[:len(fi.index)-1])
	parentForced := false
	grandParentForced := false
	if err == nil {
		if parent.Kind() == reflect.Ptr && parent.Type().Elem().Kind() == reflect.Struct && parent.IsNil() {
			parentForced = true
			parent.Set(reflect.New(parent.Type().Elem()))
		}
	} else if err != nil && len(fi.index) > 2 {
		// try to find the grandparent
		grandParent, err = v.FieldByIndexErr(fi.index[:len(fi.index)-2])
		if err != nil {
			return true, err
		}
		if grandParent.Kind() == reflect.Ptr && grandParent.Type().Elem().Kind() == reflect.Struct && grandParent.IsNil() {
			grandParentForced = true
			grandParent.Set(reflect.New(grandParent.Type().Elem()))
		}
	} else if err != nil {
		return true, err
	}
	fieldToSet, err := v.FieldByIndexErr(fi.index)
	if err != nil {
		if parentForced {
			parent.Set(reflect.Zero(parent.Type()))
		}
		if grandParentForced {
			grandParent.Set(reflect.Zero(grandParent.Type()))
		}
		return true, err
	}
	isEmpty, err := t.setParsedValue(fieldToSet, col, fi, row)
	if isEmpty {
		if parentForced {
			parent.Set(reflect.Zero(parent.Type()))
		}
		if grandParentForced {
			grandParent.Set(reflect.Zero(grandParent.Type()))
		}
	}
	if err != nil {
		return false, err
	}
	return false, nil
}
func fieldByIndexInit(v reflect.Value, index []int) (reflect.Value, error) {
	if len(index) == 1 {
		return v.Field(index[0]), nil
	}

	if v.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("reflect: FieldByIndex of non-struct type " + v.Type().Name())
	}
	for i, x := range index {
		if i > 0 {
			if v.Kind() == reflect.Ptr && v.Type().Elem().Kind() == reflect.Struct {
				if v.IsNil() {
					//return reflect.Value{}, errors.New("reflect: indirection through nil pointer to embedded struct field " + v.Type().Elem().Name())
					//initiate the pointer
					v.Set(reflect.New(v.Type().Elem()))
				}
				v = v.Elem()
			}
		}
		v = v.Field(x)
	}
	return v, nil
}
