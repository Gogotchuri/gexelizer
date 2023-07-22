package gexelizer

type Options struct {
	DataStartRow  uint
	HeaderRow     uint
	TrimEmptyRows bool
	File          ExcelFileWriter
}

func DefaultOptions() *Options {
	return &Options{
		DataStartRow:  2,
		HeaderRow:     1,
		TrimEmptyRows: true,
		File:          nil,
	}
}
