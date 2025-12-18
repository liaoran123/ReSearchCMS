package storedb

import "errors"

type TableData interface {
	GetRecord() Records
}

//迭代器返回主键的数据集
type PrimaryData struct {
	table *Table
	Data  [][]byte
}

func PrimaryDataNew(table *Table, data [][]byte) (*PrimaryData, error) {
	if data == nil {
		return nil, errors.New("data is nil!数据不能为空。")
	}
	return &PrimaryData{
		table: table,
		Data:  data,
	}, nil
}
func (p *PrimaryData) GetRecord() Records {
	rs := make(Records, len(p.Data))
	for i, d := range p.Data {
		rs[i] = Record(p.table.ParseValue(d))
	}
	return rs
}

//迭代器返回索引的数据集
type IndexData struct {
	table *Table
	Data  [][]byte
}

func IndexDataNew(table *Table, data [][]byte) (*IndexData, error) {
	if data == nil {
		return nil, errors.New("data is nil!数据不能为空。")
	}
	return &IndexData{
		table: table,
		Data:  data,
	}, nil
}
func (i *IndexData) GetRecord() Records {
	pytype, exists := i.table.fields[i.table.primary]
	if !exists {
		return nil
	}
	rs := make(Records, len(i.Data))
	var id any
	var byrecord []byte
	for j, d := range i.Data {
		id = Bytes(d).ToAny(pytype)
		byrecord = i.table.Read(id)
		if byrecord != nil {
			rs[j] = Record(i.table.ParseValue(byrecord))
		}
	}
	return rs
}
