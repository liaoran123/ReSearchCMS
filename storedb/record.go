package storedb

type Record map[string]any

// 选择显示的key字段，如sql语句中的select f0,f1,... from table
func (r Record) GetKeys(keys ...string) (rs map[string]any) {
	if r != nil {
		return nil
	}
	if len(keys) == 0 {
		return nil
	}
	rs = make(map[string]any, len(keys))
	for _, key := range keys {
		rs[key] = r[key]
	}
	return rs
}

type Records []Record

func (rs Records) GetKeys(keys ...string) (rs2 []map[string]any) {
	if len(keys) == 0 {
		return nil
	}
	rs2 = make([]map[string]any, len(rs))
	for i, r := range rs {
		rs2[i] = r.GetKeys(keys...)
	}
	return rs2
}

type DataSet map[string]Records //string表名
