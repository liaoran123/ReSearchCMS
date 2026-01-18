package files

import (
	"ReSearch/db"
	"ReSearch/pool"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/liaoran123/sfsDb/util"
)

// 全局Pool实例
var globalPool *pool.Pool

//var batch storage.Batch

func init() {
	// 初始化全局Pool，数据库插入属于IO密集型任务，使用专门的IO Pool
	globalPool = pool.NewPoolForIO()
	//batch = db.Store.GetBatch()
	currentTime = time.Now().Format("2006-01-02 15:04:05")
}

var currentTime string

const spstr = "。!?？！；;" //支持中文为主的分隔符
//const spstr = `。.!?？！；;\n` //这里\n是两个字符。双引号才是一个字符，换行符。

// TraversePathAndReadFiles 根据给定路径遍历读取所有文本文件内容
func TraversePathAndReadFiles(rootPath string) (map[string]int, map[string]int, error) {
	btime := time.Now()
	//打印btime
	fmt.Printf("开始遍历路径 %s 时间: %v\n", rootPath, btime)
	var currentID int
	//var err error
	var content string
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error
	// 定义一个map，用于统计所有的文件类型数量
	fileTypes := make(map[string]int)
	// 定义一个map，用于统计不支持的文件类型数量
	unsupportedFiles := make(map[string]int)

	_ = filepath.Walk(rootPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			/*
				if path == `E:\工具\代理\ChromeGo\chrome-user-data\Default\Extensions\bhghoamapcdpbohphigoooaddinpkbai\8.0.1_0\_locales\id` {
					fmt.Println(path)
				}*/
			ext := filepath.Ext(path)
			//统计文件类型数量
			if ext == "" {
				fileTypes["目录"]++
			} else {
				fileTypes[ext]++
			}
			//检测目录是否存在
			dirurliter := db.Tables["dir"].Search(&map[string]any{
				"url": path,
			}, util.Equal)
			defer dirurliter.Release()
			if dirurliter.Exist() {
				record := dirurliter.GetRecords(true, 1).Select("id")
				if len(record) == 0 {
					fmt.Printf("目录 %s 不存在记录\n", path)
				}
				if record[0]["id"] != nil {
					currentID = record[0]["id"].(int)
				} else {
					fmt.Printf("目录 %s 存在记录，但id字段为空\n", path)
					currentID = -1
				}
				if len(record) > 1 {
					fmt.Printf("目录 %s 存在多个记录: %v\n", path, record)
				}
			} else {
				// 插入目录到数据库
				currentID, err = db.Tables["dir"].Insert(&map[string]any{
					"name": info.Name(),
					"url":  path,
					"ext":  ext, //文件后缀
				})
				if err != nil {
					fmt.Printf("插入目录 %s 失败: %v\n", path, err)
					return err
				}
			}
			// 读取文件内容
			if !info.IsDir() {
				content, err = ReadFileContent(path)
				if err != nil {
					content = "" // 读取失败时，内容设为空字符串
					//以后缀名分组统计不支持文件数量
					if ext == "" {
						unsupportedFiles["其他"]++
					} else {
						unsupportedFiles[ext]++
					}
				}
			} else {
				content = ""
			}
			name := info.Name()
			//删除后缀
			name = strings.TrimSuffix(name, ext)
			content = "<" + name + ">\n" + content //可以使用"<"+关键词,指定搜索标题

			// 真正的多线程处理
			wg.Add(1)
			// 捕获外部变量
			//fmt.Printf("currentID: %v\n", currentID)
			localCurrentID := currentID
			localContent := content
			localName := name

			globalPool.Submit(func() {
				defer wg.Done()
				if err := AddArticle(localCurrentID, localContent); err != nil {
					fmt.Printf("插入文章 %s 失败: %v\n", localName, err)
					mu.Lock()
					errs = append(errs, err)
					mu.Unlock()
				}
			})
			return nil
		})

	// 等待所有任务完成
	wg.Wait()

	endTime := time.Now()
	//打印endTime
	fmt.Printf("遍历路径 %s 结束时间: %v\n", rootPath, endTime)
	// 打印耗时
	fmt.Printf("遍历路径 %s 耗时: %v\n", rootPath, time.Since(btime))
	// 不要停止线程池，保持其运行状态
	//globalPool.Stop()

	// 返回第一个错误
	if len(errs) > 0 {
		return fileTypes, unsupportedFiles, errs[0]
	}
	return fileTypes, unsupportedFiles, nil
}

// addSentence 保存句子到数据库
func addSentence(did, secNo int, content string) error {
	// 插入句子到数据库
	id, err := db.Tables["senc"].Insert(&map[string]any{
		"did":     did,     //文章目录ID
		"secNo":   secNo,   //文章句子序号
		"content": content, // 包含分隔符的完整句子
	})
	if err != nil {
		//打印错误信息
		fmt.Printf("插入句子失败: %v, did: %v, secNo: %v\n", err, did, secNo)
		return err
	}
	if id == -1 {
		fmt.Printf("插入句子失败: 插入ID为 -1, did: %v, secNo: %v\n", did, secNo)
		return fmt.Errorf("插入句子失败: 插入ID为 -1")
	}
	return nil
}

func AddArticle(did int, content string) error {
	//将原来的换行符替换为当前时间戳换行符标识 <setting>\n[setting]\r\nopenid=104\r\n
	content = strings.ReplaceAll(content, "\n", currentTime+"\n")
	//将所有分隔符附加上换行符
	for _, sep := range spstr {
		content = strings.ReplaceAll(content, string(sep), string(sep)+"\n")
	}
	parts := strings.Split(content, "\n")
	// 过滤空字符串和纯空白字符串，并合并分隔符
	secNo := 0
	// 遍历所有部分
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		//恢复换行符
		trimmed = strings.ReplaceAll(trimmed, currentTime, "\n")
		secNo++
		// 插入文章到数据库
		err := addSentence(did, secNo, trimmed)
		if err != nil {
			return err
		}

	}
	return nil
}
