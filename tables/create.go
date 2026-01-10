package tables

import (
	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

var Dir *engine.Table
var Article *engine.Table

func init() {
	_, err := storage.OpenDefaultDb("./rsdb")
	if err != nil {
		panic(err)
	}
	Dir, err = CreateTable_dir()
	if err != nil {
		panic(err)
	}

	Article, err = CreateTable_article()
	if err != nil {
		panic(err)
	}
}

// CreateTable_mulu 创建目录表

// CreateTable_mulu 创建目录表
func CreateTable_dir() (*engine.Table, error) {
	//创建表
	table, err := engine.TableNew("dir")
	if err != nil {
		return nil, err
	}
	//设置字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"url":  "",
		"ext":  "", //文件后缀
	}
	//设置表字段
	table.SetFields(fields)
	//创建主键
	primaryKey, err := engine.DefaultPrimaryKeyNew("id")
	if err != nil {
		return nil, err
	}
	//添加主键字段
	primaryKey.AddFields("id")
	//在表创建索引
	table.CreateIndex(primaryKey)
	//添加普通索引url
	index, err := engine.DefaultNormalIndexNew("url")
	if err != nil {
		return nil, err
	}
	//添加普通索引字段
	index.AddFields("url")
	//在表创建索引
	table.CreateIndex(index)
	//添加普通索引ext
	index, err = engine.DefaultNormalIndexNew("ext")
	if err != nil {
		return nil, err
	}
	//添加普通索引字段
	index.AddFields("ext")
	//在表创建索引
	table.CreateIndex(index)
	return table, nil
}
func CreateTable_article() (*engine.Table, error) {
	table, err := engine.TableNew("article")
	if err != nil {
		return nil, err
	}
	// 设置字段
	fields := map[string]any{
		"mid":     0,  //文章ID或目录ID
		"secNo":   0,  //文章句子序号
		"title":   "", //文章标题
		"content": "", //文章内容
	}
	//设置表字段
	table.SetFields(fields)
	//创建主键
	primaryKey, err := engine.DefaultPrimaryKeyNew("pk")
	if err != nil {
		return nil, err
	}
	primaryKey.AddFields("mid", "secNo") //创建一个mid, secNo的组合主键
	table.CreateIndex(primaryKey)
	//添加全文索引content
	FullTextIndex, err := engine.DefaultFullTextIndexNew("ft")
	if err != nil {
		return nil, err
	}
	//添加全文索引字段
	//全文索引正常情况下必须带上主键，否则后面的关键词都被覆盖，失去全文索引的意义。
	FullTextIndex.AddFields("content", "mid", "secNo")
	//指定content为全文索引字段，索引长度为5
	//如果没有指定，则等同一般索引
	err = FullTextIndex.SetFullField("content", 5)
	if err != nil {
		return nil, err
	}
	//在表创建索引
	table.CreateIndex(FullTextIndex)
	return table, nil
}
