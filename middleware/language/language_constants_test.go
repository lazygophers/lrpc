package language

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	xlanguage "golang.org/x/text/language"
)

func TestXTextSemanticConstantsAligned(t *testing.T) {
	ours := []Language{
		Und,
		Afrikaans,
		Amharic,
		Arabic,
		ModernStandardArabic,
		Azerbaijani,
		Bulgarian,
		Bengali,
		Catalan,
		Czech,
		Danish,
		German,
		Greek,
		English,
		AmericanEnglish,
		BritishEnglish,
		Spanish,
		EuropeanSpanish,
		LatinAmericanSpanish,
		Estonian,
		Persian,
		Finnish,
		Filipino,
		French,
		CanadianFrench,
		Gujarati,
		Hebrew,
		Hindi,
		Croatian,
		Hungarian,
		Armenian,
		Indonesian,
		Icelandic,
		Italian,
		Japanese,
		Georgian,
		Kazakh,
		Khmer,
		Kannada,
		Korean,
		Kirghiz,
		Lao,
		Lithuanian,
		Latvian,
		Macedonian,
		Malayalam,
		Mongolian,
		Marathi,
		Malay,
		Burmese,
		Nepali,
		Dutch,
		Norwegian,
		Punjabi,
		Polish,
		Portuguese,
		BrazilianPortuguese,
		EuropeanPortuguese,
		Romanian,
		Russian,
		Sinhala,
		Slovak,
		Slovenian,
		Albanian,
		Serbian,
		SerbianLatin,
		Swedish,
		Swahili,
		Tamil,
		Telugu,
		Thai,
		Turkish,
		Ukrainian,
		Urdu,
		Uzbek,
		Vietnamese,
		Chinese,
		SimplifiedChinese,
		TraditionalChinese,
		Zulu,
	}

	wants := []string{
		strings.ToLower(xlanguage.Und.String()),
		strings.ToLower(xlanguage.Afrikaans.String()),
		strings.ToLower(xlanguage.Amharic.String()),
		strings.ToLower(xlanguage.Arabic.String()),
		strings.ToLower(xlanguage.ModernStandardArabic.String()),
		strings.ToLower(xlanguage.Azerbaijani.String()),
		strings.ToLower(xlanguage.Bulgarian.String()),
		strings.ToLower(xlanguage.Bengali.String()),
		strings.ToLower(xlanguage.Catalan.String()),
		strings.ToLower(xlanguage.Czech.String()),
		strings.ToLower(xlanguage.Danish.String()),
		strings.ToLower(xlanguage.German.String()),
		strings.ToLower(xlanguage.Greek.String()),
		strings.ToLower(xlanguage.English.String()),
		strings.ToLower(xlanguage.AmericanEnglish.String()),
		strings.ToLower(xlanguage.BritishEnglish.String()),
		strings.ToLower(xlanguage.Spanish.String()),
		strings.ToLower(xlanguage.EuropeanSpanish.String()),
		strings.ToLower(xlanguage.LatinAmericanSpanish.String()),
		strings.ToLower(xlanguage.Estonian.String()),
		strings.ToLower(xlanguage.Persian.String()),
		strings.ToLower(xlanguage.Finnish.String()),
		strings.ToLower(xlanguage.Filipino.String()),
		strings.ToLower(xlanguage.French.String()),
		strings.ToLower(xlanguage.CanadianFrench.String()),
		strings.ToLower(xlanguage.Gujarati.String()),
		strings.ToLower(xlanguage.Hebrew.String()),
		strings.ToLower(xlanguage.Hindi.String()),
		strings.ToLower(xlanguage.Croatian.String()),
		strings.ToLower(xlanguage.Hungarian.String()),
		strings.ToLower(xlanguage.Armenian.String()),
		strings.ToLower(xlanguage.Indonesian.String()),
		strings.ToLower(xlanguage.Icelandic.String()),
		strings.ToLower(xlanguage.Italian.String()),
		strings.ToLower(xlanguage.Japanese.String()),
		strings.ToLower(xlanguage.Georgian.String()),
		strings.ToLower(xlanguage.Kazakh.String()),
		strings.ToLower(xlanguage.Khmer.String()),
		strings.ToLower(xlanguage.Kannada.String()),
		strings.ToLower(xlanguage.Korean.String()),
		strings.ToLower(xlanguage.Kirghiz.String()),
		strings.ToLower(xlanguage.Lao.String()),
		strings.ToLower(xlanguage.Lithuanian.String()),
		strings.ToLower(xlanguage.Latvian.String()),
		strings.ToLower(xlanguage.Macedonian.String()),
		strings.ToLower(xlanguage.Malayalam.String()),
		strings.ToLower(xlanguage.Mongolian.String()),
		strings.ToLower(xlanguage.Marathi.String()),
		strings.ToLower(xlanguage.Malay.String()),
		strings.ToLower(xlanguage.Burmese.String()),
		strings.ToLower(xlanguage.Nepali.String()),
		strings.ToLower(xlanguage.Dutch.String()),
		strings.ToLower(xlanguage.Norwegian.String()),
		strings.ToLower(xlanguage.Punjabi.String()),
		strings.ToLower(xlanguage.Polish.String()),
		strings.ToLower(xlanguage.Portuguese.String()),
		strings.ToLower(xlanguage.BrazilianPortuguese.String()),
		strings.ToLower(xlanguage.EuropeanPortuguese.String()),
		strings.ToLower(xlanguage.Romanian.String()),
		strings.ToLower(xlanguage.Russian.String()),
		strings.ToLower(xlanguage.Sinhala.String()),
		strings.ToLower(xlanguage.Slovak.String()),
		strings.ToLower(xlanguage.Slovenian.String()),
		strings.ToLower(xlanguage.Albanian.String()),
		strings.ToLower(xlanguage.Serbian.String()),
		strings.ToLower(xlanguage.SerbianLatin.String()),
		strings.ToLower(xlanguage.Swedish.String()),
		strings.ToLower(xlanguage.Swahili.String()),
		strings.ToLower(xlanguage.Tamil.String()),
		strings.ToLower(xlanguage.Telugu.String()),
		strings.ToLower(xlanguage.Thai.String()),
		strings.ToLower(xlanguage.Turkish.String()),
		strings.ToLower(xlanguage.Ukrainian.String()),
		strings.ToLower(xlanguage.Urdu.String()),
		strings.ToLower(xlanguage.Uzbek.String()),
		strings.ToLower(xlanguage.Vietnamese.String()),
		strings.ToLower(xlanguage.Chinese.String()),
		strings.ToLower(xlanguage.SimplifiedChinese.String()),
		strings.ToLower(xlanguage.TraditionalChinese.String()),
		strings.ToLower(xlanguage.Zulu.String()),
	}

	if !assert.Equal(t, len(wants), len(ours)) {
		return
	}

	for i := range ours {
		assert.Equal(t, wants[i], string(ours[i]))
	}
}

func TestISO6391ConstantsCoverAllXTextTwoLetterCodes(t *testing.T) {
	content, err := os.ReadFile("language_constants.go")
	if !assert.NoError(t, err) {
		return
	}

	re := regexp.MustCompile(`ISO6391[A-Z]{2}\s+Language\s*=\s*"([a-z]{2})"`)
	matches := re.FindAllStringSubmatch(string(content), -1)

	isoSet := make(map[string]struct{}, len(matches))
	for _, m := range matches {
		isoSet[m[1]] = struct{}{}
	}

	expected := make(map[string]struct{})
	for a := 'a'; a <= 'z'; a++ {
		for b := 'a'; b <= 'z'; b++ {
			code := string([]rune{a, b})
			base, parseErr := xlanguage.ParseBase(code)
			if parseErr != nil {
				continue
			}
			if base.String() != code {
				continue
			}
			expected[code] = struct{}{}
		}
	}

	var missing []string
	for code := range expected {
		if _, ok := isoSet[code]; !ok {
			missing = append(missing, code)
		}
	}

	var extra []string
	for code := range isoSet {
		if _, ok := expected[code]; !ok {
			extra = append(extra, code)
		}
	}

	sort.Strings(missing)
	sort.Strings(extra)

	assert.True(t, len(missing) == 0, "missing ISO639-1 constants: %s", strings.Join(missing, ","))
	assert.True(t, len(extra) == 0, "unexpected ISO639-1 constants: %s", strings.Join(extra, ","))
}
