package i18n

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAcceptLanguage(t *testing.T) {
	// Save original default language
	originalDefault := DefaultI18n.DefaultLang()
	defer DefaultI18n.SetDefaultLang(originalDefault)

	tests := []struct {
		name         string
		defaultLang  string
		inputLang    []string
		expected     string
	}{
		{
			name:        "no language specified - use default",
			defaultLang: "en",
			inputLang:   []string{},
			expected:    "en;q=/1.0",
		},
		{
			name:        "language specified",
			defaultLang: "en",
			inputLang:   []string{"zh-CN"},
			expected:    "zh-CN;q=/1.0",
		},
		{
			name:        "multiple languages - use first",
			defaultLang: "en",
			inputLang:   []string{"fr", "de", "es"},
			expected:    "fr;q=/1.0",
		},
		{
			name:        "different default language",
			defaultLang: "zh-cn",
			inputLang:   []string{},
			expected:    "zh-cn;q=/1.0",
		},
		{
			name:        "override default with parameter",
			defaultLang: "zh-CN",
			inputLang:   []string{"ja"},
			expected:    "ja;q=/1.0",
		},
		{
			name:        "empty string language",
			defaultLang: "en",
			inputLang:   []string{""},
			expected:    ";q=/1.0",
		},
		{
			name:        "complex language code",
			defaultLang: "en",
			inputLang:   []string{"zh-Hans-CN"},
			expected:    "zh-Hans-CN;q=/1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the default language for this test
			DefaultI18n.SetDefaultLang(tt.defaultLang)
			
			result := GetAcceptLanguage(tt.inputLang...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAcceptLanguage_DefaultBehavior(t *testing.T) {
	// Test that it uses DefaultI18n.DefaultLang() when no language is provided
	
	// Set a specific default language
	DefaultI18n.SetDefaultLang("test-lang")
	
	result := GetAcceptLanguage()
	expected := "test-lang;q=/1.0"
	
	assert.Equal(t, expected, result)
}

func TestGetAcceptLanguage_ParameterOverride(t *testing.T) {
	// Test that parameter overrides the default
	
	DefaultI18n.SetDefaultLang("default-lang")
	
	result := GetAcceptLanguage("override-lang")
	expected := "override-lang;q=/1.0"
	
	assert.Equal(t, expected, result)
}

func TestGetAcceptLanguage_EmptyParameter(t *testing.T) {
	// Test behavior with empty parameter
	
	DefaultI18n.SetDefaultLang("default-lang")
	
	result := GetAcceptLanguage("")
	expected := ";q=/1.0"
	
	assert.Equal(t, expected, result)
}

func TestGetAcceptLanguage_MultipleParameters(t *testing.T) {
	// Test that only the first parameter is used
	
	result := GetAcceptLanguage("first", "second", "third")
	expected := "first;q=/1.0"
	
	assert.Equal(t, expected, result)
}

func TestGetAcceptLanguage_RealWorldLanguageCodes(t *testing.T) {
	// Test with real-world language codes
	
	realWorldCases := []struct {
		lang     string
		expected string
	}{
		{"en-US", "en-US;q=/1.0"},
		{"zh-CN", "zh-CN;q=/1.0"},
		{"zh-TW", "zh-TW;q=/1.0"},
		{"ja-JP", "ja-JP;q=/1.0"},
		{"ko-KR", "ko-KR;q=/1.0"},
		{"fr-FR", "fr-FR;q=/1.0"},
		{"de-DE", "de-DE;q=/1.0"},
		{"es-ES", "es-ES;q=/1.0"},
		{"pt-BR", "pt-BR;q=/1.0"},
		{"it-IT", "it-IT;q=/1.0"},
		{"ru-RU", "ru-RU;q=/1.0"},
		{"ar-SA", "ar-SA;q=/1.0"},
		{"hi-IN", "hi-IN;q=/1.0"},
		{"th-TH", "th-TH;q=/1.0"},
		{"vi-VN", "vi-VN;q=/1.0"},
	}
	
	for _, tc := range realWorldCases {
		t.Run(tc.lang, func(t *testing.T) {
			result := GetAcceptLanguage(tc.lang)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetAcceptLanguage_SpecialCharacters(t *testing.T) {
	// Test with special characters and edge cases
	
	specialCases := []struct {
		lang     string
		expected string
	}{
		{"en-US-POSIX", "en-US-POSIX;q=/1.0"},
		{"zh-Hant-HK", "zh-Hant-HK;q=/1.0"},
		{"zh-Hans-CN", "zh-Hans-CN;q=/1.0"},
		{"sr-Latn-RS", "sr-Latn-RS;q=/1.0"},
		{"az-Cyrl-AZ", "az-Cyrl-AZ;q=/1.0"},
	}
	
	for _, tc := range specialCases {
		t.Run(tc.lang, func(t *testing.T) {
			result := GetAcceptLanguage(tc.lang)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetAcceptLanguage_ConsistentFormat(t *testing.T) {
	// Test that the format is always consistent: "lang;q=/1.0"
	
	// This tests the format specification in the function
	result := GetAcceptLanguage("test-lang")
	expected := "test-lang;q=/1.0"
	
	// Should match exact format
	assert.Equal(t, expected, result)
	
	// Should always end with ;q=/1.0
	assert.True(t, strings.HasSuffix(result, ";q=/1.0"))
	
	// Should start with the language
	assert.True(t, strings.HasPrefix(result, "test-lang"))
}

func TestGetAcceptLanguage_Integration(t *testing.T) {
	// Test integration with DefaultI18n
	
	// Verify that it actually uses DefaultI18n.DefaultLang()
	testLang := "integration-test-lang"
	DefaultI18n.SetDefaultLang(testLang)
	
	result := GetAcceptLanguage()
	expected := testLang + ";q=/1.0"
	
	assert.Equal(t, expected, result)
	
	// Verify the default can be retrieved
	retrievedDefault := DefaultI18n.DefaultLang()
	assert.Equal(t, testLang, retrievedDefault)
}