package storedb

import (
	"researchCms/config"
	"sync"
	"time"
)

// 全局数据迭代器缓存，默认超时时间为5分钟
var IterCaches *IterCache

func init() {
	// 使用配置初始化缓存
	IterCacheNew(config.Cfg.IterCache.Max, config.Cfg.IterCache.Timeout)
	/*// 启动定时器，每5分钟执行一次CheckAllExpire
	// 定时器太耗资源
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
type IterCache struct {
	iterMap sync.Map
	//最新命中时间
	hit map[string]time.Time
	//超时时间
	timeout time.Duration
	max     int
}

func IterCacheNew(max int, timeout time.Duration) *IterCache {
	// 初始化全局数据迭代器缓存IterCaches，保证在多线程环境下安全，只有唯一一个实例
	IterCaches = &IterCache{
		iterMap: sync.Map{},
		hit:     make(map[string]time.Time),
		timeout: timeout,
		max:     max,
	}
	return IterCaches
}

// 存储数据迭代器
func (c *IterCache) Store(key string, iter *Iter) {
	// 当缓存中的数据迭代器数量超过最大容量的90%时，触发过期检查
	if len(c.hit) > c.max/10*9 {
		go c.CheckAllExpire()
	}
	c.iterMap.Store(key, iter)
	c.hit[key] = time.Now()
}

// 加载数据迭代器
func (c *IterCache) Load(key string) (*Iter, bool) {
	iter, ok := c.iterMap.Load(key)
	if ok {
		c.hit[key] = time.Now()
		return iter.(*Iter), ok
	}
	return nil, ok
}

// 检查数据迭代器是否过期，过期则删除
func (c *IterCache) CheckExpire(key string) bool {
	if hit, ok := c.hit[key]; ok {
		if time.Since(hit) > c.timeout {
			if val, ok := c.iterMap.Load(key); ok {
				if iter, ok := val.(*Iter); ok {
					iter.Release() // 删除前先释放数据迭代器
				}
			}
			c.iterMap.Delete(key)
			delete(c.hit, key)
			return true
		}
	}
	return false
}

// 检查所有数据迭代器是否过期，过期则删除
func (c *IterCache) CheckAllExpire() {
	for key := range c.hit {
		c.CheckExpire(key)
	}
}
