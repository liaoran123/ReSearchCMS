package services

import (
	"ReSearch/db"

	"github.com/liaoran123/sfsDb/util"
)

// SearchResult 搜索结果结构
type SearchResult struct {
	ID      int     `json:"id"`
	Content string  `json:"content"`
	Score   float64 `json:"score"`
	Path    string  `json:"path,omitempty"`
}

// Search 执行搜索
func Search(query string) ([]SearchResult, error) {
	// 在 senc 表中搜索包含查询词的记录
	iter := db.Tables["senc"].Search(&map[string]any{"content": query}, util.Like)
	defer iter.Release()

	var results []SearchResult
	for iter.Next() {
		// 解析记录
		record := iter.ParseRecord(iter.Key(), iter.Value())
		if record == nil {
			continue
		}

		result := SearchResult{
			ID:      record["did"].(int),
			Content: record["content"].(string),
			Score:   1.0, // 简化处理，实际应该计算相关性得分
		}

		// 尝试获取文件路径
		if did, ok := record["did"].(int); ok {
			dirIter := db.Tables["dir"].Search(&map[string]any{"id": did}, util.Equal)
			defer dirIter.Release()
			if dirIter.Exist() {
				dirRecord := dirIter.ParseRecord(dirIter.Key(), dirIter.Value())
				if dirRecord != nil {
					if path, ok := dirRecord["url"].(string); ok {
						result.Path = path
					}
				}
			}
		}

		results = append(results, result)
	}

	return results, nil
}
