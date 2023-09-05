package gexelizer

import (
	"github.com/xuri/excelize/v2"
	"testing"
	"time"
)

func TestExcelize_FileWriter(t *testing.T) {
	excel := excelize.NewFile()
	err := excel.SetCellValue("Sheet1", "A1", "test")
	if err != nil {
		return
	}
	err = excel.SetCellValue("Sheet1", "B1", 100)
	if err != nil {
		return
	}
	err = excel.SetCellValue("Sheet1", "B1", 100)
	err = excel.SetCellValue("Sheet1", "C1", 100)
	err = excel.SetCellValue("Sheet1", "K1", 100)
	if err != nil {
		return
	}
	if err != nil {
		return
	}
	if err != nil {
		return
	}
	err = excel.SetCellValue("Sheet1", "C1", time.Now())
	if err != nil {
		return
	}
	err = excel.SaveAs("test.xlsx")
	if err != nil {
		t.Fatal(err)
	}
}

func TestExcel_WriteBasic(t *testing.T) {
	type row struct {
		Name string
		Age  int
	}
	writer, err := NewTypeWriter[row]()
	if err != nil {
		t.Fatal(err)
	}
	err = writer.writeSingle(row{
		Name: "John",
		Age:  20,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = writer.writeSingle(row{
		Name: "Jane",
		Age:  21,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = writer.WriteToFile("test.xlsx")
	if err != nil {
		t.Fatal(err)
	}

}

func TestExcel_WriteStructWSlice(t *testing.T) {
	type sliceStruct struct {
		Position string
	}
	type row struct {
		Name string `gex:"column:name,primary"`
		Age  int
		Job  []sliceStruct
		Date time.Time
	}
	writer, err := NewTypeWriter[row]()
	if err != nil {
		t.Fatal(err)
	}
	err = writer.writeSingle(row{
		Name: "John",
		Date: time.Now(),
		Age:  20,
		Job:  []sliceStruct{{"A"}, {"B"}},
	})
	if err != nil {
		return
	}
	err = writer.writeSingle(row{
		Name: "Jane",
		Age:  21,
		Job:  []sliceStruct{{"C"}, {"D"}},
	})
	if err != nil {
		return
	}
}

func TestExcel_WriteStructWSliceInMiddle(t *testing.T) {
	type sliceStruct struct {
		Position string
	}
	type row struct {
		Name string `gex:"column:name,primary"`
		Age  int
		Job  []sliceStruct
		Job2 string
	}
	writer, err := NewTypeWriter[row]()
	if err != nil {
		t.Fatal(err)
	}
	err = writer.writeSingle(row{
		Name: "John",
		Age:  20,
		Job:  []sliceStruct{{"A"}, {"B"}},
		Job2: "C",
	})
	if err != nil {
		return
	}
	err = writer.writeSingle(row{
		Name: "Jane",
		Age:  21,
		Job:  []sliceStruct{{"C"}, {"D"}},
		Job2: "D",
	})
	if err != nil {
		return
	}
}

func TestTypeWriter_WriteToFile(t *testing.T) {
	type sliceStruct struct {
		Position string
	}
	type row struct {
		Name string `gex:"column:name,primary"`
		SL   []sliceStruct
		Age  int
		Date time.Time
	}

	writer, err := NewTypeWriter[row]()
	if err != nil {
		t.Fatal(err)
	}
	err = writer.Write([]row{
		{
			Name: "John",
			Age:  20,
			SL:   []sliceStruct{{"A"}, {"B"}},
		},
		{
			Name: "Jane",
			Age:  21,
			Date: time.Now(),
			SL:   []sliceStruct{{"C"}, {"D"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = writer.WriteToFile("test.xlsx")
	if err != nil {
		t.Fatal(err)
	}
}

func TestTypeWriter_WriteToFileMultiSheet(t *testing.T) {
	type sliceStruct struct {
		Position string
	}
	type row struct {
		Name string `gex:"column:name,primary"`
		SL   []sliceStruct
		Age  int
		Date time.Time
	}
	type row2 struct {
		Name   string `gex:"column:name,primary"`
		SL     []sliceStruct
		Age    int
		Height float64
	}
	rows1 := []row{
		{
			Name: "John",
			Age:  20,
			SL:   []sliceStruct{{"A"}, {"B"}},
		},
		{
			Name: "Jane",
			Age:  21,
			Date: time.Now(),
			SL:   []sliceStruct{{"C"}, {"D"}},
		},
		{
			Name: "Jane",
			Age:  21,
		},
	}
	rows2 := []row2{
		{
			Name:   "John",
			Age:    20,
			Height: 1.2,
		},
		{
			Name:   "Jane",
			Age:    21,
			Height: 1.3,
		},
		{
			Name:   "Jane",
			Age:    21,
			SL:     []sliceStruct{{"C"}, {"D"}},
			Height: 1.4,
		},
	}

	writer, err := WriteExcelSheet(nil, "sheet1", rows1, Options{
		DataStartRow:  5,
		HeaderRow:     4,
		TrimEmptyRows: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = WriteExcelSheet(writer, "sheet2", rows2)
	if err != nil {
		t.Fatal(err)
	}
	err = writer.SaveAs("test.xlsx")
	if err != nil {
		t.Fatal(err)
	}
}

func TestTypeWriter_TestOverwriteEmbedded(t *testing.T) {
	type embType struct {
		Position string
	}
	type typeBefore struct {
		Name string `gex:"column:name,primary"`
		embType
		Position uint
		Age      int
	}
	type typeAfter struct {
		Name     string `gex:"column:name,primary"`
		Position uint
		embType
		Age int
	}
	//should not work
	type typeSameDepth struct {
		Name     string `gex:"column:name,primary"`
		Position uint
		Pst      string `gex:"column:Position"`
	}

	writerBefore, err := NewTypeWriter[typeBefore]()
	if err != nil {
		t.Fatal(err)
	}
	err = writerBefore.Write([]typeBefore{
		{
			Name:     "John",
			Age:      20,
			Position: 1,
			embType:  embType{"A"},
		},
		{
			Name:     "Jane",
			Age:      21,
			Position: 2,
			embType:  embType{"B"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	buffer, err := writerBefore.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}
	excel, err := ReadExcel[typeBefore](buffer)
	if err != nil {
		t.Fatal(err)
	}
	if excel[0].Position != 1 {
		t.Fatal("Position should be 1")
	}

	writerAfter, err := NewTypeWriter[typeAfter]()
	if err != nil {
		t.Fatal(err)
	}
	err = writerAfter.Write([]typeAfter{
		{
			Name:     "John",
			Age:      20,
			Position: 1,
			embType:  embType{"A"},
		},
		{
			Name:     "Jane",
			Age:      21,
			Position: 2,
			embType:  embType{"B"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	buffer, err = writerAfter.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}
	excel2, err := ReadExcel[typeAfter](buffer)
	if err != nil {
		t.Fatal(err)
	}
	if excel2[0].Position != 1 {
		t.Fatal("Position should be 1")
	}

	_, err = NewTypeWriter[typeSameDepth]()
	if err == nil {
		t.Fatal("Should not be able to create writer")
	}
}

func TestTypeWriter_WriteToBufferOmitempty(t *testing.T) {
	type position struct {
		Position string
	}
	type row struct {
		Name string    `gex:"column:name,primary"`
		Age  int       `gex:"omitempty"`
		P    *position `gex:"omitempty"`
	}
	writer, err := NewTypeWriter[row]()
	if err != nil {
		t.Fatal(err)
	}
	err = writer.Write([]row{
		{
			Name: "John",
		},
		{
			Name: "Jane",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	buffer, err := writer.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}

	excel, err := excelize.OpenReader(buffer)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := excel.GetRows("Sheet1")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Fatal("Should have 3 rows")
	}
	t.Logf("%+v", rows)
	if len(rows[0]) != 1 {
		t.Fatal("Should have 1 column")
	}
}

func TestTypeWriter_WriteBufferNilOmitempty(t *testing.T) {
	type address struct {
		Street string `gex:"column:street"`
		//Street string `gex:"column:street,required"`
	}
	type row struct {
		Name     string   `gex:"column:name,primary"`
		Address  *address `gex:"omitempty"`
		Shipping address  `gex:"noprefix,omitempty"`
	}
	writer, err := NewTypeWriter[row]()
	if err != nil {
		t.Fatal(err)
	}
	err = writer.Write([]row{
		{
			Name: "John",
		},
		{
			Name: "Jane",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	buffer, err := writer.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}
	//ReadExcel
	excel, err := ReadExcel[row](buffer)
	if err != nil {
		t.Fatal(err)
	}
	if excel[0].Name != "John" {
		t.Fatal("Name should be John")
	}
	//TODO: fix this, should be nil, but it gets initialized for the inner field, and never gets set to nil
	// should scan for references to the field, and set them to nil if the struct fields are all default
	if excel[0].Address != nil {
		t.Fatal("Should be nil")
	}
	if excel[0].Shipping.Street != "" {
		t.Fatal("Should be empty")
	}
}
