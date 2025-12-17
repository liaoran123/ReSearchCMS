package storedb

import (
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

// 返回表数据或索引数据
type Iter struct {
	iter  iterator.Iterator
	move  map[bool]func() bool
	top   map[bool]func() bool
	table *Table
}

func IterNew(iter iterator.Iterator, table *Table) *Iter {
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
		table: table,
	}
}

/*
// 因为每次for之后会释放迭代器，所以需要在第二次for之前设置迭代器
func (t *Iter) Setiter(iter iterator.Iterator) {
	t.iter = iter
}
*/
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
func (t *Iter) ForRecord(isIndex bool, esc bool, limit ...int) (ret []map[string]any) {
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
		ret = make([]map[string]any, 0, count)
	}
	//var val []byte
	loop := 0
	// 处理当前位置的元素
	for {
		if loop < start {
			loop++
		} else {
			// 正确接收ConvertRecord返回的record和key值
			record, _ := t.ConvertRecord(isIndex, t.iter.Key(), t.iter.Value())
			if record == nil {
				continue
			}
			ret = append(ret, record)
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

// 遍历数据获取主键值的map[any]bool集合，作为数据并集，交集等
// isIndex为true时，是索引数据，false时，是主键记录数据
func (t *Iter) ToMap(isIndex bool) (ret map[any]bool) {
	// 检查迭代器和表是否有效
	if t.iter == nil || t.table == nil {
		return nil
	}

	// 检查主键字段是否存在
	if t.table.primary == "" {
		return nil
	}

	pytype, exists := t.table.fields[t.table.primary]
	if !exists {
		return nil
	}

	// 将迭代器重置到开头
	if !t.iter.First() {
		return nil
	}

	// 预分配map，初始容量设为100，可根据实际情况调整
	ret = make(map[any]bool, 100)

	var k, v, rk []byte
	var prefix []byte

	// 只计算一次前缀，避免重复计算
	if !isIndex {
		prefix = []byte(t.table.GetPrimaryPrefix() + SPLIT)
	}

	// 遍历所有数据
	for {
		k = t.iter.Key()
		v = t.iter.Value()

		if !isIndex { // 主键记录
			// 检查切片长度，避免越界
			if len(k) <= len(prefix) {
				// 移动到下一个元素
				if !t.iter.Next() {
					break
				}
				continue
			}
			rk = k[len(prefix):]
			ret[Bytes(rk).ToAny(pytype)] = true
		} else { // 索引数据
			ret[StrToAny(string(v), pytype)] = true
		}

		// 移动到下一个元素
		if !t.iter.Next() {
			break
		}
	}

	return ret
}

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

// 主键交集
// 求所有数据集的交集，即输出所有数据集均包含的记录
func (t *Iter) IntersectPrimary(start int, count int, others ...map[any]bool) (ret []map[string]any) {
	return t.Intersect(false, start, count, others...)
}

// 索引交集
// 求所有索引数据集的交集，即输出所有索引数据集均包含的记录
func (t *Iter) IntersectIndex(start int, count int, others ...map[any]bool) (ret []map[string]any) {
	return t.Intersect(true, start, count, others...)
}

/*
数据库对两个或多个结果集进行合并、取重、剔除操作时，可以通过UNION、INTERSECT、EXCEPT来实现。

您可以通过MaxCompute对查询结果数据集执行取交集、并集或补集操作。本文为您介绍交集（INTERSECT、INTERSECT ALL、INTERSECT DISTINCT）、
并集（UNION、UNION ALL、UNION DISTINCT）和
补集（EXCEPT、EXCEPT ALL、EXCEPT DISTINCT、MINUS、MINUS ALL、MINUS DISTINCT）的使用方法。

交集：求两个数据集的交集，即输出两个数据集均包含的记录。

并集：求两个数据集的并集，即将两个数据集合并成一个数据集。

补集：求第二个数据集在第一个数据集中的补集，即输出第一个数据集包含而第二个数据集不包含的记录。
*/

// 遍历数据获取主键值的map[any]bool集合，作为数据交集等
// isIndex为true时，是索引数据，false时，是主键记录数据
// start为遍历的起始位置，count为遍历的范围，0表示从当前位置开始遍历，1表示从当前位置开始遍历，count个元素
// 2个参数表示从start位置开始遍历，count个元素
// others为其他map[any]bool集合，用于求交集等
/*
交集（∩）：

定义：两个或多个集合中共有的元素组成的集合称为这些集合的交集。
表示方法：对于任意两个集合A和B，它们的交集表示为A∩B。
实例：如果A={1,2,3}，B={2,3,4}，则A∩B={2,3}。
a AND b = a∩b={2,3}。
*/
func (t *Iter) Intersect(isIndex bool, start int, count int, others ...map[any]bool) (ret []map[string]any) {
	if t.iter == nil {
		return nil
	}
	if !t.iter.First() {
		return nil
	}

	// 预分配结果切片容量，避免频繁扩容
	if count > 0 {
		ret = make([]map[string]any, 0, count)
	} else {
		ret = make([]map[string]any, 0)
	}

	var k, v []byte
	loop := 0
	isOthersEmpty := len(others) == 0

	// 主循环：遍历所有数据
	for {
		// 每次循环开始时初始化record为nil，避免变量重用问题
		var record map[string]any
		var key any
		var match bool = true

		// 检查是否达到起始位置
		if loop >= start {
			k = t.iter.Key()
			v = t.iter.Value()
			// 正确接收ConvertRecord返回的record和key值
			record, key = t.ConvertRecord(isIndex, k, v)
			// 检查是否需要进行交集比较
			if !isOthersEmpty && record != nil {
				// 检查是否存在于所有其他map中
				for _, m := range others {
					if _, ok := m[key]; !ok {
						match = false
						break // 提前终止，只要有一个不匹配就停止检查
					}
				}
			}

			// 如果匹配且记录有效，添加到结果中
			if (isOthersEmpty || match) && record != nil {
				ret = append(ret, record)

				// 检查是否达到数量限制
				if count > 0 && len(ret) >= count {
					break
				}
			}
		}

		loop++

		// 移动到下一个元素，如果没有更多元素，退出循环
		if !t.iter.Next() {
			break
		}
	}

	return ret
}

// 主键并集
// 求所有数据集的并集，即将所有数据集合并成一个数据集
func (t *Iter) UnionPrimary(start int, count int, others ...map[any]bool) (ret []map[string]any) {
	return t.Union(false, start, count, others...)
}

// 索引并集
// 求所有索引数据集的并集，即将所有索引数据集合并成一个数据集
func (t *Iter) UnionIndex(start int, count int, others ...map[any]bool) (ret []map[string]any) {
	return t.Union(true, start, count, others...)
}

// 遍历数据获取主键值的map[any]bool集合，作为数据并集等
// isIndex为true时，是索引数据，false时，是主键记录数据
// start为遍历的起始位置，count为遍历的范围，0表示从当前位置开始遍历，1表示从当前位置开始遍历，count个元素
// 2个参数表示从start位置开始遍历，count个元素
// others为其他map[any]bool集合，用于求并集等
/*
并集（∪）：

定义：由所有属于给定集合中的元素所构成的集合称为这些集合的并集。
表示方法：对于任意两个集合A和B，它们的并集表示为A∪B。
实例：如果A={1,2,3}，B={2,3,4}，则A∪B={1,2,3,4}。
a OR b = a∪b={1,2,3,4}。
*/
func (t *Iter) Union(isIndex bool, start int, count int, others ...map[any]bool) (ret []map[string]any) {
	if t.iter == nil {
		return nil
	}
	if !t.iter.First() {
		return nil
	}
	// 预分配结果切片容量，避免频繁扩容
	if count > 0 {
		ret = make([]map[string]any, 0, count)
	} else {
		ret = make([]map[string]any, 0)
	}

	var k, v []byte
	loop := 0
	isOthersEmpty := len(others) == 0

	// 用于跟踪已添加到结果中的主键，避免重复
	addedKeys := make(map[any]bool)

	// 主循环：遍历所有数据
	for {
		// 每次循环开始时初始化record为nil，避免变量重用问题
		var record map[string]any
		var key any
		var match bool

		// 检查是否达到起始位置
		if loop >= start {
			k = t.iter.Key()
			v = t.iter.Value()
			// 正确接收ConvertRecord返回的record和key值
			record, key = t.ConvertRecord(isIndex, k, v)
			// 检查记录是否有效且未被添加过
			if record != nil && !addedKeys[key] {
				// 并集逻辑：如果没有其他map或在任何一个map中存在
				if isOthersEmpty {
					match = true
				} else {
					// 检查是否存在于任何一个其他map中
					match = false
					for _, m := range others {
						if _, ok := m[key]; ok {
							match = true
							break // 提前终止，只要有一个匹配就停止检查
						}
					}
				}

				// 如果匹配，添加到结果中并标记为已添加
				if match {
					ret = append(ret, record)
					addedKeys[key] = true

					// 检查是否达到数量限制
					if count > 0 && len(ret) >= count {
						break
					}
				}
			}
		}

		loop++

		// 移动到下一个元素，如果没有更多元素，退出循环
		if !t.iter.Next() {
			break
		}
	}

	return ret
}

// 遍历数据获取主键值的map[any]bool集合，作为数据补集等
// isIndex为true时，是索引数据，false时，是主键记录数据
// start为遍历的起始位置，count为遍历的范围，0表示从当前位置开始遍历，1表示从当前位置开始遍历，count个元素
// 2个参数表示从start位置开始遍历，count个元素
// others为其他map[any]bool集合，用于求补集等
// 补集：求第一个数据集包含而其他数据集不包含的记录
/*
# 定义两个集合
A = {1, 2, 3, 4}
B = {3, 4, 5, 6}

# 计算差集
difference = A - B

# 输出差集
print(difference) # 输出: {1, 2}
复制
通过上述代码，我们可以轻松地计算出集合A与集合B的差集。

重要注意事项

在计算差集时，需要注意以下几点：

差集操作是非对称的，即A - B ≠ B - A。例如，B - A = {5, 6}，因为5和6只存在于集合B中，而不在集合A中。

差集操作不会改变原始集合的内容，而是返回一个新的集合。
*/
func (t *Iter) Except(isIndex bool, start int, count int, others ...map[any]bool) (ret []map[string]any) {
	if t.iter == nil {
		return nil
	}
	if !t.iter.First() {
		return nil
	}
	// 预分配结果切片容量，避免频繁扩容
	if count > 0 {
		ret = make([]map[string]any, 0, count)
	} else {
		ret = make([]map[string]any, 0)
	}

	var k, v []byte
	loop := 0
	isOthersEmpty := len(others) == 0

	// 用于跟踪已添加到结果中的主键，避免重复
	addedKeys := make(map[any]bool)

	// 主循环：遍历所有数据
	for {
		// 每次循环开始时初始化record为nil，避免变量重用问题
		var record map[string]any
		var key any
		var match bool

		// 检查是否达到起始位置
		if loop >= start {
			k = t.iter.Key()
			v = t.iter.Value()
			// 正确接收ConvertRecord返回的record和key值
			record, key = t.ConvertRecord(isIndex, k, v)
			// 检查记录是否有效且未被添加过
			if record != nil && !addedKeys[key] {
				// 补集逻辑：如果没有其他map或不在任何一个map中存在
				if isOthersEmpty {
					match = true
				} else {
					// 检查是否不存在于所有其他map中
					match = true
					for _, m := range others {
						if _, ok := m[key]; ok {
							match = false
							break // 提前终止，只要在一个map中存在就不满足补集条件
						}
					}
				}

				// 如果匹配，添加到结果中并标记为已添加
				if match {
					ret = append(ret, record)
					addedKeys[key] = true

					// 检查是否达到数量限制
					if count > 0 && len(ret) >= count {
						break
					}
				}
			}
		}

		loop++

		// 移动到下一个元素，如果没有更多元素，退出循环
		if !t.iter.Next() {
			break
		}
	}

	return ret
}

// 主键差集
// 求第一个数据集包含而其他数据集不包含的记录
func (t *Iter) ExceptPrimary(start int, count int, others ...map[any]bool) (ret []map[string]any) {
	return t.Except(false, start, count, others...)
}

// 索引差集
// 求第一个索引数据集包含而其他索引数据集不包含的记录
func (t *Iter) ExceptIndex(start int, count int, others ...map[any]bool) (ret []map[string]any) {
	return t.Except(true, start, count, others...)
}

// 区别主键和索引的查询结果转换记录为map[string]any类型的函数
// 返回转换后的记录和对应的主键值
func (t *Iter) ConvertRecord(isIndex bool, k, v []byte) (map[string]any, any) {
	// 处理nil输入
	if k == nil || v == nil {
		return nil, nil
	}

	// 确保主键字段存在
	if t.table.primary == "" {
		return nil, nil
	}

	pytype, exists := t.table.fields[t.table.primary]
	if !exists {
		return nil, nil
	}

	var key any
	var record map[string]any
	if !isIndex { // 主键记录：直接解析value
		prefix := t.table.GetPrimaryPrefix() + SPLIT
		// 确保k的长度大于prefix的长度
		if len(k) <= len(prefix) {
			return nil, nil
		}
		key = Bytes(k[len(prefix):]).ToAny(pytype)
		record = t.table.ParseValue(v)
	} else { // 索引数据：通过主键读取记录
		key = Bytes(v).ToAny(pytype)
		byrecord := t.table.Read(key)
		if byrecord != nil {
			record = t.table.ParseValue(byrecord)
		}
	}
	return record, key
}
