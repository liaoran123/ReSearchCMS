package storedb

import (
	"testing"
)

// TestTableDataFor 测试TableData.For方法的功能
func TestTableDataFor(t *testing.T) {
	// 初始化测试数据
	table, err := TableNew("test_tabledata_for")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	
	// 设置主键
	table.primary = "id"
	
	// 插入多条测试数据
	for i := 1; i <= 5; i++ {
		table.SetField("id", i)
		table.SetField("name", "User"+string(rune('0'+i)))
		table.SetField("age", 20+i)
		table.SetField("city", "City"+string(rune('A'+i-1)))
		
		// 设置主键值
		table.SetPrimaryValue("id")
		
		// 插入数据
		if err := table.Insert(); err != nil {
			t.Fatalf("插入测试数据失败: %v", err)
		}
	}
	
	// 获取迭代器
	prefix := []byte("test_tabledata_for-pk-")
	iter := RsDB.GetIterator(prefix)
	if iter == nil {
		t.Fatal("获取迭代器失败")
	}
	
	// 创建TableData实例
	tableData := TableDataNew(iter, table)
	if tableData == nil {
		t.Fatal("创建TableData实例失败")
	}
	defer tableData.Release()
	
	// 测试1: 正向遍历所有数据
	result := tableData.For(true)
	if len(result) != 5 {
		t.Errorf("正向遍历所有数据失败，期望5条，实际%v条", len(result))
	}
	
	// 测试2: 正向遍历前3条数据
	result = tableData.For(true, 3)
	if len(result) != 3 {
		t.Errorf("正向遍历前3条数据失败，期望3条，实际%v条", len(result))
	}
	
	// 测试3: 正向从第2条开始遍历前2条数据
	result = tableData.For(true, 1, 2)
	if len(result) != 2 {
		t.Errorf("正向从第2条开始遍历前2条数据失败，期望2条，实际%v条", len(result))
	}
	
	// 测试4: 反向遍历所有数据
	result = tableData.For(false)
	if len(result) != 5 {
		t.Errorf("反向遍历所有数据失败，期望5条，实际%v条", len(result))
	}
	
	// 测试5: 反向遍历前2条数据
	result = tableData.For(false, 2)
	if len(result) != 2 {
		t.Errorf("反向遍历前2条数据失败，期望2条，实际%v条", len(result))
	}
	
	// 测试6: 反向从第1条开始遍历前3条数据
	result = tableData.For(false, 0, 3)
	if len(result) != 3 {
		t.Errorf("反向从第1条开始遍历前3条数据失败，期望3条，实际%v条", len(result))
	}
}

// TestTableDataForEmpty 测试TableData.For方法处理空数据的情况
func TestTableDataForEmpty(t *testing.T) {
	// 初始化测试数据
	table, err := TableNew("test_tabledata_for_empty")
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	
	// 设置主键
	table.primary = "id"
	
	// 获取迭代器（此时没有数据）
	prefix := []byte("test_tabledata_for_empty-pk-")
	iter := RsDB.GetIterator(prefix)
	if iter == nil {
		t.Fatal("获取迭代器失败")
	}
	
	// 创建TableData实例
	tableData := TableDataNew(iter, table)
	if tableData == nil {
		t.Fatal("创建TableData实例失败")
	}
	defer tableData.Release()
	
	// 测试正向遍历空数据
	result := tableData.For(true)
	if result != nil {
		t.Errorf("正向遍历空数据失败，期望nil，实际%v", result)
	}
	
	// 测试反向遍历空数据
	result = tableData.For(false)
	if result != nil {
		t.Errorf("反向遍历空数据失败，期望nil，实际%v", result)
	}
}
