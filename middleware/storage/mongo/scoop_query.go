package mongo

import (
	"fmt"
	"reflect"
	"time"

	"github.com/lazygophers/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Find finds documents matching the filter and returns a FindResult
func (s *Scoop) Find(result interface{}) *FindResult {
	begin := time.Now()
	var docsCount int64

	err := s.ensureCollection(result)
	if err != nil {
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.find(%v)", s.coll.Name(), s.filter.ToBson()), 0
		}, err)
		return &FindResult{Error: err}
	}

	// Build FindOptions from scoop fields
	opts := options.Find()
	if s.limit != nil {
		opts.SetLimit(*s.limit)
	}
	if s.offset != nil {
		opts.SetSkip(*s.offset)
	}
	if len(s.sort) > 0 {
		opts.SetSort(s.sort)
	}
	if len(s.projection) > 0 {
		opts.SetProjection(s.projection)
	}

	ctx := s.getContext()
	var cursor MongoCursor
	cursor, err = s.coll.Find(ctx, s.filter.ToBson(), opts)
	if err != nil {
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.find(%v)", s.coll.Name(), s.filter.ToBson()), 0
		}, err)
		return &FindResult{Error: err}
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, result)
	if err != nil {
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.find(%v)", s.coll.Name(), s.filter.ToBson()), 0
		}, err)
		return &FindResult{Error: err}
	}

	// Count the documents returned
	if resultVal := reflect.ValueOf(result); resultVal.Kind() == reflect.Ptr && resultVal.Elem().Kind() == reflect.Slice {
		docsCount = int64(resultVal.Elem().Len())
	}

	s.logger.Log(s.depth, begin, func() (string, int64) {
		return fmt.Sprintf("db.%s.find(%v)", s.coll.Name(), s.filter.ToBson()), docsCount
	}, nil)

	return &FindResult{
		DocsAffected: docsCount,
		Error:        nil,
	}
}

// First finds a single document and returns a FirstResult
func (s *Scoop) First(result interface{}) *FirstResult {
	begin := time.Now()

	err := s.ensureCollection(result)
	if err != nil {
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.findOne(%v)", s.coll.Name(), s.filter.ToBson()), 0
		}, err)
		return &FirstResult{Error: err}
	}

	ctx := s.getContext()
	opts := options.FindOne()
	if len(s.projection) > 0 {
		opts.SetProjection(s.projection)
	}
	sr := s.coll.FindOne(ctx, s.filter.ToBson(), opts)
	if sr.Err() != nil {
		log.Errorf("err:%v", sr.Err())
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.findOne(%v)", s.coll.Name(), s.filter.ToBson()), 0
		}, sr.Err())
		return &FirstResult{Error: sr.Err()}
	}

	err = sr.Decode(result)
	if err != nil {
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.findOne(%v)", s.coll.Name(), s.filter.ToBson()), 0
		}, err)
		return &FirstResult{Error: err}
	}

	s.logger.Log(s.depth, begin, func() (string, int64) {
		return fmt.Sprintf("db.%s.findOne(%v)", s.coll.Name(), s.filter.ToBson()), 1
	}, nil)

	return &FirstResult{Error: nil}
}

// Count counts documents matching the filter
func (s *Scoop) Count() (int64, error) {
	begin := time.Now()

	count, err := s.coll.CountDocuments(s.getContext(), s.filter.ToBson())
	if err != nil {
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.countDocuments(%v)", s.coll.Name(), s.filter.ToBson()), 0
		}, err)
		return 0, err
	}

	s.logger.Log(s.depth, begin, func() (string, int64) {
		return fmt.Sprintf("db.%s.countDocuments(%v)", s.coll.Name(), s.filter.ToBson()), count
	}, nil)

	return count, nil
}

// Exist checks if documents matching the filter exist by fetching only _id field
func (s *Scoop) Exist() (bool, error) {
	// Clone scoop and select only _id field for efficiency
	scoop := s.Clone()
	scoop.Select("_id")

	ctx := scoop.getContext()
	opts := options.FindOne()
	opts.SetProjection(scoop.projection)

	sr := scoop.coll.FindOne(ctx, scoop.filter.ToBson(), opts)
	if sr.Err() != nil {
		if sr.Err() == mongo.ErrNoDocuments {
			return false, nil
		}
		log.Errorf("err:%v", sr.Err())
		return false, sr.Err()
	}

	return true, nil
}

// Create inserts a new document
