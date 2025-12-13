package storedb

import (
	"fmt"
	"testing"
)

// TestSetField 测试SetField函数
func TestSetField(t *testing.T) {
	// 创建一个新表
	table := TableNew("test_setfield")
	if table == nil {
		t.Fatal("TableNew 失败")
	}

	// 测试基本的字段设置
	table.SetField("name", "张三")
	table.SetField("age", 25)
	table.SetField("active", true)

	// 验证字段值是否正确设置
	if name, ok := table.GetField("name"); !ok || name != "张三" {
		t.Errorf("name 字段值错误，期望: 张三, 实际: %v", name)
	}
	if age, ok := table.GetField("age"); !ok || age != 25 {
		t.Errorf("age 字段值错误，期望: 25, 实际: %v", age)
	}
	if active, ok := table.GetField("active"); !ok || active != true {
		t.Errorf("active 字段值错误，期望: true, 实际: %v", active)
	}

	// 测试连续调用
	table.SetField("city", "北京")
	table.SetField("salary", 10000.50)
	if city, ok := table.GetField("city"); !ok || city != "北京" {
		t.Errorf("city 字段值错误，期望: 北京, 实际: %v", city)
	}
	if salary, ok := table.GetField("salary"); !ok || salary != 10000.50 {
		t.Errorf("salary 字段值错误，期望: 10000.50, 实际: %v", salary)
	}

	// 测试更新已有字段
	table.SetField("age", 26)
	if age, ok := table.GetField("age"); !ok || age != 26 {
		t.Errorf("age 字段更新错误，期望: 26, 实际: %v", age)
	}

	fmt.Println("SetField 测试通过!")
}

// TestSetFieldWithPrimary 测试SetField与SetPrimaryValue结合使用
func TestSetFieldWithPrimary(t *testing.T) {
	// 创建一个新表
	table := TableNew("test_setfield_primary")
	if table == nil {
		t.Fatal("TableNew 失败")
	}

	// 先设置id字段，然后再设置为主键
	table.SetField("id", 1)
	err := table.SetPrimaryValue("id")
	if err != nil {
		t.Errorf("SetPrimaryValue 失败: %v", err)
	}

	// 验证主键设置成功
	if table.primary != "id" {
		t.Errorf("主键设置错误，期望: id, 实际: %s", table.primary)
	}

	fmt.Println("SetField 与 SetPrimaryValue 结合使用测试通过!")
}
