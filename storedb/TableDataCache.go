package storedb

import (
	"sync"
	"time"
)

// 数据迭代器缓存默认超时时间
var timeout time.Duration = time.Minute * 5

// 全局数据迭代器缓存，默认超时时间为5分钟
var TDCache *TableDataCache

// 启动定时器，每5分钟执行一次CheckAllExpire
func init() {
	TableDataCacheNew(10000, timeout)
	/*定时器太耗资源
	go func() {
		ticker := time.NewTicker(timeout)
		defer ticker.Stop()
		for range ticker.C {
			TDCache.CheckAllExpire()
		}
	}()
	*/
}

// 数据迭代器缓存
type TableDataCache struct {
	td sync.Map
	//最新命中时间
	hit map[string]time.Time
	//超时时间
	timeout time.Duration
	max     int
}

func TableDataCacheNew(max int, timeout time.Duration) *TableDataCache {
	// 初始化全局数据迭代器缓存TDCache，保证在多线程环境下安全，只有唯一一个实例
	TDCache = &TableDataCache{
		td:      sync.Map{},
		hit:     make(map[string]time.Time),
		timeout: timeout,
		max:     max,
	}
	return TDCache
}

// 存储数据迭代器
func (c *TableDataCache) Store(key string, td *TableData) {
	if len(c.hit) > c.max/10*9 {
		go c.CheckAllExpire()
	}
	c.td.Store(key, td)
	if _, ok := c.hit[key]; ok {
		c.hit[key] = time.Now()
	}
}

// 加载数据迭代器
func (c *TableDataCache) Load(key string) (*TableData, bool) {
	td, ok := c.td.Load(key)
	if ok {
		c.hit[key] = time.Now()
		return td.(*TableData), ok
	}
	return nil, ok
}

// 检查数据迭代器是否过期，过期则删除
func (c *TableDataCache) CheckExpire(key string) bool {
	if hit, ok := c.hit[key]; ok {
		if time.Since(hit) > c.timeout {
			if val, ok := c.td.Load(key); ok {
				if td, ok := val.(*TableData); ok {
					td.Release() // 删除前先释放数据迭代器
				}
			}
			c.td.Delete(key)
			delete(c.hit, key)
			return true
		}
	}
	return false
}

// 检查所有数据迭代器是否过期，过期则删除
func (c *TableDataCache) CheckAllExpire() {
	for key := range c.hit {
		c.CheckExpire(key)
	}
}
