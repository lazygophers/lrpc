package i18n

import (
	"fmt"
	"github.com/lazygophers/log"
	"golang.org/x/text/language"
	"strings"
)

type LanguageCode struct {
	Lang string
	Tag  language.Tag
}

func (l *LanguageCode) String() string {
	if l.Tag.String() == l.Lang {
		return l.Lang
	}

	return fmt.Sprintf("%s(%s)", l.Tag.String(), l.Lang)
}

func (l *LanguageCode) GoString() string {
	return l.String()
}

func (l *LanguageCode) ʔ() string {
	return l.String()
}

func ParseLangCode(lang string) (*LanguageCode, error) {
	lang = strings.ToLower(lang)

	switch lang {
	case "zh-cn":
		return &LanguageCode{
			Lang: lang,
			Tag:  language.Chinese,
		}, nil

	case "zh-hans", "zh-chs":
		return &LanguageCode{
			Lang: lang,
			Tag:  language.SimplifiedChinese,
		}, nil

	case "zh-hk", "zh-tw", "zh-mo", "zh-sg", "zh-cht":
		return &LanguageCode{
			Lang: lang,
			Tag:  language.TraditionalChinese,
		}, nil

	default:
		l, err := language.Parse(lang)
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}

		return &LanguageCode{
			Lang: lang,
			Tag:  l,
		}, nil
	}
}

func MustParseLangCode(lang string) *LanguageCode {
	l, err := ParseLangCode(lang)
	if err != nil {
		panic(err)
	}

	return l
}
