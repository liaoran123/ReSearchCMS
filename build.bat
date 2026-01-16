@echo off
setlocal enabledelayedexpansion

REM 设置项目名称
set PROJECT_NAME=ReSearch

REM 创建输出目录
if not exist "build" mkdir "build"

echo 开始编译各平台可执行文件...

REM Windows amd64
echo 编译 Windows amd64 版本...
set GOOS=windows
set GOARCH=amd64
go build -o build/%PROJECT_NAME%-windows-amd64.exe main.go

REM Windows 386
echo 编译 Windows 386 版本...
set GOOS=windows
set GOARCH=386
go build -o build/%PROJECT_NAME%-windows-386.exe main.go

REM Linux amd64
echo 编译 Linux amd64 版本...
set GOOS=linux
set GOARCH=amd64
go build -o build/%PROJECT_NAME%-linux-amd64 main.go

REM Linux arm64
echo 编译 Linux arm64 版本...
set GOOS=linux
set GOARCH=arm64
go build -o build/%PROJECT_NAME%-linux-arm64 main.go

REM Linux 386
echo 编译 Linux 386 版本...
set GOOS=linux
set GOARCH=386
go build -o build/%PROJECT_NAME%-linux-386 main.go

REM macOS amd64
echo 编译 macOS amd64 版本...
set GOOS=darwin
set GOARCH=amd64
go build -o build/%PROJECT_NAME%-darwin-amd64 main.go

REM macOS arm64 (Apple Silicon)
echo 编译 macOS arm64 版本...
set GOOS=darwin
set GOARCH=arm64
go build -o build/%PROJECT_NAME%-darwin-arm64 main.go

echo 编译完成！输出文件位于 build 目录。
dir build/

pause