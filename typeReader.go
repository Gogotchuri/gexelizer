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
		for i := 0; i < len(t.typeInfo.orderedColumns); i++ {
			col := t.typeInfo.orderedColumns[i]
			fi := t.typeInfo.nameToField[col]
			fv, err := v.FieldByIndexErr(fi.index)
			if err != nil {
				continue
			}
			if err = t.setParsedValue(fv, col, fi, row); err != nil {
				return "", err
			}
		}
		return "", nil
	}
	// if the type contains a slice, we need to read the slice elements as well
	primaryKey := ""
	sliceFI := *t.typeInfo.sliceFieldInfo
	sliceFV := v.FieldByIndex(sliceFI.index)
	firstElem := reflect.New(sliceFV.Type().Elem()).Elem()
	for i := 0; i < len(t.typeInfo.orderedColumns); i++ {
		col := t.typeInfo.orderedColumns[i]
		fi := t.typeInfo.nameToField[col]
		if fi.kind == kindSlice {
			continue
		}
		var fieldToSet reflect.Value
		var err error
		if fi.isChildOf(sliceFI) {
			fieldToSet, err = firstElem.FieldByIndexErr(fi.index[len(sliceFI.index):])
		} else {
			fieldToSet, err = v.FieldByIndexErr(fi.index)
		}
		if err != nil {
			continue
		}
		if err := t.setParsedValue(fieldToSet, col, fi, row); err != nil {
			return "", err
		}
		if fi.isPrimaryKey {
			primaryKey = strings.ToLower(strings.TrimSpace(row[t.headersToIndex[col]]))
		}
	}
	sliceFV.Set(reflect.Append(sliceFV, firstElem))
	return primaryKey, nil
}

func (t *TypeReader[T]) setParsedValue(v reflect.Value, col string, info fieldInfo, row []string) error {
	headerIndex, columnExists := t.headersToIndex[col]
	if !columnExists && info.defaultValue == "" {
		if !info.required && !info.isPrimaryKey {
			return nil
		}
		return fmt.Errorf("required column %s is not present", col)
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
			return nil
		}
		//TODO here we have an issue, if struct is not present at all, required shouldn't be taken into consideration
		return fmt.Errorf("required column %s is empty", col)
	}
	if parsed, err := parseStringIntoType(rowVal, v.Type()); err != nil {
		return fmt.Errorf("error parsing cell value: %v", err)
	} else {
		v.Set(reflect.ValueOf(parsed))
	}
	return nil
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
