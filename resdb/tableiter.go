package resdb

import (
	"sync"

	"github.com/syndtr/goleveldb/leveldb/iterator"
)

// 基础迭代器结构体，提取公共字段和方法
type baseIter struct {
	table *Table
	keys  []string
	move  map[bool]func() bool
	top   map[bool]func() bool
	mu    sync.Mutex
}

// 实现sql语句中的select f0,f1,... from table 要返回的字段
// 如果keys为空，则返回所有字段
// 在最底层转换，最大化减少内存占用
type TableIter struct {
	baseIter
	iter iterator.Iterator
}

// 解析函数
type Parser func(k, v []byte) any

func TableIterNew(table *Table, iter iterator.Iterator, keys ...string) *TableIter {
	return &TableIter{
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
func (t *TableIter) Release() {
	t.iter.Release()
}

// 遍历迭代器
func (t *TableIter) For(parser Parser, esc bool, limit ...int) (r []any) {
	//添加锁，防止并发访问
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.top[esc]() {
		return nil
	}
	var rd any
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
		r = make([]any, 0, count)
	}
	loop := 0
	// 处理当前位置的元素
	for {
		if loop < start {
			loop++
		} else {
			rd = parser(t.iter.Key(), t.iter.Value())
			r = append(r, rd)
			if count > 0 && len(r) >= count {
				break
			}
		}
		// 移动到下一个/前一个元素
		if !t.move[esc]() {
			break
		}
	}
	return r
}

// 解析主键记录
func (t *TableIter) parserRecordByPrimary(k, v []byte) any {
	return Record(t.table.ParseValue(v)).GetKeys(t.keys...)
}

// 获取所有主键记录
func (t *TableIter) GetRecordsByPrimary(esc bool) (r Records) {
	// 将 []any 断言为 []Records
	anySlice := t.For(t.parserRecordByPrimary, esc)
	r = make(Records, len(anySlice))
	for i, v := range anySlice {
		r[i] = v.(Record)
	}
	return r
}

// 解析索引记录
func (t *TableIter) parserRecordByIndex(k, v []byte) any {
	// 确保主键字段存在
	pytype, exists := t.table.fields[t.table.primary]
	if !exists {
		return nil
	}
	// 转换主键值
	id := Bytes(v).ToAny(pytype)
	if id == nil {
		return nil
	}
	// 读取完整记录
	byrecord := t.table.Read(id)
	if byrecord == nil {
		return nil
	}
	// 解析记录并提取指定字段
	return Record(t.table.ParseValue(byrecord)).GetKeys(t.keys...)
}

// 获取所有索引记录
func (t *TableIter) GetRecordsByIndex(esc bool) (r Records) {
	// 将 []any 断言为 []Records
	anySlice := t.For(t.parserRecordByIndex, esc)
	r = make(Records, len(anySlice))
	for i, v := range anySlice {
		r[i] = v.(Record)
	}
	return r
}

// 判断是否存在指定的主键记录
func (t *TableIter) Exist() bool {
	return t.iter.First()
}

// 统计索引记录数
func (t *TableIter) Count() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	i := 0
	for t.iter.Next() {
		i++
	}
	return i
}

// -----------------------------------------------
/*
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

// 遍历迭代器
func (p *PrimaryDataIter) For(fn func(k, v []byte) any, esc bool, limit ...int) (r []any) {
	//添加锁，防止并发访问
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.top[esc]() {
		return nil
	}
	var rd any
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
		r = make([]any, 0, count)
	}
	loop := 0
	// 处理当前位置的元素
	for {
		if loop < start {
			loop++
		} else {
			rd = fn(p.iter.Key(), p.iter.Value())
			r = append(r, rd)
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

// 遍历迭代器
func (i *IndexDataIter) For(fn func(k, v []byte) any, esc bool, limit ...int) (r []any) {
	//添加锁，防止并发访问
	i.mu.Lock()
	defer i.mu.Unlock()

	if !i.top[esc]() {
		return nil
	}
	var rd any
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
		r = make([]any, 0, count)
	}
	loop := 0
	// 处理当前位置的元素
	for {
		if loop < start {
			loop++
		} else {
			rd = fn(i.iter.Key(), i.iter.Value())
			r = append(r, rd)
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

// 从迭代器中提取索引的数据集
func (i *IndexDataIter) GetRecord(esc bool, limit ...int) (r Records) {
	//添加锁，防止并发访问
	i.mu.Lock()
	defer i.mu.Unlock()

	if !i.top[esc]() {
		return nil
	}
	var record Record
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
			record = i.Parse(i.iter.Value())
			record = record.GetKeys(i.keys...)
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
*/
