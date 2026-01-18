package files

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/unidoc/unioffice/document"
	"github.com/unidoc/unioffice/presentation"
	"github.com/unidoc/unioffice/spreadsheet"
)

// GetOfficeTextContent 从Office文件中提取文本内容
// 支持.docx、.xlsx、.pptx、.doc、.xls、.ppt文件
func GetOfficeTextContent(filePath string) (string, error) {
	// 获取文件扩展名并转换为小写
	ext := filepath.Ext(filePath)
	ext = strings.ToLower(ext)

	// 根据文件扩展名选择处理方式
	switch ext {
	case ".docx":
		return getDOCXTextContent(filePath)
	case ".doc":
		// 旧版DOC格式支持 - 建议使用LibreOffice转换
		return "", fmt.Errorf("旧版DOC格式(.doc)支持需要外部工具。推荐使用LibreOffice转换为DOCX格式后再处理，或考虑安装golang.org/x/sys/windows/registry包实现Windows COM自动化")
	case ".xlsx":
		return getXLSXTextContent(filePath)
	case ".xls":
		// 旧版XLS格式支持 - 可以使用excelize库
		return "", fmt.Errorf("旧版XLS格式(.xls)支持需要外部库。推荐使用github.com/360EntSecGroup-Skylar/excelize库，或使用LibreOffice转换为XLSX格式后再处理")
	case ".pptx":
		return getPPTXTextContent(filePath)
	case ".ppt":
		// 旧版PPT格式支持 - 建议使用LibreOffice转换
		return "", fmt.Errorf("旧版PPT格式(.ppt)支持需要外部工具。推荐使用LibreOffice转换为PPTX格式后再处理")
	default:
		return "", fmt.Errorf("不支持的文件类型: %s", ext)
	}
}

// getDOCXTextContent 从DOCX文件中提取文本内容
func getDOCXTextContent(filePath string) (string, error) {
	// 打开DOCX文件
	doc, err := document.Open(filePath)
	if err != nil {
		return "", err
	}
	defer doc.Close()

	var buf bytes.Buffer

	// 遍历所有段落
	for _, para := range doc.Paragraphs() {
		// 遍历段落中的所有文本运行
		for _, run := range para.Runs() {
			buf.WriteString(run.Text())
		}
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

// getXLSXTextContent 从XLSX文件中提取文本内容
func getXLSXTextContent(filePath string) (string, error) {
	// 打开XLSX文件
	xlFile, err := spreadsheet.Open(filePath)
	if err != nil {
		return "", err
	}
	defer xlFile.Close()

	var buf bytes.Buffer

	// 遍历所有工作表
	for _, sheet := range xlFile.Sheets() {
		fmt.Fprintf(&buf, "工作表: %s\n", sheet.Name())

		// 遍历所有行
		for _, row := range sheet.Rows() {
			// 遍历所有单元格
			for _, cell := range row.Cells() {
				// 获取单元格的文本内容
				text := cell.GetFormattedValue()
				fmt.Fprintf(&buf, "%s\t", text)
			}
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

// getPPTXTextContent 从PPTX文件中提取文本内容
func getPPTXTextContent(filePath string) (string, error) {
	// 打开PPTX文件
	prs, err := presentation.Open(filePath)
	if err != nil {
		return "", err
	}
	defer prs.Close()

	var buf bytes.Buffer

	// 遍历所有幻灯片
	for i, slide := range prs.Slides() {
		fmt.Fprintf(&buf, "幻灯片 %d:\n", i+1)

		// 使用ExtractText方法提取幻灯片中的所有文本
		slideText := slide.ExtractText()
		if slideText != nil {
			buf.WriteString(slideText.Text())
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

// GetHTMLTextContent 从HTML文件中提取文本内容
func GetHTMLTextContent(filePath string) (string, error) {
	// 读取HTML文件
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取HTML文件 %s 失败: %w", filePath, err)
	}

	// 使用正则表达式处理HTML内容，提取纯文本
	htmlContent := string(content)

	// 先将换行相关标签转换为换行符
	// 处理<br>、<br/>、<br />
	brRegex := regexp.MustCompile(`<\s*br\s*/?\s*>`)
	htmlContent = brRegex.ReplaceAllString(htmlContent, "\n")
	// 处理<p>、</p>
	pRegex := regexp.MustCompile(`<\s*p\s*>`)
	htmlContent = pRegex.ReplaceAllString(htmlContent, "\n")
	pCloseRegex := regexp.MustCompile(`<\s*/\s*p\s*>`)
	htmlContent = pCloseRegex.ReplaceAllString(htmlContent, "\n")
	// 处理<div>、</div>
	divRegex := regexp.MustCompile(`<\s*div\s*>`)
	htmlContent = divRegex.ReplaceAllString(htmlContent, "\n")
	divCloseRegex := regexp.MustCompile(`<\s*/\s*div\s*>`)
	htmlContent = divCloseRegex.ReplaceAllString(htmlContent, "\n")
	// 处理标题标签
	hRegex := regexp.MustCompile(`<\s*h[1-6]\s*>`)
	htmlContent = hRegex.ReplaceAllString(htmlContent, "\n")
	hCloseRegex := regexp.MustCompile(`<\s*/\s*h[1-6]\s*>`)
	htmlContent = hCloseRegex.ReplaceAllString(htmlContent, "\n")

	// 移除HTML标签，使用预编译的正则表达式
	textContent := tagRegex.ReplaceAllString(htmlContent, "")

	// 先处理多余的换行符，保留单个换行符
	newlineRegex := regexp.MustCompile(`\n+`)
	textContent = newlineRegex.ReplaceAllString(textContent, "\n")

	// 移除行首和行尾的空格，然后处理行内多余空格
	lines := strings.Split(textContent, "\n")
	var result strings.Builder
	for _, line := range lines {
		// 移除行首尾空格
		line = strings.TrimSpace(line)
		if line != "" {
			// 处理行内多余空格
			line = whitespaceRegex.ReplaceAllString(line, " ")
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return result.String(), nil
}

// IsTextFile 检查文件内容是否为文本文件
// 简单检查：如果前1000个字符中可打印字符比例超过80%，则视为文本文件
func IsTextFile(content []byte) bool {
	// 空文件视为文本文件
	if len(content) == 0 {
		return true
	}

	// 检查前1000个字符或整个文件（取较小者）
	checkLen := min(len(content), 1000)

	// 统计可打印字符数量
	printableChars := 0
	for i := range checkLen {
		c := content[i]
		// 可打印字符：32-126 或 换行符(\n)、回车符(\r)、制表符(\t)
		if (c >= 32 && c <= 126) || c == 10 || c == 13 || c == 9 {
			printableChars++
		}
	}

	// 如果可打印字符比例超过80%，视为文本文件
	return float64(printableChars)/float64(checkLen) >= 0.8
}

// 预编译正则表达式，避免重复编译开销
var (
	tagRegex        = regexp.MustCompile(`<[^>]+>`)
	whitespaceRegex = regexp.MustCompile(`\s+`)
)

// 预定义文件类型映射，提高查找效率
var (
	officeFileMap = map[string]bool{
		".docx": true,
		".xlsx": true,
		".pptx": true,
		".doc":  true,
		".xls":  true,
		".ppt":  true,
	}
	htmlFileMap = map[string]bool{
		".html": true,
		".htm":  true,
	}
	textFileMap = map[string]bool{
		// 普通文本文件
		".txt": true,
		// 配置文件
		".json": true, ".yaml": true, ".yml": true, ".toml": true, ".ini": true, ".conf": true, ".cfg": true, ".xml": true,
		// 代码文件
		".go": true, ".py": true, ".java": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true, ".css": true, ".scss": true, ".md": true,
		".rb": true, ".php": true, ".cpp": true, ".c": true, ".h": true, ".hpp": true, ".cs": true, ".swift": true, ".kt": true, ".rs": true,
		// 日志和脚本文件
		".log": true, ".sql": true, ".sh": true, ".bash": true, ".cmd": true, ".bat": true, ".ps1": true,
		// 其他常见文本文件类型
		".rst": true, ".text": true, ".markdown": true, ".adoc": true, ".asciidoc": true,
		".tex": true, ".latex": true, ".bib": true,
		".pl": true, ".pm": true, ".t": true, // Perl
		".lua": true, ".groovy": true, ".scala": true, ".erl": true, ".hrl": true, // 其他脚本和编程语言
		".diff": true, ".patch": true, ".csv": true, ".tsv": true, ".rtf": true,
		".properties": true, ".resx": true, ".strings": true, ".po": true, ".mo": true, // 资源文件
		".htaccess": true, ".htpasswd": true, ".env": true, ".dockerfile": true, ".docker-compose.yml": true,
		".gitignore": true, ".gitattributes": true, ".gitmodules": true,
		".vue": true, ".svelte": true, ".angular.json": true, ".vue.config.js": true, // 前端框架配置文件
		".gradle": true, ".gradlew": true, ".mvnw": true, ".pom.xml": true, // Java构建工具
		".csproj": true, ".vbproj": true, ".fsproj": true, // .NET项目文件
		".sln": true, ".suo": true, ".user": true, // .NET解决方案文件
		".xcodeproj": true, ".xcworkspace": true, ".swiftpm": true, // iOS/macOS开发文件
		".gradle.kts": true, ".build.gradle": true, ".settings.gradle": true, // Gradle配置文件
		".bazel": true, ".bazelrc": true, ".bzl": true, // Bazel构建工具
		".meson.build": true, ".meson_options.txt": true, // Meson构建工具
		".cmake": true, ".CMakeLists.txt": true, // CMake构建工具
		".ninja": true, ".build.ninja": true, // Ninja构建系统
		".pkg": true, ".pkg-info": true, ".egg-info": true, ".whl": true, ".tar.gz": true, ".zip": true, // 包管理文件
	}
)

// ReadFileContent 读取文件内容，根据文件类型自动选择处理方式
func ReadFileContent(filePath string) (string, error) {
	// 获取文件扩展名并转换为小写
	ext := filepath.Ext(filePath)
	ext = strings.ToLower(ext)

	// 检查是否为Office文件
	if officeFileMap[ext] {
		return GetOfficeTextContent(filePath)
	}

	// 检查是否为HTML文件
	if htmlFileMap[ext] {
		return GetHTMLTextContent(filePath)
	}

	// 检查是否为文本、配置、代码或日志文件
	if textFileMap[ext] {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("读取文件 %s 失败: %w", filePath, err)
		}
		return string(content), nil
	}
	//不支持没有定义的类型
	return "", fmt.Errorf("不支持的文件类型: %s", ext)
}
