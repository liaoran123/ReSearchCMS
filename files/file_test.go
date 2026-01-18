package files

import (
	"testing"
)

// TestIsTextFile 测试IsTextFile函数
func TestIsTextFile(t *testing.T) {
	// 测试文本文件
	textContent := []byte("This is a text file.\nIt contains multiple lines.\nAnd some special characters: !@#$%^&*()_+{}\":<>?|[]\\;',./~`\n\t\r\n")
	if !IsTextFile(textContent) {
		t.Error("Expected text file to be recognized as text")
	}

	// 测试空文件
	emptyContent := []byte("")
	if !IsTextFile(emptyContent) {
		t.Error("Expected empty file to be recognized as text")
	}

	// 测试二进制文件（包含大量不可打印字符）
	binaryContent := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F}
	if IsTextFile(binaryContent) {
		t.Error("Expected binary file to not be recognized as text")
	}

	// 测试混合文件（可打印字符比例低于80%）
	mixedContent := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 'T', 'e', 'x', 't', ' ', 'F', 'i', 'l', 'e'}
	if IsTextFile(mixedContent) {
		t.Error("Expected mixed file with low printable char ratio to not be recognized as text")
	}

	// 测试JSON文件
	jsonContent := []byte(`{
		"name": "test",
		"value": 123,
		"isValid": true
	}`)
	if !IsTextFile(jsonContent) {
		t.Error("Expected JSON file to be recognized as text")
	}

	// 测试代码文件
	codeContent := []byte(`package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`)
	if !IsTextFile(codeContent) {
		t.Error("Expected code file to be recognized as text")
	}
}
