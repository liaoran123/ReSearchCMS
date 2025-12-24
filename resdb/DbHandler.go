package resdb

// FieldHandler 定义字段处理的通用接口
type DbHandler interface {
	Parse(table *Table, value []byte) any
}
type ParseFun func(table *Table, value []byte) any
