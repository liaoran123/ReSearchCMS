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
	table.SetPrimary("id")
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
		"mid":     0,  //文章目录ID
		"title":   "", //文章标题
		"content": "", //文章内容
	}
	table.SetFields(fields)
	//设置主键
	table.SetPrimary("id") //默认主键为id
	// 添加索引
	table.AddIndex([]string{"mid"})
	table.AddIndex([]string{"title"})
	// 添加全文索引字段
	table.AddFullTextField("content")
	return table
}
