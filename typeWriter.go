package gexelizer

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
)

type TypeWriter[T any] struct {
	file           ExcelFileWriter
	headers        []string
	ignored        map[int]struct{}
	nextRowToWrite uint
	options        *Options
}

// NewTypeWriter creates a new TypeWriter[T] instance
// It returns an error if the type T cannot be written to excel
// This function is heavier, so it is recommended to create a single instance and reuse it
// Otherwise, it is recommended to make it parallel while you fetch data to write
func NewTypeWriter[T any]() (*TypeWriter[T], error) {
	w := &TypeWriter[T]{ignored: make(map[int]struct{})}
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
	fmt.Println(info)
	return nil
}

func (w *TypeWriter[T]) writeHeaders() error {
	return w.file.SetStringRow(w.options.HeaderRow, w.headers)
}

func (w *TypeWriter[T]) writeRow(row T) error {
	//TODO: implement
	return nil
}
