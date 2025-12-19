package main

import (
	"fmt"
	"researchCms/files"
)

func main1() {
	fmt.Println("=== 测试 forPath.go 优化后功能 ===")
	
	// 创建一个测试文本
	testContent := "这是第一句话。这是第二句话！这是第三句话？\n这是第四句话；这是第五句话。\n\n这是第六句话。This is the seventh sentence. This is the eighth sentence!"
	
	// 调用AddArticle函数进行测试
	fmt.Println("\n1. 测试AddArticle函数...")
	err := files.AddArticle(1, "测试文章", testContent)
	if err != nil {
		fmt.Printf("AddArticle测试失败: %v\n", err)
	} else {
		fmt.Println("AddArticle测试成功!")
	}
	
	// 测试空内容
	fmt.Println("\n2. 测试空内容...")
	err = files.AddArticle(2, "空文章", "")
	if err != nil {
		fmt.Printf("空内容测试失败: %v\n", err)
	} else {
		fmt.Println("空内容测试成功!")
	}
	
	// 测试只有空白字符的内容
	fmt.Println("\n3. 测试只有空白字符的内容...")
	err = files.AddArticle(3, "空白文章", "   \n\t  \n   ")
	if err != nil {
		fmt.Printf("空白内容测试失败: %v\n", err)
	} else {
		fmt.Println("空白内容测试成功!")
	}
	
	fmt.Println("\n=== 所有测试完成 ===")
}