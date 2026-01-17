# 考据级文档搜索引擎 - Windows 安装说明

## 安装步骤

1. **解压安装包**
   - 将安装包解压到您选择的目录，例如：`D:\ReSearch`

2. **运行启动脚本**
   - 双击运行 `start.bat` 脚本
   - 或在命令行中执行：`start.bat`

3. **访问服务**
   - 打开浏览器，访问：`http://localhost:9981`
   - 服务启动成功后会在命令行中显示访问地址

## 目录结构

```
ReSearch/
├── ReSearch-windows-amd64.exe  # 主程序可执行文件
├── start.bat                    # 启动脚本
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
    filepath: D:\Documents       # 默认索引路径

db:
    filepath: ./rsdb             # 数据库路径
```

## 启动参数

您也可以直接运行可执行文件，不使用启动脚本：

```bash
ReSearch-windows-amd64.exe
```

## 停止服务

- 在命令行窗口中按 `Ctrl+C` 停止服务

## 故障排除

1. **端口被占用**
   - 编辑 `config.yaml` 文件，修改 `port` 为其他可用端口

2. **服务无法启动**
   - 检查配置文件是否正确
   - 确保当前目录有写入权限
   - 查看命令行输出的错误信息

3. **索引创建失败**
   - 检查索引路径是否存在
   - 确保对索引路径有读写权限

## 常见问题

**Q: 如何修改默认语言？**
A: 编辑 `config.yaml` 文件，修改 `web.lang` 字段，可选值：zh（简体中文）、zh_tw（繁体中文）、en（英语）、ja（日语）、ko（韩语）、th（泰语）

**Q: 如何修改索引密码？**
A: 编辑 `config.yaml` 文件，修改 `web.password` 字段

**Q: 如何修改默认索引路径？**
A: 编辑 `config.yaml` 文件，修改 `web.filepath` 字段

## 联系方式

如有问题或建议，请通过以下方式联系：
- GitHub Issues：[https://github.com/liaoran123/ReSearch/issues](https://github.com/liaoran123/ReSearch/issues)
- 项目主页：[https://github.com/liaoran123/ReSearch](https://github.com/liaoran123/ReSearch)
