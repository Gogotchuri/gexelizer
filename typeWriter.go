package gexelizer

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
)

type TypeWriter[T any] struct {
	file           ExcelFileWriter
	typeInfo       typeInfo
	headers        []string
	nextRowToWrite uint
	options        *Options
}

// NewTypeWriter creates a new TypeWriter[T] instance
// It returns an error if the type T cannot be written to excel
// This function is heavier, so it is recommended to create a single instance and reuse it
// Otherwise, it is recommended to make it parallel while you fetch data to write
func NewTypeWriter[T any]() (*TypeWriter[T], error) {
	w := &TypeWriter[T]{}
	if err := w.analyzeType(); err != nil {
		return nil, err
	}
	w.file = NewExcel()
	w.options = DefaultOptions()
	w.nextRowToWrite = w.options.HeaderRow
	return w, nil
}

func NewOptionedTypeWriter[T any](opt Options) (*TypeWriter[T], error) {
	if opt.DataStartRow <= opt.HeaderRow {
		return nil, fmt.Errorf("data start row must be greater than header row")
	}
	writer, err := NewTypeWriter[T]()
	if err != nil {
		return nil, err
	}
	writer.options = &opt
	writer.nextRowToWrite = opt.HeaderRow
	return writer, nil
}

func (w *TypeWriter[T]) Write(data []T) error {
	if len(data) == 0 {
		return nil
	}
	if w.nextRowToWrite == w.options.HeaderRow {
		w.nextRowToWrite = w.options.DataStartRow
		if err := w.writeHeaders(); err != nil {
			return err
		}
	}
	for _, row := range data {
		if err := w.writeRow(row); err != nil {
			return err
		}
	}
	return nil
}

func (w *TypeWriter[T]) WriteTo(writer io.Writer) (int64, error) {
	return w.file.WriteTo(writer)
}

func (w *TypeWriter[T]) WriteToBuffer() (*bytes.Buffer, error) {
	return w.file.WriteToBuffer()
}

func (w *TypeWriter[T]) analyzeType() error {
	var t T
	info, err := analyzeType(reflect.TypeOf(t))
	if err != nil {
		return err
	}
	w.typeInfo = info
	w.headers = make([]string, 0, len(info.orderedColumns))
	//Include headers except for slices
	for _, col := range info.orderedColumns {
		fi := info.nameToField[col]
		if fi.kind == kindSlice {
			continue
		}
		w.headers = append(w.headers, col)
	}
	return nil
}

func (w *TypeWriter[T]) writeHeaders() error {
	return w.file.SetStringRow(w.options.HeaderRow, w.headers)
}

func (w *TypeWriter[T]) writeRow(row T) error {
	cellValues := make([]interface{}, 0, len(w.typeInfo.orderedColumns))
	for i := 0; i < len(w.typeInfo.orderedColumns); i++ {
		col := w.typeInfo.orderedColumns[i]
		fieldInfo := w.typeInfo.nameToField[col]
		fieldValue := reflect.ValueOf(row).FieldByIndex(fieldInfo.index)
		if fieldInfo.kind == kindSlice {
			for j := 0; j < fieldValue.Len(); j++ {
				i++ //TODO wrong, because, every row isn't different column
				col = w.typeInfo.orderedColumns[i]
				sliceElemInfo := w.typeInfo.nameToField[col]
				println(fieldValue.Kind(), fieldValue.Type(), fieldValue.Len(), j)
				sliceElem := fieldValue.Index(j)
				cellValues = append(cellValues, sliceElem.FieldByIndex(sliceElemInfo.index[1:]).Interface())
				w.nextRowToWrite++
			}
		} else {
			cellValues = append(cellValues, fieldValue.Interface())
		}
	}
	return w.file.SetRow(w.nextRowToWrite, cellValues)
}
