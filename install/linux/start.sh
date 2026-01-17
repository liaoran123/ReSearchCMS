#!/bin/bash
echo "考据级文档搜索引擎"

# 创建数据目录
mkdir -p rsdb

# 设置可执行权限
chmod +x ReSearch-linux-amd64

# 启动应用
if [ -f "ReSearch-linux-amd64" ]; then
    echo "正在启动服务..."
    echo "访问地址: http://localhost:9981"
    echo "按 Ctrl+C 停止服务"
    ./ReSearch-linux-amd64
else
    echo "错误：未找到可执行文件 ReSearch-linux-amd64"
    exit 1
fi
