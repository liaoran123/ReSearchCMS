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

// -----------------------------------------------------
// []byte 对象池，用于复用 []byte
var GlobalBytesPool = &sync.Pool{
	New: func() any {
		return make([]byte, 0, 128) // 初始容量 1024，可按需调整
	},
}

// 从池中获取一个 []byte
func GetBytes() []byte {
	return GlobalBytesPool.Get().([]byte)
}

// 将 []byte 归还池中，重置长度但不释放底层数组
func PutBytes(b []byte) {
	b = b[:0] // 仅截断长度，保留容量
	GlobalBytesPool.Put(b)
}

// -----------------------------------------------------
// [][]byte 对象池，用于复用二维字节数组
var GlobalBytesArrayPool = &sync.Pool{
	New: func() any {
		return make([][]byte, 0, 64) // 初始容量 64，可按需调整
	},
}

// 从池中获取一个 [][]byte
func GetBytesArray() [][]byte {
	return GlobalBytesArrayPool.Get().([][]byte)
}

// 将 [][]byte 归还池中，重置长度但不释放底层数组
func PutBytesArray(b [][]byte) {
	b = b[:0] // 仅截断长度，保留容量
	GlobalBytesArrayPool.Put(b)
}

// -----------------------------------------------------
// map[string]any 对象池，用于复用 map[string]any
var GlobalMapAnyPool = &sync.Pool{
	New: func() any {
		return make(map[string]any)
	},
}

// 从池中获取一个 map[string]any
func GetMapAny() map[string]any {
	return GlobalMapAnyPool.Get().(map[string]any)
}

// 将 map[string]any 归还池中，清空 map 内容
func PutMapAny(m map[string]any) {
	for k := range m {
		delete(m, k)
	}
	GlobalMapAnyPool.Put(m)
}

// -----------------------------------------------------
// []map[string]any 对象池，用于复用切片 map[string]any
var GlobalMapsAnyPool = &sync.Pool{
	New: func() any {
		return make([]map[string]any, 0, 16) // 初始容量 16，可按需调整
	},
}

// 从池中获取一个 []map[string]any
func GetMapsAny() []map[string]any {
	return GlobalMapsAnyPool.Get().([]map[string]any)
}

// 将 []map[string]any 归还池中，清空切片内容
func PutMapsAny(m []map[string]any) {
	// 先逐个清空内部 map
	for i := range m {
		for k := range m[i] {
			delete(m[i], k)
		}
	}
	// 截断切片长度，保留容量
	m = m[:0]
	GlobalMapsAnyPool.Put(m)
}

// -----------------------------------------------------
// -----------------------------------------------------
// []string 对象池，用于复用字符串切片
var GlobalStringSlicePool = &sync.Pool{
	New: func() any {
		return make([]string, 0, 64) // 初始容量 64，可按需调整
	},
}

// 从池中获取一个 []string
func GetStringSlice() []string {
	return GlobalStringSlicePool.Get().([]string)
}

// 将 []string 归还池中，重置长度但不释放底层数组
func PutStringSlice(s []string) {
	s = s[:0] // 仅截断长度，保留容量
	GlobalStringSlicePool.Put(s)
}
