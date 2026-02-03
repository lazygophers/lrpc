package db

import (
	"fmt"

	"github.com/lazygophers/log"
)

func (p *ModelScoop[M]) Create(m *M) error {
	if m == nil {
		return fmt.Errorf("create failed: input parameter m is nil")
	}

	p.inc()
	defer p.dec()

	err := p.Scoop.Create(m).Error
	if err != nil {
		log.Errorf("err:%v", err)
	}
	return err
}

type FirstOrCreateResult[M any] struct {
	IsCreated bool
	Error     error

	Object *M
}

func (p *ModelScoop[M]) FirstOrCreate(m *M) *FirstOrCreateResult[M] {
	if m == nil {
		return &FirstOrCreateResult[M]{
			Error: fmt.Errorf("FirstOrCreate failed: input parameter m is nil"),
		}
	}

	p.inc()
	defer p.dec()

	var mm M
	err := p.Scoop.First(&mm).Error
	if err == nil {
		return &FirstOrCreateResult[M]{
			Object: &mm,
		}
	}

	if !p.IsNotFound(err) {
		log.Errorf("err:%v", err)
		return &FirstOrCreateResult[M]{
			Error: err,
		}
	}

	err = p.Scoop.Create(m).Error
	if err == nil {
		return &FirstOrCreateResult[M]{
			IsCreated: true,
			Object:    m,
		}
	}

	if !p.IsDuplicatedKeyError(err) {
		log.Errorf("err:%v", err)
		return &FirstOrCreateResult[M]{
			Error: err,
		}
	}

	err = p.Scoop.First(&mm).Error
	if err == nil {
		return &FirstOrCreateResult[M]{
			IsCreated: false,
			Object:    &mm,
		}
	}

	log.Errorf("err:%v", err)
	return &FirstOrCreateResult[M]{
		Error: err,
	}
}

type CreateIfNotExistsResult struct {
	IsCreated bool
	Error     error
}

func (p *ModelScoop[M]) CreateIfNotExists(m *M) *CreateIfNotExistsResult {
	if m == nil {
		return &CreateIfNotExistsResult{
			Error: fmt.Errorf("CreateIfNotExists failed: input parameter m is nil"),
		}
	}

	p.inc()
	defer p.dec()

	exist, err := p.Exist()
	if err != nil {
		log.Errorf("err:%v", err)
		return &CreateIfNotExistsResult{
			Error: err,
		}
	}

	if exist {
		return &CreateIfNotExistsResult{
			IsCreated: false,
		}
	}

	err = p.Scoop.Create(m).Error
	if err != nil {
		if p.IsDuplicatedKeyError(err) {
			exist, err = p.Exist()
			if err != nil {
				log.Errorf("err:%v", err)
				return &CreateIfNotExistsResult{
					Error: err,
				}
			}

			if exist {
				return &CreateIfNotExistsResult{}
			}

			log.Errorf("err:%v", p.getDuplicatedKeyError())
			return &CreateIfNotExistsResult{
				Error: p.getDuplicatedKeyError(),
			}
		}

		log.Errorf("err:%v", err)
		return &CreateIfNotExistsResult{
			Error: err,
		}
	}

	return &CreateIfNotExistsResult{
		IsCreated: true,
	}
}

type CreateNotExistResult[M any] struct {
	IsCreated bool
	Error     error

	Object *M
}

// CreateNotExist is an alias for FirstOrCreate for backward compatibility.
// It performs the same operation: finds a record, returns it if found, creates it if not.
// Deprecated: Use FirstOrCreate instead, which provides the same functionality.
func (p *ModelScoop[M]) CreateNotExist(m *M) *CreateNotExistResult[M] {
	result := p.FirstOrCreate(m)
	return &CreateNotExistResult[M]{
		IsCreated: result.IsCreated,
		Error:     result.Error,
		Object:    result.Object,
	}
}

func (p *ModelScoop[M]) CreateInBatches(values []*M, batchSize int) *CreateInBatchesResult {
	if len(values) == 0 {
		return &CreateInBatchesResult{
			RowsAffected: 0,
			Error:        nil,
		}
	}

	p.inc()
	defer p.dec()

	return p.Scoop.CreateInBatches(values, batchSize)
}
