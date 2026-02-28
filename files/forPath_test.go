package files

import (
	"ReSearch/db"
	"fmt"
	"strings"
	"testing"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/util"
)

func Test_ForPath(T *testing.T) {
	//TraversePathAndReadFiles("D:\\MyGo\\src\\sfsApp\\qldzj")
}
func Test_ForPath1(T *testing.T) {

	TraversePathAndReadFiles("E:\\test")

	fmt.Println("----------dir--------------------")
	iter := db.Tables["dir"].ForData()
	defer engine.GlobalTableIterPool.Put(iter)
	rd := iter.GetRecords(true)
	fmt.Println(len(rd))
	fmt.Println("------------------------------")
	for _, record := range rd {
		fmt.Println(record)
	}

}

func Test_ForPath2(T *testing.T) {
	TraversePathAndReadFiles("E:\\四库全书20231110\\11-乾隆大藏经\\3-论\\1-大乘论\\010-瑜伽师地论（第001卷～第020卷）")
	fmt.Println("----------dir--------------------")
	iter := db.Tables["dir"].ForData()
	defer iter.Release() //engine.GlobalTableIterPool.Put(iter)
	rd := iter.GetRecords(true)
	defer rd.Release()
	fmt.Println(len(rd))
	fmt.Println("------------------------------")
	for _, record := range rd {
		fmt.Println(record)
	}
}
func Test_ForPath5(T *testing.T) {
	iter1, _ := db.Tables["senc"].Search(&map[string]any{
		"content": "般若",
	})
	defer iter1.Release() //engine.GlobalTableIterPool.Put(iter1)
	rlen := 11
	pk := db.Tables["senc"].GetPrimaryKey()
	//Prefix := pk.Prefix(0)
	pfxlen := len(pk.Prefix(0))
	pfxlen += 1 //去除分隔符
	fields := db.Tables["senc"].GetAllFields()
	fieldTypeLen := pk.GetfieldTypeLen(&fields)
	tylens := 0
	for _, fieldLen := range *fieldTypeLen {
		tylens += int(fieldLen) + 1
	}
	tylens -= 1 //去除最后一个分隔符
	relist := []string{}
	remap := map[string]bool{}
	loop := 0

	for iter1.Next() {
		key := iter1.Key()
		key = key[pfxlen : len(key)-tylens-pfxlen+3]

		if remap[string(key)] {
			continue
		}
		fmt.Println(string(key))
		remap[string(key)] = true
		relist = append(relist, string(key))
		loop++
		if loop > rlen {
			break
		}
	}

}

// 目录管理
func Test_ForPath6(T *testing.T) {

	//TraversePathAndReadFiles("E:\\四库全书20231110\\11-乾隆大藏经\\3-论\\1-大乘论")

	fmt.Println("----------dir--------------------")
	iter, _ := db.Tables["dir"].Search(&map[string]any{
		"url": "E:\\四库全书20231110\\11-乾隆大藏经\\3-论\\1-大乘论", //总目录="E:\\四库全书20231110\\11-乾隆大藏经\\3-论\\1-大乘论"
		//"ext": ".docx",
	}, util.Equal)
	defer engine.GlobalTableIterPool.Put(iter)

	iter1 := db.Tables["dir"].ForData()
	defer engine.GlobalTableIterPool.Put(iter1)
	rd1 := iter1.GetRecords(true)
	fmt.Println(len(rd1))
	fmt.Println("------------------------------")
	for _, record := range rd1 {
		fmt.Println(record)
	}

}
func Test_ForPath7(T *testing.T) {
	//打开目录id为1489的名称
	iterdir, _ := db.Tables["dir"].Search(&map[string]any{
		"id": 1489,
	}, util.Equal)
	defer engine.GlobalTableIterPool.Put(iterdir)
	rddir := iterdir.GetRecords(true).Select("name", "url")
	title := rddir[0]["name"].(string)
	url := rddir[0]["url"].(string)
	fmt.Println(title, url)

	//打开文章id为1489的内容
	itersenc, _ := db.Tables["senc"].Search(&map[string]any{
		"did":   1489,
		"secNo": nil,
	})
	defer engine.GlobalTableIterPool.Put(itersenc)
	rdsenc := itersenc.GetRecords(true).Select("content")
	var text strings.Builder
	content := ""
	for _, record := range rdsenc {
		content = record["content"].(string)
		//将回车符转换为<br />
		text.WriteString(strings.ReplaceAll(content, "\n", "<br />"))
	}
	fmt.Println(text.String())
}

func Test_ForPath8(T *testing.T) {
	iterdirid, _ := db.Tables["dir"].Search(&map[string]any{
		"id": 8882,
	}, util.Equal)
	defer engine.GlobalTableIterPool.Put(iterdirid)
	rddir := iterdirid.GetRecords(true).Select("id")
	id := rddir[0]["id"].(int)
	fmt.Println(id)
	//通过url匹配文章id为8882的id
	/*
		iterdir, _ := db.Tables["dir"].Search(&map[string]any{
			"url": `E:\工具\代理\ChromeGo\chrome-user-data\Default\Extensions\bhghoamapcdpbohphigoooaddinpkbai\8.0.1_0\_locales\id`,
		}, util.Equal)
		defer engine.GlobalTableIterPool.Put(iterdir)
		rddir = iterdir.GetRecords(true).Select("id")
		id = rddir[0]["id"].(int)
		fmt.Println(id)
	*/

}

func Test_ForPath9(T *testing.T) {

	TraversePathAndReadFiles("E:\\test")

	fmt.Println("----------dir--------------------")
	iter1 := db.Tables["dir"].For()
	defer iter1.Release()
	for iter1.Next() {
		fmt.Println(string(iter1.Key()), string(iter1.Value()))
	}

	fmt.Println("----------dir--------------------")
	iter := db.Tables["dir"].ForData()
	defer engine.GlobalTableIterPool.Put(iter)
	rd := iter.GetRecords(true)
	fmt.Println(len(rd))
	fmt.Println("------------------------------")
	for _, record := range rd {
		fmt.Println(record)
	}

	fmt.Println("------------------------------")
	iter2 := db.Tables["senc"].For()
	defer iter2.Release()
	for iter2.Next() {
		fmt.Println(string(iter2.Key()), string(iter2.Value()))
	}
	fmt.Println("------------------------------")
	iter3 := db.Tables["senc"].ForData()
	defer engine.GlobalTableIterPool.Put(iter3)
	rd = iter3.GetRecords(true)
	for _, record := range rd {
		fmt.Println(record)
	}
	fmt.Println("-------openid-----------------------")
	iter4, _ := db.Tables["senc"].Search(&map[string]any{
		"content": "open",
	})
	defer engine.GlobalTableIterPool.Put(iter4)
	rd1 := iter4.GetRecords(true, 21)
	for _, record := range rd1 {
		fmt.Println(record)
	}

	itemPath := "E:\\test\\abc.txt" //"\x00-\x01-E:\\test\\123.txt"  -- "\x00-\x01-E:\\test\\123.txt"
	path := itemPath
	iterdirPath, _ := db.Tables["dir"].Search(&map[string]any{
		"url": path,
	}, util.Equal)
	defer engine.GlobalTableIterPool.Put(iterdirPath)
	if iterdirPath != nil {
		if iterdirPath.First() {
			key := iterdirPath.Key()
			fmt.Println(string(key))
		}
		rddirPath := iterdirPath.GetRecords(true).Select("id")
		if len(rddirPath) > 0 {
			itemId := rddirPath[0]["id"].(int)
			fmt.Println(itemId)
		}
	}
}
func Test_ForPath11(T *testing.T) {
	TraversePathAndReadFiles("E:\\test1")
}
func Test_ForPath10(T *testing.T) {
	id := db.Tables["dir"].GetId()
	fmt.Println(id)
	fieldIdMap := db.Tables["dir"].GetAllFieldNameIdMap()
	fmt.Println(fieldIdMap)
	indexNameMap := db.Tables["dir"].GetAllIndexNameIdMap()
	fmt.Println(indexNameMap)
	did := db.Tables["senc"].GetId()
	fmt.Println(did)

	sencFieldIdMap := db.Tables["senc"].GetAllFieldNameIdMap()
	fmt.Println(sencFieldIdMap)
	sencIndexNameMap := db.Tables["senc"].GetAllIndexNameIdMap()
	fmt.Println(sencIndexNameMap)
}
func Test_ForPath12(T *testing.T) {
	/*
		TraversePathAndReadFiles("E:\\test")

		fmt.Println("----------dir--------------------")
		iter1 := db.Tables["dir"].For()
		defer iter1.Release()
		for iter1.Next() {
			fmt.Println(string(iter1.Key()), string(iter1.Value()))
		}

		fmt.Println("----------dir--------------------")
		iter := db.Tables["dir"].ForData()
		defer iter.Release()
		rd := iter.GetRecords(true)
		fmt.Println(len(rd))
		fmt.Println("------------------------------")
		for _, record := range rd {
			fmt.Println(record)
		}

		fmt.Println("------------------------------")
		iter2 := db.Tables["senc"].For()
		defer iter2.Release()
		for iter2.Next() {
			fmt.Println(string(iter2.Key()), string(iter2.Value()))
		}
		fmt.Println("------------------------------")
		iter3 := db.Tables["senc"].ForData()
		defer iter3.Release()
		rd = iter3.GetRecords(true)
		for _, record := range rd {
			fmt.Println(record)
		}
		fmt.Println("-------openid-----------------------")
		iter4 := db.Tables["senc"].Search(&map[string]any{
			"content": "open",
		})
		defer iter4.Release()
		rd1 := iter4.GetRecords(true, 21)
		for _, record := range rd1 {
			fmt.Println(record)
		}
	*/
	itemPath := "E:\\test\\abc.txt" //"\x00-\x01-E:\\test\\123.txt"  -- "\x00-\x01-E:\\test\\123.txt"
	path := itemPath
	iterdirPath, _ := db.Tables["dir"].Search(&map[string]any{
		"url": path,
	}, util.Equal)
	defer engine.GlobalTableIterPool.Put(iterdirPath)
	if iterdirPath != nil {
		if iterdirPath.First() {
			key := iterdirPath.Key()
			fmt.Println(string(key))
		}
		rddirPath := iterdirPath.GetRecords(true).Select("id")
		if len(rddirPath) > 0 {
			itemId := rddirPath[0]["id"].(int)
			fmt.Println(itemId)
		}
	}
}
