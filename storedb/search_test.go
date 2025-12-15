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

	// 测试1: 使用主键索引搜索
	table.fields["id"] = 1
	table.fields["name"] = "test"
	table.fields["age"] = 18
	table.fields["city"] = "New York"

	// 测试Search函数是否能正常返回TableData
	td := table.Search("id")
	if td != nil {
		td.Release() // 释放资源
	}

	// 测试2: 使用普通索引搜索
	t1 := table.Search("name")
	if t1 != nil {
		t1.Release() // 释放资源
	}

	// 测试3: 使用组合索引搜索
	t2 := table.Search("age", "city")
	if t2 != nil {
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

	// 设置普通索引
	table.index = [][]string{
		{"name"}, // 单列索引
	}

	// 设置字段值
	table.fields["id"] = 1
	table.fields["name"] = "test"

	// 第一次调用Search，应该创建新的TableData对象
	td1 := table.Search("id")
	if td1 == nil {
		t.Fatal("第一次调用Search返回nil")
	}

	// 第二次调用Search，应该从缓存中获取TableData对象
	td2 := table.Search("id")
	if td2 == nil {
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

	// 不设置普通索引
	table.index = nil

	// 设置字段值
	table.fields["id"] = 1
	table.fields["name"] = "test"

	// 测试Search函数在没有匹配索引时的表现
	td := table.Search("name")
	if td != nil {
		t.Errorf("Search函数在没有匹配索引时应该返回nil，实际返回了%v", td)
	}
}
