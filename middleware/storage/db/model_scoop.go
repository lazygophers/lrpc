package db

import (
	"fmt"

	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/core"
	"github.com/lazygophers/utils/candy"
	"gorm.io/gorm"
)

type ModelScoop[M any] struct {
	Scoop

	m M
}

func NewModelScoop[M any](db *gorm.DB) *ModelScoop[M] {
	scoop := &ModelScoop[M]{
		Scoop: Scoop{
			_db: db.Session(&gorm.Session{
				Initialized: true,
			}),
		},
	}

	scoop.inc()

	return scoop
}

// ——————————条件——————————

func (p *ModelScoop[M]) Select(fields ...string) *ModelScoop[M] {
	p.selects = append(p.selects, fields...)
	return p
}

func (p *ModelScoop[M]) Where(args ...interface{}) *ModelScoop[M] {
	p.cond.Where(args...)
	return p
}

func (p *ModelScoop[M]) Or(args ...interface{}) *ModelScoop[M] {
	p.cond.OrWhere(args...)
	return p
}

func (p *ModelScoop[M]) Equal(column string, value interface{}) *ModelScoop[M] {
	p.cond.where(column, value)
	return p
}

func (p *ModelScoop[M]) NotEqual(column string, value interface{}) *ModelScoop[M] {
	p.cond.where(column, " != ", value)
	return p
}

func (p *ModelScoop[M]) In(column string, values interface{}) *ModelScoop[M] {
	vo := EnsureIsSliceOrArray(values)
	if vo.Len() == 0 {
		p.cond.where(false)
		return p
	}
	p.cond.where(column, "IN", UniqueSlice(vo.Interface()))
	return p
}

func (p *ModelScoop[M]) NotIn(column string, values interface{}) *ModelScoop[M] {
	vo := EnsureIsSliceOrArray(values)
	if vo.Len() == 0 {
		return p
	}
	p.cond.where(column, "NOT IN", UniqueSlice(vo.Interface()))
	return p
}

func (p *ModelScoop[M]) Like(column string, value string) *ModelScoop[M] {
	p.cond.where(column, "LIKE", "%"+value+"%")
	return p
}

func (p *ModelScoop[M]) LeftLike(column string, value string) *ModelScoop[M] {
	p.cond.where(column, "LIKE", value+"%")
	return p
}

func (p *ModelScoop[M]) RightLike(column string, value string) *ModelScoop[M] {
	p.cond.where(column, "LIKE", "%"+value)
	return p
}

func (p *ModelScoop[M]) NotLike(column string, value string) *ModelScoop[M] {
	p.cond.where(column, "NOT LIKE", "%"+value+"%")
	return p
}

func (p *ModelScoop[M]) NotLeftLike(column string, value string) *ModelScoop[M] {
	p.cond.where(column, "NOT LIKE", value+"%")
	return p
}

func (p *ModelScoop[M]) NotRightLike(column string, value string) *ModelScoop[M] {
	p.cond.where(column, "NOT LIKE", "%"+value)
	return p
}

func (p *ModelScoop[M]) Between(column string, min, max interface{}) *ModelScoop[M] {
	p.cond.whereRaw(quoteFieldName(column)+" BETWEEN ? AND ?", min, max)
	return p
}

func (p *ModelScoop[M]) NotBetween(column string, min, max interface{}) *ModelScoop[M] {
	p.cond.whereRaw(quoteFieldName(column)+" NOT BETWEEN ? AND ?", min, max)
	return p
}

func (p *ModelScoop[M]) Unscoped(b ...bool) *ModelScoop[M] {
	if len(b) == 0 {
		p.unscoped = true
		return p
	}
	p.unscoped = b[0]
	return p
}

func (p *ModelScoop[M]) Limit(limit uint64) *ModelScoop[M] {
	p.limit = limit
	return p
}

func (p *ModelScoop[M]) Offset(offset uint64) *ModelScoop[M] {
	p.offset = offset
	return p
}

func (p *ModelScoop[M]) Group(fields ...string) *ModelScoop[M] {
	p.groups = append(p.groups, fields...)
	return p
}

func (p *ModelScoop[M]) Order(fields ...string) *ModelScoop[M] {
	p.orders = append(p.orders, fields...)
	return p
}

func (p *ModelScoop[M]) Desc(fields ...string) *ModelScoop[M] {
	p.orders = append(p.orders, candy.Map(fields, func(s string) string {
		return s + " DESC"
	})...)
	return p
}

func (p *ModelScoop[M]) Asc(fields ...string) *ModelScoop[M] {
	p.orders = append(p.orders, candy.Map(fields, func(s string) string {
		return s + " ASC"
	})...)
	return p
}

func (p *ModelScoop[M]) Ignore(b ...bool) *ModelScoop[M] {
	if len(b) == 0 {
		p.ignore = true
		return p
	}

	p.ignore = b[0]

	return p
}

// ——————————操作——————————

func (p *ModelScoop[M]) First() (*M, error) {
	p.inc()
	defer p.dec()

	var m M
	err := p.Scoop.First(&m).Error
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return &m, nil
}

func (p *ModelScoop[M]) Find() ([]*M, error) {
	p.inc()
	defer p.dec()

	var ms []*M
	err := p.Scoop.Find(&ms).Error
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return ms, nil
}

func (p *ModelScoop[M]) Create(m *M) error {
	if m == nil {
		err := fmt.Errorf("Create failed: input parameter m is nil")
		log.Errorf("err:%v", err)
		return err
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
		err := fmt.Errorf("FirstOrCreate failed: input parameter m is nil")
		log.Errorf("err:%v", err)
		return &FirstOrCreateResult[M]{
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
				if p.IsDuplicatedKeyError(err) {
					err = p.Scoop.First(&mm).Error
					if err != nil {
						if p.IsNotFound(err) {
							log.Errorf("err:%v", p.getDuplicatedKeyError())
							return &FirstOrCreateResult[M]{
								Error: p.getDuplicatedKeyError(),
							}
						}

						log.Errorf("err:%v", err)
						return &FirstOrCreateResult[M]{
							Error: err,
						}
					}

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
			return &FirstOrCreateResult[M]{
				IsCreated: true,
				Object:    m,
			}
		}
		log.Errorf("err:%v", err)
		return &FirstOrCreateResult[M]{
			Error: err,
		}
	}
	return &FirstOrCreateResult[M]{
		Object: &mm,
	}
}

type CreateIfNotExistsResult struct {
	IsCreated bool
	Error     error
}

func (p *ModelScoop[M]) CreateIfNotExists(m *M) *CreateIfNotExistsResult {
	if m == nil {
		err := fmt.Errorf("CreateIfNotExists failed: input parameter m is nil")
		log.Errorf("err:%v", err)
		return &CreateIfNotExistsResult{
			Error: err,
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

func (p *ModelScoop[M]) Chunk(size uint64, fc func(tx *Scoop, out []*M, offset uint64) error) *ChunkResult {
	p.inc()
	defer p.dec()

	var out []*M
	return p.Scoop.Chunk(&out, size, func(tx *Scoop, offset uint64) error {
		// Create a copy of the slice to avoid issues when the callback holds a reference
		// The underlying Scoop.Chunk resets the slice on each iteration
		batch := make([]*M, len(out))
		copy(batch, out)
		return fc(tx, batch, offset)
	})
}

func (p *ModelScoop[M]) CreateInBatches(values []*M, batchSize int) *CreateInBatchesResult {
	if values == nil {
		err := fmt.Errorf("CreateInBatches failed: input parameter values is nil")
		log.Errorf("err:%v", err)
		return &CreateInBatchesResult{
			Error: err,
		}
	}

	p.inc()
	defer p.dec()

	return p.Scoop.CreateInBatches(values, batchSize)
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

	// TODO: candy.DeepCopy
	//candy.DeepCopy(&mm, m)

	return &CreateOrUpdateResult[M]{
		Updated: true,
		Object:  &mm,
	}
}

func (p *ModelScoop[M]) FindByPage(opt *core.ListOption) (page *core.Paginate, values []*M, err error) {
	p.inc()
	defer p.dec()

	page, err = p.Scoop.FindByPage(opt, &values)
	if err != nil {
		log.Errorf("err:%v", err)
	}
	return page, values, err
}
