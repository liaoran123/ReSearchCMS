# ReSearch - 专业级句子精确搜索引擎

一个专注于句子级精确匹配的专业搜索引擎，专为需要精确定位特定表述的场景设计。

## 🌟 核心优势

### 句子级精确匹配
区别于传统搜索引擎的**文章级**搜索粒度，ReSearch实现**句子级**精确匹配，每个搜索结果都精确指向匹配内容所在的句子。

### 100%精确搜索
- 支持全文精确匹配，无分词遗漏问题
- 搜索词中的每个字符（包括空格、标点符号）都会被精确匹配
- 支持长句、特殊符号、古籍标点等复杂搜索场景

### 专业级搜索体验
- 支持多关键词组合搜索（最多3个关键词）
- 支持目录级搜索范围限制
- 搜索结果实时高亮显示
- 支持上下文浏览，便于理解匹配内容

## 🎯 目标用户

- **学术研究者**：精确查找文献中的特定句子、引用或证据
- **程序员**：在代码库中精确查找代码片段、API调用或错误信息
- **法律工作者**：精确查找法律条文、判例或合同条款
- **古籍研究者**：精确查找古籍中的特定表述、成语或典故
- **技术文档管理人员**：在大量技术文档中快速定位特定内容

## ✨ 功能特性

- 🔍 **句子级搜索**：精确匹配并返回句子级结果
- 📚 **多关键词搜索**：支持最多3个关键词组合搜索
- 📁 **目录级搜索限制**：可限制搜索范围在特定目录内
- 🌍 **多语言支持**：支持中英文切换
- 📱 **响应式设计**：适配各种设备
- ⚡ **高性能**：基于Go语言开发，性能优异
- 📊 **状态监控**：实时显示服务状态
- 📖 **文章浏览**：支持文章内容查看和上下文导航
- 📄 **文档索引**：支持批量文档索引创建

## 🛠️ 技术栈

- **后端框架**：Gin (Go)
- **数据库**：[sfsDb](https://github.com/liaoran123/sfsDb)（自研文档数据库）
- **前端框架**：Bootstrap 5.3
- **模板引擎**：Go html/template
- **多语言支持**：i18n
- **搜索算法**：自研句子级精确匹配算法
- **静态资源**：Bootstrap Icons

## 📁 项目结构

```
ReSearch/
├── api/              # API处理函数
├── config/           # 配置管理
├── db/               # 数据库操作
├── files/            # 文件处理与索引
├── i18n/             # 多语言支持
├── pool/             # 连接池管理
├── rsdb/             # 搜索数据库
├── services/         # 业务服务（搜索核心逻辑）
├── web/              # Web界面
│   ├── static/       # 静态资源
│   └── templates/    # 模板文件
├── build-final.bat   # 构建脚本
├── config.yaml       # 配置文件
├── go.mod            # Go模块定义
└── main.go           # 主入口文件
```

## 🚀 快速开始

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
  port: 9981  # 服务端口
  path: ""     # 默认索引路径
  password: "123456"  # 索引密码
```

## 📖 使用指南

### 1. 创建索引

- 进入索引页面（`/index`）
- 选择索引类型：目录索引或单个文件索引
- 输入要索引的文件路径或选择目录
- 输入索引密码（默认为：123456）
- 点击"创建索引"按钮开始索引过程

### 2. 搜索文档

- 在搜索框中输入关键词
- 支持多关键词搜索（用空格分隔）
- 支持使用`<`前缀搜索目录名和文件名
- 点击搜索结果查看详情
- 搜索结果精确指向匹配的句子

### 3. 高级搜索技巧

- **精确匹配**：直接输入完整句子或短语
- **多关键词**：用空格分隔多个关键词，如：`金刚经 般若 波罗蜜`
- **目录搜索**：在目录页面使用搜索框，自动限制搜索范围
- **前缀搜索**：使用`<`前缀搜索目录和文件名，如：`<心经` 或 `心经>`

### 4. 浏览目录

- 进入目录页面（`/directory`）
- 点击目录链接导航
- 支持层级目录浏览
- 可在当前目录内直接搜索

### 5. 查看文章

- 点击搜索结果中的"打开快照"按钮
- 在新窗口中查看文章内容
- 使用上下箭头按钮浏览更多内容
- 匹配内容会被高亮显示

## 📊 功能详细说明

### 搜索功能

- **句子级索引**：每句话作为独立索引单位
- **精确匹配算法**：完全匹配搜索词，无分词问题
- **搜索结果排序**：按匹配度和相关性排序
- **分页显示**：默认每页显示21条结果
- **关键词高亮**：搜索结果中的关键词会被高亮显示

### 索引功能

- **支持的文件类型**：
  - 文本文件：.txt, .md, .go, .py, .java等
  - 文档文件：.html, .htm, .pdf
  - Office文件：.docx, .xlsx, .pptx
  - 图片文件：.jpg, .jpeg, .png, .gif, .bmp, .tiff, .tif（支持OCR文本提取）
  - 配置文件：.json, .yaml, .yml, .toml, .ini等
  - 代码文件：支持多种编程语言
- **索引速度**：取决于文件大小和数量，支持多线程处理
- **索引更新**：支持重新索引，覆盖原有数据

### 多语言支持

- 支持中英文界面切换
- 支持多种语言的文档索引
- 支持国际化搜索

## 🔧 构建和部署

### 构建方法

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

# macOS amd64
go env -w GOOS=darwin GOARCH=amd64
go build -o build/ReSearch-macos-amd64 main.go
```

### 构建结果

构建完成后，可执行文件将生成在 `build/` 目录中。

### 预构建可执行文件下载

您也可以直接下载预构建的可执行文件：

- **下载地址**：[https://share.weiyun.com/VYvvYTds](https://share.weiyun.com/VYvvYTds)
- **包含版本**：Windows、Linux、macOS 各平台的可执行文件
- **更新频率**：与最新版本保持同步

### 部署方式

#### Windows 部署

1. 下载或编译Windows版本的可执行文件
2. 运行 `ReSearch-windows-amd64.exe`
3. 访问 `http://localhost:9981`

#### Linux 部署

1. 下载或编译Linux版本的可执行文件
2. 赋予执行权限：`chmod +x ReSearch-linux-amd64`
3. 运行：`./ReSearch-linux-amd64`
4. 访问 `http://服务器IP:9981`

#### macOS 部署

1. 下载或编译macOS版本的可执行文件
2. 赋予执行权限：`chmod +x ReSearch-macos-amd64`
3. 运行：`./ReSearch-macos-amd64`
4. 访问 `http://localhost:9981`

## 📱 API 接口

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

## 🐛 故障排除

### 服务无法启动
- 检查端口是否被占用
- 检查配置文件是否正确
- 查看控制台日志输出

### 搜索结果为空
- 检查是否创建了索引
- 检查索引是否包含相关文档
- 尝试调整搜索关键词
- 检查搜索词是否包含特殊字符

### 索引创建失败
- 检查目录权限
- 检查目录是否存在
- 检查文档格式是否支持
- 检查索引密码是否正确

### 搜索速度慢
- 检查索引文件是否过大
- 尝试减少搜索关键词数量
- 考虑优化硬件配置

## 📈 性能指标

- **索引速度**：取决于文件大小和数量，支持多线程处理
- **搜索响应时间**：毫秒级响应
- **并发处理能力**：支持高并发请求
- **内存占用**：轻量级设计，内存占用低

## � 基准测试

### 概述

ReSearch提供了详细的基准测试用例，用于评估搜索功能在不同场景下的性能表现。基准测试覆盖了单查询搜索、带目录限制的搜索、不同分页参数的搜索、并发搜索以及不同查询词长度的搜索性能。

### 基准测试文件

- **测试代码**：[services/search_test.go](services/search_test.go)
- **测试报告**：[services/SEARCH_BENCHMARK_REPORT.md](services/SEARCH_BENCHMARK_REPORT.md)

### 运行基准测试

```bash
# 运行所有基准测试
go test ./services -bench=BenchmarkSearch -benchmem -v

# 运行单个基准测试
go test ./services -bench=BenchmarkSearch$ -benchmem -v

# 生成CPU分析文件
go test ./services -bench=BenchmarkSearch -cpuprofile=cpu.prof

# 生成内存分析文件
go test ./services -bench=BenchmarkSearch -memprofile=mem.prof
```

### 测试场景

1. **单查询搜索性能**：测试不同查询词的搜索响应时间
2. **带目录限制的搜索性能**：测试不同目录ID下的搜索性能
3. **不同分页参数的搜索性能**：测试不同start和limit参数的搜索性能
4. **并发搜索性能**：测试1、2、4、8、16个并发请求下的搜索性能
5. **不同查询词长度的搜索性能**：测试1-8个字符长度的查询词搜索性能

### 访问规模分析

基于基准测试结果，我们生成了详细的项目访问规模分析报告，评估系统在不同并发情况下的性能表现：

- **分析报告**：[PERFORMANCE_SCALE_REPORT.md](PERFORMANCE_SCALE_REPORT.md)
- **报告内容**：包括系统能力评估、访问规模估算、性能瓶颈分析和优化建议
- **适用场景**：为项目部署、扩容和优化提供参考依据

## 🔮 未来规划

### 近期规划
- [ ] 支持PDF文档解析
- [ ] 优化句子分割算法
- [ ] 增强搜索结果排序
- [ ] 支持更多语言的文档索引

### 中期规划
- [ ] 集成AI语义搜索
- [ ] 支持图片OCR识别
- [ ] 实现智能推荐功能
- [ ] 支持云存储接入

### 长期规划
- [ ] 支持视频和音频内容搜索
- [ ] 实现自然语言问答功能
- [ ] 支持分布式部署
- [ ] 提供浏览器插件

## 📄 许可证

MIT License

## 🤝 贡献

欢迎提交Issue和Pull Request！

## 📞 联系方式

如有问题或建议，请通过以下方式联系：

- GitHub Issues：[https://github.com/liaoran123/ReSearch/issues](https://github.com/liaoran123/ReSearch/issues)
- 项目主页：[https://github.com/liaoran123/ReSearch](https://github.com/liaoran123/ReSearch)

## 🙏 致谢

感谢所有为项目做出贡献的开发者！

---

**ReSearch** - 让搜索更精确，让知识更易获取！

如果您觉得ReSearch对您有帮助，请给我们一个 ⭐ Star！