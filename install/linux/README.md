# 考据级文档搜索引擎

一个功能强大的文档搜索引擎，专注于提供高效、准确的文档搜索体验。

## 功能特性

- 🔍 **强大的搜索功能**：支持关键词搜索、多关键词搜索
- 📁 **目录导航**：支持目录结构展示和导航
- 📄 **文档索引**：支持批量文档索引创建
- 🌍 **多语言支持**：支持中英文切换
- 📱 **响应式设计**：适配各种设备
- ⚡ **高性能**：基于Go语言开发，性能优异
- 📊 **状态监控**：实时显示服务状态
- 📖 **文章浏览**：支持文章内容查看和导航

## 技术栈

- **后端框架**：Gin (Go)
- **数据库**：sfsDb
- **前端框架**：Bootstrap 5.3
- **模板引擎**：Go html/template
- **多语言支持**：i18n
- **搜索功能**：自定义搜索算法

## 项目结构

```
ReSearch/
├── api/              # API处理函数
├── config/           # 配置管理
├── db/               # 数据库操作
├── files/            # 文件处理
├── i18n/             # 多语言支持
├── pool/             # 连接池
├── rsdb/             # 搜索数据库
├── search/           # 搜索功能
├── services/         # 业务服务
├── test/             # 测试文件
├── util/             # 工具函数
├── web/              # Web界面
│   ├── static/       # 静态资源
│   └── templates/    # 模板文件
├── build-final.bat   # 构建脚本
├── config.yaml       # 配置文件
├── go.mod            # Go模块定义
├── go.sum            # Go模块依赖
└── main.go           # 主入口文件
```

## 安装和使用

### 前提条件

- Go 1.25.3 或更高版本
- 操作系统：Windows/Linux/macOS

### 安装步骤

1. **克隆项目**
   ```bash
   git clone https://github.com/liaoran123/ReSearch.git
   cd ReSearch
   ```

2. **安装依赖**
   ```bash
   go mod tidy
   ```

3. **运行项目**
   ```bash
   go run main.go
   ```

4. **访问服务**
   浏览器会自动打开：`http://localhost:9981`

### 配置文件

编辑 `config.yaml` 文件可以配置服务端口、数据库等参数：

```yaml
web:
  port: 9981
```

## 构建方法

### 使用构建脚本

项目提供了构建脚本，可以编译不同平台的可执行文件：

```bash
# Windows
build-final.bat

# 手动构建
# Windows amd64
go build -o build/ReSearch-windows-amd64.exe main.go

# Linux amd64
go env -w GOOS=linux GOARCH=amd64
go build -o build/ReSearch-linux-amd64 main.go
```

### 构建结果

构建完成后，可执行文件将生成在 `build/` 目录中。

## API 接口

### 搜索接口
```
GET /api/search?q=关键词&start=0&limit=20&lang=zh
```

### 索引接口
```
POST /api/index
POST /api/index/path
```

### 状态接口
```
GET /api/status
```

### 文章内容接口
```
GET /api/article/:id
```

### 目录接口
```
GET /api/directory
```

## 使用指南

### 1. 创建索引

- 进入索引页面
- 选择要索引的目录
- 点击"创建索引"按钮
- 等待索引创建完成

### 2. 搜索文档

- 在搜索框中输入关键词
- 支持多关键词搜索（用空格分隔）
- 支持使用`<`前缀搜索目录名和文件名
- 点击搜索结果查看详情

### 3. 浏览目录

- 进入目录页面
- 点击目录链接导航
- 支持层级目录浏览

### 4. 查看文章

- 点击搜索结果中的"打开快照"按钮
- 在新窗口中查看文章内容
- 使用上下箭头按钮浏览更多内容

## 功能说明

### 搜索功能

- 支持关键词高亮显示
- 支持多关键词搜索
- 支持目录和文件名搜索
- 支持分页显示结果

### 索引功能

- 支持批量文档索引
- 支持索引状态监控
- 支持索引路径保存

### 多语言支持

- 支持中英文切换
- 支持界面元素翻译
- 支持搜索结果翻译

## 开发说明

### 开发环境搭建

1. 安装Go 1.25.3
2. 克隆项目
3. 安装依赖：`go mod tidy`
4. 运行开发服务器：`go run main.go`

### 测试

```bash
# 运行测试
cd files
go test -v
```

## 部署说明

### Windows 部署

1. 下载或编译Windows版本的可执行文件
2. 运行 `ReSearch-windows-amd64.exe`
3. 访问 `http://localhost:9981`

### Linux 部署

1. 下载或编译Linux版本的可执行文件
2. 赋予执行权限：`chmod +x ReSearch-linux-amd64`
3. 运行：`./ReSearch-linux-amd64`
4. 访问 `http://服务器IP:9981`

## 注意事项

- 首次运行需要创建索引
- 索引创建时间取决于文档数量
- 建议定期更新索引以保持搜索结果的准确性
- 大文件可能需要较长时间索引

## 故障排除

### 服务无法启动

- 检查端口是否被占用
- 检查配置文件是否正确
- 查看日志输出

### 搜索结果为空

- 检查是否创建了索引
- 检查索引是否包含相关文档
- 尝试调整搜索关键词

### 索引创建失败

- 检查目录权限
- 检查目录是否存在
- 检查文档格式是否支持

## 更新日志

### v1.0.0

- 初始版本发布
- 支持基本搜索功能
- 支持文档索引
- 支持目录导航
- 支持多语言

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！

## 联系方式

如有问题或建议，请通过以下方式联系：

- GitHub Issues：[https://github.com/liaoran123/ReSearch/issues](https://github.com/liaoran123/ReSearch/issues)
- 项目主页：[https://github.com/liaoran123/ReSearch](https://github.com/liaoran123/ReSearch)

## 致谢

感谢所有为项目做出贡献的开发者！

---

**考据级文档搜索引擎** - 让文档搜索更简单、更高效！