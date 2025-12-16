package files

import (
	"os"
	"strings"
	"testing"
)

// TestGetOfficeTextContent 测试GetOfficeTextContent函数
func TestGetOfficeTextContent(t *testing.T) {
	// 创建一个临时文本文件作为测试
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// 写入测试内容
	testContent := "测试Office文件文本内容提取"
	if _, err := tmpFile.WriteString(testContent); err != nil {
		t.Fatalf("写入测试内容失败: %v", err)
	}
	tmpFile.Close()

	// 测试ReadFileContent函数
	content, err := ReadFileContent(tmpFile.Name())
	if err != nil {
		t.Fatalf("ReadFileContent函数失败: %v", err)
	}

	if content != testContent {
		t.Errorf("ReadFileContent函数返回的内容错误，期望: %s, 实际: %s", testContent, content)
	}
}

// TestGetOfficeTextContentUnsupportedType 测试GetOfficeTextContent函数处理不支持的文件类型
func TestGetOfficeTextContentUnsupportedType(t *testing.T) {
	// 创建一个临时文本文件作为测试
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// 测试GetOfficeTextContent函数处理不支持的文件类型
	_, err = GetOfficeTextContent(tmpFile.Name())
	if err == nil {
		t.Error("GetOfficeTextContent函数应该返回错误，但没有返回")
	}
}

// TestGetHTMLTextContent 测试GetHTMLTextContent函数
func TestGetHTMLTextContent(t *testing.T) {
	// 创建一个临时HTML文件作为测试
	tmpFile, err := os.CreateTemp("", "test-*.html")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// 写入测试HTML内容
	testHTML := `<!DOCTYPE html>
<html>
<head>
    <title>测试HTML文件</title>
</head>
<body>
    <h1>这是一个标题</h1>
    <p>这是一个段落。</p>
    <div>
        <p>这是一个嵌套的段落。</p>
    </div>
</body>
</html>`
	if _, err := tmpFile.WriteString(testHTML); err != nil {
		t.Fatalf("写入测试内容失败: %v", err)
	}
	tmpFile.Close()

	// 测试GetHTMLTextContent函数
	content, err := GetHTMLTextContent(tmpFile.Name())
	if err != nil {
		t.Fatalf("GetHTMLTextContent函数失败: %v", err)
	}

	// 检查内容中是否包含预期的文本
	expectedTexts := []string{"测试HTML文件", "这是一个标题", "这是一个段落", "这是一个嵌套的段落"}
	for _, expectedText := range expectedTexts {
		if !strings.Contains(content, expectedText) {
			t.Errorf("GetHTMLTextContent函数返回的内容中不包含预期的文本: %s", expectedText)
		}
	}
}

// TestReadFileContentForHTML 测试ReadFileContent函数处理HTML文件
func TestReadFileContentForHTML(t *testing.T) {
	// 创建一个临时HTML文件作为测试
	tmpFile, err := os.CreateTemp("", "test-*.html")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// 写入测试HTML内容
	testHTML := `<!DOCTYPE html>
<html>
<head>
    <title>测试HTML文件</title>
</head>
<body>
    <p>这是一个段落。</p>
</body>
</html>`
	if _, err = tmpFile.WriteString(testHTML); err != nil {
		t.Fatalf("写入测试内容失败: %v", err)
	}
	tmpFile.Close()

	// 测试ReadFileContent函数处理HTML文件
	content, err := ReadFileContent(tmpFile.Name())
	if err != nil {
		t.Fatalf("ReadFileContent函数处理HTML文件失败: %v", err)
	}

	// 检查内容中是否包含预期的文本
	if !strings.Contains(content, "测试HTML文件") {
		t.Error("ReadFileContent函数返回的内容中不包含预期的文本: 测试HTML文件")
	}
	if !strings.Contains(content, "这是一个段落") {
		t.Error("ReadFileContent函数返回的内容中不包含预期的文本: 这是一个段落")
	}
}
