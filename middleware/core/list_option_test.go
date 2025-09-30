package core

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewListOption(t *testing.T) {
	opt := NewListOption()
	assert.Equal(t, uint64(0), opt.Offset)
	assert.Equal(t, uint64(20), opt.Limit)
	assert.False(t, opt.ShowTotal)
	assert.Nil(t, opt.Options)
}

func TestListOption_SetOffset(t *testing.T) {
	opt := NewListOption().SetOffset(100)
	assert.Equal(t, uint64(100), opt.Offset)
}

func TestListOption_SetLimit(t *testing.T) {
	opt := NewListOption().SetLimit(50)
	assert.Equal(t, uint64(50), opt.Limit)
}

func TestListOption_SetShowTotal(t *testing.T) {
	t.Run("with true", func(t *testing.T) {
		opt := NewListOption().SetShowTotal(true)
		assert.True(t, opt.ShowTotal)
	})

	t.Run("with false", func(t *testing.T) {
		opt := NewListOption().SetShowTotal(false)
		assert.False(t, opt.ShowTotal)
	})

	t.Run("without argument defaults to true", func(t *testing.T) {
		opt := NewListOption().SetShowTotal()
		assert.True(t, opt.ShowTotal)
	})
}

func TestListOption_SetOptions(t *testing.T) {
	opt1 := &ListOption_Option{Key: 1, Value: "a"}
	opt2 := &ListOption_Option{Key: 2, Value: "b"}

	opt := NewListOption().SetOptions(opt1, opt2)
	assert.Len(t, opt.Options, 2)
	assert.Equal(t, opt1, opt.Options[0])
	assert.Equal(t, opt2, opt.Options[1])
}

func TestListOption_AddOptions(t *testing.T) {
	opt1 := &ListOption_Option{Key: 1, Value: "a"}
	opt2 := &ListOption_Option{Key: 2, Value: "b"}

	opt := NewListOption().AddOptions(opt1).AddOptions(opt2)
	assert.Len(t, opt.Options, 2)
}

func TestListOption_AddOption(t *testing.T) {
	opt := NewListOption().AddOption(1, "test")
	assert.Len(t, opt.Options, 1)
	assert.Equal(t, int32(1), opt.Options[0].Key)
	assert.Equal(t, "test", opt.Options[0].Value)
}

func TestListOption_Clone(t *testing.T) {
	t.Run("clone with values", func(t *testing.T) {
		original := NewListOption().
			SetOffset(100).
			SetLimit(50).
			SetShowTotal(true).
			AddOption(1, "test")

		cloned := original.Clone()

		assert.Equal(t, original.Offset, cloned.Offset)
		assert.Equal(t, original.Limit, cloned.Limit)
		assert.Equal(t, original.ShowTotal, cloned.ShowTotal)
		assert.Len(t, cloned.Options, 1)

		// Verify deep copy - modifying cloned shouldn't affect original
		cloned.Offset = 200
		cloned.Options[0].Value = "modified"

		assert.Equal(t, uint64(100), original.Offset)
		assert.Equal(t, "test", original.Options[0].Value)
	})

	t.Run("clone nil returns default", func(t *testing.T) {
		var opt *ListOption
		cloned := opt.Clone()

		assert.Equal(t, uint64(0), cloned.Offset)
		assert.Equal(t, uint64(20), cloned.Limit)
	})
}

func TestListOption_Paginate(t *testing.T) {
	t.Run("normal case", func(t *testing.T) {
		opt := NewListOption().SetOffset(100).SetLimit(50)
		paginate := opt.Paginate()

		assert.Equal(t, uint64(100), paginate.Offset)
		assert.Equal(t, uint64(50), paginate.Limit)
	})

	t.Run("nil returns default", func(t *testing.T) {
		var opt *ListOption
		paginate := opt.Paginate()

		assert.Equal(t, uint64(0), paginate.Offset)
		assert.Equal(t, uint64(20), paginate.Limit)
	})
}

func TestListOptionProcessor_String(t *testing.T) {
	opt := NewListOption().AddOption(1, "test")
	processor := opt.Processor()

	called := false
	var receivedValue string

	err := processor.String(1, func(value string) error {
		called = true
		receivedValue = value
		return nil
	}).Process()

	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "test", receivedValue)
}

func TestListOptionProcessor_Int(t *testing.T) {
	opt := NewListOption().AddOption(1, "42")
	processor := opt.Processor()

	var receivedValue int
	err := processor.Int(1, func(value int) error {
		receivedValue = value
		return nil
	}).Process()

	assert.NoError(t, err)
	assert.Equal(t, 42, receivedValue)
}

func TestListOptionProcessor_Bool(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"TRUE", true},
		{"1", true},
		{"yes", true},
		{"on", true},
		{"enable", true},
		{"false", false},
		{"FALSE", false},
		{"0", false},
		{"no", false},
		{"off", false},
		{"disable", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			opt := NewListOption().AddOption(1, tc.input)
			processor := opt.Processor()

			var receivedValue bool
			err := processor.Bool(1, func(value bool) error {
				receivedValue = value
				return nil
			}).Process()

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, receivedValue)
		})
	}

	t.Run("invalid value", func(t *testing.T) {
		opt := NewListOption().AddOption(1, "invalid")
		processor := opt.Processor()

		err := processor.Bool(1, func(value bool) error {
			return nil
		}).Process()

		assert.Error(t, err)
	})
}

func TestListOptionProcessor_StringSlice(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "normal",
			input:    "a,b,c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with spaces",
			input:    " a , b , c ",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with empty values",
			input:    "a,,b,,,c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opt := NewListOption().AddOption(1, tc.input)
			processor := opt.Processor()

			var receivedValue []string
			err := processor.StringSlice(1, func(value []string) error {
				receivedValue = value
				return nil
			}).Process()

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, receivedValue)
		})
	}
}

func TestListOptionProcessor_IntSlice(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []int
		hasError bool
	}{
		{
			name:     "normal",
			input:    "1,2,3",
			expected: []int{1, 2, 3},
		},
		{
			name:     "with spaces",
			input:    " 1 , 2 , 3 ",
			expected: []int{1, 2, 3},
		},
		{
			name:     "with empty values",
			input:    "1,,2,,,3",
			expected: []int{1, 2, 3},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []int{},
		},
		{
			name:     "invalid value",
			input:    "1,abc,3",
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opt := NewListOption().AddOption(1, tc.input)
			processor := opt.Processor()

			var receivedValue []int
			err := processor.IntSlice(1, func(value []int) error {
				receivedValue = value
				return nil
			}).Process()

			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, receivedValue)
			}
		})
	}
}

func TestListOptionProcessor_Order(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool // true for desc, false for asc
	}{
		{"desc", true},
		{"DESC", true},
		{"descend", true},
		{"descending", true},
		{"asc", false},
		{"ASC", false},
		{"ascend", false},
		{"ascending", false},
		{"", false},
		{"  ", false},
		{"invalid", false}, // defaults to asc
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			opt := NewListOption().AddOption(1, tc.input)
			processor := opt.Processor()

			var receivedValue bool
			err := processor.Order(1, func(isDesc bool) error {
				receivedValue = isDesc
				return nil
			}).Process()

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, receivedValue)
		})
	}
}

func TestListOptionProcessor_Has(t *testing.T) {
	opt := NewListOption().AddOption(1, "any_value")
	processor := opt.Processor()

	called := false
	err := processor.Has(1, func() error {
		called = true
		return nil
	}).Process()

	assert.NoError(t, err)
	assert.True(t, called)
}

func TestListOptionProcessor_Process(t *testing.T) {
	t.Run("adjusts negative offset", func(t *testing.T) {
		opt := &ListOption{Offset: ^uint64(0), Limit: 10} // max uint64
		processor := NewListOptionProcessor(opt)

		err := processor.Process()
		assert.NoError(t, err)
		// Offset should remain unchanged as it's unsigned
	})

	t.Run("adjusts negative limit", func(t *testing.T) {
		opt := &ListOption{Offset: 0, Limit: ^uint64(0)}
		processor := NewListOptionProcessor(opt)

		err := processor.Process()
		assert.NoError(t, err)
		// Limit > maxLimit should be capped
		assert.Equal(t, uint64(maxLimit), processor.ListOption.Limit)
	})

	t.Run("caps limit to max", func(t *testing.T) {
		opt := &ListOption{Offset: 0, Limit: 2000}
		processor := NewListOptionProcessor(opt)

		err := processor.Process()
		assert.NoError(t, err)
		assert.Equal(t, uint64(1000), processor.ListOption.Limit)
	})

	t.Run("error handling", func(t *testing.T) {
		opt := NewListOption().AddOption(1, "test")
		processor := opt.Processor()

		testErr := errors.New("test error")
		err := processor.String(1, func(value string) error {
			return testErr
		}).Process()

		assert.Error(t, err)
		assert.Equal(t, testErr, err)
	})

	t.Run("unregistered option is ignored", func(t *testing.T) {
		opt := NewListOption().AddOption(999, "test")
		processor := opt.Processor()

		err := processor.Process()
		assert.NoError(t, err)
	})
}

func TestListOptionProcessor_ChainedCalls(t *testing.T) {
	opt := NewListOption().
		AddOption(1, "test").
		AddOption(2, "42").
		AddOption(3, "true")

	processor := opt.Processor()

	var strValue string
	var intValue int
	var boolValue bool

	err := processor.
		String(1, func(value string) error {
			strValue = value
			return nil
		}).
		Int(2, func(value int) error {
			intValue = value
			return nil
		}).
		Bool(3, func(value bool) error {
			boolValue = value
			return nil
		}).
		Process()

	assert.NoError(t, err)
	assert.Equal(t, "test", strValue)
	assert.Equal(t, 42, intValue)
	assert.True(t, boolValue)
}

func Test_splitAndTrim(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "normal",
			input:    "a,b,c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with spaces",
			input:    " a , b , c ",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with empty values",
			input:    "a,,b,,,c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: []string{},
		},
		{
			name:     "single value",
			input:    "a",
			expected: []string{"a"},
		},
		{
			name:     "single value with spaces",
			input:    "  a  ",
			expected: []string{"a"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := splitAndTrim(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}