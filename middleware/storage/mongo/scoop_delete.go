package mongo

import (
	"fmt"
	"time"

	"github.com/lazygophers/log"
)

func (s *Scoop) Delete() *DeleteResult {
	begin := time.Now()

	result, err := s.coll.DeleteMany(s.getContext(), s.filter.ToBson())
	if err != nil {
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.deleteMany(%v)", s.coll.Name(), s.filter.ToBson()), 0
		}, err)
		return &DeleteResult{Error: err}
	}

	s.logger.Log(s.depth, begin, func() (string, int64) {
		return fmt.Sprintf("db.%s.deleteMany(%v)", s.coll.Name(), s.filter.ToBson()), result.DeletedCount
	}, nil)

	return &DeleteResult{
		DocsAffected: result.DeletedCount,
		Error:        nil,
	}
}

// Aggregate creates an aggregation pipeline
