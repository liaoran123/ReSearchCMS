package api

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置 API 路由
func SetupRoutes(r *gin.Engine) {
	// API 路由组
	api := r.Group("/api")
	{
		// 搜索相关路由
		api.GET("/search", SearchHandler)
		
		// 索引相关路由
		api.POST("/index", IndexHandler)
		api.POST("/index/path", IndexPathHandler)
		
		// 状态相关路由
		api.GET("/status", StatusHandler)
	}
}
