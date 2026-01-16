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

		// 上传相关路由
		api.POST("/upload/logo", UploadLogoHandler)

		// 搜索建议相关路由
		api.GET("/suggestions", GetSuggestionsHandler)

		// 内容查询路由
		api.GET("/content", GetContentByDidAndSecNoHandler)
		
		// 目录浏览路由
		api.GET("/directory", DirectoryHandler)
		
		// 文章内容路由
		api.GET("/article/:id", ArticleContentHandler)
	}
}
