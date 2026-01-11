package main

import (
	"ReSearch/api"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化索引（可选，根据需要取消注释）
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

	// 设置健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "搜索引擎服务运行正常",
		})
	})

	// 启动服务器
	port := ":8080"
	fmt.Printf("搜索引擎服务启动中，监听端口 %s...\n", port)
	log.Fatal(r.Run(port))
}
