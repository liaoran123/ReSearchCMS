package storedb

import (
	"encoding/json/v2"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb/iterator"
)

// 返回表数据或索引数据
type TableData struct {
	iter iterator.Iterator
	fn   map[bool]func() bool
	top  map[bool]func() bool
}

func TableDataNew(iter iterator.Iterator) *TableData {
	return &TableData{
		iter: iter,
		fn: map[bool]func() bool{
			true:  iter.Next,
			false: iter.Prev,
		},
		top: map[bool]func() bool{
			true:  iter.First,
			false: iter.Last,
		},
	}
}

// 因为每次for之后会释放迭代器，所以需要在第二次for之前设置迭代器
func (t *TableData) Setiter(iter iterator.Iterator) {
	t.iter = iter
}
func (t *TableData) For(esc bool, limit ...int) (ret []any) {
	if !t.top[esc]() {
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
	loop := 0

	// 处理当前位置的元素
	for {
		if loop < start {
			loop++
		} else {
			var val any
			err := json.Unmarshal(t.iter.Value(), &val)
			if err != nil {
				fmt.Printf("字段反序列化失败: %v\n", err)
				continue
			}
			ret = append(ret, val)
			if count > 0 && len(ret) >= count {
				break
			}
		}
		// 移动到下一个/前一个元素
		if !t.fn[esc]() {
			break
		}
	}

	// 释放迭代器
	t.iter.Release()
	return ret
}
func (t *TableData) First() (key []byte, value []byte) {
	if !t.iter.First() {
		return nil, nil
	}
	return t.iter.Key(), t.iter.Value()
}
func (t *TableData) Last() (key []byte, value []byte) {
	if !t.iter.Last() {
		return nil, nil
	}
	return t.iter.Key(), t.iter.Value()
}
func (t *TableData) Next() (key []byte, value []byte) {
	if !t.iter.Next() {
		return nil, nil
	}
	return t.iter.Key(), t.iter.Value()
}
func (t *TableData) Prev() (key []byte, value []byte) {
	if !t.iter.Prev() {
		return nil, nil
	}
	return t.iter.Key(), t.iter.Value()
}
func (t *TableData) Release() {
	t.iter.Release()
}
