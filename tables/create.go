package tables

import (
	"researchCms/storedb"
)

var Tables = map[string]*storedb.Table{
	"dir":     CreateTable_dir(),
	"article": CreateTable_article(),
}

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
	}
	table.SetFields(fields)
	//设置主键
	table.SetPrimary("id")
	// 添加索引
	table.AddIndex([]string{"url"})
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
		"id":          0,
		"mid":         0,  //目录ID
		"description": "", //将文章分为多个段落
	}
	table.SetFields(fields)
	//设置主键
	table.SetPrimary("mid")
	// 添加索引
	table.AddIndex([]string{"mid"})
	// 添加全文索引字段
	table.AddFullTextField("description")
	return table
}
