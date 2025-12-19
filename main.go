package main

import (
	"fmt"
	"researchCms/files"
	"researchCms/storedb"
	"researchCms/tables"
)

func main() {

	fmt.Println("考据级文档搜索引擎!免费版有广告，付费版无广告，企业版提供多维分析统计功能。")
	files.TraversePathAndReadFiles("E:\\360Downloads")
	iter := tables.Dir.ForData()
	defer iter.Release()
	pdi := storedb.PrimaryDataIterNew(tables.Dir, iter)
	rd := pdi.GetRecord(true)
	for _, record := range rd {
		fmt.Println(record)
	}
	iter1 := tables.Dir.For()
	defer iter1.Release()
	for iter1.Next() {
		fmt.Println(string(iter1.Key()), string(iter1.Value()))
	}
}
