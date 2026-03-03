package language

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	xlanguage "golang.org/x/text/language"
)

var parseLangCodeCache sync.Map

type LanguageCode struct {
	Lang string
	Tag  xlanguage.Tag
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

func Parse(lang string) (Language, error) {
	token := cleanLanguageToken(lang)
	if token == "" {
		return "", fmt.Errorf("language: empty")
	}

	if mapped, ok := mapLanguageForParse(token); ok {
		return mapped, nil
	}

	if !isWellFormedLanguageTag(token) {
		return "", fmt.Errorf("language: tag is not well-formed")
	}

	return Language(token), nil
}

func Base(lang Language) Language {
	parsed, err := Parse(string(lang))
	if err != nil {
		return ""
	}

	switch parsed {
	case ChineseMainland, ChineseHongKong, ChineseTaiwan, ChineseMacao, ChineseSingapore, ChineseTraditional, SimplifiedChinese, TraditionalChinese:
		return Chinese
	}

	text := string(parsed)
	if strings.Contains(text, "-") {
		text = text[:strings.Index(text, "-")]
	}

	return Language(text)
}

func MustParse(lang string) Language {
	v, err := Parse(lang)
	if err != nil {
		panic(err)
	}

	return v
}

func ParseLangCode(lang string) (*LanguageCode, error) {
	token := cleanLanguageToken(lang)
	if token == "" {
		return nil, fmt.Errorf("language: empty")
	}

	if cached, ok := parseLangCodeCache.Load(token); ok {
		return cached.(*LanguageCode), nil
	}

	if _, ok := mapLanguageForParse(token); !ok && !isWellFormedLanguageTag(token) {
		return nil, fmt.Errorf("language: tag is not well-formed")
	}

	tagText := tagTextForLanguage(token)
	result := &LanguageCode{
		Lang: token,
		Tag:  xlanguage.Make(tagText),
	}

	parseLangCodeCache.Store(token, result)
	return result, nil
}

func MustParseLangCode(lang string) *LanguageCode {
	v, err := ParseLangCode(lang)
	if err != nil {
		panic(err)
	}

	return v
}

func ParseAcceptLanguageList(value string) []Language {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	type preference struct {
		lang Language
		q    float64
		idx  int
	}

	items := strings.Split(value, ",")
	prefs := make([]preference, 0, len(items))

	for idx, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		parts := strings.Split(item, ";")
		token := strings.TrimSpace(parts[0])
		if token == "" {
			continue
		}

		token = strings.ReplaceAll(token, "_", "-")
		token = strings.ToLower(token)

		lang, ok := mapLanguageForAccept(token)
		if !ok {
			var err error
			lang, err = Parse(token)
			if err != nil {
				continue
			}
		}

		q := 1.0
		for _, p := range parts[1:] {
			p = strings.TrimSpace(strings.ToLower(p))
			if !strings.HasPrefix(p, "q=") {
				continue
			}

			val := strings.TrimSpace(strings.TrimPrefix(p, "q="))
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				q = 0
				break
			}
			if f < 0 {
				f = 0
			}
			if f > 1 {
				f = 1
			}
			q = f
			break
		}

		if q <= 0 {
			continue
		}

		prefs = append(prefs, preference{
			lang: lang,
			q:    q,
			idx:  idx,
		})
	}

	sort.SliceStable(prefs, func(i, j int) bool {
		if prefs[i].q == prefs[j].q {
			return prefs[i].idx < prefs[j].idx
		}
		return prefs[i].q > prefs[j].q
	})

	seen := map[Language]struct{}{}
	langs := make([]Language, 0, len(prefs))
	for _, pref := range prefs {
		if _, ok := seen[pref.lang]; ok {
			continue
		}
		seen[pref.lang] = struct{}{}
		langs = append(langs, pref.lang)
	}

	return langs
}

func ParseAcceptLanguage(value string) string {
	langs := ParseAcceptLanguageList(value)
	if len(langs) > 0 {
		return string(langs[0])
	}

	parsed, err := Parse(value)
	if err != nil {
		return ""
	}

	return string(parsed)
}

func ParseAcceptLanguageHeader(headers http.Header) string {
	return string(ParseForHeader(headers))
}

func ParseForHeader(headers http.Header) Language {
	if len(headers) == 0 {
		return Language("")
	}

	var raw string
	if values, ok := headers["Accept-Language"]; ok {
		raw = strings.Join(values, ",")
	} else if values, ok := headers["accept-language"]; ok {
		raw = strings.Join(values, ",")
	}

	if raw == "" {
		return Language("")
	}

	langs := ParseAcceptLanguageList(raw)
	if len(langs) > 0 {
		return langs[0]
	}

	parsed, err := Parse(raw)
	if err != nil {
		return Language("")
	}

	return parsed
}

func MatchSupportedByAcceptLanguage(value string, supported []Language, fallback Language) Language {
	if len(supported) == 0 {
		return fallback
	}

	acceptLangs := ParseAcceptLanguageList(value)
	if len(acceptLangs) == 0 {
		return fallback
	}

	supportedTags := make([]xlanguage.Tag, 0, len(supported))
	matchedSupported := make([]Language, 0, len(supported))
	for _, lang := range supported {
		parsed, err := Parse(string(lang))
		if err != nil {
			continue
		}

		supportedTags = append(supportedTags, xlanguage.Make(tagTextForLanguage(string(parsed))))
		matchedSupported = append(matchedSupported, lang)
	}

	if len(supportedTags) == 0 {
		return fallback
	}

	acceptTags := make([]xlanguage.Tag, 0, len(acceptLangs))
	for _, lang := range acceptLangs {
		acceptTags = append(acceptTags, xlanguage.Make(tagTextForLanguage(string(lang))))
	}
	if len(acceptTags) == 0 {
		return fallback
	}

	matcher := xlanguage.NewMatcher(supportedTags)
	_, idx, _ := matcher.Match(acceptTags...)
	if idx < 0 || idx >= len(matchedSupported) {
		return fallback
	}

	return matchedSupported[idx]
}

func cleanLanguageToken(lang string) string {
	lang = strings.TrimSpace(lang)
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

	return strings.ToLower(strings.TrimSpace(lang))
}

func isWellFormedLanguageTag(token string) bool {
	parts := strings.Split(token, "-")
	if len(parts) == 0 {
		return false
	}

	if len(parts[0]) < 2 || len(parts[0]) > 3 || !isAllLetters(parts[0]) {
		return false
	}

	idx := 1
	// extlang (common 3-letter), support one for practical compatibility.
	if idx < len(parts) && len(parts[idx]) == 3 && isAllLetters(parts[idx]) {
		idx++
	}

	// script
	if idx < len(parts) && len(parts[idx]) == 4 && isAllLetters(parts[idx]) {
		idx++
	}

	// region
	if idx < len(parts) {
		region := parts[idx]
		if (len(region) == 2 && isAllLetters(region)) || (len(region) == 3 && isAllDigits(region)) {
			idx++
		}
	}

	// variants
	for ; idx < len(parts); idx++ {
		part := parts[idx]
		if (len(part) >= 5 && len(part) <= 8 && isAlphaNum(part)) ||
			(len(part) == 4 && isAlphaNum(part) && isAllDigits(part[:1])) {
			continue
		}
		return false
	}

	return true
}

func tagTextForLanguage(token string) string {
	switch token {
	case "zh-chs":
		return "zh-Hans"
	case "zh-cht":
		return "zh-Hant"
	default:
		return canonicalTag(token)
	}
}

func canonicalTag(token string) string {
	parts := strings.Split(token, "-")
	for i := range parts {
		p := parts[i]
		switch {
		case i == 0:
			parts[i] = strings.ToLower(p)
		case len(p) == 4 && isAllLetters(p):
			parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		case len(p) == 2 && isAllLetters(p):
			parts[i] = strings.ToUpper(p)
		case len(p) == 3 && isAllDigits(p):
			parts[i] = p
		default:
			parts[i] = strings.ToLower(p)
		}
	}

	return strings.Join(parts, "-")
}

func isAllLetters(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') {
			return false
		}
	}
	return true
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func isAlphaNum(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}
