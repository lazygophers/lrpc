package clickhouse

import (
	"database/sql"
	"fmt"
)

type ModelScoop[M any] struct {
	Scoop

	m M
}

func NewModelScoop[M any](db *sql.DB) *ModelScoop[M] {
	scoop := &ModelScoop[M]{
		Scoop: Scoop{
			_db: db,
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

func (p *ModelScoop[M]) Equal(column string, value interface{}) *ModelScoop[M] {
	p.cond.where(column, value)
	return p
}

func (p *ModelScoop[M]) NotEqual(column string, value interface{}) *ModelScoop[M] {
	p.cond.where(column, "!= ", value)
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

func (p *ModelScoop[M]) Ignore(b ...bool) *ModelScoop[M] {
	if len(b) == 0 {
		p.ignore = true
		return p
	}

	p.ignore = b[0]

	return p
}

// ——————————操作——————————

//func (p *ModelScoop[M]) First() (*M, error) {
//	p.inc()
//	defer p.dec()
//
//	var m M
//	err := p.Scoop.First(&m).Error
//	if err != nil {
//		return nil, err
//	}
//
//	return &m, nil
//}

func (p *ModelScoop[M]) Find() ([]*M, error) {
	p.inc()
	defer p.dec()

	var ms []*M
	err := p.Scoop.Find(&ms).Error
	if err != nil {
		return nil, err
	}

	return ms, nil
}

//func (p *ModelScoop[M]) Create(m *M) error {
//	p.inc()
//	defer p.dec()
//
//	return p.Scoop.Create(m).Error
//}
//
//type FirstOrCreateResult[M any] struct {
//	IsCreated bool
//	Error     error
//
//	Object *M
//}
//
//func (p *ModelScoop[M]) FirstOrCreate(m *M) *FirstOrCreateResult[M] {
//	p.inc()
//	defer p.dec()
//
//	var mm M
//	err := p.Scoop.First(&mm).Error
//	if err != nil {
//		if p.IsNotFound(err) {
//			err = p.Scoop.Create(m).Error
//			if err != nil {
//				return &FirstOrCreateResult[M]{
//					Error: err,
//				}
//			}
//			return &FirstOrCreateResult[M]{
//				IsCreated: true,
//				Object:    m,
//			}
//		}
//		return &FirstOrCreateResult[M]{
//			Error: err,
//		}
//	}
//	return &FirstOrCreateResult[M]{
//		Object: &mm,
//	}
//}
//
//type UpdateOrCreateResult[M any] struct {
//	IsCreated bool
//	Error     error
//
//	Object *M
//}
//
//func (p *ModelScoop[M]) UpdateOrCreate(values map[string]interface{}, m *M) *UpdateOrCreateResult[M] {
//	p.inc()
//	defer p.dec()
//
//	var mm M
//	err := p.Scoop.First(&mm).Error
//	if err != nil {
//		if p.IsNotFound(err) {
//			err = p.Scoop.Create(m).Error
//			if err != nil {
//				return &UpdateOrCreateResult[M]{
//					Error: err,
//				}
//			}
//			return &UpdateOrCreateResult[M]{
//				IsCreated: true,
//				Object:    m,
//			}
//		}
//
//		return &UpdateOrCreateResult[M]{
//			Error: err,
//		}
//	}
//
//	err = p.Scoop.update(values).Error
//	if err != nil {
//		return &UpdateOrCreateResult[M]{
//			Error: err,
//		}
//	}
//
//	err = p.Scoop.First(&mm).Error
//	if err != nil {
//		return &UpdateOrCreateResult[M]{
//			Error: err,
//		}
//	}
//
//	candy.DeepCopy(&mm, m)
//
//	return &UpdateOrCreateResult[M]{
//		Object: &mm,
//	}
//}
//
//type CreateNotExistResult[M any] struct {
//	IsCreated bool
//	Error     error
//
//	Object *M
//}
//
//func (p *ModelScoop[M]) CreateNotExist(m *M) *CreateNotExistResult[M] {
//	p.inc()
//	defer p.dec()
//
//	var mm M
//	err := p.Scoop.First(&mm).Error
//	if err != nil {
//		if p.IsNotFound(err) {
//			err = p.Scoop.Create(m).Error
//			if err != nil {
//				return &CreateNotExistResult[M]{
//					Error: err,
//				}
//			}
//			return &CreateNotExistResult[M]{
//				IsCreated: true,
//				Object:    m,
//			}
//		}
//		return &CreateNotExistResult[M]{
//			Error: err,
//		}
//	}
//
//	candy.DeepCopy(&mm, m)
//
//	return &CreateNotExistResult[M]{
//		Object: &mm,
//	}
//}
//
//func (p *ModelScoop[M]) Chunk(size uint64, fc func(tx *Scoop, out []*M, offset uint64) error) *ChunkResult {
//	p.inc()
//	defer p.dec()
//
//	var out []*M
//	return p.Scoop.Chunk(&out, size, func(tx *Scoop, offset uint64) error {
//		return fc(tx, out, offset)
//	})
//}
//
//type CreateOrUpdateResult[M any] struct {
//	Error  error
//	Object *M
//
//	Created bool
//	Updated bool
//}
//
//func (p *ModelScoop[M]) CreateOrUpdate(values map[string]interface{}, m *M) *CreateOrUpdateResult[M] {
//	p.inc()
//	defer p.dec()
//
//	var mm M
//	err := p.Scoop.First(&mm).Error
//	if err != nil {
//		if p.IsNotFound(err) {
//			err = p.Scoop.Create(m).Error
//			if err != nil {
//				return &CreateOrUpdateResult[M]{
//					Error: err,
//				}
//			}
//			return &CreateOrUpdateResult[M]{
//				Created: true,
//				Object:  m,
//			}
//		}
//
//		return &CreateOrUpdateResult[M]{
//			Error: err,
//		}
//	}
//
//	err = p.Scoop.update(values).Error
//	if err != nil {
//		return &CreateOrUpdateResult[M]{
//			Error: err,
//		}
//	}
//
//	err = p.Scoop.First(&mm).Error
//	if err != nil {
//		return &CreateOrUpdateResult[M]{
//			Error: err,
//		}
//	}
//
//	candy.DeepCopy(&mm, m)
//
//	return &CreateOrUpdateResult[M]{
//		Object: &mm,
//	}
//}
//
//func (p *ModelScoop[M]) FindByPage(opt *core.ListOption) (page *core.Paginate, values []*M, err error) {
//	p.inc()
//	defer p.dec()
//
//	page, err = p.Scoop.FindByPage(opt, &values)
//	return page, values, err
//}
