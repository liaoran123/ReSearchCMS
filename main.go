package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("考据级文档搜索引擎!免费版有广告，付费版无广告，企业版提供多维分析统计功能。")
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)
	fmt.Println(execDir)
	configPath := filepath.Join(execDir, "config.yaml")
	fmt.Println(configPath)
}
