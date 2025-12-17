package i18n

import (
	"encoding/json"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"

	"github.com/gin-gonic/gin"
)

var bundle *i18n.Bundle
var matcher language.Matcher

func Setup() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// 支持的语言顺序：优先中文，再英文
	supported := []language.Tag{language.Chinese, language.English}
	matcher = language.NewMatcher(supported)

	// 自动遍历 asset/locales 目录并加载所有 .json 翻译文件
	localesRoot := "asset/locales"
	err := filepath.WalkDir(localesRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("walk locales error: %v", err)
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".json") {
			if _, err := bundle.LoadMessageFile(path); err != nil {
				log.Printf("Failed to load message file %s: %v", path, err)
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("Failed to scan locales directory %s: %v", localesRoot, err)
	}
}

// GetLang 根据 gin.Context 从 Accept-Language header 返回简短语言码 ("zh" 或 "en")，缺省为 "en"
func GetLang(c *gin.Context) string {

	lang := c.GetHeader("Accept-Language")
	if lang == "" {
		return "en"
	}
	return lang
	// header := c.GetHeader("Accept-Language")
	// if header == "" {
	// 	return "en"
	// }
	// tags, _, err := language.ParseAcceptLanguage(header)
	// if err != nil || len(tags) == 0 {
	// 	return "en"
	// }
	// tag, _, _ := matcher.Match(tags...)
	// base, _ := tag.Base()
	// if base.String() == "zh" {
	// 	return "zh"
	// }
	// return "en"
}

// Tr 翻译 helper（保留现有接口）
func Tr(lang, key string) string {
	localizer := i18n.NewLocalizer(bundle, lang)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: key,
	})
	if err != nil {
		return key // 如果未找到翻译，返回 key
	}
	return msg
}
