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
