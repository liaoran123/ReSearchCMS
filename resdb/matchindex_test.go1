package resdb

import (
	"testing"
)

// TestMatchIndex 测试MatchIndex函数的功能
func TestMatchIndex(t *testing.T) {
	// 创建一个Table实例用于测试
	table, err := TableNew("test_matchindex")
	if err != nil {
		t.Fatalf("TableNew 失败: %v", err)
	}
	if table == nil {
		t.Fatal("TableNew 失败")
	}

	// 设置主键
	table.primary = "id"

	// 设置全文索引
	table.fullText = []string{"description", "content"}

	// 设置普通索引
	table.index = [][]string{
		{"name"},           // 单列索引
		{"age", "city"},    // 组合索引
		{"email", "phone"}, // 组合索引
		{"name", "email"},  // 组合索引
	}

	// 测试1: 匹配主键索引
	idx, idxType := table.MatchIndex("id")
	if len(idx) != 1 || idx[0] != "id" || idxType != 0 {
		t.Errorf("主键索引匹配失败，期望: {[id], 0}, 实际: {%v, %d}", idx, idxType)
	}

	// 测试2: 匹配全文索引
	idx, idxType = table.MatchIndex("description")
	if len(idx) != 1 || idx[0] != "description" || idxType != 2 {
		t.Errorf("全文索引匹配失败，期望: {[description], 2}, 实际: {%v, %d}", idx, idxType)
	}

	idx, idxType = table.MatchIndex("content")
	if len(idx) != 1 || idx[0] != "content" || idxType != 2 {
		t.Errorf("全文索引匹配失败，期望: {[content], 2}, 实际: {%v, %d}", idx, idxType)
	}

	// 测试3: 匹配单列普通索引
	idx, idxType = table.MatchIndex("name")
	if len(idx) != 1 || idx[0] != "name" || idxType != 1 {
		t.Errorf("单列普通索引匹配失败，期望: {[name], 1}, 实际: {%v, %d}", idx, idxType)
	}

	// 测试4: 匹配组合索引
	idx, idxType = table.MatchIndex("age", "city")
	if len(idx) != 2 || idx[0] != "age" || idx[1] != "city" || idxType != 1 {
		t.Errorf("组合索引匹配失败，期望: {[age, city], 1}, 实际: {%v, %d}", idx, idxType)
	}

	// 测试5: 完全匹配组合索引
	// 匹配[name, email]索引
	idx, idxType = table.MatchIndex("name", "email")
	if len(idx) != 2 || idx[0] != "name" || idx[1] != "email" || idxType != 1 {
		t.Errorf("完全匹配组合索引失败，期望: {[name, email], 1}, 实际: {%v, %d}", idx, idxType)
	}

	// 测试6: 完全匹配另一个组合索引
	// 匹配[email, phone]索引
	idx, idxType = table.MatchIndex("email", "phone")
	if len(idx) != 2 || idx[0] != "email" || idx[1] != "phone" || idxType != 1 {
		t.Errorf("完全匹配组合索引失败，期望: {[email, phone], 1}, 实际: {%v, %d}", idx, idxType)
	}

	// 测试7: 不支持部分匹配，应该返回nil
	// 没有索引包含age和email两个字段
	idx, idxType = table.MatchIndex("age", "email")
	if idx != nil {
		t.Errorf("部分匹配组合索引测试失败，期望: {nil, -1}, 实际: {%v, %d}", idx, idxType)
	}

	// 测试8: 不支持多字段部分匹配，应该返回nil
	// 没有索引包含name, email和phone三个字段
	idx, idxType = table.MatchIndex("name", "email", "phone")
	if idx != nil {
		t.Errorf("多字段部分匹配测试失败，期望: {nil, -1}, 实际: {%v, %d}", idx, idxType)
	}

	// 测试9: 没有匹配的索引
	table.index = nil // 清空索引
	idx, idxType = table.MatchIndex("no_such_field")
	if idx != nil {
		t.Errorf("无匹配索引测试失败，期望: {nil, -1}, 实际: {%v, %d}", idx, idxType)
	}
}
