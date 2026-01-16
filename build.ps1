# 设置项目名称
$PROJECT_NAME = "ReSearch"

# 创建输出目录
if (-not (Test-Path -Path "build")) {
    New-Item -ItemType Directory -Path "build" -Force
}

Write-Host "开始编译各平台可执行文件..." -ForegroundColor Green

# 定义要编译的平台列表
$platforms = @(
    @{ GOOS = "windows"; GOARCH = "amd64"; Extension = ".exe" },
    @{ GOOS = "windows"; GOARCH = "386"; Extension = ".exe" },
    @{ GOOS = "linux"; GOARCH = "amd64"; Extension = "" },
    @{ GOOS = "linux"; GOARCH = "arm64"; Extension = "" },
    @{ GOOS = "linux"; GOARCH = "386"; Extension = "" },
    @{ GOOS = "darwin"; GOARCH = "amd64"; Extension = "" },
    @{ GOOS = "darwin"; GOARCH = "arm64"; Extension = "" }
)

# 遍历平台列表并编译
foreach ($platform in $platforms) {
    $GOOS = $platform.GOOS
    $GOARCH = $platform.GOARCH
    $Extension = $platform.Extension
    
    $outputName = "$PROJECT_NAME-$GOOS-$GOARCH$Extension"
    $outputPath = Join-Path -Path "build" -ChildPath $outputName
    
    Write-Host "编译 $GOOS $GOARCH 版本..." -ForegroundColor Cyan
    
    # 设置环境变量
    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH
    
    # 执行编译
    go build -o $outputPath main.go
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ 编译成功: $outputName" -ForegroundColor Green
    } else {
        Write-Host "✗ 编译失败: $outputName" -ForegroundColor Red
    }
}

# 清除环境变量
Remove-Item Env:GOOS
Remove-Item Env:GOARCH

Write-Host "
编译完成！输出文件位于 build 目录。" -ForegroundColor Green
Write-Host "
文件列表：" -ForegroundColor Yellow
Get-ChildItem -Path "build" | Select-Object Name, Length, LastWriteTime

# 可选：打包静态资源
Write-Host "
是否需要打包静态资源？(Y/N)" -ForegroundColor Yellow
$response = Read-Host
if ($response -eq "Y" -or $response -eq "y") {
    Write-Host "开始打包静态资源..." -ForegroundColor Cyan
    
    # 遍历所有编译好的可执行文件
    Get-ChildItem -Path "build" -File | Where-Object { $_.Name -notlike "*.zip" } | ForEach-Object {
        $exeName = $_.Name
        $zipName = $exeName -replace $_.Extension, ".zip"
        $zipPath = Join-Path -Path "build" -ChildPath $zipName
        
        Write-Host "打包 $exeName -> $zipName" -ForegroundColor Cyan
        
        # 创建临时目录
        $tempDir = Join-Path -Path $env:TEMP -ChildPath "ReSearch_$([guid]::NewGuid())"
        New-Item -ItemType Directory -Path $tempDir -Force
        
        # 复制可执行文件
        Copy-Item -Path $_.FullName -Destination $tempDir
        
        # 复制静态资源和配置文件
        if (Test-Path -Path "web") {
            Copy-Item -Path "web" -Destination $tempDir -Recurse -Force
        }
        if (Test-Path -Path "config.yaml") {
            Copy-Item -Path "config.yaml" -Destination $tempDir -Force
        }
        
        # 创建zip文件
        Compress-Archive -Path "$tempDir\*" -DestinationPath $zipPath -Force
        
        # 清理临时目录
        Remove-Item -Path $tempDir -Recurse -Force
        
        Write-Host "✓ 打包成功: $zipName" -ForegroundColor Green
    }
    
    Write-Host "
所有资源打包完成！" -ForegroundColor Green
}