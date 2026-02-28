package services

import (
	"ReSearch/db"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/match"
	"github.com/liaoran123/sfsDb/record"
	"github.com/liaoran123/sfsDb/util"
)

// SearchResult 搜索结果结构
type SearchResult struct {
	Did     int    `json:"did"`
	SecNo   int    `json:"secNo"`
	Content string `json:"content"`
	Path    string `json:"path,omitempty"`
}

// Search 执行搜索
func Search(query string, did int, start, limit int) ([]SearchResult, int, error) {
	// 开始计时
	totalStartTime := time.Now()

	// 1. 搜索词处理
	wordProcessStart := time.Now()
	//将query按空格分隔
	words := strings.Split(query, " ")
	//按词的长度排序
	sort.Slice(words, func(i, j int) bool {
		return len(words[i]) < len(words[j])
	})
	//最多支持3个查询词
	if len(words) > 3 {
		words = words[:3]
	}
	wordProcessTime := time.Since(wordProcessStart)

	// 调试输出
	fmt.Printf("搜索词处理时间: %v ms\n", wordProcessTime.Milliseconds())
	/*
		对应sql语句
		select * from senc where content like words0 and content like words1 and content like words2 and content like words3 order by content asc
	*/
	// 2. 执行搜索（数据库查询）
	searchExecStart := time.Now()
	iters := make([]*engine.TableIter, len(words))
	for i, word := range words {
		iters[i], _ = db.Tables["senc"].Search(&map[string]any{
			"content": word,
		})
		defer engine.GlobalTableIterPool.Put(iters[i])
		if i > 0 {
			data := iters[i].Map()
			defer engine.PutMap(data)
			mach := match.NewAND([]string{"did", "secNo"}, data)
			iters[0].SetMatch(mach)
		}
	}
	if did > 0 {
		/*
			限制搜索范围在当前目录内
			对应sql语句
			select id from dir where url like (select url from dir where did = [did])
			select senc.* from senc,dir where content like words0 and content like words1 and content like words2 and content like words3
			and senc.did in (select id from dir where url like (select url from dir where did = [did]))
			order by content asc

		*/
		//select url from dir where did = [did]
		dirIdIter, _ := db.Tables["dir"].Search(&map[string]any{"id": did}, util.Equal)
		defer engine.GlobalTableIterPool.Put(dirIdIter)
		dirRecords := dirIdIter.GetRecords(true, 1)
		var dirURL string
		if len(dirRecords) > 0 {
			dirURL = dirRecords[0]["url"].(string)
			//select id from dir where url like dirURL
			dirUrlIter, _ := db.Tables["dir"].Search(&map[string]any{
				"url": dirURL,
			}, util.Like)
			defer engine.GlobalTableIterPool.Put(dirUrlIter)
			mach := match.NewAND([]string{"did"}, dirUrlIter.Map()) //dir的did等于当前目录的did
			iters[0].SetMatch(mach)
		}
	}
	searchExecTime := time.Since(searchExecStart)

	// 调试输出
	fmt.Printf("搜索执行时间: %v ms\n", searchExecTime.Milliseconds())

	var results []SearchResult
	// 初始化结果切片
	results = make([]SearchResult, 0)

	// 3. 获取匹配记录
	recordsStart := time.Now()
	/*
		对应sql语句
		select senc.* from senc,dir where content like words0 and content like words1 and content like words2 and content like words3
		and senc.did in (select id from dir where url like (select url from dir where did = [did]))
		order by content asc

	*/
	records := iters[0].GetRecords(true, start, limit)
	defer record.PutRecords(records)
	totalCount := len(records)
	recordsTime := time.Since(recordsStart)

	// 调试输出
	fmt.Printf("记录获取时间: %v ms, 记录数量: %v\n", recordsTime.Milliseconds(), totalCount)

	// 4. 结果处理（包括路径获取）
	resultProcessStart := time.Now()
	// 处理分页
	currentIndex := 0
	for _, record := range records {

		result := SearchResult{
			Did:     record["did"].(int),
			SecNo:   record["secNo"].(int),
			Content: record["content"].(string),
		}
		// 尝试获取文件路径
		if recordDid, ok := record["did"].(int); ok {
			dirIter, _ := db.Tables["dir"].Search(&map[string]any{"id": recordDid}, util.Equal)
			defer engine.GlobalTableIterPool.Put(dirIter)
			dirRecords := dirIter.GetRecords(true, 1)
			if len(dirRecords) > 0 {
				if path, ok := dirRecords[0]["url"].(string); ok {
					result.Path = path
				}
			}
		}
		results = append(results, result)
		currentIndex++
	}
	resultProcessTime := time.Since(resultProcessStart)

	// 调试输出
	fmt.Printf("结果处理时间: %v ms\n", resultProcessTime.Milliseconds())

	// 总搜索时间
	totalSearchTime := time.Since(totalStartTime)
	fmt.Printf("总搜索时间: %v ms\n", totalSearchTime.Milliseconds())

	return results, totalCount, nil
}
