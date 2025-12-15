package syncpool

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
	workers := runtime.NumCPU()
	p := &Pool{
		work: make(chan func(), workers),
	}
	for range workers {
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

// Submit 向 Pool 提交一个任务
func (p *Pool) Submit(task func()) {
	p.work <- task
}

// Stop 关闭 work 通道并等待所有 worker 退出
func (p *Pool) Stop() {
	close(p.work)
	p.wg.Wait()
}
