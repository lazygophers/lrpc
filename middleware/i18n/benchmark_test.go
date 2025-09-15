package i18n

import (
	"fmt"
	"testing"
)

func BenchmarkI18n_Localize(b *testing.B) {
	i18n := NewI18n()
	
	// Setup test data
	pack := NewPack("en")
	pack.corpus["hello"] = "Hello, World!"
	pack.corpus["greeting"] = "Hello, {{.Name}}"
	i18n.packMap["en"] = pack
	i18n.SetDefaultLang("en")
	
	b.Run("simple_localize", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = i18n.localize("en", "hello")
		}
	})
	
	b.Run("localize_with_fallback", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = i18n.localize("fr", "hello") // Will fallback to default
		}
	})
	
	b.Run("localize_missing_key", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = i18n.localize("en", "missing")
		}
	})
}

func BenchmarkI18n_LocalizeWithLang(b *testing.B) {
	i18n := NewI18n()
	
	// Setup test data
	pack := NewPack("en")
	pack.corpus["greeting"] = "Hello, {{.Name}}"
	pack.corpus["simple"] = "Simple message"
	i18n.packMap["en"] = pack
	i18n.SetDefaultLang("en")
	
	b.Run("simple_message", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = i18n.LocalizeWithLang("en", "simple")
		}
	})
	
	b.Run("template_message", func(b *testing.B) {
		data := map[string]string{"Name": "World"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = i18n.LocalizeWithLang("en", "greeting", data)
		}
	})
}

func BenchmarkLanguageCache(b *testing.B) {
	i18n := NewI18n()
	
	// Setup test data
	pack := NewPack("en")
	pack.corpus["test"] = "Test message"
	i18n.packMap["en"] = pack
	i18n.SetDefaultLang("en")
	
	b.Run("set_language", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			SetLanguage("en")
		}
	})
	
	b.Run("get_language", func(b *testing.B) {
		SetLanguage("en")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = GetLanguage()
		}
	})
	
	b.Run("localize_with_cache", func(b *testing.B) {
		SetLanguage("en")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Localize("test")
		}
	})
}

func BenchmarkParseLanguage(b *testing.B) {
	testCases := []string{
		"en",
		"en-US",
		"zh-CN",
		"en-US,zh;q=0.9",
		"fr-FR;q=0.8,en;q=0.7",
		"zh-CN.UTF-8",
	}
	
	for _, tc := range testCases {
		b.Run(fmt.Sprintf("parse_%s", tc), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = ParseLanguage(tc)
			}
		})
	}
}

func BenchmarkParseLangCode(b *testing.B) {
	testCases := []string{
		"en",
		"en-US",
		"zh-CN",
		"zh-hans",
		"zh-tw", 
		"fr-FR",
		"de-DE",
	}
	
	for _, tc := range testCases {
		b.Run(fmt.Sprintf("parse_%s", tc), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = ParseLangCode(tc)
			}
		})
	}
}

func BenchmarkLocalizer(b *testing.B) {
	jsonData := []byte(`{"hello": "world", "nested": {"key": "value"}}`)
	yamlData := []byte("hello: world\nnested:\n  key: value")
	tomlData := []byte(`hello = "world"
[nested]
key = "value"`)
	
	b.Run("json_unmarshal", func(b *testing.B) {
		localizer, _ := GetLocalizer("json")
		var result map[string]interface{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = localizer.Unmarshal(jsonData, &result)
		}
	})
	
	b.Run("yaml_unmarshal", func(b *testing.B) {
		localizer, _ := GetLocalizer("yaml")
		var result map[string]interface{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = localizer.Unmarshal(yamlData, &result)
		}
	})
	
	b.Run("toml_unmarshal", func(b *testing.B) {
		localizer, _ := GetLocalizer("toml")
		var result map[string]interface{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = localizer.Unmarshal(tomlData, &result)
		}
	})
}

func BenchmarkXerrorI18n(b *testing.B) {
	i18n := NewI18n()
	
	// Setup test data
	pack := NewPack("en")
	pack.corpus["error.404"] = "Not Found"
	pack.corpus["error.500"] = "Internal Server Error"
	i18n.packMap["en"] = pack
	i18n.SetDefaultLang("en")
	
	xerrorI18n := NewI18nForXerror(i18n)
	
	b.Run("localize_existing_error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = xerrorI18n.Localize(404, "en")
		}
	})
	
	b.Run("localize_missing_error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = xerrorI18n.Localize(999, "en")
		}
	})
}

func BenchmarkPack_Parse(b *testing.B) {
	testData := map[string]any{
		"simple": "value",
		"nested": map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"deep": map[string]interface{}{
				"key3": "value3",
			},
		},
		"numbers": map[string]interface{}{
			"int":    123,
			"int64":  int64(456789),
			"float":  3.14159,
		},
	}
	
	b.Run("parse_complex_data", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pack := NewPack("en")
			pack.parse(nil, testData)
		}
	})
}

func BenchmarkGetAcceptLanguage(b *testing.B) {
	DefaultI18n.SetDefaultLang("en")
	
	b.Run("no_params", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = GetAcceptLanguage()
		}
	})
	
	b.Run("with_language", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = GetAcceptLanguage("zh-CN")
		}
	})
	
	b.Run("with_multiple_languages", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = GetAcceptLanguage("zh-CN", "en", "fr")
		}
	})
}

// BenchmarkConcurrentAccess tests performance under concurrent access
func BenchmarkConcurrentAccess(b *testing.B) {
	i18n := NewI18n()
	
	// Setup test data
	pack := NewPack("en")
	pack.corpus["hello"] = "Hello, World!"
	i18n.packMap["en"] = pack
	i18n.SetDefaultLang("en")
	
	b.Run("concurrent_localize", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = i18n.localize("en", "hello")
			}
		})
	})
	
	b.Run("concurrent_cache_access", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				SetLanguage("en")
				_ = GetLanguage()
			}
		})
	})
}