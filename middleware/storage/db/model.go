package db

import (
	"gorm.io/gorm"
	"reflect"
)

type Tabler interface {
	TableName() string
}

type Model[M any] struct {
	db *Client
	m  M

	notFoundError      error
	duplicatedKeyError error

	hasCreatedAt, hasUpdatedAt, hasDeletedAt bool

	hasId bool
	table string
}

func NewModel[M any](db *Client) *Model[M] {
	p := &Model[M]{
		db:                 db,
		notFoundError:      gorm.ErrRecordNotFound,
		duplicatedKeyError: gorm.ErrDuplicatedKey,
	}

	rt := reflect.TypeOf(new(M))
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	p.hasId = hasId(rt)
	p.hasCreatedAt = hasCreatedAt(rt)
	p.hasUpdatedAt = hasUpdatedAt(rt)
	p.hasDeletedAt = hasDeletedAt(rt)
	p.table = getTableName(rt)

	return p
}

func (p *Model[M]) SetNotFound(err error) *Model[M] {
	p.notFoundError = err
	return p
}

func (p *Model[M]) SetDuplicatedKeyError(err error) *Model[M] {
	p.duplicatedKeyError = err
	return p
}

func (p *Model[M]) IsNotFound(err error) bool {
	return err == p.notFoundError || gorm.ErrRecordNotFound == err
}

func (p *Model[M]) IsDuplicatedKeyError(err error) bool {
	return err == p.duplicatedKeyError || gorm.ErrDuplicatedKey == err
}

func (p *Model[M]) NewScoop(tx ...*Scoop) *ModelScoop[M] {
	var db *gorm.DB
	if len(tx) == 0 || tx[0] == nil {
		db = p.db.db
	} else {
		db = tx[0]._db
	}

	scoop := NewModelScoop[M](db)

	scoop.hasCreatedAt = p.hasCreatedAt
	scoop.hasUpdatedAt = p.hasUpdatedAt
	scoop.hasDeletedAt = p.hasDeletedAt

	scoop.clientType = p.db.clientType
	scoop.hasId = p.hasId
	scoop.table = p.table
	scoop.notFoundError = p.notFoundError
	scoop.duplicatedKeyError = p.duplicatedKeyError

	return scoop
}

func (p *Model[M]) TableName() string {
	return p.table
}
