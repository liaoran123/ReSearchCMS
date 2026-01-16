package services

import (
	"ReSearch/db"
	"sort"
	"strings"

	"github.com/liaoran123/sfsDb/engine"
)

// SearchResult 搜索结果结构
type SearchResult struct {
	Did     int    `json:"did"`
	SecNo   int    `json:"secNo"`
	Content string `json:"content"`
	Path    string `json:"path,omitempty"`
}

// Search 执行搜索
func Search(query string, start, limit int) ([]SearchResult, int, error) {
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
	iters := make([]*engine.TableIter, len(words))
	for i, word := range words {
		iters[i] = db.Tables["senc"].Search(&map[string]any{
			"content": word,
		})
		defer iters[i].Release()
		if i > 0 {
			mach := engine.NewAND([]string{"did", "secNo"}, iters[i].Map())
			iters[0].SetMatch(mach)
		}
	}

	var results []SearchResult

	// 初始化结果切片
	results = make([]SearchResult, 0)

	// 获取所有匹配记录
	records := iters[0].GetRecords(true, start, limit)
	totalCount := len(records)

	// 处理分页
	currentIndex := 0
	for _, record := range records {

		result := SearchResult{
			Did:     record["did"].(int),
			SecNo:   record["secNo"].(int),
			Content: record["content"].(string),
		}

		// 尝试获取文件路径
		if did, ok := record["did"].(int); ok {
			dirIter := db.Tables["dir"].Search(&map[string]any{"id": did})
			defer dirIter.Release()
			dirRecords := dirIter.GetRecords(true)
			if len(dirRecords) > 0 {
				if path, ok := dirRecords[0]["url"].(string); ok {
					result.Path = path
				}
			}
		}

		results = append(results, result)
		currentIndex++
	}

	return results, totalCount, nil
}
