package api

import (
	"ReSearch/config"
	"ReSearch/db"
	"ReSearch/files"
	"ReSearch/services"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/liaoran123/sfsDb/util"
	"gopkg.in/yaml.v3"
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

	// 获取分页参数
	startStr := c.DefaultQuery("start", "0")
	limitStr := c.DefaultQuery("limit", "21")

	start, err := strconv.Atoi(startStr)
	if err != nil {
		start = 0
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 21
	}

	// 开始计时
	startTime := time.Now()

	results, totalCount, err := services.Search(query, start, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "搜索失败: " + err.Error(),
		})
		return
	}

	// 计算搜索用时
	duration := time.Since(startTime)
	searchTime := duration.Milliseconds()

	c.JSON(http.StatusOK, gin.H{
		"query":      query,
		"results":    results,
		"count":      totalCount,
		"start":      start,
		"limit":      limit,
		"searchTime": searchTime,
	})
}

// IndexHandler 处理文件索引请求
func IndexHandler(c *gin.Context) {
	filePath := c.PostForm("path")
	password := c.PostForm("password")

	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "文件路径不能为空",
		})
		return
	}

	if password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "密码不能为空",
		})
		return
	}

	// 验证密码
	if password != config.Cfg.Web.Password {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "密码错误",
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
		Path       string `json:"path" binding:"required"`
		Password   string `json:"password" binding:"required"`
		SaveAsRoot bool   `json:"saveAsRoot"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数",
		})
		return
	}

	// 验证密码
	if request.Password != config.Cfg.Web.Password {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "密码错误",
		})
		return
	}

	// 如果需要保存为总目录
	if request.SaveAsRoot {
		config.Cfg.Web.Path = request.Path
		// 保存配置到文件
		err := saveConfigToFile("config.yaml", config.Cfg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "保存配置失败: " + err.Error(),
			})
			return
		}
	}

	err := files.TraversePathAndReadFiles(request.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "目录索引失败: " + err.Error(),
		})
		return
	}

	message := "目录索引成功"
	if request.SaveAsRoot {
		message += "，并已保存为总目录"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"path":    request.Path,
	})
}

// saveConfigToFile 保存配置到文件
func saveConfigToFile(filePath string, cfg *config.Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// StatusHandler 处理状态请求
func StatusHandler(c *gin.Context) {
	// 直接返回状态，不依赖数据库
	status := gin.H{
		"status":            "ok",
		"message":           "服务运行正常",
		"documents":         0,
		"indexed_files":     0,
		"indexed_sentences": 0,
		"last_updated":      time.Now(),
	}

	c.JSON(http.StatusOK, status)
}

// UploadLogoHandler 处理logo上传请求
func UploadLogoHandler(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("logo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "获取上传文件失败: " + err.Error(),
		})
		return
	}

	// 验证文件类型
	ext := file.Filename[len(file.Filename)-4:]
	if ext != ".jpg" && ext != ".png" && ext != ".gif" && ext != "jpeg" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "只支持 JPG、PNG、GIF 格式的图片",
		})
		return
	}

	// 保存文件到static/images目录
	dst := "./web/static/images/logo.png"
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "保存文件失败: " + err.Error(),
		})
		return
	}

	// 重定向回设置页面
	c.Redirect(http.StatusFound, "/settings?logoMessage=Logo上传成功&lang="+c.PostForm("lang"))
}

// GetSuggestionsHandler 处理搜索建议请求
func GetSuggestionsHandler(c *gin.Context) {
	// 获取查询参数
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusOK, gin.H{
			"suggestions": []string{},
		})
		return
	}

	// 生成搜索建议（这里使用模拟数据，实际应用中可以从数据库或搜索历史中获取）
	suggestions := generateSuggestions(query)

	c.JSON(http.StatusOK, gin.H{
		"suggestions": suggestions,
	})
}

// generateSuggestions 生成搜索建议
func generateSuggestions(query string) []string {
	// 初始化返回的建议列表
	iter1 := db.Tables["senc"].Search(&map[string]any{
		"content": query,
	})
	defer iter1.Release()
	rlen := 11
	pk := db.Tables["senc"].GetPrimaryKey()
	//Prefix := pk.Prefix(0)
	pfxlen := len(pk.Prefix(0))
	pfxlen += 1 //去除分隔符
	fields := db.Tables["senc"].GetAllFields()
	fieldTypeLen := pk.GetfieldTypeLen(&fields)
	tylens := 0
	for _, fieldLen := range *fieldTypeLen {
		tylens += int(fieldLen) + 1
	}
	tylens -= 1 //去除最后一个分隔符
	relist := []string{}
	remap := map[string]bool{}
	loop := 0
	for iter1.Next() {
		key := iter1.Key()
		key = key[pfxlen : len(key)-tylens-pfxlen+3]
		if remap[string(key)] {
			continue
		}
		remap[string(key)] = true
		relist = append(relist, string(key))
		loop++
		if loop > rlen {
			break
		}
	}
	return relist
}

// containsIgnoreCase 忽略大小写检查字符串是否包含子串
func containsIgnoreCase(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalsIgnoreCase(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

// equalsIgnoreCase 忽略大小写比较两个字符串是否相等
func equalsIgnoreCase(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if toLower(a[i]) != toLower(b[i]) {
			return false
		}
	}
	return true
}

// toLower 将字符转换为小写
func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}
	return c
}

// GetContentByDidAndSecNoHandler 通过did和secNo获取content
func GetContentByDidAndSecNoHandler(c *gin.Context) {
	// 获取查询参数
	didStr := c.Query("did")
	secNoStr := c.Query("secNo")

	if didStr == "" || secNoStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "缺少必要的参数did和secNo",
		})
		return
	}

	// 转换参数类型
	did, err := strconv.Atoi(didStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "did参数格式无效",
		})
		return
	}

	secNo, err := strconv.Atoi(secNoStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "secNo参数格式无效",
		})
		return
	}

	// 从数据库获取内容
	content, err := getContentByDidAndSecNo(did, secNo)
	if err != nil {
		// 检查错误类型，如果是"未找到对应的数据"，返回404状态码
		if err.Error() == "未找到对应的数据" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		} else {
			// 其他错误返回500状态码
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "获取内容失败: " + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"did":     did,
		"secNo":   secNo,
		"content": content,
	})
}

// getContentByDidAndSecNo 从数据库获取内容
func getContentByDidAndSecNo(did, secNo int) (string, error) {
	// 获取senc表
	sencTable, exists := db.Tables["senc"]
	if !exists {
		return "", fmt.Errorf("senc表不存在")
	}

	// 创建查询条件
	query := map[string]any{
		"did":   did,
		"secNo": secNo,
	}

	// 查询数据
	iter := sencTable.Search(&query, util.Equal)
	defer iter.Release()

	// 获取记录
	records := iter.GetRecords(true, 1)
	if len(records) == 0 {
		return "", fmt.Errorf("未找到对应的数据")
	}

	// 获取第一条记录
	record := records[0]

	// 从记录中获取content字段
	content, ok := record["content"].(string)
	if !ok {
		return "", fmt.Errorf("content字段类型错误")
	}

	return content, nil
}

// DirectoryHandler 处理目录浏览请求
func DirectoryHandler(c *gin.Context) {
	path := c.Query("path")

	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "目录路径不能为空",
		})
		return
	}

	// 读取目录内容
	items, err := readDirectory(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "读取目录失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
	})
}

// ArticleContentHandler 处理文章内容请求
func ArticleContentHandler(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "文章ID不能为空",
		})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "文章ID格式错误",
		})
		return
	}

	// 从数据库获取文章标题和URL
	iterdir := db.Tables["dir"].Search(&map[string]any{
		"id": id,
	}, util.Equal)
	defer iterdir.Release()
	rddir := iterdir.GetRecords(true).Select("name", "url")
	if len(rddir) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "未找到文章信息",
		})
		return
	}

	title := rddir[0]["name"].(string)
	url := rddir[0]["url"].(string)

	// 从数据库获取文章内容
	itersenc := db.Tables["senc"].Search(&map[string]any{
		"did":   id,
		"secNo": nil,
	})
	defer itersenc.Release()
	rdsenc := itersenc.GetRecords(true).Select("content")

	var text strings.Builder
	for _, record := range rdsenc {
		content := record["content"].(string)
		//将回车符转换为<br />
		text.WriteString(strings.ReplaceAll(content, "\n", "<br />"))
	}

	c.JSON(http.StatusOK, gin.H{
		"title":   title,
		"url":     url,
		"content": text.String(),
	})
}

// readDirectory 读取目录内容
func readDirectory(path string) ([]map[string]string, error) {
	var items []map[string]string

	// 读取目录内容
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		item := map[string]string{
			"name": file.Name(),
			"path": filepath.Join(path, file.Name()),
		}

		if file.IsDir() {
			item["type"] = "directory"
		} else {
			item["type"] = "file"
		}

		items = append(items, item)
	}

	return items, nil
}
