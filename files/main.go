package files

import (
	"ReSearch/db"
	"fmt"
)

func main() {

	TraversePathAndReadFiles("E:\\书")
	iter := db.Tables["dir"].ForData()
	defer iter.Release()

	rd := iter.GetRecords(true)
	for _, record := range rd {
		fmt.Println(record)
	}
	iter1 := db.Tables["dir"].For()
	defer iter1.Release()
	for iter1.Next() {
		fmt.Println(string(iter1.Key()), string(iter1.Value()))
	}
	fmt.Println("------------------------------")
	iter2 := db.Tables["article"].For()
	defer iter2.Release()
	for iter2.Next() {
		fmt.Println(string(iter2.Key()), string(iter2.Value()))
	}
}
