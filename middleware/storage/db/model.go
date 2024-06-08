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

	notFoundError error

	hasDeletedAt bool
	hasId        bool
	table        string
}

func NewModel[M any](db *Client) *Model[M] {
	p := &Model[M]{
		db:            db,
		notFoundError: gorm.ErrRecordNotFound,
	}

	rt := reflect.TypeOf(new(M))
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	p.hasId = hasId(rt)
	p.hasDeletedAt = hasDeleted(rt)
	p.table = getTableName(rt)

	return p
}

func (p *Model[M]) SetNotFound(err error) *Model[M] {
	p.notFoundError = err

	return p
}

func (p *Model[M]) IsNotFound(err error) bool {
	return err == p.notFoundError || gorm.ErrRecordNotFound == err
}

func (p *Model[M]) NewScoop(tx ...*Scoop) *ModelScoop[M] {
	var db *gorm.DB
	if len(tx) == 0 || tx[0] == nil {
		db = p.db.db
	} else {
		db = tx[0]._db
	}

	scoop := NewModelScoop[M](db)
	scoop.hasDeletedAt = p.hasDeletedAt
	scoop.hasId = p.hasId
	scoop.table = p.table
	scoop.notFoundError = p.notFoundError

	return scoop
}

func (p *Model[M]) TableName() string {
	return p.table
}
