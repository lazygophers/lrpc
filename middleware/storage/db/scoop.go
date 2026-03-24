package db

import (
	"errors"
	"reflect"
	"strings"
	"sync"

	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/xerror"
	"gorm.io/gorm"
)

// Field name constants for common database fields
const (
	fieldDeletedAt = "deleted_at"
	fieldCreatedAt = "created_at"
	fieldUpdatedAt = "updated_at"
	fieldID        = "id"
)

// SQL condition constants
const (
	condNotDeleted = "deleted_at = 0"
)

// Field name constants for Go struct fields
const (
	structFieldDeletedAt = "DeletedAt"
	structFieldCreatedAt = "CreatedAt"
	structFieldUpdatedAt = "UpdatedAt"
)

// gormTagCache caches parsed GORM tag information for struct types
// This significantly improves performance for repeated Updates operations
var (
	gormTagCache = make(map[reflect.Type]*gormTagInfo)
	gormTagMutex sync.RWMutex

	// Cache for field name conversions from snake_case to CamelCase
	fieldNameCache = make(map[string]string)
	fieldNameMutex sync.RWMutex

	// Pool for gorm.Statement objects to reduce allocations
	statementPool = sync.Pool{
		New: func() interface{} {
			return &gorm.Statement{}
		},
	}
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

type joinClause struct {
	joinType  string // INNER, LEFT, RIGHT, FULL
	table     string
	condition string
}

type Scoop struct {
	clientType string
	_db        *gorm.DB

	notFoundError      error
	duplicatedKeyError error

	hasCreatedAt, hasUpdatedAt, hasDeletedAt bool

	hasId bool
	table string

	cond          Cond
	limit, offset uint64
	selects       []string
	groups        []string
	orders        []string
	joins         []joinClause
	havingCond    *Cond
	unscoped      bool

	ignore bool

	depth int
}

func NewScoop(db *gorm.DB, clientType string) *Scoop {
	s := &Scoop{
		depth:      3,
		clientType: clientType,
		_db: db.Session(&gorm.Session{
			//NewDB: true,
			Initialized: true,
		}),
	}
	// Set clientType in cond for proper field quoting
	s.cond.clientType = clientType
	return s
}

func (p *Scoop) getNotFoundError() error {
	if p.notFoundError != nil {
		if x, ok := p.notFoundError.(*xerror.Error); ok {
			return xerror.New(x.Code)
		}

		return p.notFoundError
	}

	return gorm.ErrRecordNotFound
}

func (p *Scoop) getDuplicatedKeyError() error {
	if p.duplicatedKeyError != nil {
		if x, ok := p.duplicatedKeyError.(*xerror.Error); ok {
			return xerror.New(x.Code)
		}

		return p.duplicatedKeyError
	}

	return gorm.ErrDuplicatedKey
}

func (p *Scoop) IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	if p.notFoundError != nil {
		if x, ok := err.(*xerror.Error); ok {
			return xerror.CheckCode(err, x.Code)
		}

		return errors.Is(err, p.getNotFoundError())
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}

	return false
}

func (p *Scoop) IsDuplicatedKeyError(err error) bool {
	if err == nil {
		return false
	}

	if p.duplicatedKeyError != nil {
		if x, ok := err.(*xerror.Error); ok {
			return xerror.CheckCode(err, x.Code)
		}

		return errors.Is(err, p.duplicatedKeyError)
	}

	return IsUniqueIndexConflictErr(err)
}

func (p *Scoop) AutoMigrate(dst ...interface{}) error {
	return p._db.AutoMigrate(dst...)
}

// isValidTableName validates table name to prevent SQL injection
// Only allows alphanumeric characters, underscores, and dots (for schema.table)
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

func (p *Scoop) inc() {
	p.depth++
}

func (p *Scoop) dec() {
	p.depth--
}

func (p *Scoop) Model(m any) *Scoop {
	rt := reflect.ValueOf(m).Type()
	p.table = getTableName(rt)

	p.hasCreatedAt = hasCreatedAt(rt)
	p.hasUpdatedAt = hasUpdatedAt(rt)
	p.hasDeletedAt = hasDeletedAt(rt)

	p.hasId = hasId(rt)

	return p
}

// Table sets the table name for the query
// Validates table name to prevent SQL injection
func (p *Scoop) Table(m string) *Scoop {
	if !isValidTableName(m) {
		log.Warnf("invalid table name: %s, table name must contain only alphanumeric characters, underscores and dots", m)
	}
	p.table = m
	return p
}

func (p *Scoop) Ignore(b ...bool) *Scoop {
	if len(b) == 0 {
		p.ignore = true
		return p
	}
	p.ignore = b[0]
	return p
}
