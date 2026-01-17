@echo off
chcp 65001 >nul
echo 考据级文档搜索引擎
setlocal enabledelayedexpansion

REM 创建数据目录
if not exist "rsdb" mkdir "rsdb"

REM 启动应用
if exist "ReSearch-windows-amd64.exe" (
    echo 正在启动服务...
    echo 访问地址: http://localhost:9981
    echo 按 Ctrl+C 停止服务
    "ReSearch-windows-amd64.exe"
) else (
    echo 错误：未找到可执行文件 ReSearch-windows-amd64.exe
    pause
)
endlocal
