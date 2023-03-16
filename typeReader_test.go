package gexelizer

import "testing"

func TestTypeReader_ReadExcelFile(t *testing.T) {
	type row struct {
		Name string `gex:"primary"`
		Age  int
	}
	ts, err := ReadExcelFile[row]("test.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	if len(ts) != 4 {
		t.Fatalf("expected 4 rows, got %v", len(ts))
	}
}
