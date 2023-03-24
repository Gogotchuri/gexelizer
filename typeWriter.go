package gexelizer

import (
	"bytes"
	"io"
	"reflect"
	"strings"
)

type TypeWriter[T any] struct {
	file     ExcelFileWriter
	typeInfo typeInfo
	headers  []string
	//TODO add support for omitEmpty
	nextRowToWrite uint
	options        *Options
}

func WriteToFile[T any](filename string, data []T, opts ...Options) error {
	tw, err := NewTypeWriter[T](opts...)
	if err != nil {
		return err
	}
	err = tw.Write(data)
	if err != nil {
		return err
	}
	return tw.WriteToFile(filename)
}

func WriteExcel[T any](writer io.Writer, data []T, opts ...Options) error {
	tw, err := NewTypeWriter[T](opts...)
	if err != nil {
		return err
	}
	err = tw.Write(data)
	if err != nil {
		return err
	}
	_, err = tw.WriteTo(writer)
	return err
}

// NewTypeWriter creates a new TypeWriter[T] instance
// It returns an error if the type T cannot be written to excel
// This function is heavier, so it is recommended to create a single instance and reuse it
// Otherwise, it is recommended to make it parallel while you fetch data to writeSingle
func NewTypeWriter[T any](opts ...Options) (*TypeWriter[T], error) {
	w := &TypeWriter[T]{}
	if err := w.analyzeType(); err != nil {
		return nil, err
	}
	w.file = newExcel()
	if len(opts) > 0 {
		w.options = &opts[0]
	} else {
		w.options = DefaultOptions()
	}
	w.nextRowToWrite = w.options.HeaderRow
	return w, nil
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
		if err := w.writeSingle(row); err != nil {
			return err
		}
	}
	return nil
}

func (w *TypeWriter[T]) WriteToFile(filename string) error {
	//Create file
	if err := w.file.SaveAs(filename); err != nil {
		return err
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
	capitalized := make([]string, len(w.headers))
	for i, header := range w.headers {
		capitalized[i] = strings.ToUpper(header[0:1]) + header[1:]
	}
	return w.file.SetStringRow(w.options.HeaderRow, capitalized)
}

type singleWrite struct {
	rows       [][]any
	numColumns int
}

func newRows(numColumns int) singleWrite {
	sw := singleWrite{rows: make([][]any, 1)}
	sw.rows[0] = make([]any, numColumns)
	sw.numColumns = numColumns
	return sw
}

func (r *singleWrite) setCell(x, y int, value any) {
	currentLen := len(r.rows)
	if currentLen <= y {
		for i := currentLen; i <= y; i++ {
			r.rows = append(r.rows, make([]any, r.numColumns))
			//Fill new singleWrite with values from the previous row
			for j := 0; j < r.numColumns; j++ {
				if r.rows[i-1][j] == nil || r.rows[i][j] != nil {
					continue
				}
				if j == x {
					continue
				}
				r.rows[i][j] = r.rows[i-1][j]
			}
		}
	}
	r.rows[y][x] = value
}

func (r *singleWrite) setColumnValue(x int, value any) {
	if len(r.rows) == 1 {
		r.rows[0][x] = value
		return
	}
	for i := 0; i < len(r.rows); i++ {
		r.setCell(x, i, value)
	}
}

func (w *TypeWriter[T]) writeSingle(row T) error {
	sw := newRows(len(w.headers))
	if !w.typeInfo.containsSlice() {
		for i := 0; i < len(w.typeInfo.orderedColumns); i++ {
			col := w.typeInfo.orderedColumns[i]
			fi := w.typeInfo.nameToField[col]
			fv := reflect.ValueOf(row).FieldByIndex(fi.index)
			sw.setColumnValue(i, fv.Interface())
		}

	} else {
		passedSlice := false
		sliceFI := *w.typeInfo.sliceFieldInfo
		sliceFV := reflect.ValueOf(row).FieldByIndex(sliceFI.index)
		for i := 0; i < len(w.typeInfo.orderedColumns); i++ {
			col := w.typeInfo.orderedColumns[i]
			fi := w.typeInfo.nameToField[col]
			if fi.kind == kindSlice {
				passedSlice = true
				continue
			}
			if fi.isChildOf(sliceFI) {
				// For each slice element, write a new row, and fill the rest of the columns with the previous value
				for j := 0; j < sliceFV.Len(); j++ { //For each slice struct property fill the column with the values
					elemIndex := fi.index[len(sliceFI.index):]
					sliceElemValue := sliceFV.Index(j).FieldByIndex(elemIndex)
					x := i
					if passedSlice {
						x -= 1
					}
					sw.setCell(x, j, sliceElemValue.Interface())
				}
				continue
			}
			fv := reflect.ValueOf(row).FieldByIndex(fi.index)
			x := i
			if passedSlice {
				x -= 1
			}
			sw.setColumnValue(x, fv.Interface())
		}
	}

	for _, row := range sw.rows {
		if err := w.file.SetRow(w.nextRowToWrite, row); err != nil {
			return err
		}
		w.nextRowToWrite++
	}
	return nil
}
