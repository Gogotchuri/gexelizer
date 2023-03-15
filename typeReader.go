package gexelizer

import (
	"fmt"
	"io"
	"reflect"
)

type TypeReader[T any] struct {
	file           ExcelFileReader
	typeInfo       typeInfo
	headers        []string
	nextRowToRead  uint
	options        *Options
	headersToIndex map[string]int
	rows           [][]string
}

func ReadExcel[T any](reader io.Reader, opts ...Options) ([]T, error) {
	r, err := NewTypeReader[T](reader, opts...)
	if err != nil {
		return nil, err
	}
	return r.Read()
}

func NewTypeReader[T any](reader io.Reader, opts ...Options) (*TypeReader[T], error) {
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

func (t *TypeReader[T]) Read() ([]T, error) {
	var result []T
	for t.nextRowToRead < uint(len(t.rows)) {
		//todo
		//row := t.rows[t.nextRowToRead]
		t.nextRowToRead++
		var toRead T
		//if err := t.readSingle(row, &toRead); err != nil {
		//	return nil, err
		//}
		result = append(result, toRead)
	}
	return result, nil
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
	t.headers = t.rows[t.options.HeaderRow]
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
