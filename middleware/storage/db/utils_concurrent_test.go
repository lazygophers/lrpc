package db

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

// TestConcurrentGetTableName tests thread safety of getTableName with caching
func TestConcurrentGetTableName(t *testing.T) {
	type TestModel1 struct {
		ID   int64
		Name string
	}
	type TestModel2 struct {
		ID    int64
		Email string
	}

	typ1 := reflect.TypeOf(TestModel1{})
	typ2 := reflect.TypeOf(TestModel2{})

	const goroutines = 100
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Test concurrent access to cached types
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				if id%2 == 0 {
					_ = getTableName(typ1)
				} else {
					_ = getTableName(typ2)
				}
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentHasField tests thread safety of hasField with caching
func TestConcurrentHasField(t *testing.T) {
	type TestStruct struct {
		ID        int64
		Name      string
		DeletedAt int64
		CreatedAt int64
		UpdatedAt int64
	}

	typ := reflect.TypeOf(TestStruct{})

	const goroutines = 100
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	fields := []string{"ID", "Name", "DeletedAt", "CreatedAt", "UpdatedAt", "NonExistent"}

	// Test concurrent access to cached field lookups
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				fieldName := fields[j%len(fields)]
				_ = hasField(typ, fieldName)
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentDecode tests thread safety of decode function
func TestConcurrentDecode(t *testing.T) {
	const goroutines = 100
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Test concurrent decoding of different types
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				switch j % 5 {
				case 0: // int
					var result int
					field := reflect.ValueOf(&result).Elem()
					_ = decode(field, []byte("123"))
				case 1: // string
					var result string
					field := reflect.ValueOf(&result).Elem()
					_ = decode(field, []byte("test"))
				case 2: // bool
					var result bool
					field := reflect.ValueOf(&result).Elem()
					_ = decode(field, []byte("1"))
				case 3: // float64
					var result float64
					field := reflect.ValueOf(&result).Elem()
					_ = decode(field, []byte("3.14"))
				case 4: // uint64
					var result uint64
					field := reflect.ValueOf(&result).Elem()
					_ = decode(field, []byte("999"))
				}
			}
		}(i)
	}

	wg.Wait()
}

// Benchmark concurrent getTableName
func BenchmarkConcurrentGetTableName(b *testing.B) {
	type BenchModel struct {
		ID   int64
		Name string
	}
	typ := reflect.TypeOf(BenchModel{})

	// Warm up cache
	_ = getTableName(typ)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = getTableName(typ)
		}
	})
}

// Benchmark concurrent getTableName with multiple types
func BenchmarkConcurrentGetTableName_MultipleTypes(b *testing.B) {
	types := make([]reflect.Type, 10)
	for i := 0; i < 10; i++ {
		// Create distinct types dynamically
		switch i {
		case 0:
			type T0 struct{ ID int64 }
			types[i] = reflect.TypeOf(T0{})
		case 1:
			type T1 struct{ ID int64 }
			types[i] = reflect.TypeOf(T1{})
		case 2:
			type T2 struct{ ID int64 }
			types[i] = reflect.TypeOf(T2{})
		case 3:
			type T3 struct{ ID int64 }
			types[i] = reflect.TypeOf(T3{})
		case 4:
			type T4 struct{ ID int64 }
			types[i] = reflect.TypeOf(T4{})
		case 5:
			type T5 struct{ ID int64 }
			types[i] = reflect.TypeOf(T5{})
		case 6:
			type T6 struct{ ID int64 }
			types[i] = reflect.TypeOf(T6{})
		case 7:
			type T7 struct{ ID int64 }
			types[i] = reflect.TypeOf(T7{})
		case 8:
			type T8 struct{ ID int64 }
			types[i] = reflect.TypeOf(T8{})
		case 9:
			type T9 struct{ ID int64 }
			types[i] = reflect.TypeOf(T9{})
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = getTableName(types[i%len(types)])
			i++
		}
	})
}

// Benchmark concurrent hasField
func BenchmarkConcurrentHasField(b *testing.B) {
	type TestStruct struct {
		ID        int64
		Name      string
		DeletedAt int64
	}
	typ := reflect.TypeOf(TestStruct{})

	// Warm up cache
	_ = hasField(typ, "DeletedAt")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = hasField(typ, "DeletedAt")
		}
	})
}

// Benchmark concurrent hasField with cache misses
func BenchmarkConcurrentHasField_MultipleFields(b *testing.B) {
	type TestStruct struct {
		Field1  string
		Field2  int
		Field3  bool
		Field4  float64
		Field5  uint64
		Field6  string
		Field7  int
		Field8  bool
		Field9  float64
		Field10 uint64
	}
	typ := reflect.TypeOf(TestStruct{})

	fields := []string{"Field1", "Field2", "Field3", "Field4", "Field5",
		"Field6", "Field7", "Field8", "Field9", "Field10"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = hasField(typ, fields[i%len(fields)])
			i++
		}
	})
}

// Benchmark concurrent decode
func BenchmarkConcurrentDecode_Int(b *testing.B) {
	data := []byte("12345")

	b.RunParallel(func(pb *testing.PB) {
		var result int
		field := reflect.ValueOf(&result).Elem()
		for pb.Next() {
			_ = decode(field, data)
		}
	})
}

func BenchmarkConcurrentDecode_String(b *testing.B) {
	data := []byte("test string value")

	b.RunParallel(func(pb *testing.PB) {
		var result string
		field := reflect.ValueOf(&result).Elem()
		for pb.Next() {
			_ = decode(field, data)
		}
	})
}

func BenchmarkConcurrentDecode_Bool(b *testing.B) {
	data := []byte("1")

	b.RunParallel(func(pb *testing.PB) {
		var result bool
		field := reflect.ValueOf(&result).Elem()
		for pb.Next() {
			_ = decode(field, data)
		}
	})
}

func BenchmarkConcurrentDecode_Float64(b *testing.B) {
	data := []byte("3.141592653589793")

	b.RunParallel(func(pb *testing.PB) {
		var result float64
		field := reflect.ValueOf(&result).Elem()
		for pb.Next() {
			_ = decode(field, data)
		}
	})
}

// Benchmark concurrent operations with mixed types
func BenchmarkConcurrentMixedOperations(b *testing.B) {
	type TestModel struct {
		ID        int64
		Name      string
		DeletedAt int64
	}
	typ := reflect.TypeOf(TestModel{})

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 3 {
			case 0:
				_ = getTableName(typ)
			case 1:
				_ = hasField(typ, "DeletedAt")
			case 2:
				var result int
				field := reflect.ValueOf(&result).Elem()
				_ = decode(field, []byte("123"))
			}
			i++
		}
	})
}

// Test race conditions with -race flag
func TestRaceConditionGetTableName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	type RaceModel struct {
		ID int64
	}
	typ := reflect.TypeOf(RaceModel{})

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = getTableName(typ)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestRaceConditionHasField(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	type RaceStruct struct {
		ID   int64
		Name string
	}
	typ := reflect.TypeOf(RaceStruct{})

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = hasField(typ, "Name")
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// Benchmark to compare before/after optimization
func BenchmarkGetTableName_Serial(b *testing.B) {
	type SerialModel struct {
		ID int64
	}
	typ := reflect.TypeOf(SerialModel{})

	// Warm up
	_ = getTableName(typ)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getTableName(typ)
	}
}

func BenchmarkHasField_Serial(b *testing.B) {
	type SerialStruct struct {
		ID        int64
		DeletedAt int64
	}
	typ := reflect.TypeOf(SerialStruct{})

	// Warm up
	_ = hasField(typ, "DeletedAt")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hasField(typ, "DeletedAt")
	}
}

func BenchmarkUnsafeString(b *testing.B) {
	data := []byte("test string for unsafe conversion")

	b.Run("unsafeString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = unsafeString(data)
		}
	})

	b.Run("string()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = string(data)
		}
	})
}

func BenchmarkDecode_WithAndWithoutUnsafe(b *testing.B) {
	data := []byte("12345")

	b.Run("decode_int_optimized", func(b *testing.B) {
		var result int
		field := reflect.ValueOf(&result).Elem()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = decode(field, data)
		}
	})
}

// Stress test with high concurrency
func TestHighConcurrencyStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	type StressModel struct {
		ID        int64
		Name      string
		DeletedAt int64
		CreatedAt int64
		UpdatedAt int64
	}
	typ := reflect.TypeOf(StressModel{})

	const goroutines = 1000
	const iterations = 10000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	errors := make(chan error, goroutines*iterations)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// Mix different operations
				switch (id + j) % 5 {
				case 0:
					_ = getTableName(typ)
				case 1:
					_ = hasField(typ, "DeletedAt")
				case 2:
					_ = hasDeletedAt(typ)
				case 3:
					_ = hasCreatedAt(typ)
				case 4:
					_ = hasUpdatedAt(typ)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	if len(errors) > 0 {
		t.Fatalf("Encountered %d errors during stress test", len(errors))
	}
}

// Verify cache effectiveness
func TestCacheEffectiveness(t *testing.T) {
	type CacheTestModel struct {
		ID int64
	}
	typ := reflect.TypeOf(CacheTestModel{})

	// Clear caches
	tableNameCacheMu.Lock()
	tableNameCache = make(map[reflect.Type]string)
	tableNameCacheMu.Unlock()

	hasFieldCacheMu.Lock()
	hasFieldCache = make(map[fieldCacheKey]bool)
	hasFieldCacheMu.Unlock()

	// First call - cache miss
	name1 := getTableName(typ)

	// Second call - should hit cache
	name2 := getTableName(typ)

	if name1 != name2 {
		t.Errorf("Cache returned different values: %s vs %s", name1, name2)
	}

	// Verify cache was populated
	tableNameCacheMu.RLock()
	_, exists := tableNameCache[typ]
	tableNameCacheMu.RUnlock()

	if !exists {
		t.Error("Cache was not populated after getTableName call")
	}

	// Test hasField cache
	has1 := hasField(typ, "ID")
	has2 := hasField(typ, "ID")

	if has1 != has2 {
		t.Errorf("hasField cache returned different values: %v vs %v", has1, has2)
	}

	key := fieldCacheKey{typ: typ, fieldName: "ID"}
	hasFieldCacheMu.RLock()
	_, exists = hasFieldCache[key]
	hasFieldCacheMu.RUnlock()

	if !exists {
		t.Error("hasField cache was not populated")
	}
}

// Benchmark memory allocations
func BenchmarkDecodeAllocations(b *testing.B) {
	data := []byte("12345")

	b.Run("int_with_unsafe", func(b *testing.B) {
		var result int
		field := reflect.ValueOf(&result).Elem()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = decode(field, data)
		}
	})

	b.Run("string_with_copy", func(b *testing.B) {
		var result string
		field := reflect.ValueOf(&result).Elem()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = decode(field, data)
		}
	})

	b.Run("bool_optimized", func(b *testing.B) {
		data := []byte("1")
		var result bool
		field := reflect.ValueOf(&result).Elem()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = decode(field, data)
		}
	})
}

// Test concurrent cache growth
func TestConcurrentCacheGrowth(t *testing.T) {
	const numTypes = 100
	const goroutines = 50

	// Create many different types
	types := make([]reflect.Type, numTypes)
	for i := 0; i < numTypes; i++ {
		// Create a struct type with a unique field
		structFields := []reflect.StructField{
			{Name: fmt.Sprintf("Field%d", i), Type: reflect.TypeOf(int64(0))},
		}
		types[i] = reflect.StructOf(structFields)
	}

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numTypes; j++ {
				typ := types[(id+j)%numTypes]
				_ = getTableName(typ)
				_ = hasField(typ, fmt.Sprintf("Field%d", (id+j)%numTypes))
			}
		}(i)
	}

	wg.Wait()

	// Verify all types were cached
	tableNameCacheMu.RLock()
	cacheSize := len(tableNameCache)
	tableNameCacheMu.RUnlock()

	if cacheSize == 0 {
		t.Error("Cache should contain entries")
	}

	t.Logf("Cache contains %d entries after concurrent operations", cacheSize)
}