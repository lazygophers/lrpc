package language

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	xlanguage "golang.org/x/text/language"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Language
		expectError bool
	}{
		{"zh-cn", "zh-CN", ChineseMainland, false},
		{"zh-hans", "zh-hans", SimplifiedChinese, false},
		{"zh-hk", "zh-HK", ChineseHongKong, false},
		{"zh-tw", "zh-TW", ChineseTaiwan, false},
		{"zh-mo", "zh-MO", ChineseMacao, false},
		{"zh-sg", "zh-SG", ChineseSingapore, false},
		{"zh-cht", "zh-CHT", ChineseTraditional, false},
		{"zh-hant", "zh-Hant", TraditionalChinese, false},
		{"en-us", "en-US", AmericanEnglish, false},
		{"en", "en", English, false},
		{"fr-ca", "fr-CA", CanadianFrench, false},
		{"pt-br", "pt-BR", BrazilianPortuguese, false},
		{"hebrew old code iw", "iw", Hebrew, false},
		{"invalid", "invalid-lang-code", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			input := tt.input

			// Act
			got, err := Parse(input)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestBase(t *testing.T) {
	tests := []struct {
		name     string
		input    Language
		expected Language
	}{
		{"american english", AmericanEnglish, English},
		{"british english", BritishEnglish, English},
		{"mainland chinese", ChineseMainland, Chinese},
		{"hongkong chinese", ChineseHongKong, Chinese},
		{"simplified chinese", SimplifiedChinese, Chinese},
		{"traditional chinese", TraditionalChinese, Chinese},
		{"french", French, French},
		{"unknown", Language("abc"), Language("abc")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			input := tt.input

			// Act
			got := Base(input)

			// Assert
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestParseLangCode(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedTag xlanguage.Tag
		expectedRaw string
		expectError bool
	}{
		{"zh-cn", "zh-CN", xlanguage.Make("zh-CN"), "zh-cn", false},
		{"zh-chs", "zh-CHS", xlanguage.SimplifiedChinese, "zh-chs", false},
		{"zh-tw", "zh-TW", xlanguage.Make("zh-TW"), "zh-tw", false},
		{"zh-hk", "zh-HK", xlanguage.Make("zh-HK"), "zh-hk", false},
		{"zh-mo", "zh-MO", xlanguage.Make("zh-MO"), "zh-mo", false},
		{"zh-sg", "zh-SG", xlanguage.Make("zh-SG"), "zh-sg", false},
		{"zh-cht", "zh-CHT", xlanguage.TraditionalChinese, "zh-cht", false},
		{"zh-hant", "zh-Hant", xlanguage.TraditionalChinese, "zh-hant", false},
		{"en-us", "en-US", xlanguage.AmericanEnglish, "en-us", false},
		{"invalid", "invalid-lang-code", xlanguage.Und, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			input := tt.input

			// Act
			got, err := ParseLangCode(input)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.expectedRaw, got.Lang)
			assert.Equal(t, tt.expectedTag, got.Tag)
		})
	}
}

func TestParseLangCodeCache(t *testing.T) {
	// Arrange
	first, err := ParseLangCode("en-US")
	require.NoError(t, err)
	require.NotNil(t, first)

	// Act
	second, err := ParseLangCode("en-US")
	require.NoError(t, err)
	require.NotNil(t, second)

	// Assert
	assert.Same(t, first, second)
}

func TestParseAcceptLanguageList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Language
	}{
		{
			name:     "quality order",
			input:    "zh;q=0.1,en-US;q=1.0,fr;q=0.9",
			expected: []Language{AmericanEnglish, French, Chinese},
		},
		{
			name:     "dedupe aliases",
			input:    "zh-CN,zh-Hans;q=0.9,zh;q=0.8",
			expected: []Language{ChineseMainland, SimplifiedChinese, Chinese},
		},
		{
			name:     "fallback single",
			input:    "pt-BR",
			expected: []Language{BrazilianPortuguese},
		},
		{
			name:     "empty",
			input:    "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			input := tt.input

			// Act
			got := ParseAcceptLanguageList(input)

			// Assert
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestParseAcceptLanguage(t *testing.T) {
	// Arrange & Act & Assert
	assert.Equal(t, "en-us", ParseAcceptLanguage("en-US,en;q=0.9,zh;q=0.8"))
	assert.Equal(t, "en-us", ParseAcceptLanguage("zh;q=0.1,en-US;q=1.0"))
	assert.Equal(t, "zh-cn", ParseAcceptLanguage("zh-CN"))
}

func TestParseForHeader(t *testing.T) {
	t.Run("accept-language", func(t *testing.T) {
		// Arrange
		headers := http.Header{
			"Accept-Language": []string{"zh-CN,zh;q=0.9"},
		}

		// Act
		got := ParseForHeader(headers)

		// Assert
		assert.Equal(t, ChineseMainland, got)
	})

	t.Run("lower-case", func(t *testing.T) {
		// Arrange
		headers := http.Header{
			"accept-language": []string{"en-US,en;q=0.9"},
		}

		// Act
		got := ParseForHeader(headers)

		// Assert
		assert.Equal(t, AmericanEnglish, got)
	})

	t.Run("missing", func(t *testing.T) {
		// Arrange
		headers := http.Header{}

		// Act
		got := ParseForHeader(headers)

		// Assert
		assert.Equal(t, Language(""), got)
	})
}

func TestParseAcceptLanguageHeader(t *testing.T) {
	// Arrange
	headers := http.Header{
		"Accept-Language": []string{"zh-CN,zh;q=0.9"},
	}

	// Act
	got := ParseAcceptLanguageHeader(headers)

	// Assert
	assert.Equal(t, "zh-cn", got)

	// Ensure backward-compatible wrapper follows ParseForHeader.
	assert.Equal(t, string(ParseForHeader(headers)), got)
}

func TestParseForHeader_Invalid(t *testing.T) {
	// Arrange
	headers := http.Header{
		"Accept-Language": []string{"$$$"},
	}

	// Act
	got := ParseForHeader(headers)

	// Assert
	assert.Equal(t, Language(""), got)
}

func TestMatchSupportedByAcceptLanguage(t *testing.T) {
	t.Run("best match", func(t *testing.T) {
		// Arrange
		supported := []Language{SimplifiedChinese, AmericanEnglish, Japanese}
		accept := "en-GB,en;q=0.9,zh-CN;q=0.8"

		// Act
		got := MatchSupportedByAcceptLanguage(accept, supported, SimplifiedChinese)

		// Assert
		assert.Equal(t, AmericanEnglish, got)
	})

	t.Run("fallback empty supported", func(t *testing.T) {
		// Arrange
		accept := "en-US,en;q=0.9"

		// Act
		got := MatchSupportedByAcceptLanguage(accept, nil, Japanese)

		// Assert
		assert.Equal(t, Japanese, got)
	})

	t.Run("fallback parse failure", func(t *testing.T) {
		// Arrange
		supported := []Language{Language("bad@@"), French}

		// Act
		got := MatchSupportedByAcceptLanguage("$$$", supported, French)

		// Assert
		assert.Equal(t, French, got)
	})
}

func TestMustParse(t *testing.T) {
	assert.Equal(t, English, MustParse("en"))
	assert.Panics(t, func() {
		_ = MustParse("invalid-lang-code")
	})
}

func TestMustParseLangCode(t *testing.T) {
	assert.Equal(t, "en", MustParseLangCode("en").Lang)
	assert.Panics(t, func() {
		_ = MustParseLangCode("invalid-lang-code")
	})
}
