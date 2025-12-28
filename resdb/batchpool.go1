package resdb

import (
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

func init() {
	GlobalBatchPool = BatchPoolNew()
}

// BatchPool 自定义的 Pool，自动处理 Reset
type BatchPool struct {
	pool      sync.Pool
	Allocated int64 // 累计分配次数
	Returned  int64 // 累计归还次数
}

// 全局 BatchPool 实例
var GlobalBatchPool *BatchPool

// BatchPoolNew 创建一个新的 BatchPool
func BatchPoolNew() *BatchPool {
	return &BatchPool{
		pool: sync.Pool{
			New: func() any {
				return new(leveldb.Batch)
			},
		},
	}
}

// Get 从对象池获取一个 Batch
func (p *BatchPool) Get() *leveldb.Batch {
	p.Allocated++
	return p.pool.Get().(*leveldb.Batch)
}

// Put 将 Batch 放回对象池，并自动清空
func (p *BatchPool) Put(b *leveldb.Batch) {
	b.Reset() // 直接调用，不需要 defer
	p.Returned++
	p.pool.Put(b)
}

// Clear 清空对象池中的所有对象
func (p *BatchPool) Clear() {
	// 清空对象池的方法，通过不断获取直到返回 nil（sync.Pool 没有直接的清空方法）
	for {
		if b := p.pool.Get(); b == nil {
			break
		}
		// 直接丢弃，不进行统计
	}
}

// Stats 返回对象池的统计信息
func (p *BatchPool) Stats() map[string]int64 {
	return map[string]int64{
		"allocated": p.Allocated,
		"returned":  p.Returned,
	}
}
