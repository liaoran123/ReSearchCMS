package tables

import (
	storedb "researchCms/resdb"
)

var Dir = CreateTable_dir()
var Article = CreateTable_article()

// CreateTable_mulu 创建目录表
func CreateTable_dir() *storedb.Table {
	table, err := storedb.TableNew("dir")
	if err != nil {
		return nil
	}
	// 必须先为表预设字段和数据类型
	fields := map[string]any{
		"id":   0,
		"name": "",
		"url":  "",
		"ext":  "", //文件后缀
	}
	table.SetFields(fields)
	//设置主键
	table.SetPrimary([]string{"id"})
	// 添加索引
	table.AddIndex([]string{"url"})
	table.AddIndex([]string{"ext"})
	//table.AddFullTextField("description")
	return table
}
func CreateTable_article() *storedb.Table {
	table, err := storedb.TableNew("article")
	if err != nil {
		return nil
	}
	// 必须先为表预设字段和数据类型
	fields := map[string]any{
		"id":      0,
		"mid":     0,  //文章ID或目录ID
		"secNo":   0,  //文章句子序号
		"title":   "", //文章标题
		"content": "", //文章内容
	}
	table.SetFields(fields)
	//设置主键
	table.SetPrimary([]string{"id"}) //默认主键为id
	// 添加全文索引字段
	table.SetFullTextField("content")
	// 添加索引
	//用于查询文章的所有句子
	table.AddIndex([]string{"mid", "secNo"})
	//用于查询文章中包含指定关键词的句子。
	// 同时包含文章ID和句子序号，这样当需要在指定文章或目录下查询包含关键词的句子时，就可以通过索引快速定位到相关记录。
	table.AddIndex([]string{"content", "mid", "secNo"})

	return table
}
