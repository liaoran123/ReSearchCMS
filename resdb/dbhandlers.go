package resdb

/*
// 将主键转为字符串 ，主要是用于组合主键需要进行map比较时使用
func PrimaryToStr(t *Table, k []byte) string {
	var key any
	rstr := ""
	for _, p := range t.primary {
		key = Bytes(k).ToAny(t.fields[p])
		rstr += string(AnyToBytes(key)) + SPLIT
	}
	if rstr == "" {
		return ""
	}
	return rstr[:len(rstr)-1]
}
*/
