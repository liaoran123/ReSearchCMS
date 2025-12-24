package main

import (
	"fmt"
)

func main() {
	fmt.Println("考据级文档搜索引擎!免费版有广告，付费版无广告，企业版提供多维分析统计功能。")
	//files.TraversePathAndReadFiles("D:\\MyGo\\src\\sfsApp\\qldzj")
	/*
		files.TraversePathAndReadFiles("E:\\书")
		iter := tables.Dir.ForData()
		defer iter.Release()
		pdi := resdb.TableIterNew(tables.Dir, iter)
		rd := pdi.GetRecordsByPrimary(true)
		for _, record := range rd {
			fmt.Println(record)
		}
		iter1 := tables.Dir.For()
		defer iter1.Release()
		for iter1.Next() {
			fmt.Println(string(iter1.Key()), string(iter1.Value()))
		}
		fmt.Println("------------------------------")
		iter2 := tables.Article.For()
		defer iter2.Release()
		for iter2.Next() {
			fmt.Println(string(iter2.Key()), string(iter2.Value()))
		}*/
}
