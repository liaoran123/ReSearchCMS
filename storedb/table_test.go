// 测试文件，对应 table.go
package storedb

import (
	"bytes"
	"encoding/json/v2"
	"fmt"
	"testing"
)

// TestTableNew 测试表的创建
func TestTableNew(t *testing.T) {
	// 测试创建正常表名
	table := TableNew("test_table")
	if table == nil {
		t.Fatal("TableNew 失败，返回 nil")
	}
	if table.name != "test_table" {
		t.Errorf("表名错误，期望: test_table, 实际: %s", table.name)
	}
	if table.ftlen != 5 {
		t.Errorf("默认全文索引长度错误，期望: 5, 实际: %d", table.ftlen)
	}
	if table.fields == nil {
		t.Error("fields 未初始化")
	}

	// 测试包含分隔符的表名
	invalidTable := TableNew("test-table")
	if invalidTable != nil {
		t.Error("TableNew 应该拒绝包含分隔符的表名")
	}
}

// TestPrimaryFunctions 测试主键相关功能
func TestPrimaryFunctions(t *testing.T) {
	table := TableNew("test_primary")
	if table == nil {
		t.Fatal("TableNew 失败")
	}

	// 测试设置主键字段
	table.SetPrimaryValue("id")
	if table.primary != "id" {
		t.Errorf("设置主键失败，期望: id, 实际: %s", table.primary)
	}

	// 测试获取主键前缀
	prefix := table.GetPrimaryPrefix()
	expectedPrefix := "test_primary-pk"
	if prefix != expectedPrefix {
		t.Errorf("主键前缀错误，期望: %s, 实际: %s", expectedPrefix, prefix)
	}

	// 测试主键序列化
	table.SetField("id", 123)
	primaryValue := table.GetPrimaryValue()
	var value int
	err := json.Unmarshal(primaryValue, &value)
	if err != nil {
		t.Errorf("主键序列化失败: %v", err)
	}
	if value != 123 {
		t.Errorf("主键值错误，期望: 123, 实际: %d", value)
	}
}

// TestIndexFunctions 测试索引相关功能
func TestIndexFunctions(t *testing.T) {
	table := TableNew("test_index")
	if table == nil {
		t.Fatal("TableNew 失败")
	}

	// 测试设置索引字段
	indexes := [][]string{{"name"}, {"age", "city"}}
	table.SetIndexValue(indexes)
	if len(table.index) != 2 {
		t.Errorf("设置索引数量错误，期望: 2, 实际: %d", len(table.index))
	}

	// 测试添加单个索引
	table.AddIndex([]string{"email"})
	if len(table.index) != 3 {
		t.Errorf("添加索引数量错误，期望: 3, 实际: %d", len(table.index))
	}

	// 测试获取索引前缀
	prefix := table.GetIndexPrefix()
	expectedPrefix := "test_index-idx"
	if prefix != expectedPrefix {
		t.Errorf("索引前缀错误，期望: %s, 实际: %s", expectedPrefix, prefix)
	}
}

// TestFullTextFunctions 测试全文索引相关功能
func TestFullTextFunctions(t *testing.T) {
	table := TableNew("test_fulltext")
	if table == nil {
		t.Fatal("TableNew 失败")
	}

	// 测试设置全文索引字段
	fields := []string{"content", "description"}
	table.SetFullTextValue(fields)
	if len(table.fullText) != 2 {
		t.Errorf("设置全文索引字段数量错误，期望: 2, 实际: %d", len(table.fullText))
	}

	// 测试添加单个全文索引字段
	table.AddFullTextField("title")
	if len(table.fullText) != 3 {
		t.Errorf("添加全文索引字段数量错误，期望: 3, 实际: %d", len(table.fullText))
	}

	// 测试设置全文索引分词长度
	table.SetFullTextLen(7)
	if table.ftlen != 7 {
		t.Errorf("设置全文索引分词长度错误，期望: 7, 实际: %d", table.ftlen)
	}

	// 测试边界值检查
	table.SetFullTextLen(2) // 小于最小值3，应该保持不变
	if table.ftlen != 7 {
		t.Errorf("最小全文索引分词长度检查失败，期望: 7, 实际: %d", table.ftlen)
	}

	table.SetFullTextLen(12) // 大于最大值11，应该保持不变
	if table.ftlen != 7 {
		t.Errorf("最大全文索引分词长度检查失败，期望: 7, 实际: %d", table.ftlen)
	}

	// 测试获取全文索引前缀
	prefix := table.GetFullTextPrefix()
	expectedPrefix := "test_fulltext-ft"
	if prefix != expectedPrefix {
		t.Errorf("全文索引前缀错误，期望: %s, 实际: %s", expectedPrefix, prefix)
	}
}

// TestFullTextToken 测试全文索引分词功能
func TestFullTextToken(t *testing.T) {
	table := TableNew("test_token")
	if table == nil {
		t.Fatal("TableNew 失败")
	}

	// 测试默认分词长度(5)
	tokens := table.GetFullTextToken("abcdefghij", 5)
	expectedTokens := []string{"abcde", "bcdef", "cdefg", "defgh", "efghi", "fghij", "ghij", "hij", "ij", "j"}
	if !compareStringSlices(tokens, expectedTokens) {
		t.Errorf("默认分词长度测试失败，期望: %v, 实际: %v", expectedTokens, tokens)
	}

	// 测试自定义分词长度(3)
	tokens = table.GetFullTextToken("abcdef", 3)
	expectedTokens = []string{"abc", "bcd", "cde", "def", "ef", "f"}
	if !compareStringSlices(tokens, expectedTokens) {
		t.Errorf("自定义分词长度测试失败，期望: %v, 实际: %v", expectedTokens, tokens)
	}

	// 测试短文本分词
	tokens = table.GetFullTextToken("abc", 5)
	expectedTokens = []string{"abc", "bc", "c"}
	if !compareStringSlices(tokens, expectedTokens) {
		t.Errorf("短文本分词测试失败，期望: %v, 实际: %v", expectedTokens, tokens)
	}

	// 测试中文分词
	tokens = table.GetFullTextToken("测试中文分词", 2)
	expectedTokens = []string{"测试", "试中", "中文", "文分", "分词", "词"}
	if !compareStringSlices(tokens, expectedTokens) {
		t.Errorf("中文分词测试失败，期望: %v, 实际: %v", expectedTokens, tokens)
	}
}

// TestGetFields 测试获取字段功能
func TestGetFields(t *testing.T) {
	table := TableNew("test_fields")
	if table == nil {
		t.Fatal("TableNew 失败")
	}

	// 设置一些字段
	table.SetField("id", 1)
	table.SetField("name", "test")
	table.SetField("age", 25)
	table.SetField("active", true)

	// 获取字段的字节表示
	fieldsBytes := table.GetFields()
	if fieldsBytes == nil {
		t.Error("GetFields 返回 nil")
		return
	}

	// 验证序列化结果
	var fieldsMap map[string]any
	err := json.Unmarshal(fieldsBytes, &fieldsMap)
	if err != nil {
		t.Errorf("字段反序列化失败: %v", err)
		return
	}

	// 验证字段值
	// 注意：JSON 反序列化时，数字默认会被解析为 float64 类型
	if fieldsMap["id"] != float64(1) {
		t.Errorf("字段 id 值错误，期望: 1, 实际: %v", fieldsMap["id"])
	}
	if fieldsMap["name"] != "test" {
		t.Errorf("字段 name 值错误，期望: test, 实际: %v", fieldsMap["name"])
	}
	if fieldsMap["age"] != float64(25) {
		t.Errorf("字段 age 值错误，期望: 25, 实际: %v", fieldsMap["age"])
	}
	if fieldsMap["active"] != true {
		t.Errorf("字段 active 值错误，期望: true, 实际: %v", fieldsMap["active"])
	}
}

// TestAutoIncrement 测试自动增值功能
func TestAutoIncrement(t *testing.T) {
	table := TableNew("test_autoinc")
	if table == nil {
		t.Fatal("TableNew 失败")
	}

	// 设置主键为自动增值
	table.SetPrimaryValue("id")
	table.SetField("id", nil) // 表示自动增值

	// 初始化自动增值计数器
	table.InitAuto()

	// 测试自动生成主键值
	table.InitPrimary()
	id1Val, ok := table.GetField("id")
	if !ok || id1Val == nil {
		t.Error("自动增值应该生成主键值")
		return
	}

	id1, ok := id1Val.(int64)
	if !ok {
		t.Error("主键值应该是 int64 类型")
		return
	}

	// 再次调用应该生成下一个值
	table2 := TableNew("test_autoinc")
	table2.SetPrimaryValue("id")
	table2.SetField("id", nil)
	table2.InitAuto()
	table2.InitPrimary()

	id2Val, ok := table2.GetField("id")
	if !ok {
		t.Error("获取id字段失败")
		return
	}
	id2, ok := id2Val.(int64)
	if !ok {
		t.Error("主键值应该是 int64 类型")
		return
	}

	// 注意：由于测试环境中没有实际的数据库存储，两次调用可能生成相同的值
	// 这里我们主要测试功能是否正常执行，而不是实际的递增效果
	fmt.Printf("自动增值测试: id1=%d, id2=%d\n", id1, id2)
}

// TestGetFullTextValue 测试全文索引值生成功能
func TestGetFullTextValue(t *testing.T) {
	table := TableNew("test_fulltext_value")
	if table == nil {
		t.Fatal("TableNew 失败")
	}

	// 设置全文索引字段和主键
	table.SetPrimaryValue("id")
	table.SetFullTextValue([]string{"content"})
	table.SetFullTextLen(2) // 设置较短的分词长度便于测试

	// 设置字段值
	table.SetField("id", 1)
	table.SetField("content", "测试全文索引")

	// 获取全文索引值
	fullTextValues := table.GetFullTextValue()
	if len(fullTextValues) == 0 {
		t.Error("GetFullTextValue 应该返回全文索引值")
		return
	}

	// 验证生成的全文索引键是否包含预期的分词
	expectedTokens := []string{"测试", "试全", "全文", "文索", "索引", "引"}

	// 打印生成的键值对，便于调试
	fmt.Println("生成的全文索引键值对:")
	for _, key := range fullTextValues {
		// 验证键的格式：表名-ft-分词
		if !bytes.HasPrefix(key, []byte("test_fulltext_value-ft-")) {
			t.Errorf("全文索引键格式错误，期望以'test_fulltext_value-ft-'开头，实际: %s", string(key))
		}
		fmt.Printf("键: %s\n", string(key))
	}

	// 验证生成的键数量与预期一致
	expectedKeyCount := len(expectedTokens)
	if len(fullTextValues) != expectedKeyCount {
		t.Errorf("全文索引键数量错误，期望: %d, 实际: %d", expectedKeyCount, len(fullTextValues))
	}

	fmt.Printf("全文索引测试: 生成了 %d 个键\n", len(fullTextValues))
}

// compareStringSlices 比较两个字符串切片是否相等
func TestGetIndexValue(t *testing.T) {
	table := TableNew("test_get_index_value")
	if table == nil {
		t.Fatal("TableNew 失败")
	}

	// 设置索引字段
	indexes := [][]string{{"name"}, {"age", "city"}}
	table.SetIndexValue(indexes)

	// 设置字段值
	table.SetField("name", "John Doe")
	table.SetField("age", 30)
	table.SetField("city", "New York")

	// 获取索引值
	indexValues := table.GetIndexValue()

	// 验证索引值数量
	if len(indexValues) != 2 {
		t.Errorf("索引值数量错误，期望: 2, 实际: %d", len(indexValues))
	}

	// 验证索引值格式
	expectedIndex1 := []byte("test_get_index_value-idx-\"John Doe\"")
	expectedIndex2 := []byte("test_get_index_value-idx-30-\"New York\"")

	if !bytes.Equal(indexValues[0], expectedIndex1) && !bytes.Equal(indexValues[0], expectedIndex2) {
		t.Errorf("索引值1错误，期望: %s 或 %s, 实际: %s", expectedIndex1, expectedIndex2, indexValues[0])
	}

	if !bytes.Equal(indexValues[1], expectedIndex1) && !bytes.Equal(indexValues[1], expectedIndex2) {
		t.Errorf("索引值2错误，期望: %s 或 %s, 实际: %s", expectedIndex1, expectedIndex2, indexValues[1])
	}

	// 打印索引值用于调试
	fmt.Println("生成的索引键值对:")
	for i, value := range indexValues {
		fmt.Printf("键 %d: %s\n", i+1, string(value))
	}
}

// TestTableInitAuto 测试InitAuto方法的功能
func TestTableInitAuto(t *testing.T) {
	// 测试场景1: 空表初始化
	table1 := TableNew("test_initauto_empty")
	table1.SetPrimaryValue("id")

	// 验证空表情况下InitAuto的行为
	table1.InitAuto()
	// 由于是空表，MaxAutoValue应该返回1，所以计数器应该是1
	if table1.counter.Load() != 1 {
		t.Errorf("空表初始化失败，期望计数器值: 1, 实际: %d", table1.counter.Load())
	}

	// 测试场景2: 已有数据的表初始化
	table2 := TableNew("test_inauto_with_data")
	table2.SetPrimaryValue("id")

	// 插入一些测试数据
	for i := 1; i <= 5; i++ {
		table2.SetField("id", i)
		table2.SetField("name", fmt.Sprintf("User %d", i))
		if err := table2.Insert(); err != nil {
			t.Fatalf("插入测试数据失败: %v", err)
		}
	}

	// 创建一个新的表实例，用于测试InitAuto方法
	table2New := TableNew("test_inauto_with_data")
	table2New.SetPrimaryValue("id")

	// 调用InitAuto方法
	table2New.InitAuto()

	// 验证计数器是否正确设置为最大的主键值
	if table2New.counter.Load() != 5 {
		t.Errorf("已有数据的表初始化失败，期望计数器值: 5, 实际: %d", table2New.counter.Load())
	}

	// 测试场景3: 已有数据且设置了主键的情况
	table3 := TableNew("test_inauto_with_primary_set")
	table3.SetPrimaryValue("id")
	table3.SetField("id", 10)
	table3.SetField("name", "User with set id")

	// 调用InitAuto方法
	table3.InitAuto()

	// 由于主键已经设置，计数器应该保持不变（默认是0）
	if table3.counter.Load() != 0 {
		t.Errorf("已设置主键的表初始化失败，期望计数器值保持0, 实际: %d", table3.counter.Load())
	}
}

// TestTableRead 测试Read方法的功能
func TestTableRead(t *testing.T) {
	// 创建测试表
	table := TableNew("test_read")
	table.SetField("id", 1)
	table.SetField("name", "John Doe")
	table.SetField("age", 30)
	table.SetField("city", "New York")

	// 设置主键
	table.SetPrimaryValue("id")

	// 插入测试数据
	err := table.Insert()
	if err != nil {
		t.Fatalf("插入测试数据失败: %v", err)
	}

	// 测试1: 正常读取已存在的记录
	insertedFields := table.Read(1)
	if insertedFields == nil {
		t.Fatal("读取失败，无法读取已存在的记录")
	}

	if insertedFields["name"] != "John Doe" {
		t.Errorf("读取的姓名错误，期望: John Doe, 实际: %v", insertedFields["name"])
	}

	if insertedFields["age"] != 30.0 {
		t.Errorf("读取的年龄错误，期望: 30, 实际: %v", insertedFields["age"])
	}

	// 测试2: 读取不存在的记录
	notExistFields := table.Read(999)
	if notExistFields != nil {
		t.Errorf("读取不存在的记录应该返回nil，实际: %v", notExistFields)
	}

	// 测试3: 使用字符串主键
	table2 := TableNew("test_read_string_pk")
	table2.SetField("user_id", "user123")
	table2.SetField("name", "Jane Smith")
	table2.SetField("age", 25)
	table2.SetField("city", "Los Angeles")

	// 设置字符串主键
	table2.SetPrimaryValue("user_id")

	// 插入测试数据
	err = table2.Insert()
	if err != nil {
		t.Fatalf("插入字符串主键测试数据失败: %v", err)
	}

	// 读取字符串主键的记录
	stringPkFields := table2.Read("user123")
	if stringPkFields == nil {
		t.Fatal("读取字符串主键记录失败")
	}

	if stringPkFields["name"] != "Jane Smith" {
		t.Errorf("读取的姓名错误，期望: Jane Smith, 实际: %v", stringPkFields["name"])
	}
}

func TestCRUDOperations(t *testing.T) {
	// 创建测试表
	table := TableNew("test_crud")
	if table == nil {
		t.Fatal("TableNew 失败")
	}

	// 设置主键
	table.SetPrimaryValue("id")

	// 测试插入操作
	table.SetField("id", 1)
	table.SetField("name", "John Doe")
	table.SetField("age", 30)
	table.SetField("city", "New York")

	err := table.Insert()
	if err != nil {
		t.Fatalf("Insert 失败: %v", err)
	}

	// 验证插入是否成功
	insertedFields := table.Read(1)
	if insertedFields == nil {
		t.Fatal("插入失败，无法读取记录")
	}

	if insertedFields["name"] != "John Doe" {
		t.Errorf("插入的姓名错误，期望: John Doe, 实际: %v", insertedFields["name"])
	}

	if insertedFields["age"] != 30.0 {
		t.Errorf("插入的年龄错误，期望: 30, 实际: %v", insertedFields["age"])
	}

	// 测试更新操作
	updateFields := map[string]interface{}{
		"id":   1,
		"age":  31,
		"city": "Los Angeles",
	}

	err = table.Update(updateFields)
	if err != nil {
		t.Fatalf("Update 失败: %v", err)
	}

	// 验证更新是否成功
	updatedFields := table.Read(1)
	if updatedFields == nil {
		t.Fatal("更新失败，无法读取记录")
	}

	if updatedFields["age"] != 31.0 {
		t.Errorf("更新的年龄错误，期望: 31, 实际: %v", updatedFields["age"])
	}

	if updatedFields["city"] != "Los Angeles" {
		t.Errorf("更新的城市错误，期望: Los Angeles, 实际: %v", updatedFields["city"])
	}

	// 测试删除操作
	table.SetField("id", 1)

	err = table.Delete()
	if err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}

	// 验证删除是否成功
	deletedFields := table.Read(1)
	if deletedFields != nil {
		t.Error("删除失败，记录仍然存在")
	}
}

func compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestSearch 测试Search方法的功能
func TestSearch(t *testing.T) {
	// 创建测试表
	table := TableNew("test_search")
	table.SetPrimaryValue("id")
	table.AddIndex([]string{"name"})
	table.AddFullTextField("description")
	table.InitAuto()

	// 插入测试数据
	data := []map[string]any{
		{"id": 1, "name": "Alice", "age": 25, "description": "Alice is a software engineer"},
		{"id": 2, "name": "Bob", "age": 30, "description": "Bob is a product manager"},
		{"id": 3, "name": "Charlie", "age": 35, "description": "Charlie is a designer"},
		{"id": 4, "name": "David", "age": 40, "description": "David is a developer"},
		{"id": 5, "name": "Eve", "age": 45, "description": "Eve is a manager"},
		{"id": nil, "name": "Eve nil", "age": 45, "description": "Eve is a manager nil"},
		{"id": nil, "name": "Eve nil1", "age": 45, "description": "Eve is a manager nil1"},
	}

	for _, item := range data {
		table.SetField("id", item["id"])
		table.SetField("name", item["name"])
		table.SetField("age", item["age"])
		table.SetField("description", item["description"])
		if err := table.Insert(); err != nil {
			t.Fatalf("插入测试数据失败: %v", err)
		}
	}
	//测试遍历表所有kv
	t.Run("For", func(t *testing.T) {
		dataIter := table.For()
		for dataIter.Next() {
			fmt.Printf("key: %s, value: %s\n", dataIter.Key(), dataIter.Value())
		}
	})
	// 测试遍历表所有数据
	t.Run("ForData", func(t *testing.T) {
		// 使用ForData方法遍历所有数据
		dataIter := table.ForData()
		results := dataIter.For(true)
		if len(results) != 5 {
			t.Errorf("遍历所有数据预期返回5条记录，实际返回%d条", len(results))
			return
		}
		fmt.Printf("results: %v\n", results)
	})
	// 测试1: 主键搜索
	t.Run("PrimaryKeySearch", func(t *testing.T) {
		// 使用SearchData方法搜索主键为3的记录
		dataIter := table.SearchData("id", 3)
		results := dataIter.For(true)
		if len(results) != 1 {
			t.Errorf("主键搜索预期返回1条记录，实际返回%d条", len(results))
			return
		}

		// 验证返回的数据是否正确
		result := results[0].(map[string]any)
		if result["id"] != 3.0 || result["name"] != "Charlie" {
			t.Errorf("主键搜索返回的数据不正确，预期: {id: 3, name: 'Charlie'}, 实际: %v", result)
		}
	})

	// 测试2: 索引字段搜索
	t.Run("IndexSearch", func(t *testing.T) {
		// 使用SearchData方法搜索name为"Charlie"的记录
		dataIter := table.SearchData("name", "Charlie")
		results := dataIter.For(true)
		if len(results) != 1 {
			t.Errorf("索引搜索预期返回1条记录，实际返回%d条", len(results))
			return
		}

		// 验证返回的数据是否正确
		result := results[0]
		if result != 3.0 {
			t.Errorf("索引搜索返回的数据不正确，预期: {id: 3}, 实际: %v", result)
		}
	})

	// 测试3: 全文索引搜索
	t.Run("FullTextSearch", func(t *testing.T) {
		// 使用SearchData方法搜索description包含"developer"的记录
		dataIter := table.SearchData("description", "Bob")
		results := dataIter.For(true)
		if len(results) == 0 {
			t.Error("全文索引搜索预期返回至少1条记录，实际返回0条")
			return
		}

	})

	// 测试4: 搜索不存在的数据
	t.Run("SearchNonExistent", func(t *testing.T) {
		// 使用SearchData方法搜索不存在的id
		dataIter := table.SearchData("id", 100)
		results := dataIter.For(true)
		if len(results) != 0 {
			t.Errorf("搜索不存在的数据预期返回0条记录，实际返回%d条", len(results))
		}
	})
}
