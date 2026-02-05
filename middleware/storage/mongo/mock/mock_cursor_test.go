package mock

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

// TestMockCursor_All tests the All method
func TestMockCursor_All(t *testing.T) {
	type TestDoc struct {
		Name string
		Age  int
	}

	tests := []struct {
		name      string
		documents []bson.M
		target    interface{}
		wantErr   bool
		wantCount int
		setup     func(*MockCursor)
	}{
		{
			name: "get all documents from start",
			documents: []bson.M{
				{"name": "Alice", "age": 30},
				{"name": "Bob", "age": 25},
				{"name": "Charlie", "age": 35},
			},
			target:    &[]TestDoc{},
			wantErr:   false,
			wantCount: 3,
		},
		{
			name: "get all documents after advancing cursor",
			documents: []bson.M{
				{"name": "Alice", "age": 30},
				{"name": "Bob", "age": 25},
				{"name": "Charlie", "age": 35},
			},
			target:    &[]TestDoc{},
			wantErr:   false,
			wantCount: 2,
			setup: func(c *MockCursor) {
				ctx := context.Background()
				c.Next(ctx)
			},
		},
		{
			name:      "empty documents",
			documents: []bson.M{},
			target:    &[]TestDoc{},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "error when cursor is closed",
			documents: []bson.M{
				{"name": "Alice", "age": 30},
			},
			target:  &[]TestDoc{},
			wantErr: true,
			setup: func(c *MockCursor) {
				c.closed = true
			},
		},
		{
			name: "error when target is nil",
			documents: []bson.M{
				{"name": "Alice", "age": 30},
			},
			target:  nil,
			wantErr: true,
		},
		{
			name: "get all to bson.M slice",
			documents: []bson.M{
				{"name": "Alice", "age": 30},
				{"name": "Bob", "age": 25},
			},
			target:    &[]bson.M{},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "single document",
			documents: []bson.M{
				{"name": "Alice", "age": 30},
			},
			target:    &[]TestDoc{},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "error when target is not a pointer to slice",
			documents: []bson.M{
				{"name": "Alice", "age": 30},
			},
			target:  new(TestDoc), // pointer to struct, not slice
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := NewMockCursor(tt.documents)
			if tt.setup != nil {
				tt.setup(cursor)
			}

			ctx := context.Background()
			err := cursor.All(ctx, tt.target)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify count
				if docs, ok := tt.target.(*[]TestDoc); ok {
					assert.Equal(t, tt.wantCount, len(*docs))
					// Verify first document if exists
					if tt.wantCount > 0 {
						assert.NotEmpty(t, (*docs)[0].Name)
					}
				} else if docs, ok := tt.target.(*[]bson.M); ok {
					assert.Equal(t, tt.wantCount, len(*docs))
				}

				// Verify cursor is closed after All
				assert.True(t, cursor.closed)
			}
		})
	}
}

// TestMockCursor_Decode tests the Decode method
func TestMockCursor_Decode(t *testing.T) {
	type TestDoc struct {
		Name string
		Age  int
	}

	tests := []struct {
		name      string
		documents []bson.M
		position  int
		target    interface{}
		wantErr   bool
		setup     func(*MockCursor)
	}{
		{
			name: "decode current document successfully",
			documents: []bson.M{
				{"name": "Alice", "age": 30},
				{"name": "Bob", "age": 25},
			},
			position: 0,
			target:   &TestDoc{},
			wantErr:  false,
			setup: func(c *MockCursor) {
				c.position = 0
			},
		},
		{
			name: "decode second document successfully",
			documents: []bson.M{
				{"name": "Alice", "age": 30},
				{"name": "Bob", "age": 25},
			},
			position: 1,
			target:   &TestDoc{},
			wantErr:  false,
			setup: func(c *MockCursor) {
				c.position = 1
			},
		},
		{
			name: "error when cursor is closed",
			documents: []bson.M{
				{"name": "Alice", "age": 30},
			},
			target:  &TestDoc{},
			wantErr: true,
			setup: func(c *MockCursor) {
				c.closed = true
			},
		},
		{
			name: "error when target is nil",
			documents: []bson.M{
				{"name": "Alice", "age": 30},
			},
			target:  nil,
			wantErr: true,
			setup: func(c *MockCursor) {
				c.position = 0
			},
		},
		{
			name: "error when position is invalid (negative)",
			documents: []bson.M{
				{"name": "Alice", "age": 30},
			},
			target:  &TestDoc{},
			wantErr: true,
			setup: func(c *MockCursor) {
				c.position = -1
			},
		},
		{
			name: "error when position is out of bounds",
			documents: []bson.M{
				{"name": "Alice", "age": 30},
			},
			target:  &TestDoc{},
			wantErr: true,
			setup: func(c *MockCursor) {
				c.position = 10
			},
		},
		{
			name: "decode to bson.M",
			documents: []bson.M{
				{"name": "Alice", "age": 30},
			},
			target:  &bson.M{},
			wantErr: false,
			setup: func(c *MockCursor) {
				c.position = 0
			},
		},
		{
			name: "decode with complex nested data",
			documents: []bson.M{
				{
					"name": "Alice",
					"age":  30,
					"address": bson.M{
						"city":    "NYC",
						"country": "USA",
					},
					"tags": []string{"admin", "user"},
				},
			},
			target: &struct {
				Name    string
				Age     int
				Address struct {
					City    string
					Country string
				}
				Tags []string
			}{},
			wantErr: false,
			setup: func(c *MockCursor) {
				c.position = 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := NewMockCursor(tt.documents)
			if tt.setup != nil {
				tt.setup(cursor)
			}

			err := cursor.Decode(tt.target)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify decoded content
				if doc, ok := tt.target.(*TestDoc); ok {
					assert.NotEmpty(t, doc.Name)
				}
			}
		})
	}
}

// TestMockCursor_Err tests the Err method
func TestMockCursor_Err(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*MockCursor)
		wantErr  bool
		checkErr error
	}{
		{
			name: "no error initially",
			setup: func(c *MockCursor) {
				// Default state
			},
			wantErr: false,
		},
		{
			name: "returns configured error",
			setup: func(c *MockCursor) {
				c.err = ErrInvalidArgument
			},
			wantErr:  true,
			checkErr: ErrInvalidArgument,
		},
		{
			name: "returns nil after clearing error",
			setup: func(c *MockCursor) {
				c.err = ErrInvalidArgument
				c.err = nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := NewMockCursor([]bson.M{})
			if tt.setup != nil {
				tt.setup(cursor)
			}

			err := cursor.Err()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.checkErr != nil {
					assert.Equal(t, tt.checkErr, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestMockCursor_ID tests the ID method
func TestMockCursor_ID(t *testing.T) {
	tests := []struct {
		name      string
		documents []bson.M
		want      int64
	}{
		{
			name:      "empty cursor returns 0",
			documents: []bson.M{},
			want:      0,
		},
		{
			name: "cursor with documents returns 0",
			documents: []bson.M{
				{"name": "Alice"},
			},
			want: 0,
		},
		{
			name: "closed cursor returns 0",
			documents: []bson.M{
				{"name": "Alice"},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := NewMockCursor(tt.documents)
			id := cursor.ID()
			assert.Equal(t, tt.want, id)
		})
	}
}

// TestMockCursor_Next tests the Next method
func TestMockCursor_Next(t *testing.T) {
	tests := []struct {
		name       string
		documents  []bson.M
		iterations int
		wantTrue   int // how many iterations should return true
		setup      func(*MockCursor)
	}{
		{
			name:       "empty cursor",
			documents:  []bson.M{},
			iterations: 1,
			wantTrue:   0,
		},
		{
			name: "single document",
			documents: []bson.M{
				{"name": "Alice"},
			},
			iterations: 2,
			wantTrue:   1,
		},
		{
			name: "multiple documents",
			documents: []bson.M{
				{"name": "Alice"},
				{"name": "Bob"},
				{"name": "Charlie"},
			},
			iterations: 5,
			wantTrue:   3,
		},
		{
			name: "closed cursor returns false",
			documents: []bson.M{
				{"name": "Alice"},
			},
			iterations: 1,
			wantTrue:   0,
			setup: func(c *MockCursor) {
				c.closed = true
			},
		},
		{
			name: "cursor with error returns false",
			documents: []bson.M{
				{"name": "Alice"},
			},
			iterations: 1,
			wantTrue:   0,
			setup: func(c *MockCursor) {
				c.err = ErrInvalidArgument
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := NewMockCursor(tt.documents)
			if tt.setup != nil {
				tt.setup(cursor)
			}

			ctx := context.Background()
			trueCount := 0
			for i := 0; i < tt.iterations; i++ {
				if cursor.Next(ctx) {
					trueCount++
				}
			}

			assert.Equal(t, tt.wantTrue, trueCount)
		})
	}
}

// TestMockCursor_NextAndDecode tests Next and Decode together
func TestMockCursor_NextAndDecode(t *testing.T) {
	type TestDoc struct {
		Name string
		Age  int
	}

	documents := []bson.M{
		{"name": "Alice", "age": 30},
		{"name": "Bob", "age": 25},
		{"name": "Charlie", "age": 35},
	}

	cursor := NewMockCursor(documents)
	ctx := context.Background()

	var results []TestDoc
	for cursor.Next(ctx) {
		var doc TestDoc
		err := cursor.Decode(&doc)
		assert.NoError(t, err)
		results = append(results, doc)
	}

	assert.Equal(t, 3, len(results))
	assert.Equal(t, "Alice", results[0].Name)
	assert.Equal(t, 30, results[0].Age)
	assert.Equal(t, "Bob", results[1].Name)
	assert.Equal(t, 25, results[1].Age)
	assert.Equal(t, "Charlie", results[2].Name)
	assert.Equal(t, 35, results[2].Age)
}

// TestMockCursor_RemainingBatchLength tests the RemainingBatchLength method
func TestMockCursor_RemainingBatchLength(t *testing.T) {
	tests := []struct {
		name      string
		documents []bson.M
		advance   int // how many times to call Next before checking
		want      int
		setup     func(*MockCursor)
	}{
		{
			name:      "empty cursor",
			documents: []bson.M{},
			advance:   0,
			want:      0,
		},
		{
			name: "cursor at start",
			documents: []bson.M{
				{"name": "Alice"},
				{"name": "Bob"},
				{"name": "Charlie"},
			},
			advance: 0,
			want:    3,
		},
		{
			name: "cursor after one advance",
			documents: []bson.M{
				{"name": "Alice"},
				{"name": "Bob"},
				{"name": "Charlie"},
			},
			advance: 1,
			want:    2,
		},
		{
			name: "cursor at end",
			documents: []bson.M{
				{"name": "Alice"},
				{"name": "Bob"},
			},
			advance: 2,
			want:    0,
		},
		{
			name: "cursor beyond end",
			documents: []bson.M{
				{"name": "Alice"},
			},
			advance: 5,
			want:    0,
		},
		{
			name: "closed cursor",
			documents: []bson.M{
				{"name": "Alice"},
			},
			advance: 0,
			want:    0,
			setup: func(c *MockCursor) {
				c.closed = true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := NewMockCursor(tt.documents)
			if tt.setup != nil {
				tt.setup(cursor)
			}

			ctx := context.Background()
			for i := 0; i < tt.advance; i++ {
				cursor.Next(ctx)
			}

			remaining := cursor.RemainingBatchLength()
			assert.Equal(t, tt.want, remaining)
		})
	}
}

// TestMockCursor_SetBatchSize tests the SetBatchSize method
func TestMockCursor_SetBatchSize(t *testing.T) {
	tests := []struct {
		name      string
		batchSize int32
	}{
		{
			name:      "set batch size to 10",
			batchSize: 10,
		},
		{
			name:      "set batch size to 100",
			batchSize: 100,
		},
		{
			name:      "set batch size to 0",
			batchSize: 0,
		},
		{
			name:      "set negative batch size",
			batchSize: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := NewMockCursor([]bson.M{{"name": "Alice"}})

			// Should not panic
			assert.NotPanics(t, func() {
				cursor.SetBatchSize(tt.batchSize)
			})

			// Cursor should still work normally
			ctx := context.Background()
			assert.True(t, cursor.Next(ctx))
		})
	}
}

// TestMockCursor_SetComment tests the SetComment method
func TestMockCursor_SetComment(t *testing.T) {
	tests := []struct {
		name    string
		comment interface{}
	}{
		{
			name:    "set string comment",
			comment: "test comment",
		},
		{
			name:    "set nil comment",
			comment: nil,
		},
		{
			name:    "set number comment",
			comment: 123,
		},
		{
			name:    "set map comment",
			comment: map[string]string{"key": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := NewMockCursor([]bson.M{{"name": "Alice"}})

			// Should not panic
			assert.NotPanics(t, func() {
				cursor.SetComment(tt.comment)
			})

			// Cursor should still work normally
			ctx := context.Background()
			assert.True(t, cursor.Next(ctx))
		})
	}
}

// TestMockCursor_SetMaxTime tests the SetMaxTime method
func TestMockCursor_SetMaxTime(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
	}{
		{
			name:     "set max time to 1 second",
			duration: time.Second,
		},
		{
			name:     "set max time to 0",
			duration: 0,
		},
		{
			name:     "set max time to negative",
			duration: -time.Second,
		},
		{
			name:     "set max time to 1 hour",
			duration: time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := NewMockCursor([]bson.M{{"name": "Alice"}})

			// Should not panic
			assert.NotPanics(t, func() {
				cursor.SetMaxTime(tt.duration)
			})

			// Cursor should still work normally
			ctx := context.Background()
			assert.True(t, cursor.Next(ctx))
		})
	}
}

// TestMockCursor_TryNext tests the TryNext method
func TestMockCursor_TryNext(t *testing.T) {
	tests := []struct {
		name       string
		documents  []bson.M
		iterations int
		wantTrue   int
		setup      func(*MockCursor)
	}{
		{
			name:       "empty cursor",
			documents:  []bson.M{},
			iterations: 1,
			wantTrue:   0,
		},
		{
			name: "single document",
			documents: []bson.M{
				{"name": "Alice"},
			},
			iterations: 2,
			wantTrue:   1,
		},
		{
			name: "multiple documents",
			documents: []bson.M{
				{"name": "Alice"},
				{"name": "Bob"},
				{"name": "Charlie"},
			},
			iterations: 5,
			wantTrue:   3,
		},
		{
			name: "closed cursor returns false",
			documents: []bson.M{
				{"name": "Alice"},
			},
			iterations: 1,
			wantTrue:   0,
			setup: func(c *MockCursor) {
				c.closed = true
			},
		},
		{
			name: "cursor with error returns false",
			documents: []bson.M{
				{"name": "Alice"},
			},
			iterations: 1,
			wantTrue:   0,
			setup: func(c *MockCursor) {
				c.err = ErrInvalidArgument
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := NewMockCursor(tt.documents)
			if tt.setup != nil {
				tt.setup(cursor)
			}

			ctx := context.Background()
			trueCount := 0
			for i := 0; i < tt.iterations; i++ {
				if cursor.TryNext(ctx) {
					trueCount++
				}
			}

			assert.Equal(t, tt.wantTrue, trueCount)
		})
	}
}

// TestMockCursor_TryNextAndNext_Consistency tests that TryNext and Next behave the same
func TestMockCursor_TryNextAndNext_Consistency(t *testing.T) {
	documents := []bson.M{
		{"name": "Alice"},
		{"name": "Bob"},
		{"name": "Charlie"},
	}

	// Test with Next
	cursor1 := NewMockCursor(documents)
	ctx := context.Background()
	nextResults := []bool{}
	for i := 0; i < 5; i++ {
		nextResults = append(nextResults, cursor1.Next(ctx))
	}

	// Test with TryNext
	cursor2 := NewMockCursor(documents)
	tryNextResults := []bool{}
	for i := 0; i < 5; i++ {
		tryNextResults = append(tryNextResults, cursor2.TryNext(ctx))
	}

	// Results should be identical
	assert.Equal(t, nextResults, tryNextResults)
}

// TestMockCursor_Current tests the Current method
func TestMockCursor_Current(t *testing.T) {
	type TestDoc struct {
		Name string
		Age  int
	}

	tests := []struct {
		name      string
		documents []bson.M
		advance   int
		wantNil   bool
		checkDoc  *TestDoc
		setup     func(*MockCursor)
	}{
		{
			name:      "before first document",
			documents: []bson.M{{"name": "Alice", "age": 30}},
			advance:   0,
			wantNil:   true,
		},
		{
			name:      "at first document",
			documents: []bson.M{{"name": "Alice", "age": 30}},
			advance:   1,
			wantNil:   false,
			checkDoc:  &TestDoc{Name: "Alice", Age: 30},
		},
		{
			name: "at second document",
			documents: []bson.M{
				{"name": "Alice", "age": 30},
				{"name": "Bob", "age": 25},
			},
			advance:  2,
			wantNil:  false,
			checkDoc: &TestDoc{Name: "Bob", Age: 25},
		},
		{
			name:      "empty cursor",
			documents: []bson.M{},
			advance:   0,
			wantNil:   true,
		},
		{
			name:      "beyond last document",
			documents: []bson.M{{"name": "Alice", "age": 30}},
			advance:   2,
			wantNil:   true,
		},
		{
			name:      "closed cursor",
			documents: []bson.M{{"name": "Alice", "age": 30}},
			advance:   1,
			wantNil:   true,
			setup: func(c *MockCursor) {
				c.closed = true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := NewMockCursor(tt.documents)

			ctx := context.Background()
			for i := 0; i < tt.advance; i++ {
				cursor.Next(ctx)
			}

			if tt.setup != nil {
				tt.setup(cursor)
			}

			current := cursor.Current()
			if tt.wantNil {
				assert.Nil(t, current)
			} else {
				assert.NotNil(t, current)

				// Verify we can decode the current document
				if tt.checkDoc != nil {
					var doc TestDoc
					err := bson.Unmarshal(current, &doc)
					assert.NoError(t, err)
					assert.Equal(t, tt.checkDoc.Name, doc.Name)
					assert.Equal(t, tt.checkDoc.Age, doc.Age)
				}
			}
		})
	}
}

// TestMockCursor_Current_UnmarshalError tests Current with marshal errors
func TestMockCursor_Current_UnmarshalError(t *testing.T) {
	// Create a cursor with a document that might cause issues
	cursor := NewMockCursor([]bson.M{
		{"name": "Alice", "age": 30},
	})

	ctx := context.Background()
	cursor.Next(ctx)

	// Get current document
	current := cursor.Current()
	assert.NotNil(t, current)

	// Should be valid BSON
	var result bson.M
	err := bson.Unmarshal(current, &result)
	assert.NoError(t, err)
	assert.Equal(t, "Alice", result["name"])
}

// TestMockCursor_CompleteWorkflow tests a complete cursor workflow
func TestMockCursor_CompleteWorkflow(t *testing.T) {
	type User struct {
		Name  string
		Email string
		Age   int
	}

	documents := []bson.M{
		{"name": "Alice", "email": "alice@example.com", "age": 30},
		{"name": "Bob", "email": "bob@example.com", "age": 25},
		{"name": "Charlie", "email": "charlie@example.com", "age": 35},
	}

	cursor := NewMockCursor(documents)
	ctx := context.Background()

	// Check initial state
	assert.Equal(t, int64(0), cursor.ID())
	assert.NoError(t, cursor.Err())
	assert.Equal(t, 3, cursor.RemainingBatchLength())

	// Set options (should not affect behavior)
	cursor.SetBatchSize(10)
	cursor.SetComment("test query")
	cursor.SetMaxTime(time.Second * 5)

	// Iterate through documents
	var users []User
	count := 0
	for cursor.Next(ctx) {
		count++

		// Check remaining batch length
		expectedRemaining := 3 - count
		assert.Equal(t, expectedRemaining, cursor.RemainingBatchLength())

		// Decode using Decode method
		var user User
		err := cursor.Decode(&user)
		assert.NoError(t, err)
		users = append(users, user)

		// Also check Current method
		current := cursor.Current()
		assert.NotNil(t, current)
	}

	// Verify we got all documents
	assert.Equal(t, 3, count)
	assert.Equal(t, 3, len(users))
	assert.Equal(t, "Alice", users[0].Name)
	assert.Equal(t, "Bob", users[1].Name)
	assert.Equal(t, "Charlie", users[2].Name)

	// Check final state
	assert.Equal(t, 0, cursor.RemainingBatchLength())
	assert.NoError(t, cursor.Err())

	// Close cursor
	err := cursor.Close(ctx)
	assert.NoError(t, err)

	// After closing, Next should return false
	assert.False(t, cursor.Next(ctx))
	assert.False(t, cursor.TryNext(ctx))
}

// TestMockCursor_EdgeCases tests various edge cases
func TestMockCursor_EdgeCases(t *testing.T) {
	t.Run("decode after close should fail", func(t *testing.T) {
		cursor := NewMockCursor([]bson.M{{"name": "Alice"}})
		ctx := context.Background()

		cursor.Next(ctx)
		cursor.Close(ctx)

		var doc bson.M
		err := cursor.Decode(&doc)
		assert.Error(t, err)
	})

	t.Run("current after close should return nil", func(t *testing.T) {
		cursor := NewMockCursor([]bson.M{{"name": "Alice"}})
		ctx := context.Background()

		cursor.Next(ctx)
		cursor.Close(ctx)

		current := cursor.Current()
		assert.Nil(t, current)
	})

	t.Run("multiple close calls should work", func(t *testing.T) {
		cursor := NewMockCursor([]bson.M{{"name": "Alice"}})
		ctx := context.Background()

		err1 := cursor.Close(ctx)
		err2 := cursor.Close(ctx)

		assert.NoError(t, err1)
		assert.NoError(t, err2)
	})

	t.Run("remaining batch length never negative", func(t *testing.T) {
		cursor := NewMockCursor([]bson.M{{"name": "Alice"}})
		ctx := context.Background()

		// Advance beyond end
		for i := 0; i < 10; i++ {
			cursor.Next(ctx)
		}

		remaining := cursor.RemainingBatchLength()
		assert.Equal(t, 0, remaining)
	})

	t.Run("all setter methods work without panic", func(t *testing.T) {
		cursor := NewMockCursor([]bson.M{{"name": "Alice"}})

		// Call all setter methods
		assert.NotPanics(t, func() {
			cursor.SetBatchSize(10)
			cursor.SetBatchSize(100)
			cursor.SetBatchSize(0)
			cursor.SetBatchSize(-1)

			cursor.SetComment("test comment")
			cursor.SetComment(nil)
			cursor.SetComment(123)
			cursor.SetComment(map[string]string{"key": "value"})

			cursor.SetMaxTime(time.Second)
			cursor.SetMaxTime(0)
			cursor.SetMaxTime(-time.Second)
			cursor.SetMaxTime(time.Hour)
		})

		// Cursor should still work normally after all these calls
		ctx := context.Background()
		assert.True(t, cursor.Next(ctx))
	})
}

// TestNewMockCursor tests the constructor
func TestNewMockCursor(t *testing.T) {
	tests := []struct {
		name      string
		documents []bson.M
	}{
		{
			name:      "nil documents",
			documents: nil,
		},
		{
			name:      "empty documents",
			documents: []bson.M{},
		},
		{
			name: "single document",
			documents: []bson.M{
				{"name": "Alice"},
			},
		},
		{
			name: "multiple documents",
			documents: []bson.M{
				{"name": "Alice"},
				{"name": "Bob"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := NewMockCursor(tt.documents)

			assert.NotNil(t, cursor)
			assert.Equal(t, -1, cursor.position)
			assert.False(t, cursor.closed)
			assert.NoError(t, cursor.Err())
		})
	}
}
