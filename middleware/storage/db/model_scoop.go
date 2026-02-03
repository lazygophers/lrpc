package db

import (
	"gorm.io/gorm"
)

type ModelScoop[M any] struct {
	Scoop

	m M
}

func NewModelScoop[M any](db *gorm.DB, clientType string) *ModelScoop[M] {
	scoop := &ModelScoop[M]{
		Scoop: Scoop{
			clientType: clientType,
			_db: db.Session(&gorm.Session{
				Initialized: true,
			}),
		},
	}

	// Set clientType in cond for proper field quoting
	scoop.cond.clientType = clientType
	scoop.inc()

	return scoop
}
