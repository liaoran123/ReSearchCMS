package files

import (
	"os"
	"path/filepath"
	"researchCms/tables"

	"strconv"
)

// TraversePathAndReadFiles 根据给定路径遍历读取所有文本文件内容
func TraversePathAndReadFiles(rootPath string) error {
	err := filepath.Walk(rootPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			//打印目录名称
			println(path)

			// 插入目录到数据库
			tables.Dir.Insert(&map[string]any{
				"name": info.Name(),
				"url":  path,
				"ext":  filepath.Ext(path), //文件后缀
			})
			// 跳过目录
			if info.IsDir() {
				return nil
			}
			// 读取文件内容
			content, readErr := ReadFileContent(path)
			if readErr != nil {
				return readErr
			}
			// 保存到结果中，key 为文件路径，value 为文本内容
			println("文件路径：" + path)
			println("文件内容长度：" + strconv.Itoa(len(content)))
			return nil
		})

	return err
}
