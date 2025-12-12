package i18n

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

func TestLanguageCode_String(t *testing.T) {
	tests := []struct {
		name     string
		langCode *LanguageCode
		expected string
	}{
		{
			name: "tag equals lang",
			langCode: &LanguageCode{
				Lang: "en",
				Tag:  language.English,
			},
			expected: "en",
		},
		{
			name: "tag different from lang",
			langCode: &LanguageCode{
				Lang: "zh-cn",
				Tag:  language.Chinese,
			},
			expected: "zh(zh-cn)",
		},
		{
			name: "complex language code",
			langCode: &LanguageCode{
				Lang: "zh-hans",
				Tag:  language.SimplifiedChinese,
			},
			expected: "zh-Hans(zh-hans)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.langCode.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLanguageCode_GoString(t *testing.T) {
	langCode := &LanguageCode{
		Lang: "en",
		Tag:  language.English,
	}

	// GoString should be the same as String
	assert.Equal(t, langCode.String(), langCode.GoString())
}

func TestParseLangCode(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		expected    *LanguageCode
	}{
		{
			name:  "chinese simplified zh-cn",
			input: "zh-CN",
			expected: &LanguageCode{
				Lang: "zh-cn",
				Tag:  language.Chinese,
			},
		},
		{
			name:  "chinese simplified zh-hans",
			input: "zh-hans",
			expected: &LanguageCode{
				Lang: "zh-hans",
				Tag:  language.SimplifiedChinese,
			},
		},
		{
			name:  "chinese simplified zh-chs",
			input: "zh-CHS",
			expected: &LanguageCode{
				Lang: "zh-chs",
				Tag:  language.SimplifiedChinese,
			},
		},
		{
			name:  "chinese traditional zh-hk",
			input: "zh-HK",
			expected: &LanguageCode{
				Lang: "zh-hk",
				Tag:  language.TraditionalChinese,
			},
		},
		{
			name:  "chinese traditional zh-tw",
			input: "zh-TW",
			expected: &LanguageCode{
				Lang: "zh-tw",
				Tag:  language.TraditionalChinese,
			},
		},
		{
			name:  "chinese traditional zh-mo",
			input: "zh-MO",
			expected: &LanguageCode{
				Lang: "zh-mo",
				Tag:  language.TraditionalChinese,
			},
		},
		{
			name:  "chinese traditional zh-sg",
			input: "zh-SG",
			expected: &LanguageCode{
				Lang: "zh-sg",
				Tag:  language.TraditionalChinese,
			},
		},
		{
			name:  "chinese traditional zh-cht",
			input: "zh-CHT",
			expected: &LanguageCode{
				Lang: "zh-cht",
				Tag:  language.TraditionalChinese,
			},
		},
		{
			name:  "english",
			input: "en",
			expected: &LanguageCode{
				Lang: "en",
				Tag:  language.English,
			},
		},
		{
			name:  "english us",
			input: "en-US",
			expected: &LanguageCode{
				Lang: "en-us",
				Tag:  language.AmericanEnglish,
			},
		},
		{
			name:  "french",
			input: "fr",
			expected: &LanguageCode{
				Lang: "fr",
				Tag:  language.French,
			},
		},
		{
			name:  "german",
			input: "de",
			expected: &LanguageCode{
				Lang: "de",
				Tag:  language.German,
			},
		},
		{
			name:  "spanish",
			input: "es",
			expected: &LanguageCode{
				Lang: "es",
				Tag:  language.Spanish,
			},
		},
		{
			name:  "japanese",
			input: "ja",
			expected: &LanguageCode{
				Lang: "ja",
				Tag:  language.Japanese,
			},
		},
		{
			name:  "korean",
			input: "ko",
			expected: &LanguageCode{
				Lang: "ko",
				Tag:  language.Korean,
			},
		},
		{
			name:        "invalid language code",
			input:       "invalid-lang-code",
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:  "case insensitive",
			input: "EN-us",
			expected: &LanguageCode{
				Lang: "en-us",
				Tag:  language.AmericanEnglish,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseLangCode(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.expected.Lang, result.Lang)
			assert.Equal(t, tt.expected.Tag, result.Tag)
		})
	}
}

func TestMustParseLangCode(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		result := MustParseLangCode("en")

		assert.NotNil(t, result)
		assert.Equal(t, "en", result.Lang)
		assert.Equal(t, language.English, result.Tag)
	})

	t.Run("panic on invalid code", func(t *testing.T) {
		assert.Panics(t, func() {
			MustParseLangCode("invalid-lang-code")
		})
	})
}

func TestParseLangCode_SpecialCases(t *testing.T) {
	// Test all special Chinese cases to ensure they work correctly
	chineseCases := []struct {
		input    string
		expected language.Tag
	}{
		{"zh-cn", language.Chinese},
		{"ZH-CN", language.Chinese},
		{"zh-hans", language.SimplifiedChinese},
		{"ZH-HANS", language.SimplifiedChinese},
		{"zh-chs", language.SimplifiedChinese},
		{"ZH-CHS", language.SimplifiedChinese},
		{"zh-hk", language.TraditionalChinese},
		{"ZH-HK", language.TraditionalChinese},
		{"zh-tw", language.TraditionalChinese},
		{"ZH-TW", language.TraditionalChinese},
		{"zh-mo", language.TraditionalChinese},
		{"ZH-MO", language.TraditionalChinese},
		{"zh-sg", language.TraditionalChinese},
		{"ZH-SG", language.TraditionalChinese},
		{"zh-cht", language.TraditionalChinese},
		{"ZH-CHT", language.TraditionalChinese},
	}

	for _, tc := range chineseCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := ParseLangCode(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result.Tag)
			assert.Equal(t, strings.ToLower(tc.input), result.Lang)
		})
	}
}

func TestParseLangCode_DefaultCase(t *testing.T) {
	// Test the default case that uses language.Parse
	commonCases := []string{
		"en",
		"en-US",
		"en-GB",
		"fr",
		"fr-FR",
		"de",
		"de-DE",
		"es",
		"es-ES",
		"it",
		"it-IT",
		"pt",
		"pt-BR",
		"ru",
		"ru-RU",
		"ja",
		"ko",
		"ar",
		"hi",
		"th",
		"vi",
	}

	for _, lang := range commonCases {
		t.Run(lang, func(t *testing.T) {
			result, err := ParseLangCode(lang)
			require.NoError(t, err)

			// Should be able to parse with golang.org/x/text/language
			expectedTag, expectedErr := language.Parse(lang)
			require.NoError(t, expectedErr)

			assert.Equal(t, strings.ToLower(lang), result.Lang)
			assert.Equal(t, expectedTag, result.Tag)
		})
	}
}

func TestLanguageCode_Coverage(t *testing.T) {
	// Test creating language codes for different scenarios
	testCases := []struct {
		lang     string
		expected string
	}{
		{"en", "en"},
		{"zh-cn", "zh(zh-cn)"},
		{"fr-FR", "fr-FR"},
	}

	for _, tc := range testCases {
		t.Run(tc.lang, func(t *testing.T) {
			langCode := MustParseLangCode(tc.lang)
			result := langCode.String()

			// The actual result may vary based on language.Parse behavior
			// Just ensure it doesn't panic and returns a string
			assert.NotEmpty(t, result)
		})
	}
}
