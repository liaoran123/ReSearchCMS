package resdb

import "sync"

type IndexJoinValue struct {
	PrimaryLen     int
	PrimaryIDValue []byte
	PrimaryValue   []byte
	OtherValues    [][]byte
}

// IndexJoinValuePool 结构体用于封装 sync.Pool，提供更好的封装和扩展性
// 用于复用 IndexJoinValue 对象，减少 GC 压力
type IndexJoinValuePool struct {
	pool sync.Pool
}

// NewIndexJoinValuePool 创建一个新的 IndexJoinValuePool 实例
func NewIndexJoinValuePool() *IndexJoinValuePool {
	return &IndexJoinValuePool{
		pool: sync.Pool{
			New: func() any {
				return &IndexJoinValue{
					PrimaryIDValue: make([]byte, 0, 64),
					PrimaryValue:   make([]byte, 0, 256),
					OtherValues:    make([][]byte, 0, 4),
				}
			},
		},
	}
}

// Acquire 从池中获取一个 IndexJoinValue
func (p *IndexJoinValuePool) Get() *IndexJoinValue {
	return p.pool.Get().(*IndexJoinValue)
}

// Release 将 IndexJoinValue 归还池中
func (p *IndexJoinValuePool) Release(v *IndexJoinValue) {
	if v == nil {
		return
	}
	// 重置字段，避免内存泄漏
	v.PrimaryLen = 0
	v.PrimaryIDValue = v.PrimaryIDValue[:0]
	v.PrimaryValue = v.PrimaryValue[:0]
	// 清空 OtherValues 中的切片，但不释放底层数组
	for i := range v.OtherValues {
		v.OtherValues[i] = v.OtherValues[i][:0]
	}
	v.OtherValues = v.OtherValues[:0]
	p.pool.Put(v)
}

// 全局默认池，兼容原有代码
var DefaultIndexJoinValuePool = NewIndexJoinValuePool()

/*
// AcquireIndexJoinValue 从全局默认池中获取一个 IndexJoinValue（兼容原有代码）
func AcquireIndexJoinValue() *IndexJoinValue {
	return DefaultIndexJoinValuePool.Get()
}

// ReleaseIndexJoinValue 将 IndexJoinValue 归还全局默认池中（兼容原有代码）
func ReleaseIndexJoinValue(v *IndexJoinValue) {
	DefaultIndexJoinValuePool.Release(v)
}
*/
