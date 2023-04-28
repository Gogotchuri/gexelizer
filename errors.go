package gexelizer

import "fmt"

type RowError struct {
	RowNumber int
	Err       error
}

func (e RowError) Error() string {
	return fmt.Sprintf("row %d: %v", e.RowNumber, e.Err)
}

func newRowError(rowNumber int, err error) RowError {
	return RowError{
		RowNumber: rowNumber,
		Err:       err,
	}
}

func newNonIndexedRowError(err error) RowError {
	return RowError{
		RowNumber: -1,
		Err:       err,
	}
}
