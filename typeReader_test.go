package gexelizer

import (
	"bytes"
	"testing"
	"time"
)

func TestTypeReader_ReadExcelFile(t *testing.T) {
	type positionStruct struct {
		Position string `gex:""`
	}
	type row struct {
		Name string `gex:",primary"`
		Sl   []positionStruct
		Age  int
	}
	ts, err := ReadExcelFile[row]("test.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", ts)
	if len(ts) != 2 {
		t.Fatalf("expected 2 rows, got %v", len(ts))
	}
}

func TestWriteAndRead(t *testing.T) {
	type positionStruct struct {
		Position string `gex:""`
	}
	type row struct {
		Name string `gex:",primary"`
		Sl   []positionStruct
		Age  int
	}
	ts := []row{
		{
			Name: "John",
			Sl: []positionStruct{
				{
					Position: "CEO",
				},
				{
					Position: "CTO",
				},
			},
			Age: 30,
		},
		{
			Name: "Jane",
			Sl: []positionStruct{
				{
					Position: "CFO",
				},
			},
			Age: 25,
		},
	}
	buffer := &bytes.Buffer{}
	if err := WriteExcel(buffer, ts); err != nil {
		t.Fatal(err)
	}
	tsR, err := ReadExcel[row](buffer)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", ts)
	if len(tsR) != 2 {
		t.Fatalf("expected 2 rows, got %v", len(tsR))
	}
	for i := range tsR {
		if tsR[i].Name != ts[i].Name {
			t.Fatalf("expected %v, got %v", ts[i].Name, tsR[i].Name)
		}
		if tsR[i].Age != ts[i].Age {
			t.Fatalf("expected %v, got %v", ts[i].Age, tsR[i].Age)
		}
		if len(tsR[i].Sl) != len(ts[i].Sl) {
			t.Fatalf("expected %v, got %v", len(ts[i].Sl), len(tsR[i].Sl))
		}
		for j := range tsR[i].Sl {
			if tsR[i].Sl[j].Position != ts[i].Sl[j].Position {
				t.Fatalf("expected %v, got %v", ts[i].Sl[j].Position, tsR[i].Sl[j].Position)
			}
		}
	}
}

func TestWriteAndReadComplex(t *testing.T) {
	type positionStruct struct {
		Position string `gex:""`
		Salary   float64
	}
	type dates struct {
		StartDate string `gex:"Start Date"`
		EndDate   time.Time
	}
	type Employee struct {
		Name string `gex:"primary"`
		Sl   []positionStruct
		Age  int
	}
	type manager struct {
		Employee
		Department string
		Dates      dates
	}
	ts := []manager{
		{
			Employee: Employee{
				Name: "John",
				Sl: []positionStruct{
					{
						Position: "CEO",
						Salary:   100000.0,
					},
					{
						Position: "CTO",
						Salary:   90000.0,
					},
				},
				Age: 30,
			},
			Department: "IT",
			Dates: dates{
				StartDate: "01/01/2020",
				EndDate:   time.Now(),
			},
		},
		{
			Employee: Employee{
				Name: "Jane",
				Sl: []positionStruct{
					{
						Position: "CFO",
						Salary:   80000.0,
					},
				},
				Age: 25,
			},
			Department: "Finance",
			Dates: dates{
				StartDate: "01/01/2020",
				EndDate:   time.Now(),
			},
		},
	}
	buffer := &bytes.Buffer{}
	if err := WriteExcel(buffer, ts); err != nil {
		t.Fatal(err)
	}
	tsR, err := ReadExcel[manager](buffer)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", ts)
	if len(tsR) != 2 {
		t.Fatalf("expected 2 rows, got %v", len(tsR))
	}
	for i := range tsR {
		if tsR[i].Name != ts[i].Name {
			t.Fatalf("expected %v, got %v", ts[i].Name, tsR[i].Name)
		}
		if tsR[i].Age != ts[i].Age {
			t.Fatalf("expected %v, got %v", ts[i].Age, tsR[i].Age)
		}
		if len(tsR[i].Sl) != len(ts[i].Sl) {
			t.Fatalf("expected %v, got %v", len(ts[i].Sl), len(tsR[i].Sl))
		}
		for j := range tsR[i].Sl {
			if tsR[i].Sl[j].Position != ts[i].Sl[j].Position {
				t.Fatalf("expected %v, got %v", ts[i].Sl[j].Position, tsR[i].Sl[j].Position)
			}
			if tsR[i].Sl[j].Salary != ts[i].Sl[j].Salary {
				t.Fatalf("expected %v, got %v", ts[i].Sl[j].Salary, tsR[i].Sl[j].Salary)
			}
		}
		if tsR[i].Department != ts[i].Department {
			t.Fatalf("expected %v, got %v", ts[i].Department, tsR[i].Department)
		}
		if tsR[i].Dates.StartDate != ts[i].Dates.StartDate {
			t.Fatalf("expected %v, got %v", ts[i].Dates.StartDate, tsR[i].Dates.StartDate)
		}
		if tsR[i].Dates.EndDate.Equal(ts[i].Dates.EndDate) {
			t.Fatalf("expected %v, got %v", ts[i].Dates.EndDate, tsR[i].Dates.EndDate)
		}
	}
}
