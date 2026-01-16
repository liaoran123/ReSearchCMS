package i18n

import (
	"encoding/json/v2"
	"os"
	"path/filepath"
)

// Translator 翻译器结构体
type Translator struct {
	locales map[string]map[string]string
}

// NewTranslator 创建一个新的翻译器实例
func NewTranslator() *Translator {
	return &Translator{
		locales: make(map[string]map[string]string),
	}
}

// LoadLocales 加载指定目录下的所有语言文件
func (t *Translator) LoadLocales(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			lang := filepath.Base(file.Name())[:len(filepath.Base(file.Name()))-5]
			data, err := os.ReadFile(filepath.Join(dir, file.Name()))
			if err != nil {
				return err
			}

			var locale map[string]string
			if err := json.Unmarshal(data, &locale); err != nil {
				return err
			}

			t.locales[lang] = locale
		}
	}

	return nil
}

// Translate 翻译指定的键到指定的语言
func (t *Translator) Translate(key, lang string) string {
	if locale, ok := t.locales[lang]; ok {
		if value, ok := locale[key]; ok {
			return value
		}
	}
	// 如果指定语言不存在或键不存在，返回默认语言（中文）
	if locale, ok := t.locales["zh"]; ok {
		if value, ok := locale[key]; ok {
			return value
		}
	}
	// 如果默认语言也不存在，返回键本身
	return key
}

// GetSupportedLanguages 获取支持的语言列表
func (t *Translator) GetSupportedLanguages() []string {
	languages := make([]string, 0, len(t.locales))
	for lang := range t.locales {
		languages = append(languages, lang)
	}
	return languages
}
