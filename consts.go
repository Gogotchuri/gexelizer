package gexelizer

const mainTag = "gex"
const mainSeparator = ","
const listSeparator = "|"

type kind int

const (
	kindPrimitive kind = iota
	kindSlice
	kindStruct
	kindPrimitivePtr
	kindStructPtr
)

const (
	ignoreTag     = "-"
	primaryKeyTag = "primary"
	omitEmptyTag  = "omitempty"
	noprefixTag   = "noprefix"
	requiredTag   = "required"
	columnTag     = "column:"
	prefixTag     = "prefix:"
	defaultTag    = "default:"
	aliasesTag    = "aliases:"
	orderTag      = "order:"
)
