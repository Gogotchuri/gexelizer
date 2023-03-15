package gexelizer

import "testing"

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

}

func TestExcel_WriteStructWSlice(t *testing.T) {
	type sliceStruct struct {
		Position string
	}
	type row struct {
		Name string
		Age  int
		Job  []sliceStruct
	}
	writer, err := NewTypeWriter[row]()
	if err != nil {
		t.Fatal(err)
	}
	err = writer.writeSingle(row{
		Name: "John",
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
		Name string
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
		Name string
		SL   []sliceStruct
		Age  int
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
