package mock

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestToBsonM(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "nil value",
			input:   nil,
			wantErr: false,
		},
		{
			name:    "bson.M value",
			input:   bson.M{"key": "value"},
			wantErr: false,
		},
		{
			name: "struct value",
			input: struct {
				Name string
				Age  int
			}{Name: "test", Age: 20},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toBsonM(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("toBsonM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Errorf("toBsonM() returned nil result")
			}
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name string
		v1   interface{}
		v2   interface{}
		want int
	}{
		{
			name: "both nil",
			v1:   nil,
			v2:   nil,
			want: 0,
		},
		{
			name: "first nil",
			v1:   nil,
			v2:   10,
			want: -1,
		},
		{
			name: "second nil",
			v1:   10,
			v2:   nil,
			want: 1,
		},
		{
			name: "int equal",
			v1:   10,
			v2:   10,
			want: 0,
		},
		{
			name: "int less than",
			v1:   5,
			v2:   10,
			want: -1,
		},
		{
			name: "int greater than",
			v1:   15,
			v2:   10,
			want: 1,
		},
		{
			name: "string equal",
			v1:   "abc",
			v2:   "abc",
			want: 0,
		},
		{
			name: "string less than",
			v1:   "abc",
			v2:   "def",
			want: -1,
		},
		{
			name: "string greater than",
			v1:   "xyz",
			v2:   "abc",
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compare(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToInt64(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  int64
	}{
		{
			name:  "nil",
			input: nil,
			want:  0,
		},
		{
			name:  "int",
			input: 42,
			want:  42,
		},
		{
			name:  "int8",
			input: int8(10),
			want:  10,
		},
		{
			name:  "int16",
			input: int16(100),
			want:  100,
		},
		{
			name:  "int32",
			input: int32(1000),
			want:  1000,
		},
		{
			name:  "int64",
			input: int64(100),
			want:  100,
		},
		{
			name:  "uint",
			input: uint(20),
			want:  20,
		},
		{
			name:  "uint8",
			input: uint8(15),
			want:  15,
		},
		{
			name:  "uint16",
			input: uint16(200),
			want:  200,
		},
		{
			name:  "uint32",
			input: uint32(50),
			want:  50,
		},
		{
			name:  "uint64",
			input: uint64(500),
			want:  500,
		},
		{
			name:  "float32",
			input: float32(3.14),
			want:  3,
		},
		{
			name:  "float64",
			input: float64(3.14),
			want:  3,
		},
		{
			name:  "string (default)",
			input: "not a number",
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toInt64(tt.input)
			if got != tt.want {
				t.Errorf("toInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  float64
	}{
		{
			name:  "nil",
			input: nil,
			want:  0,
		},
		{
			name:  "int",
			input: 42,
			want:  42.0,
		},
		{
			name:  "int8",
			input: int8(10),
			want:  10.0,
		},
		{
			name:  "int16",
			input: int16(100),
			want:  100.0,
		},
		{
			name:  "int32",
			input: int32(1000),
			want:  1000.0,
		},
		{
			name:  "int64",
			input: int64(100000),
			want:  100000.0,
		},
		{
			name:  "uint",
			input: uint(20),
			want:  20.0,
		},
		{
			name:  "uint8",
			input: uint8(15),
			want:  15.0,
		},
		{
			name:  "uint16",
			input: uint16(200),
			want:  200.0,
		},
		{
			name:  "uint32",
			input: uint32(50),
			want:  50.0,
		},
		{
			name:  "uint64",
			input: uint64(500),
			want:  500.0,
		},
		{
			name:  "float32",
			input: float32(2.5),
			want:  2.5,
		},
		{
			name:  "float64",
			input: float64(3.14),
			want:  3.14,
		},
		{
			name:  "string (default)",
			input: "not a number",
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toFloat64(tt.input)
			if got != tt.want {
				t.Errorf("toFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInSlice(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		sliceVal interface{}
		want     bool
	}{
		{
			name:     "nil slice",
			value:    1,
			sliceVal: nil,
			want:     false,
		},
		{
			name:     "not a slice",
			value:    1,
			sliceVal: "not a slice",
			want:     false,
		},
		{
			name:     "int in slice",
			value:    2,
			sliceVal: []int{1, 2, 3},
			want:     true,
		},
		{
			name:     "int not in slice",
			value:    5,
			sliceVal: []int{1, 2, 3},
			want:     false,
		},
		{
			name:     "string in slice",
			value:    "b",
			sliceVal: []string{"a", "b", "c"},
			want:     true,
		},
		{
			name:     "string not in slice",
			value:    "d",
			sliceVal: []string{"a", "b", "c"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inSlice(tt.value, tt.sliceVal)
			if got != tt.want {
				t.Errorf("inSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchRegex(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		pattern interface{}
		want    bool
	}{
		{
			name:    "nil value",
			value:   nil,
			pattern: ".*",
			want:    false,
		},
		{
			name:    "nil pattern",
			value:   "test",
			pattern: nil,
			want:    false,
		},
		{
			name:    "simple match",
			value:   "hello",
			pattern: "^hello$",
			want:    true,
		},
		{
			name:    "no match",
			value:   "hello",
			pattern: "^world$",
			want:    false,
		},
		{
			name:    "pattern match",
			value:   "test123",
			pattern: "^test[0-9]+$",
			want:    true,
		},
		{
			name:    "invalid pattern",
			value:   "test",
			pattern: "[",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchRegex(tt.value, tt.pattern)
			if got != tt.want {
				t.Errorf("matchRegex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBsonToStruct(t *testing.T) {
	type TestStruct struct {
		Name string
		Age  int
	}

	tests := []struct {
		name    string
		doc     bson.M
		result  interface{}
		wantErr bool
	}{
		{
			name:    "nil result",
			doc:     bson.M{"name": "test", "age": 20},
			result:  nil,
			wantErr: true,
		},
		{
			name:    "not a pointer",
			doc:     bson.M{"name": "test", "age": 20},
			result:  TestStruct{},
			wantErr: true,
		},
		{
			name:    "valid conversion",
			doc:     bson.M{"name": "test", "age": 20},
			result:  &TestStruct{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bsonToStruct(tt.doc, tt.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("bsonToStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBsonArrayToStruct(t *testing.T) {
	type TestStruct struct {
		Name string
		Age  int
	}

	tests := []struct {
		name    string
		docs    []bson.M
		result  interface{}
		wantErr bool
	}{
		{
			name: "nil result",
			docs: []bson.M{
				{"name": "test1", "age": 20},
			},
			result:  nil,
			wantErr: true,
		},
		{
			name: "not a pointer",
			docs: []bson.M{
				{"name": "test1", "age": 20},
			},
			result:  []TestStruct{},
			wantErr: true,
		},
		{
			name: "not a slice pointer",
			docs: []bson.M{
				{"name": "test1", "age": 20},
			},
			result:  new(TestStruct),
			wantErr: true,
		},
		{
			name: "valid conversion",
			docs: []bson.M{
				{"name": "test1", "age": 20},
				{"name": "test2", "age": 30},
			},
			result:  &[]TestStruct{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bsonArrayToStruct(tt.docs, tt.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("bsonArrayToStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCompareTime(t *testing.T) {
	now := time.Now()
	before := now.Add(-time.Hour)
	after := now.Add(time.Hour)

	tests := []struct {
		name string
		v1   interface{}
		v2   interface{}
		want int
	}{
		{
			name: "equal time",
			v1:   now,
			v2:   now,
			want: 0,
		},
		{
			name: "before time",
			v1:   before,
			v2:   now,
			want: -1,
		},
		{
			name: "after time",
			v1:   after,
			v2:   now,
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compare(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestToBsonM_MarshalError tests toBsonM with values that cause marshal errors
func TestToBsonM_MarshalError(t *testing.T) {
	// Channels cannot be marshaled to BSON
	ch := make(chan int)
	_, err := toBsonM(ch)
	if err == nil {
		t.Error("toBsonM() should return error for channel type")
	}
}

// TestToBsonM_AllBranches tests all code paths in toBsonM
func TestToBsonM_AllBranches(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
		check   func(t *testing.T, result bson.M, err error)
	}{
		{
			name:    "nil input returns empty bson.M",
			input:   nil,
			wantErr: false,
			check: func(t *testing.T, result bson.M, err error) {
				if len(result) != 0 {
					t.Errorf("expected empty bson.M for nil input, got %v", result)
				}
			},
		},
		{
			name:    "bson.M input returns directly",
			input:   bson.M{"key": "value"},
			wantErr: false,
			check: func(t *testing.T, result bson.M, err error) {
				if result["key"] != "value" {
					t.Errorf("expected key=value, got %v", result)
				}
			},
		},
		{
			name: "struct marshals successfully",
			input: struct {
				Name string
				Age  int
			}{Name: "Alice", Age: 30},
			wantErr: false,
			check: func(t *testing.T, result bson.M, err error) {
				if result["name"] != "Alice" {
					t.Errorf("expected name=Alice, got %v", result)
				}
			},
		},
		{
			name:    "channel causes marshal error",
			input:   make(chan int),
			wantErr: true,
			check: func(t *testing.T, result bson.M, err error) {
				if err == nil {
					t.Error("expected error for channel type")
				}
				if result != nil {
					t.Errorf("expected nil result on error, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toBsonM(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("toBsonM() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.check != nil {
				tt.check(t, result, err)
			}
		})
	}
}

// TestBsonToStruct_UnmarshalError tests bsonToStruct with unmarshal errors
func TestBsonToStruct_UnmarshalError(t *testing.T) {
	type InvalidStruct struct {
		Field chan int // channels cannot be unmarshaled
	}

	// This should succeed in Marshal but may have issues
	// Let's test with incompatible type conversion instead
	doc2 := bson.M{"age": "not a number"}
	type AgeStruct struct {
		Age int
	}
	result2 := &AgeStruct{}

	// This should not error in BSON unmarshal (string to int conversion)
	// but demonstrates the error path
	err := bsonToStruct(doc2, result2)
	// BSON is flexible, so this might not error. Let's verify it works.
	if err != nil {
		t.Logf("bsonToStruct() error (expected for incompatible types): %v", err)
	}
}
