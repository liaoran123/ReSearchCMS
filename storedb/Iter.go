package storedb

import (
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

// 返回表数据或索引数据
type Iter struct {
	iter iterator.Iterator
	move map[bool]func() bool
	top  map[bool]func() bool
}

func IterNew(iter iterator.Iterator) *Iter {
	if iter == nil {
		return nil
	}
	return &Iter{
		iter: iter,
		move: map[bool]func() bool{
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
func (t *Iter) Setiter(iter iterator.Iterator) {
	t.iter = iter
}

// 遍历数据，esc为true时，从前往后遍历，false时，从后往前遍历
// limit为遍历的范围，0表示从当前位置开始遍历，1表示从当前位置开始遍历，count个元素
// 2个参数表示从start位置开始遍历，count个元素
func (t *Iter) For(esc bool, limit ...int) (ret [][]byte) {
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
	if count > 0 {
		ret = make([][]byte, 0, count)
	}
	//var val []byte
	loop := 0
	// 处理当前位置的元素
	for {
		if loop < start {
			loop++
		} else {
			/*
				在Golang中，当使用 append 将同一个slice引用多次添加到 [][]byte 切片时，
				会出现所有元素都是最后一个值的问题。
				这是因为slice是引用类型，所有append的元素实际上指向同一个底层数组
				方法1：
				copyVal := make([]byte, len(val))
				copy(copyVal, val)
				ret = append(ret, copyVal)
			*/
			// 方法2：
			// 创建值的副本，避免引用同一底层数组
			ret = append(ret, append([]byte(nil), t.iter.Value()...))
			if count > 0 && len(ret) >= count {
				break
			}
		}
		// 移动到下一个/前一个元素
		if !t.move[esc]() {
			break
		}
	}
	// 释放迭代器
	return ret
}

// 遍历数据，esc为true时，从前往后遍历，false时，从后往前遍历
// limit为遍历的范围，0表示从当前位置开始遍历，1表示从当前位置开始遍历，count个元素
// 2个参数表示从start位置开始遍历，count个元素
// fn为遍历每个元素时调用数据的函数
func (t *Iter) ForFn(fn func(k, v []byte), esc bool, limit ...int) {
	if !t.top[esc]() {
		return
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
	rcount := 0
	// 处理当前位置的元素
	for {
		if loop < start {
			loop++
		} else {
			fn(t.iter.Key(), t.iter.Value())
			rcount++
			if count > 0 && rcount >= count {
				break
			}
		}
		// 移动到下一个/前一个元素
		if !t.move[esc]() {
			break
		}
	}
	// 释放迭代器
	//t.iter.Release()
}

/*
// 遍历数据获取主键值的map[any]bool集合，作为数据并集，交集等

	func (t *Iter) Map() (ret map[any]bool) {
		if t.iter == nil {
			return nil
		}
		if !t.iter.First() {
			return nil
		}
		ret = make(map[any]bool)
		var k, v, rk []byte
		pytype := t.table.fields[t.table.primary]
		for t.iter.Next() {
			k = t.iter.Key()
			v = t.iter.Value()
			if bytes.Contains(v, []byte{':'}) { // 检查值是否包含冒号，有，则是记录
				rk = k[len(t.table.GetPrimaryPrefix()+SPLIT):]
				ret[Bytes(rk).ToAny(pytype)] = true
			} else { // 索引数据
				ret[StrToAny(string(v), pytype)] = true
			}
		}
		return ret
	}
*/
func (t *Iter) First() (key []byte, value []byte) {
	if !t.iter.First() {
		return nil, nil
	}
	return t.iter.Key(), t.iter.Value()
}
func (t *Iter) Last() (key []byte, value []byte) {
	if !t.iter.Last() {
		return nil, nil
	}
	return t.iter.Key(), t.iter.Value()
}
func (t *Iter) Next() (key []byte, value []byte) {
	if !t.iter.Next() {
		return nil, nil
	}
	return t.iter.Key(), t.iter.Value()
}
func (t *Iter) Prev() (key []byte, value []byte) {
	if !t.iter.Prev() {
		return nil, nil
	}
	return t.iter.Key(), t.iter.Value()
}
func (t *Iter) Release() {
	t.iter.Release()
}
