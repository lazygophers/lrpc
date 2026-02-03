package db

import (
	"fmt"

	"github.com/lazygophers/log"
)

type UpdateOrCreateResult[M any] struct {
	IsCreated bool
	Error     error

	Object *M
}

// UpdateOrCreate is an alias for CreateOrUpdate for backward compatibility.
// It performs the same operation: finds a record, updates it if found, creates it if not.
// Deprecated: Use CreateOrUpdate instead, which provides more detailed status information.
func (p *ModelScoop[M]) UpdateOrCreate(values map[string]interface{}, m *M) *UpdateOrCreateResult[M] {
	result := p.CreateOrUpdate(values, m)
	return &UpdateOrCreateResult[M]{
		IsCreated: result.Created,
		Error:     result.Error,
		Object:    result.Object,
	}
}

type CreateOrUpdateResult[M any] struct {
	Error  error
	Object *M

	Created bool
	Updated bool
}

func (p *ModelScoop[M]) CreateOrUpdate(values map[string]interface{}, m *M) *CreateOrUpdateResult[M] {
	if m == nil {
		err := fmt.Errorf("CreateOrUpdate failed: input parameter m is nil")
		log.Errorf("err:%v", err)
		return &CreateOrUpdateResult[M]{
			Error: err,
		}
	}

	p.inc()
	defer p.dec()

	var mm M
	err := p.Scoop.First(&mm).Error
	if err != nil {
		if p.IsNotFound(err) {
			err = p.Scoop.Create(m).Error
			if err != nil {
				log.Errorf("err:%v", err)
				return &CreateOrUpdateResult[M]{
					Error: err,
				}
			}
			return &CreateOrUpdateResult[M]{
				Created: true,
				Object:  m,
			}
		}

		log.Errorf("err:%v", err)
		return &CreateOrUpdateResult[M]{
			Error: err,
		}
	}

	err = p.Scoop.Updates(values).Error
	if err != nil {
		log.Errorf("err:%v", err)
		return &CreateOrUpdateResult[M]{
			Error: err,
		}
	}

	err = p.Scoop.First(&mm).Error
	if err != nil {
		log.Errorf("err:%v", err)
		return &CreateOrUpdateResult[M]{
			Error: err,
		}
	}

	return &CreateOrUpdateResult[M]{
		Updated: true,
		Object:  &mm,
	}
}
