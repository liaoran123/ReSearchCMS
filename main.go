package main

import (
	"ReSearch/files"
	"ReSearch/tables"
	"fmt"
)

func main() {

	//files.TraversePathAndReadFiles("D:\\MyGo\\src\\sfsApp\\qldzj")

	files.TraversePathAndReadFiles("E:\\书")

	iter1 := tables.Dir.For()
	defer iter1.Release()
	fmt.Println("-----------Dir-------------------")
	for iter1.Next() {
		fmt.Println(string(iter1.Key()), string(iter1.Value()))
	}
	fmt.Println("------------------------------")
	iter2 := tables.Article.For()
	defer iter2.Release()
	fmt.Println("-----------Article-------------------")
	for iter2.Next() {
		fmt.Println(string(iter2.Key()), string(iter2.Value()))
	}

	fmt.Println("-----------DirData-------------------")
	iter := tables.Dir.ForData()
	defer iter.Release()
	dirRd := iter.GetRecords(true)
	for _, record := range dirRd {
		fmt.Println(record)
	}
	fmt.Println("-----------ArticleData-------------------")
	articleIter := tables.Article.ForData()
	defer articleIter.Release()
	articleRd := articleIter.GetRecords(true)
	for _, record := range articleRd {
		fmt.Println(record)
	}

}
