package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocalizerHandle(t *testing.T) {
	unmarshalFunc := func(body []byte, v any) error {
		return nil
	}
	
	handle := NewLocalizerHandle(unmarshalFunc)
	
	assert.NotNil(t, handle)
	assert.NotNil(t, handle.unmarshal)
}

func TestLocalizerHandle_Unmarshal(t *testing.T) {
	// Test successful unmarshal
	unmarshalFunc := func(body []byte, v any) error {
		if m, ok := v.(*map[string]interface{}); ok {
			*m = map[string]interface{}{"test": "value"}
		}
		return nil
	}
	
	handle := NewLocalizerHandle(unmarshalFunc)
	
	var result map[string]interface{}
	err := handle.Unmarshal([]byte("{}"), &result)
	
	assert.NoError(t, err)
	assert.Equal(t, "value", result["test"])
}

func TestLocalizerHandle_UnmarshalError(t *testing.T) {
	// Test error case
	expectedErr := assert.AnError
	unmarshalFunc := func(body []byte, v any) error {
		return expectedErr
	}
	
	handle := NewLocalizerHandle(unmarshalFunc)
	
	var result map[string]interface{}
	err := handle.Unmarshal([]byte("{}"), &result)
	
	assert.Equal(t, expectedErr, err)
}

func TestGetLocalizer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "json localizer",
			input:    "json",
			expected: true,
		},
		{
			name:     "yaml localizer",
			input:    "yaml",
			expected: true,
		},
		{
			name:     "yml localizer",
			input:    "yml",
			expected: true,
		},
		{
			name:     "toml localizer",
			input:    "toml",
			expected: true,
		},
		{
			name:     "with dot prefix",
			input:    ".json",
			expected: true,
		},
		{
			name:     "unsupported format",
			input:    "xml",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localizer, found := GetLocalizer(tt.input)
			
			assert.Equal(t, tt.expected, found)
			if tt.expected {
				assert.NotNil(t, localizer)
			} else {
				assert.Nil(t, localizer)
			}
		})
	}
}

func TestRegisterLocalizer(t *testing.T) {
	// Create a custom localizer
	customLocalizer := NewLocalizerHandle(func(body []byte, v any) error {
		return nil
	})
	
	// Register it
	RegisterLocalizer("custom", customLocalizer)
	
	// Verify it's registered
	retrieved, found := GetLocalizer("custom")
	assert.True(t, found)
	assert.Equal(t, customLocalizer, retrieved)
	
	// Test overriding existing localizer
	newCustomLocalizer := NewLocalizerHandle(func(body []byte, v any) error {
		return assert.AnError
	})
	
	RegisterLocalizer("custom", newCustomLocalizer)
	
	// Verify it's updated
	retrieved, found = GetLocalizer("custom")
	assert.True(t, found)
	assert.Equal(t, newCustomLocalizer, retrieved)
}

func TestBuiltinLocalizers(t *testing.T) {
	tests := []struct {
		name string
		ext  string
		data string
	}{
		{
			name: "json localizer",
			ext:  "json",
			data: `{"hello": "world", "nested": {"key": "value"}}`,
		},
		{
			name: "yaml localizer", 
			ext:  "yaml",
			data: "hello: world\nnested:\n  key: value",
		},
		{
			name: "yml localizer",
			ext:  "yml", 
			data: "hello: world\nnested:\n  key: value",
		},
		{
			name: "toml localizer",
			ext:  "toml",
			data: `hello = "world"
[nested]
key = "value"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localizer, found := GetLocalizer(tt.ext)
			require.True(t, found)
			require.NotNil(t, localizer)
			
			var result map[string]interface{}
			err := localizer.Unmarshal([]byte(tt.data), &result)
			require.NoError(t, err)
			
			// Verify basic structure
			assert.Equal(t, "world", result["hello"])
			
			// Verify nested structure
			if nested, ok := result["nested"]; ok {
				if nestedMap, ok := nested.(map[string]interface{}); ok {
					assert.Equal(t, "value", nestedMap["key"])
				}
			}
		})
	}
}

func TestBuiltinLocalizers_ErrorCases(t *testing.T) {
	tests := []struct {
		name string
		ext  string
		data string
	}{
		{
			name: "invalid json",
			ext:  "json",
			data: `{"invalid": json}`,
		},
		{
			name: "invalid yaml",
			ext:  "yaml",
			data: "invalid:\n\tyaml:\nstructure",
		},
		{
			name: "invalid toml",
			ext:  "toml",
			data: "[invalid\ntoml = structure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localizer, found := GetLocalizer(tt.ext)
			require.True(t, found)
			require.NotNil(t, localizer)
			
			var result map[string]interface{}
			err := localizer.Unmarshal([]byte(tt.data), &result)
			assert.Error(t, err)
		})
	}
}

func TestLocalizerInterface(t *testing.T) {
	// Verify that all built-in localizers implement the interface
	builtinLocalizers := []string{"json", "yaml", "yml", "toml"}
	
	for _, name := range builtinLocalizers {
		t.Run(name, func(t *testing.T) {
			l, found := GetLocalizer(name)
			require.True(t, found)
			
			// Verify it implements the interface
			var _ Localizer = l
			
			// Test with empty data
			var result map[string]interface{}
			err := l.Unmarshal([]byte("{}"), &result)
			
			// Should not panic and should handle empty data gracefully
			// (Error or success is acceptable depending on the format)
			_ = err
		})
	}
}

func TestLocalizer_ConcreteTypes(t *testing.T) {
	// Verify that the built-in localizers are the expected concrete types
	jsonLoc, _ := GetLocalizer("json")
	yamlLoc, _ := GetLocalizer("yaml")
	ymlLoc, _ := GetLocalizer("yml")
	tomlLoc, _ := GetLocalizer("toml")
	
	// Should all be LocalizerHandle instances
	assert.IsType(t, &LocalizerHandle{}, jsonLoc)
	assert.IsType(t, &LocalizerHandle{}, yamlLoc)
	assert.IsType(t, &LocalizerHandle{}, ymlLoc)
	assert.IsType(t, &LocalizerHandle{}, tomlLoc)
	
	// yaml and yml should be the same instance
	assert.Same(t, yamlLoc, ymlLoc)
}

func TestGetLocalizer_DotPrefix(t *testing.T) {
	// Test various dot prefix scenarios
	cases := []struct {
		input       string
		expected    string
		shouldExist bool
	}{
		{".json", "json", true},
		{".yaml", "yaml", true},
		{".yml", "yml", true},
		{".toml", "toml", true},
		{"..json", ".json", false}, // Only removes first dot, results in invalid localizer
		{".custom", "custom", true},
	}
	
	// First register a custom localizer to test dot prefix behavior
	customLocalizer := NewLocalizerHandle(func(body []byte, v any) error {
		return nil
	})
	RegisterLocalizer("custom", customLocalizer)
	
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			result, found := GetLocalizer(tc.input)
			
			assert.Equal(t, tc.shouldExist, found)
			if tc.shouldExist {
				// Compare with getting the localizer without dot
				expected, expectedFound := GetLocalizer(tc.expected)
				assert.True(t, expectedFound)
				assert.Equal(t, expected, result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}