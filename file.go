package gexelizer

import (
	"bytes"
	"fmt"
	"github.com/shakinm/xlsReader/xls"
	"github.com/xuri/excelize/v2"
	"golang.org/x/text/encoding/unicode"
	"io"
	"time"
	"unicode/utf8"
)

type ExcelFileWriter interface {
	SaveAs(path string) error
	WriteTo(w io.Writer) (int64, error)
	WriteToBuffer() (*bytes.Buffer, error)
	SetCellValueOfSheet(sheet, axis string, value any) error
	SetRow(row uint, values []any) error
	SetStringRow(row uint, values []string) error
	RemoveColumn(column string) error
	GetDefaultSheet() string
	SetDefaultSheet(sheet string) error
	GetBaseFile() *excelize.File
}
type ExcelFileReader interface {
	GetDefaultSheetRows() ([][]string, error)
}

var _ ExcelFileWriter = (*excelFile)(nil)
var _ ExcelFileReader = (*excelFile)(nil)
var _ ExcelFileReader = (*xlsFile)(nil)

type excelFile struct {
	file              *excelize.File
	rows              [][]string
	defaultSheetIndex int
}

func (f *excelFile) GetBaseFile() *excelize.File {
	return f.file
}

type xlsFile struct {
	file *xls.Workbook
	rows [][]string
}

func (x xlsFile) GetDefaultSheetRows() ([][]string, error) {
	sh, err := x.file.GetSheet(0)
	if err != nil {
		return nil, err
	}
	fancyRows := sh.GetRows()
	rows := make([][]string, len(fancyRows))
	for i, row := range fancyRows {
		cols := row.GetCols()
		rows[i] = make([]string, len(cols))
		for j, cell := range cols {
			s := cell.GetString()
			if utf8.ValidString(s) {
				rows[i][j] = s
			} else {
				decoder := unicode.UTF8.NewDecoder()
				s, _ = decoder.String(s)
				//Decode to UTF-8
				rows[i][j] = s
			}
		}
	}
	return rows, nil
}

func (f *excelFile) RemoveColumn(column string) error {
	return f.file.RemoveCol(f.GetDefaultSheet(), column)
}

func NewExcelizeWriter() ExcelFileWriter {
	return &excelFile{
		file: excelize.NewFile(),
	}
}

func readExcelFile(path string) (ExcelFileReader, error) {
	file, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	excel := &excelFile{
		file: file,
	}
	excel.rows, err = excel.GetDefaultSheetRows()
	return excel, nil
}

func readXLSExcel(reader io.ReadSeeker) (efr ExcelFileReader, err error) {
	//panic recover
	defer func() {
		if r := recover(); r != nil {
			efr = nil
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	workbook, err := xls.OpenReader(reader)
	if err != nil {
		return nil, err
	}
	excel := &xlsFile{
		file: &workbook,
	}
	excel.rows, err = excel.GetDefaultSheetRows()
	return excel, nil
}

func readExcel(reader io.Reader) (efr ExcelFileReader, err error) {
	//panic recover
	defer func() {
		if r := recover(); r != nil {
			efr = nil
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	file, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, err
	}
	excel := &excelFile{
		file: file,
	}
	excel.rows, err = excel.GetDefaultSheetRows()
	return excel, nil
}

func (f *excelFile) SetCellValueOfSheet(sheet, axis string, value any) error {
	return f.file.SetCellValue(sheet, axis, value)
}
func (f *excelFile) SetStringRow(row uint, values []string) error {
	return f.file.SetSheetRow(f.GetDefaultSheet(), fmt.Sprintf("A%d", row), &values)
}

var DateTimeFormat = time.RFC3339

func (f *excelFile) SetRow(row uint, values []any) error {
	for i, v := range values {
		if gv, ok := v.(GexValuer); ok {
			values[i] = gv.GexelizerValue()
		} else if t, ok := v.(time.Time); ok {
			values[i] = t.Format(DateTimeFormat)
		} else if t, ok := v.(fmt.Stringer); ok {
			values[i] = t.String()
		}
	}
	return f.file.SetSheetRow(f.GetDefaultSheet(), fmt.Sprintf("A%d", row), &values)
}

func (f *excelFile) WriteTo(w io.Writer) (int64, error) {
	return f.file.WriteTo(w)
}

func (f *excelFile) WriteToBuffer() (*bytes.Buffer, error) {
	return f.file.WriteToBuffer()
}

func (f *excelFile) SaveAs(path string) error {
	return f.file.SaveAs(path)
}

func (f *excelFile) GetRows(sheet string) ([][]string, error) {
	if f.rows != nil {
		return f.rows, nil
	}
	return f.file.GetRows(sheet)
}

func (f *excelFile) SetDefaultSheet(sheet string) error {
	index, err := f.file.GetSheetIndex(sheet)
	if err == nil && index >= 0 {
		f.defaultSheetIndex = index
		return nil
	}
	index, err = f.file.NewSheet(sheet)
	if err == nil {
		f.defaultSheetIndex = index
		return nil
	}
	return err
}

func (f *excelFile) GetDefaultSheet() string {
	return f.file.GetSheetName(f.defaultSheetIndex)
}

func (f *excelFile) GetDefaultSheetRows() ([][]string, error) {
	sheet := f.GetDefaultSheet()
	return f.GetRows(sheet)
}
