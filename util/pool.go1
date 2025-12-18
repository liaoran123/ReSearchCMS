package util

import (
	"runtime"
	"sync"
)

// Pool 进程池
type Pool struct {
	wg   sync.WaitGroup
	work chan func()
}

// NewPool 创建并返回一个拥有最佳并发数的 Pool
func NewPool() *Pool {
	// 根据 CPU 核心数设置最佳并发数
	// NewPool 创建并返回一个拥有最佳并发数的 Pool
	// 对于计算密集型任务，最佳并发数为 CPU 核心数
	// 对于 IO 密集型任务，最佳并发数为 CPU 核心数的 8-16 倍
	// 对于混合型任务，最佳并发数为 CPU 核心数的 2-4 倍
	wNum := runtime.NumCPU()
	p := &Pool{
		work: make(chan func(), wNum),
	}
	for range wNum {
		p.wg.Add(1)
		go p.worker()
	}
	return p
}

// worker 持续从 work 通道中取出任务并执行
func (p *Pool) worker() {
	defer p.wg.Done()
	for task := range p.work {
		task()
	}
}

// Submit 向 Pool 提交一个	任务
func (p *Pool) Submit(task func()) {
	p.work <- task
}

// Stop 关闭 work 通道并等待所有 worker 退出
func (p *Pool) Stop() {
	close(p.work)
	p.wg.Wait()
}
