@echo off
chcp 65001 >nul
echo 开始编译各平台可执行文件...
echo.

REM 创建输出目录
if not exist "build" mkdir "build"

REM Windows amd64
echo 编译 Windows amd64 版本...
go env -w GOOS=windows GOARCH=amd64
go build -o build/ReSearch-windows-amd64.exe main.go
echo Windows amd64 版本编译完成！
echo.

REM Windows 386
echo 编译 Windows 386 版本...
go env -w GOOS=windows GOARCH=386
go build -o build/ReSearch-windows-386.exe main.go
echo Windows 386 版本编译完成！
echo.

REM Linux amd64
echo 编译 Linux amd64 版本...
go env -w GOOS=linux GOARCH=amd64
go build -o build/ReSearch-linux-amd64 main.go
echo Linux amd64 版本编译完成！
echo.

REM Linux arm64
echo 编译 Linux arm64 版本...
go env -w GOOS=linux GOARCH=arm64
go build -o build/ReSearch-linux-arm64 main.go
echo Linux arm64 版本编译完成！
echo.

REM Linux 386
echo 编译 Linux 386 版本...
go env -w GOOS=linux GOARCH=386
go build -o build/ReSearch-linux-386 main.go
echo Linux 386 版本编译完成！
echo.

REM macOS amd64
echo 编译 macOS amd64 版本...
go env -w GOOS=darwin GOARCH=amd64
go build -o build/ReSearch-darwin-amd64 main.go
echo macOS amd64 版本编译完成！
echo.

REM macOS arm64
echo 编译 macOS arm64 版本...
go env -w GOOS=darwin GOARCH=arm64
go build -o build/ReSearch-darwin-arm64 main.go
echo macOS arm64 版本编译完成！
echo.

REM 恢复默认环境
go env -w GOOS=windows GOARCH=amd64
echo.
echo 所有平台编译完成！输出文件位于 build 目录。
echo 正在显示编译结果...
echo.
dir build
echo.
echo 编译脚本执行完毕！