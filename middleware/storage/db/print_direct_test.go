package db

import (
	"testing"
)

// TestPrintMethodDirectInternal - Test Print function from within the same package
func TestPrintMethodDirectInternal(t *testing.T) {
	t.Run("internal_print_test", func(t *testing.T) {
		// Create mysqlLogger instance directly (no type alias)
		logger := &mysqlLogger{}
		
		// Call Print method directly
		logger.Print()
		logger.Print("test string")
		logger.Print("test", 123, true, 3.14)
		logger.Print(nil)
		logger.Print([]interface{}{"a", "b", "c"})
		logger.Print("multiple", "arguments", "test", 456)
		
		t.Logf("Internal Print method test completed successfully")
	})
}