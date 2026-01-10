package files

import (
	"ReSearch/pool"
	"ReSearch/tables"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

// 全局Pool实例
var globalPool *pool.Pool
var batch storage.Batch

func init() {
	// 初始化全局Pool，数据库插入属于IO密集型任务，使用专门的IO Pool
	globalPool = pool.NewPoolForIO()
	batch = tables.Dir.GetBatch()
}

// 定义中英文句子分隔符的正则表达式
var re = regexp.MustCompile(`[。.!?？！；;\n]+`)

// TraversePathAndReadFiles 根据给定路径遍历读取所有文本文件内容
func TraversePathAndReadFiles(rootPath string) error {
	btime := time.Now()
	//打印btime
	fmt.Printf("开始遍历路径 %s 时间: %v\n", rootPath, btime)
	var currentID int
	var err error
	err = filepath.Walk(rootPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			//查询url是否存在
			iter := tables.Dir.Search(&map[string]any{"url": path}, util.Equal)
			defer iter.Release()
			if iter.Exist() {
				fmt.Printf("目录 %s 已存在，跳过\n", path)
				// 目录已存在，跳过
				//return nil //目录存在不插入，但是文章可能是更新过，所以需要更新文章
			} else {
				// 插入目录到数据库
				currentID, err = tables.Dir.Insert(&map[string]any{
					"name": info.Name(),
					"url":  path,
					"ext":  filepath.Ext(path), //文件后缀
				}, batch)
				if err != nil {
					fmt.Printf("插入目录 %s 失败: %v\n", path, err)
					return err
				}
			}
			// 跳过目录
			if info.IsDir() {
				return nil
			}
			// 读取文件内容
			content, readErr := ReadFileContent(path)
			if readErr != nil {
				return readErr
			}
			content += "<" + info.Name() + ">\n" + content //可以使用"<"+关键词,指定搜索标题
			// 调用AddArticle并检查错误
			if err := AddArticle(currentID, info.Name(), content); err != nil {
				fmt.Printf("插入文章 %s 失败: %v\n", info.Name(), err)
				return err
			}
			return nil
		})

	endTime := time.Now()
	//打印endTime
	fmt.Printf("遍历路径 %s 结束时间: %v\n", rootPath, endTime)
	// 打印耗时
	fmt.Printf("遍历路径 %s 耗时: %v\n", rootPath, time.Since(btime))
	globalPool.Stop()
	return err
}

func AddArticle(mid int, title, content string) error {
	// 定义中英文句子分隔符的正则表达式
	parts := re.Split(content, -1)
	var wg sync.WaitGroup
	// 过滤空字符串和纯空白字符串
	secNo := 0
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			secNo++
			wg.Add(1) //+1，外面加1
			// 捕获当前循环变量
			sentence := trimmed
			globalPool.Submit(func() {
				defer wg.Done() //-1，里面减1
				// 插入文章到数据库
				_, err := tables.Article.Insert(&map[string]any{
					"mid":     mid,   //文章目录ID
					"secNo":   secNo, //文章句子序号
					"title":   title,
					"content": sentence,
				}, batch)
				//fmt.Printf("secNo: %v\n", secNo)
				if err != nil {
					fmt.Println("插入文章失败:", err)
				}
			})
		}
	}
	// 等待所有任务完成
	wg.Wait()
	return nil
}
