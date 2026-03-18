package mongo

import (
	"fmt"
	"time"

	"github.com/lazygophers/log"
)

func (s *Scoop) Create(doc interface{}) error {
	begin := time.Now()

	err := s.ensureCollection(doc)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	// Auto fill id, _id, created_at, updated_at fields
	autoFillCreateFields(doc)

	_, err = s.coll.InsertOne(s.getContext(), doc)
	if err != nil {
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.insertOne(...)", s.coll.Name()), 0
		}, err)
		return err
	}

	s.logger.Log(s.depth, begin, func() (string, int64) {
		return fmt.Sprintf("db.%s.insertOne(...)", s.coll.Name()), 1
	}, nil)

	return nil
}

// BatchCreate inserts multiple documents
func (s *Scoop) BatchCreate(docs ...interface{}) error {
	begin := time.Now()

	if len(docs) == 0 {
		err := fmt.Errorf("no documents to insert")
		log.Errorf("err:%v", err)
		return err
	}

	err := s.ensureCollection(docs[0])
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	// Auto fill id, _id, created_at, updated_at fields for each document
	for _, doc := range docs {
		autoFillCreateFields(doc)
	}

	result, err := s.coll.InsertMany(s.getContext(), docs)
	if err != nil {
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.insertMany(...) [%d docs]", s.coll.Name(), len(docs)), 0
		}, err)
		return err
	}

	s.logger.Log(s.depth, begin, func() (string, int64) {
		return fmt.Sprintf("db.%s.insertMany(...) [%d docs]", s.coll.Name(), len(docs)), int64(len(result.InsertedIDs))
	}, nil)

	return nil
}

// Updates updates documents matching the filter and returns an UpdateResult
