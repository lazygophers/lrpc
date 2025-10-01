package i18n

import (
	"strconv"

	"github.com/lazygophers/lrpc/middleware/xerror"
)

var _ xerror.I18n = (*I18nForXerror)(nil)

type I18nForXerror struct {
	i18n   *I18n
	prefix string
}

func (p *I18nForXerror) Localize(key int32, langs ...string) (string, bool) {
	var lang string
	if len(langs) > 0 {
		lang = langs[0]
	}

	return p.i18n.localize(lang, p.prefix+strconv.FormatInt(int64(key), 10))
}

// Register 追加注册指定语言的错误码翻译
// 如果语言包不存在，将自动创建
// 如果错误码已存在，将会覆盖原有值
func (p *I18nForXerror) Register(lang string, code int32, msg string) {
	p.i18n.Register(lang, p.prefix+strconv.FormatInt(int64(code), 10), msg)
}

// RegisterBatch 批量追加注册指定语言的多个错误码翻译
// 如果语言包不存在，将自动创建
// 如果错误码已存在，将会覆盖原有值
// data 的 key 为错误码，value 为翻译文本
func (p *I18nForXerror) RegisterBatch(lang string, data map[int32]string) {
	batch := make(map[string]any, len(data))
	for code, msg := range data {
		batch[p.prefix+strconv.FormatInt(int64(code), 10)] = msg
	}
	p.i18n.RegisterBatch(lang, batch)
}

func NewI18nForXerror(i *I18n, prefix ...string) *I18nForXerror {
	p := &I18nForXerror{
		i18n: i,
	}

	if len(prefix) > 0 {
		p.prefix = prefix[0]
	} else {
		p.prefix = "error."
	}

	return p
}
