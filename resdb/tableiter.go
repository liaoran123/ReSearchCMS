package resdb

import (
	"sync"

	"github.com/syndtr/goleveldb/leveldb/iterator"
)

// 基础迭代器结构体，提取公共字段和方法
type baseIter struct {
	table      *Table
	showfields []string
	index      Index //搜索时使用的索引
	move       map[bool]func() bool
	top        map[bool]func() bool
	mu         sync.Mutex
}

// 实现sql语句中的select f0,f1,... from table 要返回的字段
// 如果keys为空，则返回所有字段
// 在最底层转换，最大化减少内存占用
type TableIter struct {
	baseIter
	iter iterator.Iterator
}

// 解析函数
type Parser func(k, v []byte) Record

// 导出函数
type Export func(k, v []byte) bool

func TableIterNew(table *Table, iter iterator.Iterator, index Index, showfields ...string) *TableIter {
	return &TableIter{
		baseIter: baseIter{
			table:      table,
			showfields: showfields,
			index:      index,
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

// 将主键转为字符串 ，主要是用于组合主键需要进行map比较时使用
func (t *TableIter) PrimaryToStr(k []byte) string {
	/*
		isComposite := len(t.table.primary) > 1
		if !isComposite {
			return ""
		}*/
	var key any
	rstr := ""
	for _, p := range t.index.GetFields() {
		key = Bytes(k).ToAny(t.table.fields[p])
		rstr += string(AnyToBytes(key)) + SPLIT
	}
	return rstr[:len(rstr)-1]
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

// 分页变量
type Page struct {
	Start int
	Count int
}

func PageNew(No ...int) Page {
	start := 0
	count := -1
	switch len(No) {
	case 0:
	case 1:
		count = No[0]
	default:
		start = No[0]
		count = No[1]
	}
	return Page{
		Start: start,
		Count: count,
	}
}

// 遍历迭代器返回解析后的记录
// 单个简单查询直接使用
// handler可以对value进行各种处理
func (t *TableIter) GerRecords(esc bool, limit ...int) (r Records) {
	//添加锁，防止并发访问
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.top[esc]() {
		return nil
	}
	var rd Record

	page := PageNew(limit...)
	var recordGetFun RecordGetFun

	switch t.index.(type) { //需要区别主键和索引，根据handler命名规则
	case PrimaryKey:
		recordGetFun = ByPrimaryGet
	case NormalIndex, FullTextIndex:
		recordGetFun = ByIndexGet
	}
	if page.Count > 0 {
		r = make(Records, 0, page.Count)
	}

	if recordGetFun == nil {
		return nil
	}
	loop := 0
	// 处理当前位置的元素
	for {
		if loop < page.Start {
			loop++
		} else {

			rd = recordGetFun(t.table, t.iter.Value())
			rd = rd.GetKeys(t.showfields...)
			r = append(r, rd)
			if page.Count > 0 && len(r) >= page.Count {
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

// 遍历迭代器返回解析后的记录
// 复杂组合查询使用
func (t *TableIter) ForMatchRecord(esc bool, page Page, match ...MatchRule) (r Records) {
	//添加锁，防止并发访问
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.top[esc]() {
		return nil
	}
	var rd Record
	if page.Count > 0 {
		r = make(Records, 0, page.Count)
	}
	loop := 0
	// 处理当前位置的元素
	for {
		if loop < page.Start {
			loop++
		} else {
			for _, m := range match {
				if m.Match(t.table, t.iter.Value()) {
					r = append(r, rd)
					if page.Count > 0 && len(r) >= page.Count {
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
// export导出函数，返回false则停止导出
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

// 提取转换主键值
// 单主键则返回转换的值，组合主键则返回字符串
// field用于对组合主键提取其中的部分值
func (t *TableIter) GetPrimaryKeys(field ...string) (r any) {
	//是否组合主键
	isComposite := len(t.table.primaryKey.GetFields()) > 1
	switch isComposite {
	case true:
		// 组合主键，返回字符串
		keys := Bytes(t.iter.Key()).Split()
		if len(field) > 0 {
			newkeys := ""
			for _, f := range field {
				for i, p := range t.table.primaryKey.GetFields() {
					if f == p {
						newkeys += string(keys[i]) + SPLIT
					}
				}
			}
			r = newkeys[:len(newkeys)-1]
		}
	default:
		// 单主键，返回转换的值
		r = Bytes(t.iter.Value()).ToAny(t.table.fields[t.table.primaryKey.GetFields()[0]])
	}
	return r
}
