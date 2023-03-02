package gexelizer

type Options struct {
	DataStartRow uint
	HeaderRow    uint
}

func DefaultOptions() *Options {
	return &Options{
		DataStartRow: 2,
		HeaderRow:    1,
	}
}
