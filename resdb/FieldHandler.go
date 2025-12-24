package resdb

// FieldHandler 定义字段处理的通用接口
type FieldHandler interface {
	// Parse 将原始值进行各种解析处理的可能
	Parse(record *Record)
}
type FieldParseFunc func(record *Record)
