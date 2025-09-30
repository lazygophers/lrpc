package db

import (
	"testing"

	"gotest.tools/v3/assert"
)

// Test all sub_cond.go functions

func TestWhere(t *testing.T) {
	t.Run("simple where", func(t *testing.T) {
		cond := Where("id", 1)
		result := cond.ToString()
		assert.Equal(t, "(\"id\" = 1)", result)
	})

	t.Run("multiple conditions", func(t *testing.T) {
		cond := Where(map[string]interface{}{
			"id":   1,
			"name": "test",
		})
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
}

func TestOrWhere(t *testing.T) {
	t.Run("simple or where", func(t *testing.T) {
		cond := OrWhere("id", 1)
		result := cond.ToString()
		assert.Equal(t, "("id" = 1)", result)
	})

	t.Run("multiple or conditions", func(t *testing.T) {
		cond := OrWhere("id", 1).OrWhere("id", 2)
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
}

func TestOr(t *testing.T) {
	t.Run("or alias", func(t *testing.T) {
		cond := Or("id", 1)
		result := cond.ToString()
		assert.Equal(t, "("id" = 1)", result)
	})
}

func TestLike(t *testing.T) {
	t.Run("like condition", func(t *testing.T) {
		cond := Like("name", "test")
		result := cond.ToString()
		assert.Equal(t, "("name" LIKE \"%test%\")", result)
	})

	t.Run("empty value", func(t *testing.T) {
		cond := Like("name", "")
		result := cond.ToString()
		assert.Equal(t, "", result)
	})
}

func TestLeftLike(t *testing.T) {
	t.Run("left like condition", func(t *testing.T) {
		cond := LeftLike("name", "test")
		result := cond.ToString()
		assert.Equal(t, "("name" LIKE \"test%\")", result)
	})

	t.Run("empty value", func(t *testing.T) {
		cond := LeftLike("name", "")
		result := cond.ToString()
		assert.Equal(t, "", result)
	})
}

func TestRightLike(t *testing.T) {
	t.Run("right like condition", func(t *testing.T) {
		cond := RightLike("name", "test")
		result := cond.ToString()
		assert.Equal(t, "("name" LIKE \"%test\")", result)
	})

	t.Run("empty value", func(t *testing.T) {
		cond := RightLike("name", "")
		result := cond.ToString()
		assert.Equal(t, "", result)
	})
}

func TestNotLike(t *testing.T) {
	t.Run("not like condition", func(t *testing.T) {
		cond := NotLike("name", "test")
		result := cond.ToString()
		assert.Equal(t, "("name" NOT LIKE \"%test%\")", result)
	})

	t.Run("empty value", func(t *testing.T) {
		cond := NotLike("name", "")
		result := cond.ToString()
		assert.Equal(t, "", result)
	})
}

func TestNotLeftLike(t *testing.T) {
	t.Run("not left like condition", func(t *testing.T) {
		cond := NotLeftLike("name", "test")
		result := cond.ToString()
		assert.Equal(t, "("name" NOT LIKE \"test%\")", result)
	})

	t.Run("empty value", func(t *testing.T) {
		cond := NotLeftLike("name", "")
		result := cond.ToString()
		assert.Equal(t, "", result)
	})
}

func TestNotRightLike(t *testing.T) {
	t.Run("not right like condition", func(t *testing.T) {
		cond := NotRightLike("name", "test")
		result := cond.ToString()
		assert.Equal(t, "("name" NOT LIKE \"%test\")", result)
	})

	t.Run("empty value", func(t *testing.T) {
		cond := NotRightLike("name", "")
		result := cond.ToString()
		assert.Equal(t, "", result)
	})
}

func TestBetween(t *testing.T) {
	t.Run("between condition", func(t *testing.T) {
		cond := Between("age", 18, 65)
		result := cond.ToString()
		assert.Equal(t, ""age" BETWEEN 18 AND 65", result)
	})
}

func TestNotBetween(t *testing.T) {
	t.Run("not between condition", func(t *testing.T) {
		cond := NotBetween("age", 18, 65)
		result := cond.ToString()
		assert.Equal(t, ""age" NOT BETWEEN 18 AND 65", result)
	})
}

// Test initialization

func TestCondInitialization(t *testing.T) {
	t.Run("where initializes correctly", func(t *testing.T) {
		cond := Where("id", 1)
		assert.Assert(t, cond != nil)
	})

	t.Run("like initializes correctly", func(t *testing.T) {
		cond := Like("name", "test")
		assert.Assert(t, cond != nil)
	})
}

// Benchmark tests

func BenchmarkWhere(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Where("id", 1)
	}
}

func BenchmarkLike(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Like("name", "test")
	}
}

func BenchmarkBetween(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Between("age", 18, 65)
	}
}

func BenchmarkOrWhere(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = OrWhere("id", 1)
	}
}

func BenchmarkMultipleConditions(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Where("id", 1).Where("name", "test").Where("age >", 18)
	}
}

// Benchmark concurrent usage

func BenchmarkConcurrentWhere(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Where("id", 1)
		}
	})
}

func BenchmarkConcurrentLike(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Like("name", "test")
		}
	})
}

func BenchmarkConcurrentBetween(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Between("age", 18, 65)
		}
	})
}

func BenchmarkConcurrentMixed(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 4 {
			case 0:
				_ = Where("id", 1)
			case 1:
				_ = Like("name", "test")
			case 2:
				_ = Between("age", 18, 65)
			case 3:
				_ = OrWhere("status", "active")
			}
			i++
		}
	})
}

// Test concurrent stress

func TestConcurrentStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	const goroutines = 100
	const iterations = 1000

	done := make(chan bool)

	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				switch j % 10 {
				case 0:
					_ = Where("id", j)
				case 1:
					_ = Like("name", "test")
				case 2:
					_ = LeftLike("email", "user")
				case 3:
					_ = RightLike("domain", ".com")
				case 4:
					_ = NotLike("status", "deleted")
				case 5:
					_ = NotLeftLike("prefix", "tmp")
				case 6:
					_ = NotRightLike("suffix", ".tmp")
				case 7:
					_ = Between("age", 18, 65)
				case 8:
					_ = NotBetween("score", 0, 10)
				case 9:
					_ = OrWhere("active", true)
				}
			}
			done <- true
		}()
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}
}

// Benchmark memory usage

func BenchmarkMemoryUsage(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cond := Where("id", i)
		_ = cond.ToString()
	}
}

// Test API completeness

func TestAPICompleteness(t *testing.T) {
	tests := []struct {
		name string
		fn   func() *Cond
		want string
	}{
		{"Where", func() *Cond { return Where("id", 1) }, "("id" = 1)"},
		{"OrWhere", func() *Cond { return OrWhere("id", 1) }, "("id" = 1)"},
		{"Or", func() *Cond { return Or("id", 1) }, "("id" = 1)"},
		{"Like", func() *Cond { return Like("name", "test") }, "("name" LIKE \"%test%\")"},
		{"LeftLike", func() *Cond { return LeftLike("name", "test") }, "("name" LIKE \"test%\")"},
		{"RightLike", func() *Cond { return RightLike("name", "test") }, "("name" LIKE \"%test\")"},
		{"NotLike", func() *Cond { return NotLike("name", "test") }, "("name" NOT LIKE \"%test%\")"},
		{"NotLeftLike", func() *Cond { return NotLeftLike("name", "test") }, "("name" NOT LIKE \"test%\")"},
		{"NotRightLike", func() *Cond { return NotRightLike("name", "test") }, "("name" NOT LIKE \"%test\")"},
		{"Between", func() *Cond { return Between("age", 18, 65) }, ""age" BETWEEN 18 AND 65"},
		{"NotBetween", func() *Cond { return NotBetween("age", 18, 65) }, ""age" NOT BETWEEN 18 AND 65"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := tt.fn()
			result := cond.ToString()
			assert.Equal(t, tt.want, result)
		})
	}
}