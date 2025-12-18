package storedb

import (
	"fmt"
	"testing"
)

// TestSearch 测试Search函数的基本功能
func TestSearch(t *testing.T) {
	// 创建一个Table实例用于测试
	table, err := TableNew("test_search")
	if err != nil {
		t.Fatalf("TableNew 失败: %v", err)
	}
	if table == nil {
		t.Fatal("TableNew 失败")
	}

	// 设置主键
	table.primary = "id"

	// 设置普通索引
	table.index = [][]string{
		{"name"},        // 单列索引
		{"age", "city"}, // 组合索引
	}

	//预设表的字段和类型
	table.fields["id"] = 0
	table.fields["name"] = "0"
	table.fields["age"] = uint8(0)
	table.fields["city"] = ""

	fields := table.GetAllFields()
	table.Insert(&fields)
	// 测试1: 使用主键索引搜索
	// 测试Search函数是否能正常返回TableData
	fields = map[string]any{
		"id": 1,
	}
	td := table.Search(&fields)
	if td.iter != nil {
		td.Release() // 释放资源
	}

	// 测试2: 使用普通索引搜索
	fields = map[string]any{
		"name": "test",
	}
	t1 := table.Search(&fields)
	if t1.iter != nil {
		t1.Release() // 释放资源
	}

	// 测试3: 使用组合索引搜索
	fields = map[string]any{
		"age":  18,
		"city": "New York",
	}
	t2 := table.Search(&fields)
	if t2.iter != nil {
		t2.Release() // 释放资源
	}
}

// TestSearchCache 测试Search函数的缓存机制
func TestSearchCache(t *testing.T) {
	// 创建一个Table实例用于测试
	table, err := TableNew("test_search_cache")
	if err != nil {
		t.Fatalf("TableNew 失败: %v", err)
	}
	if table == nil {
		t.Fatal("TableNew 失败")
	}

	// 设置主键
	table.primary = "id"

	//预设表的字段和类型
	table.fields["id"] = 0
	table.fields["name"] = "0"
	table.fields["age"] = uint8(0)
	table.fields["city"] = ""
	fields := table.GetAllFields()
	table.Insert(&fields)
	// 设置普通索引
	table.index = [][]string{
		{"name"}, // 单列索引
	}

	// 设置字段值
	table.fields["id"] = 1
	table.fields["name"] = "test"

	// 第一次调用Search，应该创建新的TableData对象
	fields = map[string]any{
		"id": 1,
	}
	td1 := table.Search(&fields)
	if td1.iter == nil {
		t.Fatal("第一次调用Search返回nil")
	}

	// 第二次调用Search，应该从缓存中获取TableData对象
	td2 := table.Search(&fields)
	if td2.iter == nil {
		t.Fatal("第二次调用Search返回nil")
	}
	rs := td2.For(true)
	fmt.Printf("rs: %v\n", rs)
	// 释放资源
	td1.Release()
}

// TestSearchNoIndex 测试Search函数在没有匹配索引时的表现
func TestSearchNoIndex(t *testing.T) {
	// 创建一个Table实例用于测试
	table, err := TableNew("test_search_no_index")
	if err != nil {
		t.Fatalf("TableNew 失败: %v", err)
	}
	if table == nil {
		t.Fatal("TableNew 失败")
	}

	// 设置主键
	table.primary = "id"

	//预设表的字段和类型
	table.fields["id"] = 0
	table.fields["name"] = "0"
	table.fields["age"] = uint8(0)
	table.fields["city"] = ""
	fields := table.GetAllFields()
	table.Insert(&fields)
	// 不设置普通索引
	table.index = nil

	// 设置字段值
	table.fields["id"] = 1
	table.fields["name"] = "test"

	// 测试Search函数在没有匹配索引时的表现
	fields = map[string]any{
		"name": "test",
	}
	td := table.Search(&fields)
	if td.iter != nil {
		t.Errorf("Search函数在没有匹配索引时应该返回nil，实际返回了%v", td)
	}
}
