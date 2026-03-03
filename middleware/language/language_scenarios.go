package language

func mapLanguageForParse(lang string) (Language, bool) {
	switch lang {
	case "en":
		return English, true
	case "en-us":
		return AmericanEnglish, true
	case "en-gb":
		return BritishEnglish, true
	case "en-ca":
		return CanadianEnglish, true

	case "zh":
		return Chinese, true
	case "zh-cn":
		return ChineseMainland, true
	case "zh-hk":
		return ChineseHongKong, true
	case "zh-tw":
		return ChineseTaiwan, true
	case "zh-mo":
		return ChineseMacao, true
	case "zh-sg":
		return ChineseSingapore, true
	case "zh-cht":
		return ChineseTraditional, true
	case "zh-hant":
		return TraditionalChinese, true
	case "zh-hans", "zh-chs":
		return SimplifiedChinese, true

	case "ja":
		return Japanese, true
	case "ko":
		return Korean, true
	case "fr":
		return French, true
	case "fr-ca":
		return CanadianFrench, true
	case "de":
		return German, true
	case "de-ch":
		return SwissGerman, true
	case "es":
		return Spanish, true
	case "es-es":
		return EuropeanSpanish, true
	case "es-mx":
		return MexicanSpanish, true
	case "it":
		return Italian, true
	case "pt":
		return Portuguese, true
	case "pt-br":
		return BrazilianPortuguese, true
	case "pt-pt":
		return EuropeanPortuguese, true
	case "ru":
		return Russian, true
	case "uk":
		return Ukrainian, true
	case "ar":
		return Arabic, true
	case "fa":
		return Persian, true
	case "tr":
		return Turkish, true
	case "hi":
		return Hindi, true
	case "bn":
		return Bengali, true
	case "ta":
		return Tamil, true
	case "te":
		return Telugu, true
	case "mr":
		return Marathi, true
	case "gu":
		return Gujarati, true
	case "pa":
		return Punjabi, true
	case "th":
		return Thai, true
	case "vi":
		return Vietnamese, true
	case "id":
		return Indonesian, true
	case "ms":
		return Malay, true
	case "fil":
		return Filipino, true
	case "nl":
		return Dutch, true
	case "sv":
		return Swedish, true
	case "no":
		return Norwegian, true
	case "da":
		return Danish, true
	case "fi":
		return Finnish, true
	case "pl":
		return Polish, true
	case "cs":
		return Czech, true
	case "sk":
		return Slovak, true
	case "hu":
		return Hungarian, true
	case "ro":
		return Romanian, true
	case "bg":
		return Bulgarian, true
	case "el":
		return Greek, true
	case "sr":
		return Serbian, true
	case "hr":
		return Croatian, true
	case "sl":
		return Slovenian, true
	case "lt":
		return Lithuanian, true
	case "lv":
		return Latvian, true
	case "et":
		return Estonian, true
	case "he", "iw":
		return Hebrew, true
	case "is":
		return Icelandic, true
	case "ga":
		return Irish, true
	case "eu":
		return Basque, true
	case "ca":
		return Catalan, true
	case "gl":
		return Galician, true
	case "af":
		return Afrikaans, true
	case "sw":
		return Swahili, true
	default:
		return "", false
	}
}

func mapLanguageForAccept(lang string) (Language, bool) {
	if lang == "" || lang == "*" {
		return "", false
	}

	return mapLanguageForParse(lang)
}
