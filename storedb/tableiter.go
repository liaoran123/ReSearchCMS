package storedb

import (
	"sync"

	"github.com/syndtr/goleveldb/leveldb/iterator"
)

type TableIter interface {
	//实现sql语句中的select f0,f1,... from table 要返回的字段
	//如果keys为空，则返回所有字段
	// 在最底层转换，最大化减少内存占用
	GetRecord(esc bool, limit ...int) Records
}

// 基础迭代器结构体，提取公共字段和方法
type baseIter struct {
	table *Table
	keys  []string
	move  map[bool]func() bool
	top   map[bool]func() bool
	mu    sync.Mutex
}

// 直接使用游标迭代器提取主键的数据集，减少内存占用
type PrimaryDataIter struct {
	baseIter
	iter iterator.Iterator
}

func PrimaryDataIterNew(table *Table, iter iterator.Iterator, keys ...string) *PrimaryDataIter {
	return &PrimaryDataIter{
		baseIter: baseIter{
			table: table,
			keys:  keys,
			move: map[bool]func() bool{
				true:  iter.Next,
				false: iter.Prev,
			},
			top: map[bool]func() bool{
				true:  iter.First,
				false: iter.Last,
			},
		},
		iter: iter,
	}
}

// 设置要返回的字段
func (p *PrimaryDataIter) SetKeys(keys ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.keys = keys
}
func (p *PrimaryDataIter) Release() {
	p.iter.Release()
}

// 从迭代器中提取主键的数据集
func (p *PrimaryDataIter) GetRecord(esc bool, limit ...int) (r Records) {
	//添加锁，防止并发访问
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.top[esc]() {
		return nil
	}
	var start int
	var count int
	switch len(limit) {

	case 0:
		start = 0
		count = -1
	case 1:
		count = limit[0]
	default:
		start = limit[0]
		count = limit[1]
	}
	if count > 0 {
		r = make(Records, 0, count)
	}
	loop := 0
	// 处理当前位置的元素
	for {
		if loop < start {
			loop++
		} else {
			record := Record(p.table.ParseValue(p.iter.Value())).GetKeys(p.keys...)
			r = append(r, record)
			if count > 0 && len(r) >= count {
				break
			}
		}
		// 移动到下一个/前一个元素
		if !p.move[esc]() {
			break
		}
	}
	// 释放迭代器
	return r
}

type IndexDataIter struct {
	baseIter
	iter iterator.Iterator
}

func IndexDataIterNew(table *Table, iter iterator.Iterator, keys ...string) *IndexDataIter {
	return &IndexDataIter{
		baseIter: baseIter{
			table: table,
			keys:  keys,
			move: map[bool]func() bool{
				true:  iter.Next,
				false: iter.Prev,
			},
			top: map[bool]func() bool{
				true:  iter.First,
				false: iter.Last,
			},
		},
		iter: iter,
	}
}

// 设置要返回的字段
func (i *IndexDataIter) SetKeys(keys ...string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.keys = keys
}
func (i *IndexDataIter) Release() {
	i.iter.Release()
}

// 从迭代器中提取索引的数据集
func (i *IndexDataIter) GetRecord(esc bool, limit ...int) (r Records) {
	//添加锁，防止并发访问
	i.mu.Lock()
	defer i.mu.Unlock()

	if !i.top[esc]() {
		return nil
	}
	var start int
	var count int
	switch len(limit) {

	case 0:
		start = 0
		count = -1
	case 1:
		count = limit[0]
	default:
		start = limit[0]
		count = limit[1]
	}
	if count > 0 {
		r = make(Records, 0, count)
	}
	loop := 0
	// 处理当前位置的元素
	for {
		if loop < start {
			loop++
		} else {
			record := i.Parse(i.iter.Value())
			r = append(r, record)
			if count > 0 && len(r) >= count {
				break
			}
		}
		// 移动到下一个/前一个元素
		if !i.move[esc]() {
			break
		}
	}
	// 释放迭代器
	return r
}

// 在最底层转换，最大化减少内存占用
func (i *IndexDataIter) Parse(value []byte) Record {
	// 确保主键字段存在
	pytype, exists := i.table.fields[i.table.primary]
	if !exists {
		return nil
	}

	// 转换主键值
	id := Bytes(value).ToAny(pytype)
	if id == nil {
		return nil
	}

	// 读取完整记录
	byrecord := i.table.Read(id)
	if byrecord == nil {
		return nil
	}

	// 解析记录并提取指定字段
	return Record(i.table.ParseValue(byrecord)).GetKeys(i.keys...)

}
