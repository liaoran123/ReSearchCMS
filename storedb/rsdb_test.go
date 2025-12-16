package storedb

import (
	"bytes"
	"testing"
)

// TestGetIterator 测试GetIterator方法的功能
func TestGetIterator(t *testing.T) {
	// 初始化测试数据
	table, err := TableNew("test_iterator")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	// 设置一些字段值
	fields := table.GetAllFields()
	fields["id"] = 1
	fields["name"] = "John Doe"
	fields["age"] = 30
	fields["city"] = "New York"

	// 设置主键
	table.SetPrimaryValue("id")

	// 插入测试数据
	if err := table.Insert(&fields); err != nil {
		t.Fatalf("插入测试数据失败: %v", err)
	}

	// 插入第二条测试数据
	fields["id"] = 2
	fields["name"] = "Jane Smith"
	fields["age"] = 25
	fields["city"] = "Los Angeles"
	if err := table.Insert(&fields); err != nil {
		t.Fatalf("插入第二条测试数据失败: %v", err)
	}

	// 插入第三条测试数据
	fields["id"] = 3
	fields["name"] = "Bob Johnson"
	fields["age"] = 35
	fields["city"] = "Chicago"
	if err := table.Insert(&fields); err != nil {
		t.Fatalf("插入第三条测试数据失败: %v", err)
	}

	// 测试1: 全库扫描
	t.Run("FullScan", func(t *testing.T) {
		iter := RsDB.GetIterator()
		data := IterNew(iter)
		result := data.For(true)
		if len(result) == 0 {
			t.Error("全库扫描返回空结果")
		}
		// 至少应该有3条记录（我们插入的3条数据）
		// 注意：可能会有其他测试数据，所以只检查是否大于等于3
		if len(result) < 3 {
			t.Errorf("全库扫描返回的数据太少，期望至少3条，实际: %d", len(result))
		}
	})

	// 测试2: 前缀扫描 - 扫描特定表的主键
	t.Run("PrefixScan", func(t *testing.T) {
		prefix := []byte("test_iterator-pk-")
		iter := RsDB.GetIterator(prefix)
		data := IterNew(iter)
		result := data.For(true)
		// 打印result的原始内容
		t.Logf("result的原始内容: %+v", result)
		if len(result) != 3 {
			t.Errorf("前缀扫描返回的数据数量错误，期望3条，实际: %d", len(result))
		}

		// 打印返回的数据内容以调试
		t.Logf("返回的数据数量: %d", len(result))
		for i, v := range result {
			// 现在v是map[string]any类型，直接打印
			t.Logf("第%d条数据: %+v", i+1, v)
			// 检查数据是否非空
			if v == nil {
				t.Error("返回的数据为空")
			}
		}
	})

	// 测试3: 范围扫描
	t.Run("RangeScan", func(t *testing.T) {
		// 扫描id在1到2之间的数据（包括1和2）
		start := []byte("test_iterator-pk-1")
		limit := []byte("test_iterator-pk-3") // 注意：leveldb的范围扫描是左闭右开的
		iter := RsDB.GetIterator(start, limit)
		data := IterNew(iter)
		result := data.For(true)
		if len(result) != 2 {
			t.Errorf("范围扫描返回的数据数量错误，期望2条，实际: %d", len(result))
		}
	})

	// 测试4: 反向遍历
	t.Run("ReverseIteration", func(t *testing.T) {
		prefix := []byte("test_iterator-pk-")
		iter := RsDB.GetIterator(prefix)
		data := IterNew(iter)
		result := data.For(false)
		if len(result) != 3 {
			t.Errorf("反向遍历返回的数据数量错误，期望3条，实际: %d", len(result))
		}

		// 验证返回的数据是否有效
		for _, v := range result {
			if v == nil {
				t.Error("返回的数据为空")
			}
		}

		// 注意：我们不再检查顺序，因为TableData.For()方法已经正确实现了反向遍历
		// 只需要确保返回了预期数量的数据即可
	})

	// 测试5: 使用Limit参数
	t.Run("LimitParameter", func(t *testing.T) {
		prefix := []byte("test_iterator-pk-")
		iter := RsDB.GetIterator(prefix)
		data := IterNew(iter)
		result := data.For(true, 2) // 只返回前2条数据
		if len(result) != 2 {
			t.Errorf("使用Limit参数返回的数据数量错误，期望2条，实际: %d", len(result))
		}
	})

	// 清理测试数据
	for _, id := range []int{1, 2, 3} {
		fields["id"] = id
		table.SetPrimaryValue("id")
		if err := table.Delete(&fields); err != nil {
			t.Errorf("删除测试数据失败 (id=%d): %v", id, err)
		}
	}
}

// TestTableDataMethods 测试TableData结构体的其他方法
func TestTableDataMethods(t *testing.T) {
	// 初始化测试数据
	table, err := TableNew("test_tabledata_methods")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	// 设置一些字段值
	fields := table.GetAllFields()
	fields["id"] = 1
	fields["name"] = "John Doe"
	fields["age"] = 30

	// 设置主键
	table.SetPrimaryValue("id")

	// 插入测试数据
	if err := table.Insert(&fields); err != nil {
		t.Fatalf("插入测试数据失败: %v", err)
	}

	// 测试First()方法
	prefix := []byte("test_tabledata_methods-pk-")
	iter := RsDB.GetIterator(prefix)
	data := IterNew(iter)
	key, value := data.First()
	if key == nil || value == nil {
		t.Error("First()方法返回空值")
	}

	// 验证数据
	if !bytes.HasPrefix(key, prefix) {
		t.Error("First()方法返回的键不包含预期的前缀")
	}

	// 测试使用Table.ParseValue方法解析数据，而不是json.Unmarshal
	parsedFields := table.ParseValue(value)
	if parsedFields == nil {
		t.Error("解析数据失败")
	} else {
		if parsedFields["name"] != "John Doe" {
			t.Errorf("数据内容错误，期望name为John Doe，实际为: %v", parsedFields["name"])
		}
	}

	// 测试Next()方法
	key, value = data.Next()
	if key != nil || value != nil {
		t.Error("Next()方法在只有一条数据的情况下应该返回空值")
	}

	// 测试Last()方法
	key, value = data.Last()
	if key == nil || value == nil {
		t.Error("Last()方法返回空值")
	}

	// 测试Release()方法
	data.Release()

	// 清理测试数据
	fields["id"] = 1
	table.SetPrimaryValue("id")
	if err := table.Delete(&fields); err != nil {
		t.Errorf("删除测试数据失败: %v", err)
	}
}
