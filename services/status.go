package services

import (
	"ReSearch/db"
	"time"

	"github.com/liaoran123/sfsDb/engine"
)

// Status 搜索引擎状态结构
type Status struct {
	IndexedFiles     int       `json:"indexed_files"`
	IndexedSentences int       `json:"indexed_sentences"`
	LastUpdated      time.Time `json:"last_updated"`
	Status           string    `json:"status"`
}

// GetStatus 获取搜索引擎状态
func GetStatus() (Status, error) {
	// 获取目录数量
	dirIter := db.Tables["dir"].ForData()
	defer engine.GlobalTableIterPool.Put(dirIter)
	dirRecords := dirIter.GetRecords(true)
	indexedFiles := len(dirRecords)

	// 获取句子数量
	sencIter := db.Tables["senc"].ForData()
	defer engine.GlobalTableIterPool.Put(sencIter)
	sencRecords := sencIter.GetRecords(true)
	indexedSentences := len(sencRecords)

	status := Status{
		IndexedFiles:     indexedFiles,
		IndexedSentences: indexedSentences,
		LastUpdated:      time.Now(),
		Status:           "running",
	}

	return status, nil
}
