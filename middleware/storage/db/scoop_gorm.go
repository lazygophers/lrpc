package db

import (
	"reflect"
	"strings"
)

// gormTagInfo stores parsed information about updatable fields
type gormTagInfo struct {
	updatableFields map[string]string // fieldName -> dbName mapping
}

// parseGormTags parses GORM tags for a struct type and caches the result
// Returns information about which fields are updatable and their DB names
func parseGormTags(t reflect.Type) *gormTagInfo {
	// Fast path: read lock for cache lookup
	gormTagMutex.RLock()
	if info, ok := gormTagCache[t]; ok {
		gormTagMutex.RUnlock()
		return info
	}
	gormTagMutex.RUnlock()

	// Slow path: write lock for cache update
	gormTagMutex.Lock()
	defer gormTagMutex.Unlock()

	// Double-check in case another goroutine already cached it
	if info, ok := gormTagCache[t]; ok {
		return info
	}

	// Parse struct tags
	info := &gormTagInfo{
		updatableFields: make(map[string]string),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		gormTag := field.Tag.Get("gorm")

		// Skip ignored fields
		if gormTag == "-" {
			continue
		}

		// Skip fields that should not be updated
		if isSkippableField(gormTag, field.Name) {
			continue
		}

		// Extract DB column name
		dbName := extractDBName(gormTag, field.Name)
		info.updatableFields[field.Name] = dbName
	}

	gormTagCache[t] = info
	return info
}

// isSkippableField checks if a field should be skipped during updates
func isSkippableField(gormTag, fieldName string) bool {
	// Check for special GORM tags
	if strings.Contains(gormTag, "primaryKey") ||
		strings.Contains(gormTag, "autoCreateTime") ||
		strings.Contains(gormTag, "autoUpdateTime") {
		return true
	}

	// Check for time tracking fields by name
	if fieldName == structFieldCreatedAt || fieldName == structFieldUpdatedAt {
		return true
	}

	return false
}

// extractDBName extracts the database column name from GORM tag
func extractDBName(gormTag, fieldName string) string {
	if gormTag == "" {
		return Camel2UnderScore(fieldName)
	}

	// Parse column name from tag
	parts := strings.Split(gormTag, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}

	return Camel2UnderScore(fieldName)
}

// isValidTableName checks if a table name contains only valid characters
func isValidTableName(tableName string) bool {
	if tableName == "" {
		return false
	}
	for _, char := range tableName {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '.') {
			return false
		}
	}
	return true
}
