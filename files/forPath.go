package files

import (
	"os"
	"path/filepath"
)

// TraversePathAndReadFiles 根据给定路径遍历读取所有文本文件内容
func TraversePathAndReadFiles(rootPath string) (map[string]string, error) {
	result := make(map[string]string)
	err := filepath.Walk(rootPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
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
			// 保存到结果中，key 为文件路径，value 为文本内容
			result[path] = string(content)
			return nil
		})

	return result, err
}
