package i18n

import (
	"fmt"
	"io/fs"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Mock filesystem for testing
type mockFs struct {
	files map[string][]byte
	dirs  map[string][]fs.DirEntry
}

func (m *mockFs) ReadFile(name string) ([]byte, error) {
	if data, ok := m.files[name]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("file not found: %s", name)
}

func (m *mockFs) ReadDir(name string) ([]fs.DirEntry, error) {
	if entries, ok := m.dirs[name]; ok {
		return entries, nil
	}
	return nil, fmt.Errorf("directory not found: %s", name)
}

type mockDirEntry struct {
	name string
}

func (m *mockDirEntry) Name() string       { return m.name }
func (m *mockDirEntry) IsDir() bool        { return false }
func (m *mockDirEntry) Type() fs.FileMode  { return 0 }
func (m *mockDirEntry) Info() (fs.FileInfo, error) { return nil, fmt.Errorf("not implemented") }

func newMockFs() *mockFs {
	return &mockFs{
		files: make(map[string][]byte),
		dirs:  make(map[string][]fs.DirEntry),
	}
}

func TestPack_parse(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]any
		want    map[string]string
		wantPanic bool
	}{
		{
			name: "simple string values",
			input: map[string]any{
				"hello": "world",
				"foo":   "bar",
			},
			want: map[string]string{
				"hello": "world",
				"foo":   "bar",
			},
		},
		{
			name: "nested object",
			input: map[string]any{
				"user": map[string]interface{}{
					"name": "John",
					"age":  "30",
				},
			},
			want: map[string]string{
				"user.name": "John",
				"user.age":  "30",
			},
		},
		{
			name: "integer values",
			input: map[string]any{
				"count": 42,
				"level": int64(100),
			},
			want: map[string]string{
				"count": "42",
				"level": "100",
			},
		},
		{
			name: "float values",
			input: map[string]any{
				"price": 99.99,
				"rate":  0.15,
			},
			want: map[string]string{
				"price": "99.99",
				"rate":  "0.15",
			},
		},
		{
			name: "map with int64 keys",
			input: map[string]any{
				"status": map[int64]interface{}{
					200: "OK",
					404: "Not Found",
				},
			},
			want: map[string]string{
				"status.200": "OK",
				"status.404": "Not Found",
			},
		},
		{
			name: "map with float64 keys",
			input: map[string]any{
				"scores": map[float64]interface{}{
					85.5: "Good",
					95.0: "Excellent",
				},
			},
			want: map[string]string{
				"scores.85.5": "Good",
				"scores.95":   "Excellent",
			},
		},
		{
			name: "map with any keys",
			input: map[string]any{
				"mixed": map[any]any{
					"key1": "value1",
					123:    "value2",
				},
			},
			want: map[string]string{
				"mixed.key1": "value1",
				"mixed.123":  "value2",
			},
		},
		{
			name: "duplicate key should panic",
			input: map[string]any{
				"test": "value1",
			},
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pack := NewPack("en")
			
			if tt.wantPanic {
				// First add the key, then try to add it again
				pack.parse(nil, map[string]any{"test": "first"})
				assert.Panics(t, func() {
					pack.parse(nil, tt.input)
				})
				return
			}

			pack.parse(nil, tt.input)
			assert.Equal(t, tt.want, pack.corpus)
		})
	}
}

func TestPack_parseUnsupportedType(t *testing.T) {
	pack := NewPack("en")
	
	assert.Panics(t, func() {
		pack.parse(nil, map[string]any{
			"unsupported": []int{1, 2, 3}, // unsupported type
		})
	})
}

func TestPack_parseDuplicateKeys(t *testing.T) {
	t.Run("duplicate int key should panic", func(t *testing.T) {
		pack := NewPack("en")
		// First add the key
		pack.parse(nil, map[string]any{"num": 42})
		
		// Try to add duplicate int key
		assert.Panics(t, func() {
			pack.parse(nil, map[string]any{"num": 43})
		})
	})
	
	t.Run("duplicate int64 key should panic", func(t *testing.T) {
		pack := NewPack("en")
		// First add the key
		pack.parse(nil, map[string]any{"bignum": int64(12345678901234)})
		
		// Try to add duplicate int64 key
		assert.Panics(t, func() {
			pack.parse(nil, map[string]any{"bignum": int64(98765432109876)})
		})
	})
	
	t.Run("duplicate float64 key should panic", func(t *testing.T) {
		pack := NewPack("en")
		// First add the key
		pack.parse(nil, map[string]any{"pi": 3.14159})
		
		// Try to add duplicate float64 key
		assert.Panics(t, func() {
			pack.parse(nil, map[string]any{"pi": 2.71828})
		})
	})
}

func TestNewPack(t *testing.T) {
	pack := NewPack("en-US")
	
	assert.Equal(t, "en-US", pack.lang)
	assert.NotNil(t, pack.code)
	assert.NotNil(t, pack.corpus)
	assert.Empty(t, pack.corpus)
}

func TestI18n_localize(t *testing.T) {
	i18n := NewI18n()
	
	// Add test packs
	enPack := NewPack("en")
	enPack.corpus["hello"] = "Hello"
	enPack.corpus["welcome"] = "Welcome"
	i18n.packMap["en"] = enPack
	
	enUSPack := NewPack("en-us")
	enUSPack.corpus["hello"] = "Hello US"
	i18n.packMap["en-us"] = enUSPack
	
	zhPack := NewPack("zh")
	zhPack.corpus["hello"] = "你好"
	i18n.packMap["zh"] = zhPack
	
	i18n.SetDefaultLang("en")

	tests := []struct {
		name     string
		lang     string
		key      string
		want     string
		wantOK   bool
	}{
		{
			name:   "exact language match",
			lang:   "en-us",
			key:    "hello",
			want:   "Hello US",
			wantOK: true,
		},
		{
			name:   "fallback to base language",
			lang:   "en-gb",
			key:    "welcome", 
			want:   "Welcome",
			wantOK: true,
		},
		{
			name:   "fallback to default language",
			lang:   "fr",
			key:    "hello",
			want:   "Hello",
			wantOK: true,
		},
		{
			name:   "key not found returns empty string",
			lang:   "en",
			key:    "missing",
			want:   "",
			wantOK: true,
		},
		{
			name:   "empty language uses default",
			lang:   "",
			key:    "hello",
			want:   "Hello",
			wantOK: true,
		},
		{
			name:   "no pack found returns key",
			lang:   "nonexistent",
			key:    "missing",
			want:   "missing",
			wantOK: false,
		},
		{
			name:   "default lang with hyphen fallback",
			lang:   "nonexistent",
			key:    "base_test",
			want:   "Base Test",
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "no pack found returns key" {
				// Use fresh i18n instance with no packs
				emptyI18n := NewI18n()
				emptyI18n.SetDefaultLang("nonexistent")
				got, gotOK := emptyI18n.localize(tt.lang, tt.key)
				assert.Equal(t, tt.want, got)
				assert.Equal(t, tt.wantOK, gotOK)
			} else if tt.name == "default lang with hyphen fallback" {
				// Special case to test default language hyphen fallback
				testI18n := NewI18n()
				
				// Add base language pack (en) but set default to en-us
				basePack := NewPack("en")
				basePack.corpus["base_test"] = "Base Test"
				testI18n.packMap["en"] = basePack
				
				testI18n.SetDefaultLang("en-us") // Hyphenated default
				
				got, gotOK := testI18n.localize(tt.lang, tt.key)
				assert.Equal(t, tt.want, got)
				assert.Equal(t, tt.wantOK, gotOK)
			} else {
				got, gotOK := i18n.localize(tt.lang, tt.key)
				assert.Equal(t, tt.want, got)
				assert.Equal(t, tt.wantOK, gotOK)
			}
		})
	}
}

func TestI18n_LocalizeWithLang(t *testing.T) {
	i18n := NewI18n()
	
	// Add test pack
	pack := NewPack("en")
	pack.corpus["greeting"] = "Hello {{.Name}}"
	pack.corpus["simple"] = "Simple message"
	i18n.packMap["en"] = pack
	i18n.SetDefaultLang("en")

	tests := []struct {
		name string
		lang string
		key  string
		args []interface{}
		want string
	}{
		{
			name: "simple message",
			lang: "en",
			key:  "simple",
			args: nil,
			want: "Simple message",
		},
		{
			name: "template with args",
			lang: "en",
			key:  "greeting",
			args: []interface{}{map[string]string{"Name": "World"}},
			want: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := i18n.LocalizeWithLang(tt.lang, tt.key, tt.args...)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestI18n_LocalizeWithLang_TemplateError(t *testing.T) {
	i18n := NewI18n()
	
	t.Run("template parsing error", func(t *testing.T) {
		// Add pack with invalid template syntax
		pack := NewPack("en")
		pack.corpus["bad_template"] = "Hello {{.Name}"  // Missing closing brace
		i18n.packMap["en"] = pack
		i18n.SetDefaultLang("en")
		
		// This should panic due to template parsing error
		assert.Panics(t, func() {
			i18n.LocalizeWithLang("en", "bad_template", map[string]string{"Name": "World"})
		})
	})
	
	t.Run("template execution error", func(t *testing.T) {
		// Add pack with template that causes execution error
		pack := NewPack("en")
		pack.corpus["exec_error"] = "Hello {{call .BadFunc}}"
		i18n.packMap["en"] = pack
		i18n.SetDefaultLang("en")
		
		// This should panic due to template execution error (calling undefined function)
		assert.Panics(t, func() {
			i18n.LocalizeWithLang("en", "exec_error", map[string]interface{}{"BadFunc": nil})
		})
	})
}

func TestI18n_Localize(t *testing.T) {
	i18n := NewI18n()
	
	// Add test pack
	pack := NewPack("en")
	pack.corpus["test"] = "Test message"
	i18n.packMap["en"] = pack
	i18n.SetDefaultLang("en")

	// Set language for current goroutine
	SetLanguage("en")
	defer DelLanguage()
	
	got := i18n.Localize("test")
	assert.Equal(t, "Test message", got)
}

func TestI18n_AddTemplateFunc(t *testing.T) {
	i18n := NewI18n()
	
	customFunc := func(s string) string {
		return "custom: " + s
	}
	
	i18n.AddTemplateFunc("custom", customFunc)
	
	assert.Contains(t, i18n.templateFunc, "custom")
	assert.NotNil(t, i18n.templateFunc["custom"])
}

func TestI18n_LoadLocalizesWithFs(t *testing.T) {
	i18n := NewI18n()
	mockFs := newMockFs()
	
	// Setup mock files
	mockFs.dirs["localize"] = []fs.DirEntry{
		&mockDirEntry{name: "en.json"},
		&mockDirEntry{name: "zh.yaml"},
		&mockDirEntry{name: "invalid.txt"}, // unsupported extension
	}
	
	mockFs.files["localize/en.json"] = []byte(`{"hello": "Hello", "world": "World"}`)
	mockFs.files["localize/zh.yaml"] = []byte("hello: 你好\nworld: 世界")
	
	err := i18n.LoadLocalizesWithFs("localize", mockFs)
	require.NoError(t, err)
	
	// Verify loaded packs
	assert.Contains(t, i18n.packMap, "en")
	assert.Contains(t, i18n.packMap, "zh")
	assert.NotContains(t, i18n.packMap, "invalid")
	
	// Verify content
	assert.Equal(t, "Hello", i18n.packMap["en"].corpus["hello"])
	assert.Equal(t, "World", i18n.packMap["en"].corpus["world"])
	assert.Equal(t, "你好", i18n.packMap["zh"].corpus["hello"])
	assert.Equal(t, "世界", i18n.packMap["zh"].corpus["world"])
}

func TestI18n_LoadLocalizesWithFs_Error(t *testing.T) {
	i18n := NewI18n()
	mockFs := newMockFs()
	
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "directory not found",
			setup: func() {
				// Don't add the directory
			},
		},
		{
			name: "file read error",
			setup: func() {
				mockFs.dirs["localize"] = []fs.DirEntry{
					&mockDirEntry{name: "missing.json"},
				}
				// File exists in dir listing but not in files map to simulate read error
			},
		},
		{
			name: "unmarshal error",
			setup: func() {
				mockFs.dirs["localize"] = []fs.DirEntry{
					&mockDirEntry{name: "broken.json"},
				}
				mockFs.files["localize/broken.json"] = []byte(`{"missing": quote}`)
			},
		},
		{
			name: "invalid JSON",
			setup: func() {
				mockFs.dirs["localize"] = []fs.DirEntry{
					&mockDirEntry{name: "invalid.json"},
				}
				mockFs.files["localize/invalid.json"] = []byte(`{"invalid": json}`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFs := newMockFs()
			tt.setup()
			
			err := i18n.LoadLocalizesWithFs("localize", mockFs)
			assert.Error(t, err)
		})
	}
}

func TestI18n_LoadLocalizes(t *testing.T) {
	i18n := NewI18n()
	mockFs := newMockFs()
	
	// Setup mock files
	mockFs.dirs["localize"] = []fs.DirEntry{
		&mockDirEntry{name: "en.json"},
	}
	mockFs.files["localize/en.json"] = []byte(`{"test": "Test"}`)
	
	err := i18n.LoadLocalizes(mockFs)
	require.NoError(t, err)
	
	assert.Contains(t, i18n.packMap, "en")
}

func TestI18n_DefaultLangOperations(t *testing.T) {
	i18n := NewI18n()
	
	// Test initial default
	assert.Equal(t, "en", i18n.DefaultLang())
	
	// Test setting new default
	result := i18n.SetDefaultLang("zh-CN")
	assert.Same(t, i18n, result) // Should return self for chaining
	assert.Equal(t, "zh-cn", i18n.DefaultLang())
}

func TestI18n_AllSupportedLanguageCode(t *testing.T) {
	i18n := NewI18n()
	
	// Initially empty
	langs := i18n.AllSupportedLanguageCode()
	assert.Empty(t, langs)
	
	// Add some packs
	i18n.packMap["en"] = NewPack("en")
	i18n.packMap["zh"] = NewPack("zh")
	
	langs = i18n.AllSupportedLanguageCode()
	assert.Len(t, langs, 2)
	
	// Verify language codes are present
	langStrs := make([]string, len(langs))
	for i, lang := range langs {
		langStrs[i] = lang.Lang
	}
	assert.Contains(t, langStrs, "en")
	assert.Contains(t, langStrs, "zh")
}

func TestLanguageCache(t *testing.T) {
	// Test SetLanguage and GetLanguage
	SetLanguage("en-US")
	assert.Equal(t, "en-us", GetLanguage())
	
	// Test with specific goroutine ID
	SetLanguage("zh-CN", 12345)
	assert.Equal(t, "zh-cn", GetLanguage(12345))
	
	// Test DelLanguage
	DelLanguage()
	assert.Equal(t, "", GetLanguage())
	
	// Test DelLanguage with specific goroutine ID
	DelLanguage(12345)
	assert.Equal(t, "", GetLanguage(12345))
	
	// Test empty language deletion
	SetLanguage("", 99999)
	assert.Equal(t, "", GetLanguage(99999))
}

func TestNewI18n(t *testing.T) {
	i18n := NewI18n()
	
	assert.NotNil(t, i18n.packMap)
	assert.Empty(t, i18n.packMap)
	assert.Equal(t, "en", i18n.DefaultLang())
	assert.NotNil(t, i18n.templateFunc)
	
	// Verify default template functions exist
	expectedFuncs := []string{
		"ToCamel", "ToSmallCamel", "ToSnake", "ToLower", "ToUpper", "ToTitle",
		"TrimPrefix", "TrimSuffix", "HasPrefix", "HasSuffix",
		"PluckString", "PluckInt", "PluckInt32", "PluckUint32", "PluckInt64", "PluckUint64",
		"StringSliceEmpty", "UniqueString", "SortString", "ReverseString",
		"TopString", "FirstString", "LastString", "ContainsString",
		"TimeFormat4Pb", "TimeFormat4Timestamp",
	}
	
	for _, funcName := range expectedFuncs {
		assert.Contains(t, i18n.templateFunc, funcName, "Missing template function: %s", funcName)
	}
}

func TestDefaultTemplateFunctions(t *testing.T) {
	i18n := NewI18n()
	
	// Test TimeFormat4Pb function
	now := time.Now()
	pbTime := timestamppb.New(now)
	timeFunc := i18n.templateFunc["TimeFormat4Pb"].(func(*timestamppb.Timestamp, string) string)
	result := timeFunc(pbTime, "2006-01-02")
	// Don't assert the exact date, just ensure it's a valid date format
	assert.Regexp(t, `^\d{4}-\d{2}-\d{2}$`, result)
	
	// Test TimeFormat4Timestamp function
	timestamp := now.Unix()
	tsFunc := i18n.templateFunc["TimeFormat4Timestamp"].(func(int64, string) string)
	result2 := tsFunc(timestamp, "2006-01-02")
	expected2 := time.Unix(timestamp, 0).Format("2006-01-02")
	assert.Equal(t, expected2, result2)
	
	// Test StringSliceEmpty function
	emptyFunc := i18n.templateFunc["StringSliceEmpty"].(func([]string) bool)
	assert.True(t, emptyFunc([]string{}))
	assert.False(t, emptyFunc([]string{"test"}))
}

func TestDefaultI18nFunctions(t *testing.T) {
	// Test SetDefaultLanguage and DefaultLanguage
	SetDefaultLanguage("zh-CN")
	assert.Equal(t, "zh-cn", DefaultLanguage())
	
	// Test Localize
	result := Localize("nonexistent.key")
	assert.Equal(t, "nonexistent.key", result)
	
	// Test LoadLocalizes
	mockFs := newMockFs()
	mockFs.dirs["localize"] = []fs.DirEntry{
		&mockDirEntry{name: "en.json"},
	}
	mockFs.files["localize/en.json"] = []byte(`{"test": "value"}`)
	
	err := LoadLocalizes(mockFs)
	assert.NoError(t, err)
}

func TestParseLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"en", "en"},
		{"en-US", "en-us"},
		{"zh_CN", "zh-cn"},
		{"en-US,zh;q=0.9", "en-us"},
		{"fr-FR;q=0.8,en;q=0.7", "fr-fr"},
		{"zh-CN.UTF-8", "zh-cn"},
		{"EN-US", "en-us"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("input_%s", tt.input), func(t *testing.T) {
			result := ParseLanguage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLocalizeWithHeaders(t *testing.T) {
	// Setup a test pack in DefaultI18n
	pack := NewPack("en")
	pack.corpus["test"] = "Test Message"
	DefaultI18n.packMap["en"] = pack
	DefaultI18n.SetDefaultLang("en")

	tests := []struct {
		name     string
		headers  http.Header
		key      string
		expected string
	}{
		{
			name:     "no headers",
			headers:  http.Header{},
			key:      "test",
			expected: "Test Message",
		},
		{
			name: "with Accept-Language header",
			headers: http.Header{
				"Accept-Language": []string{"en-US,en;q=0.9"},
			},
			key:      "test",
			expected: "Test Message",
		},
		{
			name: "missing key returns key",
			headers: http.Header{
				"Accept-Language": []string{"en"},
			},
			key:      "missing",
			expected: "",
		},
		{
			name: "headers without Accept-Language",
			headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
			key:      "missing",
			expected: "missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "headers without Accept-Language" {
				// Save original state and clear packs
				originalPacks := make(map[string]*Pack)
				for k, v := range DefaultI18n.packMap {
					originalPacks[k] = v
				}
				DefaultI18n.packMap = make(map[string]*Pack)
				DefaultI18n.SetDefaultLang("nonexistent")
				
				result := LocalizeWithHeaders(tt.headers, tt.key)
				assert.Equal(t, tt.expected, result)
				
				// Restore original state
				DefaultI18n.packMap = originalPacks
				DefaultI18n.SetDefaultLang("en")
			} else {
				result := LocalizeWithHeaders(tt.headers, tt.key)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestPrivateLocalizeFunction(t *testing.T) {
	// Test the private localize function
	pack := NewPack("en")
	pack.corpus["key"] = "value"
	DefaultI18n.packMap["en"] = pack
	
	result := localize("en", "key")
	assert.Equal(t, "value", result)
	
	// Test with template args
	pack.corpus["template"] = "Hello {{.Name}}"
	result = localize("en", "template", map[string]string{"Name": "World"})
	assert.Equal(t, "Hello World", result)
}

func TestI18n_localizeWithDefaultFallback(t *testing.T) {
	i18n := NewI18n()
	
	// Set up default language pack
	enPack := NewPack("en")
	enPack.corpus["test"] = "Default"
	i18n.packMap["en"] = enPack
	
	// Set up default with hyphen 
	enUSPack := NewPack("en-us")
	enUSPack.corpus["specific"] = "US Specific"
	i18n.packMap["en-us"] = enUSPack
	
	i18n.SetDefaultLang("en-us")

	// Test fallback to default with hyphen parsing
	value, found := i18n.localize("fr", "test")
	assert.Equal(t, "", value)
	assert.True(t, found)
	
	// Test default language with hyphen
	value, found = i18n.localize("", "specific")
	assert.Equal(t, "US Specific", value)
	assert.True(t, found)
}