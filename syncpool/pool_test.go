package syncpool

import (
	"sync"
	"testing"
	"time"
)

// TestPool 测试Pool的基本功能
func TestPool(t *testing.T) {
	// 创建工作池
	pool := NewPool()
	defer pool.Stop()

	// 测试任务数量
	const taskCount = 100
	var wg sync.WaitGroup
	results := make([]int, taskCount)

	// 提交任务
	for i := range taskCount {
		wg.Add(1)
		index := i
		pool.Submit(func() {
			defer wg.Done()
			// 简单的任务：睡眠1毫秒，然后设置结果
			time.Sleep(time.Millisecond)
			results[index] = index
		})
	}

	// 等待所有任务完成
	wg.Wait()

	// 验证结果
	for i := range taskCount {
		if results[i] != i {
			t.Errorf("任务%d执行失败，结果为%d，期望为%d", i, results[i], i)
		}
	}
}

// TestPoolConcurrency 测试Pool的并发性能
func TestPoolConcurrency(t *testing.T) {
	// 创建工作池
	pool := NewPool()
	defer pool.Stop()

	// 测试并发性能
	const taskCount = 1000
	var wg sync.WaitGroup
	start := time.Now()

	// 提交任务
	for range taskCount {
		wg.Add(1)
		pool.Submit(func() {
			defer wg.Done()
			// 简单的计算任务
			for j := range 100000 {
				_ = j * j
			}
		})
	}

	// 等待所有任务完成
	wg.Wait()
	duration := time.Since(start)
	t.Logf("执行%d个任务耗时%v", taskCount, duration)
}

// TestPoolStop 测试Pool的Stop方法
func TestPoolStop(t *testing.T) {
	// 创建工作池
	pool := NewPool()

	// 提交少量任务
	const taskCount = 10
	var wg sync.WaitGroup

	for range taskCount {
		wg.Add(1)
		pool.Submit(func() {
			defer wg.Done()
			time.Sleep(time.Millisecond * 50)
		})
	}

	// 等待任务开始执行
	time.Sleep(time.Millisecond * 10)

	// 停止工作池
	pool.Stop()

	// 确保所有任务都已完成
	wg.Wait()
	t.Log("工作池已成功停止")
}
