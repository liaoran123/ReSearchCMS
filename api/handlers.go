package api

import (
	"ReSearch/files"
	"ReSearch/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SearchHandler 处理搜索请求
func SearchHandler(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "搜索查询不能为空",
		})
		return
	}

	results, err := services.Search(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "搜索失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   query,
		"results": results,
		"count":   len(results),
	})
}

// IndexHandler 处理文件索引请求
func IndexHandler(c *gin.Context) {
	filePath := c.PostForm("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "文件路径不能为空",
		})
		return
	}

	err := files.TraversePathAndReadFiles(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "索引失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "文件索引成功",
		"path":    filePath,
	})
}

// IndexPathHandler 处理目录索引请求
func IndexPathHandler(c *gin.Context) {
	var request struct {
		Path string `json:"path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数",
		})
		return
	}

	err := files.TraversePathAndReadFiles(request.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "目录索引失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "目录索引成功",
		"path":    request.Path,
	})
}

// StatusHandler 处理状态请求
func StatusHandler(c *gin.Context) {
	status, err := services.GetStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取状态失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}
