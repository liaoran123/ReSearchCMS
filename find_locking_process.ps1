# 查找占用数据库文件的进程
$rsdbPath = "D:\MyGo\src\ReSearch\rsdb"

Write-Host "正在查找占用数据库文件的进程，请稍候..."
Write-Host "扫描目录: $rsdbPath"

# 获取目录中的所有文件
$files = Get-ChildItem -Path $rsdbPath -Recurse -File

# 遍历所有文件，尝试打开它们以检测锁定
foreach ($file in $files) {
    try {
        $stream = [System.IO.File]::Open($file.FullName, [System.IO.FileMode]::Open, [System.IO.FileAccess]::ReadWrite, [System.IO.FileShare]::None)
        $stream.Close()
        # 文件未被锁定，跳过
    } catch {
        # 文件被锁定，尝试查找占用进程
        Write-Host "文件 $($file.FullName) 被锁定"
        
        # 使用Get-Process命令查找可能的占用进程
        # 注意：这种方法可能不够准确，但可以尝试
        Write-Host "可能的占用进程："
        Get-Process | Where-Object {$_.ProcessName -like "*ReSearch*" -or $_.ProcessName -like "*go*" -or $_.ProcessName -like "*sfs*"} | Format-Table -AutoSize
        
        # 只显示第一个被锁定的文件，然后退出
        break
    }
}

Write-Host "查找完成！"
