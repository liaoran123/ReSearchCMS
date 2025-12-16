package config

import (
	"os"
	"path/filepath"

	"time"

	"gopkg.in/yaml.v3"
)

var Cfg *Config

func init() {
	// 加载配置文件，尝试多个可能的路径
	execPath, loadErr := os.Executable() //测试运行时，os.Executable() 是当前目录的路径
	if loadErr != nil {
		os.Stderr.WriteString("警告：获取当前执行文件路径失败，使用默认值\n")
	}
	execDir := filepath.Dir(execPath)
	Cfg, loadErr = LoadConfig(filepath.Join(execDir, "config.yaml"))
	if loadErr != nil {
		// 尝试当前目录
		currentDir, err := os.Getwd()
		if err != nil {
			os.Stderr.WriteString("警告：获取当前工作目录失败，使用默认值\n")
		}
		Cfg, loadErr = LoadConfig(filepath.Join(currentDir, "config.yaml"))
		if loadErr != nil {
			// 尝试上级目录
			Cfg, loadErr = LoadConfig("../config.yaml")
		}
	}
}

// Config 配置结构体
type Config struct {
	System    SystemConfig    `yaml:"system"`
	IterCache IterCacheConfig `yaml:"iter_cache"`
}

// SystemConfig 系统配置
type SystemConfig struct {
	Port     int    `yaml:"port"`
	DbPath   string `yaml:"dbpath"`
	Password string `yaml:"password"`
}

// IterCacheConfig 数据迭代器缓存配置
type IterCacheConfig struct {
	Timeout time.Duration `yaml:"timeout"`
	Max     int           `yaml:"max"`
}

// LoadConfig 从yaml文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 解析yaml
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
