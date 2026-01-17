# 考据级文档搜索引擎 - macOS 安装说明

## 安装步骤

1. **解压安装包**
   - 将安装包解压到您选择的目录，例如：`/Applications/ReSearch`
   - 使用命令：`tar -xzf ReSearch-macos.tar.gz -C /Applications/ReSearch`

2. **运行启动脚本**
   - 打开终端（Terminal）应用
   - 进入安装目录：`cd /Applications/ReSearch`
   - 给启动脚本添加执行权限：`chmod +x start.sh`
   - 运行启动脚本：`./start.sh`

3. **访问服务**
   - 打开浏览器，访问：`http://localhost:9981`
   - 服务启动成功后会在终端中显示访问地址

## 目录结构

```
ReSearch/
├── ReSearch-macos-amd64        # 主程序可执行文件
├── start.sh                    # 启动脚本
├── config.yaml                  # 配置文件
├── README.md                    # 项目说明文档
├── INSTALL.md                   # 安装说明
├── static/                      # 静态资源（CSS、JavaScript、图片）
└── templates/                   # 模板文件
└── rsdb/                        # 数据目录（自动创建）
```

## 配置说明

编辑 `config.yaml` 文件可以修改服务配置：

```yaml
web:
    port: 9981                   # 服务端口
    password: "123456"           # 索引密码
    lang: zh                     # 默认语言
    filepath: ~/Documents        # 默认索引路径

db:
    filepath: ./rsdb             # 数据库路径
```

## 启动参数

您也可以直接运行可执行文件，不使用启动脚本：

```bash
./ReSearch-macos-amd64
```

## 停止服务

- 在终端窗口中按 `Ctrl+C` 停止服务
- 或使用 `kill` 命令：`kill <PID>`

## 后台运行

使用 `nohup` 命令可以让服务在后台运行：

```bash
nohup ./ReSearch-macos-amd64 > research.log 2>&1 &
```

## 故障排除

1. **端口被占用**
   - 编辑 `config.yaml` 文件，修改 `port` 为其他可用端口
   - 使用 `lsof -i :9981` 查看占用端口的进程

2. **服务无法启动**
   - 检查配置文件是否正确
   - 确保当前目录有写入权限
   - 查看终端输出的错误信息
   - 检查依赖是否完整

3. **索引创建失败**
   - 检查索引路径是否存在
   - 确保对索引路径有读写权限
   - 检查目录权限：`chmod -R 755 /Applications/ReSearch`

## 常见问题

**Q: 如何修改默认语言？**
A: 编辑 `config.yaml` 文件，修改 `web.lang` 字段，可选值：zh（简体中文）、zh_tw（繁体中文）、en（英语）、ja（日语）、ko（韩语）、th（泰语）

**Q: 如何修改索引密码？**
A: 编辑 `config.yaml` 文件，修改 `web.password` 字段

**Q: 如何修改默认索引路径？**
A: 编辑 `config.yaml` 文件，修改 `web.filepath` 字段

**Q: 如何设置开机自启动？**
A: 使用 Launchpad 创建一个启动项，或者使用 LaunchAgents 配置

## LaunchAgents 配置示例

创建启动配置文件 `~/Library/LaunchAgents/com.research.documentsearch.plist`：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.research.documentsearch</string>
    <key>ProgramArguments</key>
    <array>
        <string>/Applications/ReSearch/ReSearch-macos-amd64</string>
    </array>
    <key>WorkingDirectory</key>
    <string>/Applications/ReSearch</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/Applications/ReSearch/research.log</string>
    <key>StandardErrorPath</key>
    <string>/Applications/ReSearch/research.log</string>
</dict>
</plist>
```

启用服务：
```bash
launchctl load ~/Library/LaunchAgents/com.research.documentsearch.plist
```

停止服务：
```bash
launchctl unload ~/Library/LaunchAgents/com.research.documentsearch.plist
```

## 联系方式

如有问题或建议，请通过以下方式联系：
- GitHub Issues：[https://github.com/liaoran123/ReSearch/issues](https://github.com/liaoran123/ReSearch/issues)
- 项目主页：[https://github.com/liaoran123/ReSearch](https://github.com/liaoran123/ReSearch)
