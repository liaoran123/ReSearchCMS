package web

import (
	"ReSearch/config"
	"ReSearch/i18n"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
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
	lang := c.DefaultQuery("lang", "zh")
	logoMessage := c.Query("logoMessage")
	c.HTML(http.StatusOK, "settings.html", gin.H{
		"title":       w.translator.Translate("settings", lang),
		"active":      "settings",
		"lang":        lang,
		"translator":  w.translator,
		"languages":   w.translator.GetSupportedLanguages(),
		"logoMessage": logoMessage,
	})
}

// SaveSettingsHandler 处理保存设置请求
func (w *WebServer) SaveSettingsHandler(c *gin.Context) {
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

	var directoryItems []gin.H
	if configPath != "" {
		// 读取目录内容
		items, err := w.readDirectory(configPath)
		if err == nil {
			directoryItems = items
		}
	}

	c.HTML(http.StatusOK, "directory.html", gin.H{
		"title":          w.translator.Translate("directory", lang),
		"active":         "directory",
		"lang":           lang,
		"configPath":     configPath,
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

	// 处理目录项
	for _, file := range files {
		itemType := "file"
		if file.IsDir() {
			itemType = "directory"
		}

		items = append(items, gin.H{
			"name": file.Name(),
			"type": itemType,
			"path": filepath.Join(path, file.Name()),
		})
	}

	return items, nil
}

// Translate 翻译函数，用于模板中
func (w *WebServer) Translate(key, lang string) string {
	return w.translator.Translate(key, lang)
}
