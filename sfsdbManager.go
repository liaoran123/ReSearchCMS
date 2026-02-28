package main

import (
	"ReSearch/db"
	"fmt"

	"github.com/liaoran123/sfsDb/management"
	"github.com/liaoran123/sfsDb/web"
	"github.com/spf13/cobra"
)

// StartWebServer 启动Web服务器
func StartWebServer() {
	/*
		// 创建管理器
		manager := management.NewManager(db.Store)

		// 核心配置（只需这两行）
		configMgr := manager.ConfigManager()
		configMgr.SetConfig("web_enable", "true")
		configMgr.SetConfig("web_port", ":8083")

		// 启动Web服务器
		server := web.NewServer(":8083", manager)
		server.Start()
	*/
	// 初始化管理模块
	manager := management.NewManager(db.Store)

	// 创建web服务器
	server := web.NewServer(":8080", manager)

	// 使用固定的API密钥（管理员权限）
	fixedAPIKey := "test_api_key_for_development"
	server.AuthManager().AddAPIKey(fixedAPIKey, web.RoleAdmin)
	err := server.Start()
	if err != nil {
		fmt.Printf("启动服务器失败: %v\n", err)
		return
	}
}

var rootCmd = &cobra.Command{
	Use:   "research",
	Short: "ReSearch application",
	Run: func(cmd *cobra.Command, args []string) {
		// 应用程序主逻辑
	},
}
