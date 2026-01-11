package services

import (
	"ReSearch/db"
	"time"
)

// Status 搜索引擎状态结构
type Status struct {
	IndexedFiles int       `json:"indexed_files"`
	IndexedSentences int    `json:"indexed_sentences"`
	LastUpdated  time.Time `json:"last_updated"`
	Status       string    `json:"status"`
}

// GetStatus 获取搜索引擎状态
func GetStatus() (Status, error) {
	// 获取目录数量
	dirIter := db.Tables["dir"].ForData()
	defer dirIter.Release()
	dirRecords := dirIter.GetRecords(true)
	indexedFiles := len(dirRecords)

	// 获取句子数量
	sencIter := db.Tables["senc"].ForData()
	defer sencIter.Release()
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
