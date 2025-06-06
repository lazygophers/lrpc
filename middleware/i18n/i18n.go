package i18n

import (
	"fmt"
	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/anyx"
	"github.com/lazygophers/utils/candy"
	"github.com/lazygophers/utils/routine"
	"github.com/lazygophers/utils/stringx"
	"github.com/petermattis/goid"
	"go.uber.org/atomic"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io/fs"
	"maps"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type LocalizeFs interface {
	ReadFile(name string) ([]byte, error)
	ReadDir(name string) ([]fs.DirEntry, error)
}

// Pack 语言包
type Pack struct {
	lang string
	code *LanguageCode

	corpus map[string]string
}

func (p *Pack) parse(prefixs []string, m map[string]any) {
	for k, v := range m {
		keys := make([]string, len(prefixs)+1)
		copy(keys, prefixs)
		keys[len(keys)-1] = k

		switch x := v.(type) {
		case string:
			key := strings.Join(keys, ".")
			if _, ok := p.corpus[key]; !ok {
				p.corpus[key] = x
			} else {
				log.Panicf("%s duplicate", key)
				panic(fmt.Sprintf("%s: duplicate", key))
			}

		case int:
			key := strings.Join(keys, ".")
			if _, ok := p.corpus[key]; !ok {
				p.corpus[key] = strconv.Itoa(x)
			} else {
				log.Panicf("%s duplicate", key)
				panic(fmt.Sprintf("%s: duplicate", key))
			}

		case int64:
			key := strings.Join(keys, ".")
			if _, ok := p.corpus[key]; !ok {
				p.corpus[key] = strconv.FormatInt(x, 10)
			} else {
				log.Panicf("%s duplicate", key)
				panic(fmt.Sprintf("%s: duplicate", key))
			}

		case float64:
			key := strings.Join(keys, ".")
			if _, ok := p.corpus[key]; !ok {
				p.corpus[key] = strconv.FormatFloat(x, 'f', -1, 64)
			} else {
				log.Panicf("%s duplicate", key)
				panic(fmt.Sprintf("%s: duplicate", key))
			}

		case map[string]interface{}:
			p.parse(keys, x)

		case map[int64]interface{}:
			mm := make(map[string]interface{}, len(x))
			for k, v := range x {
				mm[strconv.FormatInt(k, 10)] = v
			}

			p.parse(keys, mm)

		case map[float64]interface{}:
			mm := make(map[string]interface{}, len(x))
			for k, v := range x {
				mm[strconv.FormatFloat(k, 'f', -1, 64)] = v
			}

			p.parse(keys, mm)

		case map[any]any:
			mm := make(map[string]interface{}, len(x))
			for k, v := range x {
				mm[anyx.ToString(k)] = v
			}

			p.parse(keys, mm)

		default:
			log.Panicf("unsupported type %T", x)
		}
	}
}

func NewPack(lang string) *Pack {
	return &Pack{
		lang:   lang,
		code:   MustParseLangCode(lang),
		corpus: map[string]string{},
	}
}

type I18n struct {
	packMap map[string]*Pack

	templateFunc template.FuncMap
	defaultLang  *atomic.String
}

func (p *I18n) AddTemplateFunc(key string, a any) {
	p.templateFunc[key] = a
}

func (p *I18n) localize(lang string, key string) (string, bool) {
	if lang != "" {
		lang = strings.ToLower(lang)
		if pack, ok := p.packMap[lang]; ok {
			return pack.corpus[key], true
		}

		if strings.Contains(lang, "-") {
			if pack, ok := p.packMap[lang[:strings.Index(lang, "-")]]; ok {
				return pack.corpus[key], true
			}
		}
	}

	defaultLang := p.defaultLang.Load()

	if pack, ok := p.packMap[defaultLang]; ok {
		return pack.corpus[key], true
	}

	if strings.Contains(defaultLang, "-") {
		if pack, ok := p.packMap[defaultLang[:strings.Index(defaultLang, "-")]]; ok {
			return pack.corpus[key], true
		}
	}

	return key, false
}

func (p *I18n) LocalizeWithLang(lang string, key string, args ...interface{}) string {
	value, _ := p.localize(lang, key)
	if len(args) == 0 {
		return value
	}

	b := log.GetBuffer()
	defer log.PutBuffer(b)

	err := template.Must(template.New("").Parse(value)).Execute(b, args[0])
	if err != nil {
		log.Panicf("err:%v", err)
		return value
	}

	return b.String()
}

func (p *I18n) Localize(key string, args ...interface{}) string {
	return p.LocalizeWithLang(GetLanguage(), key, args...)
}

func (p *I18n) LoadLocalizesWithFs(dirPath string, embedFs LocalizeFs) error {
	dirs, err := embedFs.ReadDir(dirPath)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	for _, dir := range dirs {
		log.Debugf("try loading localize %s", dir.Name())

		loclizer, ok := GetLocalizer(filepath.Ext(dir.Name()))
		if !ok {
			log.Warnf("unsupported ext %s", filepath.Ext(dir.Name()))
			continue
		}

		lang := strings.TrimSuffix(dir.Name(), filepath.Ext(dir.Name()))
		lang = strings.ToLower(lang)

		pack := NewPack(lang)

		buf, err := embedFs.ReadFile(filepath.ToSlash(filepath.Join(dirPath, dir.Name())))
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		var m map[string]any
		err = loclizer.Unmarshal(buf, &m)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		pack.parse(nil, m)

		p.packMap[lang] = pack
	}

	return nil
}

func (p *I18n) LoadLocalizes(embedFs LocalizeFs) error {
	return p.LoadLocalizesWithFs("localize", embedFs)
}

func (p *I18n) DefaultLang() string {
	return p.defaultLang.Load()
}

func (p *I18n) SetDefaultLang(lang string) *I18n {
	lang = strings.ToLower(lang)
	log.Infof("set default language: %s", lang)
	p.defaultLang.Store(lang)
	return p
}

func (p *I18n) AllSupportedLanguageCode() []*LanguageCode {
	langs := make([]*LanguageCode, 0, len(p.packMap))
	for _, pack := range p.packMap {
		langs = append(langs, pack.code)
	}

	return langs
}

var (
	DefaultI18n = NewI18n()

	defaultTemplateFunc = template.FuncMap{
		"ToCamel":      stringx.ToCamel,
		"ToSmallCamel": stringx.ToSmallCamel,
		"ToSnake":      stringx.ToSnake,
		"ToLower":      strings.ToLower,
		"ToUpper":      strings.ToUpper,
		"ToTitle":      strings.ToTitle,

		"TrimPrefix": strings.TrimPrefix,
		"TrimSuffix": strings.TrimSuffix,
		"HasPrefix":  strings.HasPrefix,
		"HasSuffix":  strings.HasSuffix,

		"PluckString": anyx.PluckString,
		"PluckInt":    anyx.PluckInt,
		"PluckInt32":  anyx.PluckInt32,
		"PluckUint32": anyx.PluckUint32,
		"PluckInt64":  anyx.PluckInt64,
		"PluckUint64": anyx.PluckUint64,

		"StringSliceEmpty": func(ss []string) bool {
			return len(ss) == 0
		},

		"UniqueString":   candy.Unique[string],
		"SortString":     candy.Sort[string],
		"ReverseString":  candy.Reverse[string],
		"TopString":      candy.Top[string],
		"FirstString":    candy.First[string],
		"LastString":     candy.Last[string],
		"ContainsString": candy.Contains[string],
		"TimeFormat4Pb": func(t *timestamppb.Timestamp, layout string) string {
			return t.AsTime().Format(layout)
		},
		"TimeFormat4Timestamp": func(t int64, layout string) string {
			return time.Unix(t, 0).Format(layout)
		},
	}
	cache = routine.NewCache[int64, string]()
)

func SetLanguage(language string, gid ...int64) {
	language = ParseLanguage(language)

	if language == "" {
		DelLanguage(gid...)
	} else {
		if len(gid) > 0 {
			cache.Set(gid[0], language)
		} else {
			cache.Set(goid.Get(), language)
		}
	}
}

func DelLanguage(gid ...int64) {
	if len(gid) > 0 {
		cache.Delete(gid[0])
	} else {
		cache.Delete(goid.Get())
	}
}

func GetLanguage(gid ...int64) string {
	if len(gid) > 0 {
		return cache.GetWithDef(gid[0], "")
	} else {
		return cache.GetWithDef(goid.Get(), "")
	}
}

func NewI18n() *I18n {
	p := &I18n{
		packMap:     map[string]*Pack{},
		defaultLang: atomic.NewString(language.English.String()),

		templateFunc: maps.Clone(defaultTemplateFunc),
	}

	return p
}

func SetDefaultLanguage(lang string) {
	DefaultI18n.SetDefaultLang(lang)
}

func DefaultLanguage() string {
	return DefaultI18n.DefaultLang()
}

func localize(lang string, key string, args ...interface{}) string {
	return DefaultI18n.LocalizeWithLang(lang, key, args...)
}

func ParseLanguage(lang string) string {
	if lang == "" {
		return lang
	}

	if strings.Contains(lang, ",") {
		lang = lang[:strings.Index(lang, ",")]
	}

	if strings.Contains(lang, ";") {
		lang = lang[:strings.Index(lang, ";")]
	}

	if strings.Contains(lang, ".") {
		lang = lang[:strings.Index(lang, ".")]
	}

	if strings.Count(lang, "_") == 1 {
		lang = strings.ReplaceAll(lang, "_", "-")
	}

	lang = strings.ToLower(lang)

	return lang
}

func LocalizeWithHeaders(headers http.Header, key string, args ...interface{}) string {
	if len(headers) == 0 {
		return localize("", key, args...)
	}

	if value, ok := headers["Accept-Language"]; ok {
		return localize(ParseLanguage(strings.Join(value, "; ")), key, args...)
	}

	return localize("", key, args...)
}

func Localize(key string, args ...interface{}) string {
	return localize("", key, args...)
}

func LoadLocalizes(embedFs LocalizeFs) error {
	return DefaultI18n.LoadLocalizes(embedFs)
}
