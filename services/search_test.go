package services

import (
	"fmt"
	"sync"
	"testing"
)

// BenchmarkSearch 测试单查询搜索性能
func BenchmarkSearch(b *testing.B) {
	// 测试数据
	testQueries := []string{
		"test",
		"文档搜索",
		"性能测试",
		"考据级文档搜索引擎",
	}

	// 对每个测试查询运行基准测试
	for _, query := range testQueries {
		b.Run(query, func(b *testing.B) {
			// 重置计时器
			b.ResetTimer()

			// 运行b.N次搜索
			for i := 0; i < b.N; i++ {
				_, _, err := Search(query, 0, 0, 10)
				if err != nil {
					b.Fatalf("搜索失败: %v", err)
				}
			}
		})
	}
}

// BenchmarkSearchWithDirectory 测试带目录限制的搜索性能
func BenchmarkSearchWithDirectory(b *testing.B) {
	// 测试数据
	query := "test"
	testDirectories := []int{0, 1, 2}

	// 对每个测试目录运行基准测试
	for _, did := range testDirectories {
		b.Run(fmt.Sprintf("dir_%d", did), func(b *testing.B) {
			// 重置计时器
			b.ResetTimer()

			// 运行b.N次搜索
			for i := 0; i < b.N; i++ {
				_, _, err := Search(query, did, 0, 10)
				if err != nil {
					b.Fatalf("搜索失败: %v", err)
				}
			}
		})
	}
}

// BenchmarkSearchWithPagination 测试不同分页参数的搜索性能
func BenchmarkSearchWithPagination(b *testing.B) {
	// 测试数据
	query := "test"
	testPagination := []struct {
		name  string
		start int
		limit int
	}{
		{"start_0_limit_5", 0, 5},
		{"start_0_limit_10", 0, 10},
		{"start_10_limit_10", 10, 10},
		{"start_100_limit_20", 100, 20},
	}

	// 对每个分页参数运行基准测试
	for _, pag := range testPagination {
		b.Run(pag.name, func(b *testing.B) {
			// 重置计时器
			b.ResetTimer()

			// 运行b.N次搜索
			for i := 0; i < b.N; i++ {
				_, _, err := Search(query, 0, pag.start, pag.limit)
				if err != nil {
					b.Fatalf("搜索失败: %v", err)
				}
			}
		})
	}
}

// BenchmarkSearchConcurrent 测试并发搜索性能
func BenchmarkSearchConcurrent(b *testing.B) {
	// 测试数据
	query := "test"
	testConcurrency := []int{1, 2, 4, 8, 16}

	// 对每个并发数运行基准测试
	for _, concurrency := range testConcurrency {
		b.Run(fmt.Sprintf("concurrency_%d", concurrency), func(b *testing.B) {
			// 重置计时器
			b.ResetTimer()

			// 运行b.N次搜索，每次使用指定的并发数
			for i := 0; i < b.N; i++ {
				// 使用WaitGroup等待所有goroutine完成
				var wg sync.WaitGroup
				wg.Add(concurrency)

				// 启动指定数量的goroutine
				for j := 0; j < concurrency; j++ {
					go func() {
						defer wg.Done()
						_, _, err := Search(query, 0, 0, 10)
						if err != nil {
							b.Fatalf("并发搜索失败: %v", err)
						}
					}()
				}

				// 等待所有goroutine完成
				wg.Wait()
			}
		})
	}
}

// BenchmarkSearchQueryLength 测试不同查询词长度的搜索性能
func BenchmarkSearchQueryLength(b *testing.B) {
	// 测试数据
	testQueries := []string{
		"a",
		"ab",
		"abc",
		"abcd",
		"abcde",
		"abcdef",
		"abcdefg",
		"abcdefgh",
	}

	// 对每个查询词长度运行基准测试
	for _, query := range testQueries {
		b.Run(fmt.Sprintf("length_%d", len(query)), func(b *testing.B) {
			// 重置计时器
			b.ResetTimer()

			// 运行b.N次搜索
			for i := 0; i < b.N; i++ {
				_, _, err := Search(query, 0, 0, 10)
				if err != nil {
					b.Fatalf("搜索失败: %v", err)
				}
			}
		})
	}
}
