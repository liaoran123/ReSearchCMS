package storedb

import "errors"

// TableData 定义表数据接口
// 实现sql语句中的select f0,f1,... from table 要返回的字段
// 如果keys为空，则返回所有字段
// 在转换为Records时，即执行该操作，可以减少内存占用
type TableData interface {
	GetRecord(keys ...string) Records
}

// baseData 基础数据结构体，提取公共字段
type baseData struct {
	table *Table
	Data  [][]byte
}

// PrimaryData 迭代器返回主键的数据集
type PrimaryData struct {
	baseData
}

// PrimaryDataNew 创建PrimaryData实例
func PrimaryDataNew(table *Table, data [][]byte) (*PrimaryData, error) {
	if data == nil {
		return nil, errors.New("data is nil!数据不能为空。")
	}
	return &PrimaryData{
		baseData: baseData{
			table: table,
			Data:  data,
		},
	}, nil
}

// GetRecord 获取主键数据集的记录
func (p *PrimaryData) GetRecord(keys ...string) Records {
	// 处理空数据情况
	if len(p.Data) == 0 {
		return Records{}
	}

	// 预分配切片，避免频繁扩容
	rs := make(Records, len(p.Data))
	for i, d := range p.Data {
		// 解析数据并提取指定字段
		rs[i] = Record(p.table.ParseValue(d)).GetKeys(keys...)
	}
	return rs
}

// IndexData 迭代器返回索引的数据集
type IndexData struct {
	baseData
}

// IndexDataNew 创建IndexData实例
func IndexDataNew(table *Table, data [][]byte) (*IndexData, error) {
	if data == nil {
		return nil, errors.New("data is nil!数据不能为空。")
	}
	return &IndexData{
		baseData: baseData{
			table: table,
			Data:  data,
		},
	}, nil
}

// GetRecord 获取索引数据集的记录
func (i *IndexData) GetRecord(keys ...string) Records {
	// 检查主键字段是否存在
	pytype, exists := i.table.fields[i.table.primary]
	if !exists {
		return Records{}
	}

	// 处理空数据情况
	if len(i.Data) == 0 {
		return Records{}
	}

	// 预分配切片，避免频繁扩容
	rs := make(Records, len(i.Data))
	var id any
	var byrecord []byte

	for j, d := range i.Data {
		// 转换索引值为主键值
		id = Bytes(d).ToAny(pytype)
		if id == nil {
			continue // 跳过无效的主键值
		}

		// 通过主键读取完整记录
		byrecord = i.table.Read(id)
		if byrecord != nil {
			// 解析记录并提取指定字段
			rs[j] = Record(i.table.ParseValue(byrecord)).GetKeys(keys...)
		}
	}

	return rs
}
