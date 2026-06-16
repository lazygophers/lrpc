package db

import (
	"reflect"
	"sync"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/xerror"
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

var (
	// gormTagCache caches parsed GORM tag information for struct types
	gormTagCache = make(map[reflect.Type]*gormTagInfo)
	gormTagMutex sync.RWMutex

	// fieldNameCache caches field name conversions from snake_case to CamelCase
	fieldNameCache = make(map[string]string)
	fieldNameMutex sync.RWMutex

	// Pool for gorm.Statement objects to reduce allocations
	statementPool = sync.Pool{
		New: func() interface{} {
			return &gorm.Statement{}
		},
	}
)

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
	return s
}

func (p *Scoop) getNotFoundError() error {
	if p.notFoundError != nil {
		return p.notFoundError
	}
	return xerror.New(xerror.CodeNoData, "record not found")
}

func (p *Scoop) getDuplicatedKeyError() error {
	if p.duplicatedKeyError != nil {
		return p.duplicatedKeyError
	}
	return xerror.New(xerror.CodeConflict, "duplicate key error")
}

func (p *Scoop) IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	return xerror.Code(err) == xerror.CodeNoData || err == gorm.ErrRecordNotFound || err == p.notFoundError
}

func (p *Scoop) IsDuplicatedKeyError(err error) bool {
	return xerror.Code(err) == xerror.CodeConflict
}

func (p *Scoop) AutoMigrate(dst ...interface{}) error {
	return p._db.AutoMigrate(dst...)
}

func (p *Scoop) inc() {
	p.depth++
}

func (p *Scoop) dec() {
	p.depth--
}

func (p *Scoop) Model(m any) *Scoop {
	p._db = p._db.Model(m)

	rt := reflect.TypeOf(m)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	if rt.Kind() == reflect.Struct {
		p.hasCreatedAt = hasCreatedAt(rt)
		p.hasUpdatedAt = hasUpdatedAt(rt)
		p.hasDeletedAt = hasDeletedAt(rt)
		p.hasId = hasId(rt)

		p.table = getTableName(rt)
	}

	return p
}

func (p *Scoop) Table(m string) *Scoop {
	if !isValidTableName(m) {
		log.Errorf("invalid table name:%s", m)
		return p
	}
	p.table = m
	p._db = p._db.Table(m)
	return p
}

func (p *Scoop) Ignore(b ...bool) *Scoop {
	if len(b) > 0 {
		p.ignore = b[0]
	} else {
		p.ignore = true
	}
	return p
}
