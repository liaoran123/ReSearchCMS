package config

import (
	"os"
	"path/filepath"

	"time"

	"gopkg.in/yaml.v3"
)

var Cfg *Config

// Config 配置结构体
type Config struct {
	System         SystemConfig         `yaml:"system"`
	TableDataCache TableDataCacheConfig `yaml:"table_data_cache"`
}

// SystemConfig 系统配置
type SystemConfig struct {
	Port     int    `yaml:"port"`
	DbPath   string `yaml:"dbpath"`
	Password string `yaml:"password"`
}

// TableDataCacheConfig 表数据缓存配置
type TableDataCacheConfig struct {
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

func init() {
	// 加载配置文件，尝试多个可能的路径
	var loadErr error
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
	if loadErr != nil {
		// 尝试项目根目录（固定路径）
		//获取当前执行文件的路径
		execPath, execErr := os.Executable()
		if execErr != nil {
			os.Stderr.WriteString("警告：获取当前执行文件路径失败，使用默认值\n")
		}
		// 提取目录部分
		execDir := filepath.Dir(execPath)
		Cfg, loadErr = LoadConfig(filepath.Join(execDir, "config.yaml"))
		if loadErr != nil {
			os.Stderr.WriteString("警告：配置文件加载失败，使用默认值\n")
		}
	}
}
