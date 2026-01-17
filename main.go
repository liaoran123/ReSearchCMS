package main

import (
	"ReSearch/api"
	"ReSearch/config"
	"ReSearch/i18n"
	"ReSearch/web"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 获取可执行文件所在目录的绝对路径
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error getting executable path: %v", err)
	}
	baseDir := filepath.Dir(execPath)

	// 初始化多语言支持
	translator := i18n.NewTranslator()
	localesDir := filepath.Join(baseDir, "web", "templates", "locales")
	if err := translator.LoadLocales(localesDir); err != nil {
		log.Printf("Error loading locales: %v", err)
	}

	// 创建 Gin 引擎
	r := gin.Default()

	// 设置 CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 设置 API 路由
	api.SetupRoutes(r)

	// 设置 Web 路由
	webServer := web.NewWebServer(translator, baseDir)
	webServer.SetupRoutes(r)

	// 设置健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "考据级文档搜索引擎服务运行正常",
		})
	})

	// 启动服务器
	// 读取配置文件
	port := ":" + strconv.Itoa(config.Cfg.Web.Port)
	fmt.Printf("考据级文档搜索引擎服务启动中，监听端口 %s...\n", port)

	// 自动调用浏览器打开地址
	go func() {
		// 等待服务器启动
		time.Sleep(1 * time.Second)
		url := "http://localhost" + port
		var cmd string
		var args []string

		switch runtime.GOOS {
		case "windows":
			cmd = "cmd"
			args = []string{"/c", "start", url}
		case "darwin":
			cmd = "open"
			args = []string{url}
		case "linux":
			cmd = "xdg-open"
			args = []string{url}
		default:
			return
		}

		if err := exec.Command(cmd, args...).Start(); err != nil {
			log.Printf("Error opening browser: %v", err)
		}
	}()

	// 启动服务器
	log.Fatal(r.Run(port))
}
