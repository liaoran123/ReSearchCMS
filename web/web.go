package web

import (
	"ReSearch/config"
	"ReSearch/db"
	"ReSearch/i18n"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/liaoran123/sfsDb/util"
	"gopkg.in/yaml.v3"
)

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

// WebServer web服务器结构体
type WebServer struct {
	translator *i18n.Translator
}

// NewWebServer 创建一个新的web服务器实例
func NewWebServer(translator *i18n.Translator) *WebServer {
	return &WebServer{
		translator: translator,
	}
}

// SetupRoutes 设置web路由
func (w *WebServer) SetupRoutes(r *gin.Engine) {
	// 设置静态文件目录
	r.Static("/static", "./web/static")

	// 加载主模板文件
	r.LoadHTMLFiles(
		"./web/templates/index.html",
		"./web/templates/search.html",
		"./web/templates/index-page.html",
		"./web/templates/status.html",
		"./web/templates/settings.html",
		"./web/templates/article.html",
		"./web/templates/help.html",
		"./web/templates/pricing.html",
		"./web/templates/contact.html",
		"./web/templates/directory.html",
		"./web/templates/partials/searchinput.html",
		"./web/templates/partials/navbar.html",
		"./web/templates/partials/footer.html",
		"./web/templates/partials/static.html",
	)

	// 主路由
	r.GET("/", w.HomeHandler)
	r.GET("/search", w.SearchPageHandler)
	r.GET("/index", w.IndexPageHandler)
	r.GET("/status", w.StatusPageHandler)
	r.GET("/settings", w.SettingsPageHandler)
	r.POST("/settings", w.SaveSettingsHandler)
	r.GET("/pricing", w.PricingPageHandler)
	r.GET("/contact", w.ContactPageHandler)
	r.GET("/help", w.HelpPageHandler)
	r.GET("/directory", w.DirectoryPageHandler)
	r.GET("/article/:id", w.ArticlePageHandler)
}

// HomeHandler 处理首页请求
func (w *WebServer) HomeHandler(c *gin.Context) {
	lang := c.DefaultQuery("lang", "zh")
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":      w.translator.Translate("home", lang),
		"active":     "home",
		"lang":       lang,
		"translator": w.translator,
		"languages":  w.translator.GetSupportedLanguages(),
	})
}

// SearchPageHandler 处理搜索页面请求
func (w *WebServer) SearchPageHandler(c *gin.Context) {
	lang := c.DefaultQuery("lang", "zh")
	query := c.Query("q")
	c.HTML(http.StatusOK, "search.html", gin.H{
		"title":      w.translator.Translate("search", lang),
		"active":     "search",
		"lang":       lang,
		"query":      query,
		"translator": w.translator,
		"languages":  w.translator.GetSupportedLanguages(),
	})
}

// IndexPageHandler 处理索引页面请求
func (w *WebServer) IndexPageHandler(c *gin.Context) {
	// 只允许本机访问索引页面
	clientIP := c.ClientIP()
	if clientIP != "127.0.0.1" && clientIP != "::1" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied. Only local access allowed.",
		})
		return
	}

	lang := c.DefaultQuery("lang", "zh")
	c.HTML(http.StatusOK, "index-page.html", gin.H{
		"title":      w.translator.Translate("index", lang),
		"active":     "index",
		"lang":       lang,
		"translator": w.translator,
		"languages":  w.translator.GetSupportedLanguages(),
		"configPath": config.Cfg.Web.Path,
	})
}

// StatusPageHandler 处理状态页面请求
func (w *WebServer) StatusPageHandler(c *gin.Context) {
	lang := c.DefaultQuery("lang", "zh")
	c.HTML(http.StatusOK, "status.html", gin.H{
		"title":      w.translator.Translate("status", lang),
		"active":     "status",
		"lang":       lang,
		"translator": w.translator,
		"languages":  w.translator.GetSupportedLanguages(),
	})
}

// SettingsPageHandler 处理设置页面请求
func (w *WebServer) SettingsPageHandler(c *gin.Context) {
	// 只允许本机访问设置页面
	clientIP := c.ClientIP()
	if clientIP != "127.0.0.1" && clientIP != "::1" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied. Only local access allowed.",
		})
		return
	}

	lang := c.DefaultQuery("lang", "zh")
	logoMessage := c.Query("logoMessage")
	v := c.Query("v")
	c.HTML(http.StatusOK, "settings.html", gin.H{
		"title":       w.translator.Translate("settings", lang),
		"active":      "settings",
		"lang":        lang,
		"translator":  w.translator,
		"languages":   w.translator.GetSupportedLanguages(),
		"logoMessage": logoMessage,
		"v":           v,
	})
}

// SaveSettingsHandler 处理保存设置请求
func (w *WebServer) SaveSettingsHandler(c *gin.Context) {
	// 只允许本机访问保存设置功能
	clientIP := c.ClientIP()
	if clientIP != "127.0.0.1" && clientIP != "::1" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied. Only local access allowed.",
		})
		return
	}

	lang := c.DefaultQuery("lang", "zh")
	defaultLanguage := c.PostForm("defaultLanguage")

	if defaultLanguage != "" {
		// 更新配置
		config.Cfg.Web.Lang = defaultLanguage

		// 保存配置到文件
		if err := saveConfigToFile("config.yaml", config.Cfg); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "保存配置失败: " + err.Error(),
			})
			return
		}
		// 使用新选择的语言作为当前语言
		lang = defaultLanguage
	}

	c.HTML(http.StatusOK, "settings.html", gin.H{
		"title":      w.translator.Translate("settings", lang),
		"active":     "settings",
		"lang":       lang,
		"translator": w.translator,
		"languages":  w.translator.GetSupportedLanguages(),
		"message":    "设置保存成功",
	})
}

// HelpPageHandler 处理帮助页面请求
func (w *WebServer) HelpPageHandler(c *gin.Context) {
	lang := c.DefaultQuery("lang", "zh")
	c.HTML(http.StatusOK, "help.html", gin.H{
		"title":      "使用帮助",
		"active":     "help",
		"lang":       lang,
		"translator": w.translator,
		"languages":  w.translator.GetSupportedLanguages(),
	})
}

// PricingPageHandler 处理价格页面请求
func (w *WebServer) PricingPageHandler(c *gin.Context) {
	lang := c.DefaultQuery("lang", "zh")
	c.HTML(http.StatusOK, "pricing.html", gin.H{
		"title":      "价格方案",
		"active":     "pricing",
		"lang":       lang,
		"translator": w.translator,
		"languages":  w.translator.GetSupportedLanguages(),
	})
}

// ContactPageHandler 处理联系我们页面请求
func (w *WebServer) ContactPageHandler(c *gin.Context) {
	lang := c.DefaultQuery("lang", "zh")
	c.HTML(http.StatusOK, "contact.html", gin.H{
		"title":      "联系我们",
		"active":     "contact",
		"lang":       lang,
		"translator": w.translator,
		"languages":  w.translator.GetSupportedLanguages(),
	})
}

// ArticlePageHandler 处理文章阅读页面请求
func (w *WebServer) ArticlePageHandler(c *gin.Context) {
	lang := c.DefaultQuery("lang", "zh")
	articleId := c.Param("id")
	c.HTML(http.StatusOK, "article.html", gin.H{
		"title":      w.translator.Translate("article", lang),
		"active":     "",
		"lang":       lang,
		"articleId":  articleId,
		"translator": w.translator,
		"languages":  w.translator.GetSupportedLanguages(),
	})
}

// DirectoryPageHandler 处理目录浏览页面请求
func (w *WebServer) DirectoryPageHandler(c *gin.Context) {
	lang := c.DefaultQuery("lang", "zh")
	configPath := config.Cfg.Web.Path

	// 获取当前目录路径参数
	currentPath := c.Query("path")
	if currentPath == "" {
		// 如果没有提供路径，使用配置路径作为默认路径
		currentPath = configPath
	}

	var directoryItems []gin.H
	if currentPath != "" {
		// 直接调用本地readDirectory函数，返回固定的articleId=1，确保超链接正常显示
		items, err := w.readDirectory(currentPath)
		if err == nil {
			directoryItems = items
			// 调试输出：目录内容数量
			fmt.Printf("Directory items count: %d\n", len(directoryItems))
		} else {
			// 调试输出：错误信息
			fmt.Printf("Error reading directory %s: %v\n", currentPath, err)
		}
	}

	// 查询当前目录的ID
	var currentDirId int

	// 调试输出：当前路径
	fmt.Printf("[DEBUG] Raw current path: %s\n", currentPath)

	// 处理路径格式，确保与数据库中的格式匹配
	processedPath := currentPath
	// 确保路径使用正确的分隔符（根据数据库存储格式调整）
	processedPath = strings.ReplaceAll(processedPath, "\\", "/")

	fmt.Printf("[DEBUG] Processed path: %s\n", processedPath)

	// 1. 先尝试使用完整路径匹配当前目录
	iterDirCurrent := db.Tables["dir"].Search(&map[string]any{
		"url": processedPath,
	}, util.Equal)
	rdDirCurrent := iterDirCurrent.GetRecords(true).Select("id", "url")
	iterDirCurrent.Release()

	fmt.Printf("[DEBUG] Full path search found %d records\n", len(rdDirCurrent))
	if len(rdDirCurrent) > 0 {
		currentDirId = rdDirCurrent[0]["id"].(int)
		fmt.Printf("[DEBUG] Found ID: %d for path: %s\n", currentDirId, rdDirCurrent[0]["url"])
	} else {
		// 2. 尝试使用原始路径格式（反斜杠）匹配
		iterDirCurrentRaw := db.Tables["dir"].Search(&map[string]any{
			"url": currentPath,
		}, util.Equal)
		rdDirCurrentRaw := iterDirCurrentRaw.GetRecords(true).Select("id", "url")
		iterDirCurrentRaw.Release()

		fmt.Printf("[DEBUG] Raw path search found %d records\n", len(rdDirCurrentRaw))
		if len(rdDirCurrentRaw) > 0 {
			currentDirId = rdDirCurrentRaw[0]["id"].(int)
			fmt.Printf("[DEBUG] Found ID: %d for raw path: %s\n", currentDirId, rdDirCurrentRaw[0]["url"])
		} else {
			// 3. 如果完整路径匹配失败，尝试使用文件名匹配
			currentDirName := filepath.Base(currentPath)
			fmt.Printf("[DEBUG] Searching by directory name: %s\n", currentDirName)

			iterDirCurrentName := db.Tables["dir"].Search(&map[string]any{
				"name": currentDirName,
			}, util.Equal)
			rdDirCurrentName := iterDirCurrentName.GetRecords(true).Select("id", "name", "url")
			iterDirCurrentName.Release()

			fmt.Printf("[DEBUG] Name search found %d records\n", len(rdDirCurrentName))
			for _, rec := range rdDirCurrentName {
				fmt.Printf("[DEBUG] Name match: ID=%d, Name=%s, URL=%s\n", rec["id"], rec["name"], rec["url"])
			}

			if len(rdDirCurrentName) > 0 {
				currentDirId = rdDirCurrentName[0]["id"].(int)
			}
		}
	}

	// 调试输出：配置路径和当前路径
	fmt.Printf("Config path: %s\n", configPath)
	fmt.Printf("Current path: %s\n", currentPath)
	fmt.Printf("Current directory ID: %d\n", currentDirId)

	c.HTML(http.StatusOK, "directory.html", gin.H{
		"title":          w.translator.Translate("directory", lang),
		"active":         "directory",
		"lang":           lang,
		"configPath":     configPath,
		"currentPath":    currentPath,
		"currentDirId":   currentDirId,
		"directoryItems": directoryItems,
		"translator":     w.translator,
		"languages":      w.translator.GetSupportedLanguages(),
	})
}

// readDirectory 读取目录内容
func (w *WebServer) readDirectory(path string) ([]gin.H, error) {
	var items []gin.H

	// 读取目录
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		itemType := "file"
		if file.IsDir() {
			itemType = "directory"
		}

		itemPath := filepath.Join(path, file.Name())
		var itemId int

		// 为所有项（目录和文件）查询对应的id
		// 1. 先尝试使用完整路径匹配
		iterdirPath := db.Tables["dir"].Search(&map[string]any{
			"url": itemPath,
		}, util.Equal)
		rddirPath := iterdirPath.GetRecords(true).Select("id")
		iterdirPath.Release()
		if len(rddirPath) > 0 {
			itemId = rddirPath[0]["id"].(int)
		} else {
			// 2. 如果完整路径匹配失败，尝试使用文件名匹配
			iterdirName := db.Tables["dir"].Search(&map[string]any{
				"name": file.Name(),
			}, util.Equal)
			rddirName := iterdirName.GetRecords(true).Select("id")
			iterdirName.Release()
			if len(rddirName) > 0 {
				itemId = rddirName[0]["id"].(int)
			}
		}

		items = append(items, gin.H{
			"name":      file.Name(),
			"type":      itemType,
			"path":      itemPath,
			"articleId": itemId,
			"id":        itemId,
		})
	}

	return items, nil
}

// Translate 翻译函数，用于模板中
func (w *WebServer) Translate(key, lang string) string {
	return w.translator.Translate(key, lang)
}
