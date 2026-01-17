@echo off
chcp 65001 >nul
echo 开始编译可执行文件...
echo.

REM 创建输出目录
if not exist "build" mkdir "build"

REM Windows amd64 (已测试成功)
echo 编译 Windows amd64 版本...
go env -w GOOS=windows GOARCH=amd64
go build -o build/ReSearch-windows-amd64.exe main.go
echo ✓ Windows amd64 版本编译完成！
echo.

REM Linux amd64 (已测试成功)
echo 编译 Linux amd64 版本...
go env -w GOOS=linux GOARCH=amd64
go build -o build/ReSearch-linux-amd64 main.go
echo ✓ Linux amd64 版本编译完成！
echo.

REM macOS amd64 (支持)
echo 编译 macOS amd64 版本...
go env -w GOOS=darwin GOARCH=amd64
go build -o build/ReSearch-macos-amd64 main.go
echo ✓ macOS amd64 版本编译完成！
echo.

REM 恢复默认环境
go env -w GOOS=windows GOARCH=amd64
echo.
echo 编译完成！输出文件位于 build 目录。
echo 成功构建的平台：
echo - Windows amd64
echo - Linux amd64
echo - macOS amd64
echo.
dir build
echo.
echo 编译脚本执行完毕！