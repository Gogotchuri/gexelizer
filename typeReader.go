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

// ReadExcelReader reads an excel file and returns a parsed slice of T objects or an error
func ReadExcelReader[T any](reader io.Reader, opts ...Options) ([]T, error) {
	r, err := NewTypeReader[T](reader, opts...)
	if err != nil {
		return nil, err
	}
	return r.Read()
}

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
func (t *TypeReader[T]) Read() ([]T, error) {
	var result []T
	for t.nextRowToRead < uint(len(t.rows)) {
		row := t.rows[t.nextRowToRead]
		t.nextRowToRead++
		var toRead T
		if err := t.readSingle(row, &toRead); err != nil {
			return nil, err
		}
		result = append(result, toRead)
	}
	return result, nil
}

// ReadSingle reads a single row from the prepared excel file and returns the row parsed into T type object or an error
// This function may advance the internal row counter, and parse other rows, if T contains a slice and the following row has the same primary key
func (t *TypeReader[T]) readSingle(row []string, toRead *T) error {
	v := reflect.ValueOf(toRead).Elem()
	primaryKey := ""
	for _, col := range t.typeInfo.orderedColumns {
		fi := t.typeInfo.nameToField[col]
		if fi.kind == kindSlice {
			continue
		}
		headerIndex, ok := t.headersToIndex[col]
		if !ok {
			if !fi.required && !fi.isPrimaryKey {
				continue
			}
			return fmt.Errorf("required column %s is not present", col)
		}
		rowVal := row[headerIndex]
		// check if the field is optional and the value is empty
		if rowVal == "" {
			if !fi.required && !fi.isPrimaryKey {
				continue
			}
			return fmt.Errorf("required column %s is empty", col)
		}
		if fi.isPrimaryKey {
			primaryKey = rowVal
		}
		field := v.FieldByIndex(fi.index)
		if parsed, err := parseStringIntoType(rowVal, field.Type()); err != nil {
			return fmt.Errorf("error parsing cell value: %v", err)
		} else {
			field.Set(reflect.ValueOf(parsed))
		}
	}
	if t.previousPrimaryKey == primaryKey {
		// we are reading a slice
		//TODO Implement slice reading
	}
	if primaryKey != "" {
		t.previousPrimaryKey = primaryKey
	}
	return nil
}

func (t *TypeReader[T]) analyzeType() error {
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
	return nil
}
