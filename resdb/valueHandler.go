package resdb

/*
// FieldHandler 定义字段处理的通用接口
type ValueHandler interface {
	Parse(table *Table, value []byte) any
}
*/
type ValueParseFun func(table *Table, value []byte) any

// 将组合主键转为字符串 ，主要是用于组合主键需要进行map比较时使用
func CompositePrimaryParse(t *Table, k []byte) any {
	var key any
	rstr := ""
	for _, p := range t.primaryKey.GetFields() {
		key = Bytes(k).ToAny(t.fields[p])
		rstr += string(AnyToBytes(key)) + SPLIT
	}
	if rstr == "" {
		return ""
	}
	return rstr[:len(rstr)-1]
}

//单主键
func SinglePrimaryParse(t *Table, k []byte) any {
	return Bytes(k).ToAny(t.fields[t.primaryKey.GetFields()[0]])
}
