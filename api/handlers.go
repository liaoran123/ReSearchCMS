package api

import (
	"ReSearch/config"
	"ReSearch/db"
	"ReSearch/files"
	"ReSearch/services"
	"crypto/rand"
	"encoding/base64"
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

	// 从查询参数获取did，默认值为0
	didStr := c.DefaultQuery("did", "0")
	did, err := strconv.Atoi(didStr)
	if err != nil {
		did = 0
	}

	results, totalCount, err := services.Search(query, did, start, limit)
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
	// 调用文件索引函数
	fileTypes, unsupportedFiles, err := files.TraversePathAndReadFiles(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "索引失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "文件索引成功",
		"path":             filePath,
		"fileTypes":        fileTypes,
		"unsupportedFiles": unsupportedFiles,
	})
}

// IndexPathHandler 处理目录索引请求
func IndexPathHandler(c *gin.Context) {
	var request struct {
		Path       string `json:"path" binding:"required"`
		Password   string `json:"password" binding:"required"`
		SaveAsRoot bool   `json:"saveAsRoot"`
		Lang       string `json:"lang"`
	}

	// 尝试解析请求体
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数: " + err.Error(),
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

	// 同步执行索引
	fileTypes, unsupportedFiles, err := files.TraversePathAndReadFiles(request.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "索引失败: " + err.Error(),
		})
		return
	}

	// 索引成功后返回响应
	message := "目录索引成功"
	if request.SaveAsRoot {
		message += "，并已保存为总目录"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          message,
		"path":             request.Path,
		"fileTypes":        fileTypes,
		"unsupportedFiles": unsupportedFiles,
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

// SaveRootPathHandler 处理保存根路径请求
func SaveRootPathHandler(c *gin.Context) {
	var request struct {
		Path     string `json:"path" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	// 尝试解析请求体
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 验证密码
	if request.Password != config.Cfg.Web.Password {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "密码错误",
		})
		return
	}

	// 保存路径到配置
	config.Cfg.Web.Path = request.Path
	// 保存配置到文件
	err := saveConfigToFile("config.yaml", config.Cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "保存配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "路径已成功保存为总目录",
		"path":    request.Path,
	})
}

// UploadLogoHandler 处理logo上传请求
func UploadLogoHandler(c *gin.Context) {
	// 只允许本机访问上传logo功能
	clientIP := c.ClientIP()
	if clientIP != "127.0.0.1" && clientIP != "::1" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied. Only local access allowed.",
		})
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("logo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "获取上传文件失败: " + err.Error(),
		})
		return
	}

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "只支持 JPG、PNG、GIF 格式的图片",
		})
		return
	}

	// 使用绝对路径保存logo，确保在Windows上正常工作
	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取当前工作目录失败: " + err.Error(),
		})
		return
	}

	// 构建logo保存目录的绝对路径
	dstDir := filepath.Join(cwd, "web", "static", "images")
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "创建保存目录失败: " + err.Error(),
				"path":  dstDir,
			})
			return
		}
	}

	// 构建logo文件的绝对路径
	dst := filepath.Join(dstDir, "logo.png")

	// 保存文件到static/images目录
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "保存文件失败: " + err.Error(),
			"path":  dst,
		})
		return
	}

	// 生成随机字符串作为版本号，防止浏览器缓存旧的logo
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		// 如果生成随机数据失败，使用当前时间戳
		version := strconv.FormatInt(time.Now().UnixNano(), 10)
		c.Redirect(http.StatusFound, "/settings?logoMessage=Logo上传成功&lang="+c.PostForm("lang")+"&v="+version)
		return
	}

	// 将随机数据转换为base64字符串，用于清除缓存
	version := base64.URLEncoding.EncodeToString(randomBytes)

	// 重定向回设置页面，添加版本参数以防止缓存
	c.Redirect(http.StatusFound, "/settings?logoMessage=Logo上传成功&lang="+c.PostForm("lang")+"&v="+version)
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
func readDirectory(path string) ([]map[string]any, error) {
	var items []map[string]any

	// 读取目录内容
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	// 查询所有目录记录，用于匹配文件路径
	iterdir := db.Tables["dir"].ForData()
	allDirs := iterdir.GetRecords(true).Select("id", "url")
	iterdir.Release()

	for _, file := range files {
		itemType := "file"
		if file.IsDir() {
			itemType = "directory"
		}

		itemPath := filepath.Join(path, file.Name())
		var articleId int

		// 如果是文件，查找匹配的文章ID
		if !file.IsDir() {
			// 遍历所有目录记录，查找匹配的文件路径
			for _, dirRecord := range allDirs {
				dbUrl := dirRecord["url"].(string)
				// 尝试不同的匹配方式
				if dbUrl == itemPath || filepath.ToSlash(dbUrl) == filepath.ToSlash(itemPath) {
					articleId = dirRecord["id"].(int)
					break
				}
			}
		}

		items = append(items, map[string]any{
			"name":      file.Name(),
			"type":      itemType,
			"path":      itemPath,
			"articleId": articleId,
		})
	}

	return items, nil
}
