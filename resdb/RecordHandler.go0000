package resdb

/*
//添加接口
type RecordHandler interface {
	Get(table *Table, value []byte, fieldParseFunc ...FieldParseFunc) Record
}
*/
//通过主键和索引返回记录，只有2种情况。
type RecordGetFun func(table *Table, value []byte, fieldParseFunc ...FieldParseFunc) Record

// 解析索引记录
// 组合主键需要反转义
// 组合主键时，所有索引对应的主键值都是经过转义的。
// 这是为了提取主键中的某个字段值作为匹配索引，必须转义才能正确根据转义符分割提取。
// 所以在用索引的主键值回表记录时，需要对索引值进行反转义。
/*
type ByIndex struct{}

func (p *ByIndex) Get(t *Table, v []byte, fieldParseFunc ...FieldParseFunc) Record {
	value := v
	if len(t.primary) > 1 { //只有组合主键才需要转义
		value = Bytes(v).UnEscape()
	}
	// 读取完整记录
	byrecord := t.ReadByBytes(value)
	if byrecord == nil {
		return nil
	}
	// 解析记录并提取指定字段
	rd := Record(t.ParseValue(byrecord))
	for _, fieldParseFunc := range fieldParseFunc {
		fieldParseFunc(&rd)
	}
	return rd
}
*/
//ByIndexGet=ByIndex.Get 两种方式都可
//简单直接函数类型ByIndexGet，需要外部数据组合操作，才需要使用结构。
func ByIndexGet(t *Table, v []byte, fieldParseFunc ...FieldParseFunc) Record {
	value := v
	if len(t.primary) > 1 { //只有组合主键才需要转义
		value = Bytes(v).UnEscape()
	}
	// 读取完整记录
	byrecord := t.ReadByBytes(value)
	if byrecord == nil {
		return nil
	}
	// 解析记录并提取指定字段
	rd := Record(t.ParseValue(byrecord))
	for _, fieldParseFunc := range fieldParseFunc {
		fieldParseFunc(&rd)
	}
	return rd
}

/*
// 解析主键记录
type ByPrimary struct{}

func (p *ByPrimary) Get(t *Table, v []byte, fieldParseFunc ...FieldParseFunc) Record {
	rd := Record(t.ParseValue(v))
	for _, fieldParseFunc := range fieldParseFunc {
		fieldParseFunc(&rd)
	}
	return rd
}
*/
// 解析主键记录
//ByPrimaryGet=ByPrimary.Get 两种方式都可
func ByPrimaryGet(t *Table, v []byte, fieldParseFunc ...FieldParseFunc) Record {
	rd := Record(t.ParseValue(v))
	for _, fieldParseFunc := range fieldParseFunc {
		fieldParseFunc(&rd)
	}
	return rd
}
