package db

import (
	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/core"
)

// ——————————操作——————————

func (p *ModelScoop[M]) First() (*M, error) {
	p.inc()
	defer p.dec()

	var m M
	err := p.Scoop.First(&m).Error
	if err != nil {
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
		return nil, err
	}

	return ms, nil
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

func (p *ModelScoop[M]) FindByPage(opt *core.ListOption) (page *core.Paginate, values []*M, err error) {
	p.inc()
	defer p.dec()

	page, err = p.Scoop.FindByPage(opt, &values)
	if err != nil {
		log.Errorf("err:%v", err)
	}
	return page, values, err
}

// Scan executes the query and scans the result into dest.
// The dest parameter can be:
//   - A pointer to a basic type (int, string, etc.) for single value queries
//   - A pointer to a slice for multiple values
//   - A pointer to M or []*M for model queries
//
// Returns ScanResult containing any error that occurred.
//
// Example:
//
//	var count int64
//	result := scoop.Select("COUNT(*)").Where("age > ?", 18).Scan(&count)
//
//	var ids []uint64
//	result := scoop.Select("id").Where("status = ?", "active").Scan(&ids)
func (p *ModelScoop[M]) Scan(dest interface{}) *ScanResult {
	p.inc()
	defer p.dec()

	return p.Scoop.Scan(dest)
}
