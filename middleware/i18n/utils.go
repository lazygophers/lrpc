package i18n

import "fmt"

func GetAcceptLanguage(language ...string) string {
	lang := DefaultI18n.DefaultLang()
	if len(language) > 0 {
		lang = language[0]
	}

	// 根据语言代码生成 Accept-Language 的值
	return fmt.Sprintf("%s;q=/1.0", lang)
}
