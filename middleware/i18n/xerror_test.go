package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestI18nForXerror_Localize(t *testing.T) {
	// Create test I18n instance
	i18n := NewI18n()

	// Add test packs
	enPack := NewPack("en")
	enPack.corpus["error.100"] = "Not Found"
	enPack.corpus["error.200"] = "Success"
	enPack.corpus["error.500"] = "Internal Server Error"
	i18n.packMap["en"] = enPack

	zhPack := NewPack("zh")
	zhPack.corpus["error.100"] = "未找到"
	zhPack.corpus["error.500"] = "内部服务器错误"
	i18n.packMap["zh"] = zhPack

	i18n.SetDefaultLang("en")

	// Create I18nForXerror instance
	xerrorI18n := NewI18nForXerror(i18n)

	tests := []struct {
		name     string
		key      int32
		langs    []string
		expected string
		found    bool
	}{
		{
			name:     "existing error in English",
			key:      100,
			langs:    []string{"en"},
			expected: "Not Found",
			found:    true,
		},
		{
			name:     "existing error in Chinese",
			key:      100,
			langs:    []string{"zh"},
			expected: "未找到",
			found:    true,
		},
		{
			name:     "fallback to default language",
			key:      200,
			langs:    []string{"zh"}, // Chinese doesn't have this key
			expected: "Success",
			found:    true,
		},
		{
			name:     "non-existing error",
			key:      999,
			langs:    []string{"en"},
			expected: "error.999",
			found:    false,
		},
		{
			name:     "no language specified",
			key:      500,
			langs:    []string{},
			expected: "Internal Server Error",
			found:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, found := xerrorI18n.Localize(tt.key, tt.langs...)

			assert.Equal(t, tt.expected, message)
			assert.Equal(t, tt.found, found)
		})
	}
}

func TestNewI18nForXerror(t *testing.T) {
	i18n := NewI18n()

	t.Run("default prefix", func(t *testing.T) {
		xerrorI18n := NewI18nForXerror(i18n)

		assert.NotNil(t, xerrorI18n)
		assert.Equal(t, i18n, xerrorI18n.i18n)
		assert.Equal(t, "error.", xerrorI18n.prefix)
	})

	t.Run("custom prefix", func(t *testing.T) {
		customPrefix := "custom.error."
		xerrorI18n := NewI18nForXerror(i18n, customPrefix)

		assert.NotNil(t, xerrorI18n)
		assert.Equal(t, i18n, xerrorI18n.i18n)
		assert.Equal(t, customPrefix, xerrorI18n.prefix)
	})

	t.Run("multiple prefixes uses first", func(t *testing.T) {
		xerrorI18n := NewI18nForXerror(i18n, "first.", "second.")

		assert.NotNil(t, xerrorI18n)
		assert.Equal(t, "first.", xerrorI18n.prefix)
	})
}

func TestI18nForXerror_LocalizeWithCustomPrefix(t *testing.T) {
	// Create test I18n instance
	i18n := NewI18n()

	// Add test pack with custom prefix
	enPack := NewPack("en")
	enPack.corpus["custom.1001"] = "Custom Error"
	enPack.corpus["api.error.2001"] = "API Error"
	i18n.packMap["en"] = enPack
	i18n.SetDefaultLang("en")

	tests := []struct {
		name     string
		prefix   string
		key      int32
		expected string
		found    bool
	}{
		{
			name:     "custom prefix",
			prefix:   "custom.",
			key:      1001,
			expected: "Custom Error",
			found:    true,
		},
		{
			name:     "api error prefix",
			prefix:   "api.error.",
			key:      2001,
			expected: "API Error",
			found:    true,
		},
		{
			name:     "non-existing with custom prefix",
			prefix:   "custom.",
			key:      9999,
			expected: "custom.9999",
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xerrorI18n := NewI18nForXerror(i18n, tt.prefix)
			message, found := xerrorI18n.Localize(tt.key, "en")

			assert.Equal(t, tt.expected, message)
			assert.Equal(t, tt.found, found)
		})
	}
}

func TestI18nForXerror_LocalizeEdgeCases(t *testing.T) {
	i18n := NewI18n()

	// Test with empty corpus
	emptyI18n := NewI18nForXerror(i18n)
	message, found := emptyI18n.Localize(123)
	assert.Equal(t, "error.123", message)
	assert.False(t, found)

	// Test with negative error code
	negativeMessage, negativeFound := emptyI18n.Localize(-1)
	assert.Equal(t, "error.-1", negativeMessage)
	assert.False(t, negativeFound)

	// Test with zero error code
	zeroMessage, zeroFound := emptyI18n.Localize(0)
	assert.Equal(t, "error.0", zeroMessage)
	assert.False(t, zeroFound)
}

func TestI18nForXerror_LocalizeMultipleLanguages(t *testing.T) {
	i18n := NewI18n()

	// Add multiple language packs
	enPack := NewPack("en")
	enPack.corpus["error.404"] = "Page Not Found"
	i18n.packMap["en"] = enPack

	frPack := NewPack("fr")
	frPack.corpus["error.404"] = "Page non trouvée"
	i18n.packMap["fr"] = frPack

	esPack := NewPack("es")
	esPack.corpus["error.404"] = "Página no encontrada"
	i18n.packMap["es"] = esPack

	i18n.SetDefaultLang("en")

	xerrorI18n := NewI18nForXerror(i18n)

	// Test different languages
	enMessage, enFound := xerrorI18n.Localize(404, "en")
	assert.Equal(t, "Page Not Found", enMessage)
	assert.True(t, enFound)

	frMessage, frFound := xerrorI18n.Localize(404, "fr")
	assert.Equal(t, "Page non trouvée", frMessage)
	assert.True(t, frFound)

	esMessage, esFound := xerrorI18n.Localize(404, "es")
	assert.Equal(t, "Página no encontrada", esMessage)
	assert.True(t, esFound)

	// Test fallback for unsupported language
	deMessage, deFound := xerrorI18n.Localize(404, "de")
	assert.Equal(t, "Page Not Found", deMessage) // Should fallback to English
	assert.True(t, deFound)
}

func TestI18nForXerror_InterfaceCompliance(t *testing.T) {
	i18n := NewI18n()
	xerrorI18n := NewI18nForXerror(i18n)

	// Verify it implements the xerror.I18n interface
	// This is tested at compile time with the var _ xerror.I18n = (*I18nForXerror)(nil) declaration
	// But we can also test it at runtime
	assert.NotNil(t, xerrorI18n)

	// Test the interface method
	message, found := xerrorI18n.Localize(123)
	assert.IsType(t, "", message)
	assert.IsType(t, false, found)
}

func TestI18nForXerror_EmptyPrefix(t *testing.T) {
	i18n := NewI18n()

	// Add test pack with keys without prefix
	enPack := NewPack("en")
	enPack.corpus["404"] = "Direct Error"
	i18n.packMap["en"] = enPack
	i18n.SetDefaultLang("en")

	// Create with empty prefix
	xerrorI18n := NewI18nForXerror(i18n, "")

	message, found := xerrorI18n.Localize(404, "en")
	assert.Equal(t, "Direct Error", message)
	assert.True(t, found)
}

func TestI18nForXerror_Register(t *testing.T) {
	i18n := NewI18n()
	xerrorI18n := NewI18nForXerror(i18n)

	// Test registering error messages
	xerrorI18n.Register("en", 1001, "Invalid parameter")
	xerrorI18n.Register("zh", 1001, "参数无效")

	// Verify English translation
	msg, found := xerrorI18n.Localize(1001, "en")
	assert.True(t, found)
	assert.Equal(t, "Invalid parameter", msg)

	// Verify Chinese translation
	msg, found = xerrorI18n.Localize(1001, "zh")
	assert.True(t, found)
	assert.Equal(t, "参数无效", msg)

	// Test overwriting existing error message
	xerrorI18n.Register("en", 1001, "Invalid param")
	msg, found = xerrorI18n.Localize(1001, "en")
	assert.True(t, found)
	assert.Equal(t, "Invalid param", msg)

	// Test registering multiple error codes
	xerrorI18n.Register("en", 1002, "Unauthorized")
	xerrorI18n.Register("en", 1003, "Not found")

	msg1, found1 := xerrorI18n.Localize(1002, "en")
	msg2, found2 := xerrorI18n.Localize(1003, "en")
	assert.True(t, found1)
	assert.True(t, found2)
	assert.Equal(t, "Unauthorized", msg1)
	assert.Equal(t, "Not found", msg2)
}

func TestI18nForXerror_RegisterBatch(t *testing.T) {
	i18n := NewI18n()
	i18n.SetDefaultLang("en") // Set default language for fallback
	xerrorI18n := NewI18nForXerror(i18n)

	// Test batch registration
	enErrors := map[int32]string{
		1001: "Invalid parameter",
		1002: "Unauthorized",
		1003: "Not found",
		1004: "Conflict",
	}

	xerrorI18n.RegisterBatch("en", enErrors)

	// Verify all errors were registered
	for code, expectedMsg := range enErrors {
		msg, found := xerrorI18n.Localize(code, "en")
		assert.True(t, found, "Error code %d not found", code)
		assert.Equal(t, expectedMsg, msg, "Error code %d has wrong message", code)
	}

	// Test batch registration for another language
	zhErrors := map[int32]string{
		1001: "参数无效",
		1002: "未授权",
		1003: "未找到",
	}

	xerrorI18n.RegisterBatch("zh", zhErrors)

	// Verify Chinese translations
	for code, expectedMsg := range zhErrors {
		msg, found := xerrorI18n.Localize(code, "zh")
		assert.True(t, found, "Error code %d not found in Chinese", code)
		assert.Equal(t, expectedMsg, msg, "Error code %d has wrong Chinese message", code)
	}

	// Verify fallback for error code not in Chinese
	msg, found := xerrorI18n.Localize(1004, "zh")
	assert.True(t, found)
	assert.Equal(t, "Conflict", msg) // Should fallback to English
}

func TestI18nForXerror_RegisterWithCustomPrefix(t *testing.T) {
	i18n := NewI18n()
	xerrorI18n := NewI18nForXerror(i18n, "api.error.")

	// Register with custom prefix
	xerrorI18n.Register("en", 5001, "API Error")

	msg, found := xerrorI18n.Localize(5001, "en")
	assert.True(t, found)
	assert.Equal(t, "API Error", msg)

	// Verify the actual key used includes the prefix
	pack := i18n.packMap["en"]
	value, ok := pack.Get("api.error.5001")
	assert.True(t, ok)
	assert.Equal(t, "API Error", value)
}

func TestI18nForXerror_RegisterBatchWithCustomPrefix(t *testing.T) {
	i18n := NewI18n()
	xerrorI18n := NewI18nForXerror(i18n, "custom.")

	errors := map[int32]string{
		100: "Custom Error 1",
		200: "Custom Error 2",
	}

	xerrorI18n.RegisterBatch("en", errors)

	// Verify errors are accessible
	msg1, found1 := xerrorI18n.Localize(100, "en")
	msg2, found2 := xerrorI18n.Localize(200, "en")

	assert.True(t, found1)
	assert.True(t, found2)
	assert.Equal(t, "Custom Error 1", msg1)
	assert.Equal(t, "Custom Error 2", msg2)

	// Verify the actual keys include the prefix
	pack := i18n.packMap["en"]
	value1, ok1 := pack.Get("custom.100")
	value2, ok2 := pack.Get("custom.200")

	assert.True(t, ok1)
	assert.True(t, ok2)
	assert.Equal(t, "Custom Error 1", value1)
	assert.Equal(t, "Custom Error 2", value2)
}

func TestI18nForXerror_RegisterEmptyBatch(t *testing.T) {
	i18n := NewI18n()
	xerrorI18n := NewI18nForXerror(i18n)

	// Register empty batch (should not panic)
	xerrorI18n.RegisterBatch("en", map[int32]string{})

	// Verify pack was created even though no errors were added
	_, ok := i18n.packMap["en"]
	assert.True(t, ok)
}
