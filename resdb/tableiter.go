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

// 导出函数
type Export func(k, v []byte) bool

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

// 遍历迭代器返回解析后的记录
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
	// 读取完整记录
	byrecord := t.table.ReadByBytes(v)
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

// 遍历迭代器进行匹配
func (t *TableIter) ForMatch(esc bool, start, count int, match ...MatchRule) (r []any) {
	//添加锁，防止并发访问
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.top[esc]() {
		return nil
	}
	var md any
	if count > 0 {
		r = make([]any, 0, count)
	}
	loop := 0
	// 处理当前位置的元素
	for {
		if loop < start {
			loop++
		} else {
			for _, m := range match {
				//根据m.Fields获取rd的对应字段值，待续.............
				if len(m.Fields) == 0 { //默认是主键值
					md = Bytes(t.iter.Value()).ToAny(t.table.fields[t.table.primary[0]])
				}
				if m.Match(md) == m.Rule {
					r = append(r, md)
					if count > 0 && len(r) >= count {
						break
					}
				}
			}
		}
		// 移动到下一个/前一个元素
		if !t.move[esc]() {
			break
		}
	}
	return r
}

// 遍历迭代器导出数据
func (t *TableIter) ForExport(esc bool, export Export) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.top[esc]() {
		return
	}
	for {
		if !export(t.iter.Key(), t.iter.Value()) {
			return
		}
		if !t.move[esc]() {
			return
		}
	}
}
