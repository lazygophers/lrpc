package mongo

import (
	"context"
	"fmt"

	"github.com/kamva/mgm/v3"
	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/core"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *Scoop) Begin() (*Scoop, error) {
	// Lazy initialization: create session only when needed
	if s.session == nil {
		_, mongoClient, _, err := mgm.DefaultConfigs()
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}

		session, err := mongoClient.StartSession()
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}

		s.session = session
	}

	// Start transaction on the session
	err := s.session.StartTransaction()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	// Return a new Scoop with the same session (for transactional operations)
	newScoop := &Scoop{
		client:     s.client,
		coll:       s.coll,
		filter:     NewCond(),
		sort:       bson.M{},
		projection: bson.M{},
		session:    s.session,
	}

	return newScoop, nil
}

// Commit commits the transaction
func (s *Scoop) Commit() error {
	if s.session == nil {
		return fmt.Errorf("no active transaction")
	}

	err := s.session.CommitTransaction(context.Background())
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	s.session.EndSession(context.Background())
	return nil
}

// Rollback aborts/rolls back the transaction
func (s *Scoop) Rollback() error {
	if s.session == nil {
		return fmt.Errorf("no active transaction")
	}

	err := s.session.AbortTransaction(context.Background())
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	s.session.EndSession(context.Background())
	return nil
}

// inc increments the depth counter
func (s *Scoop) inc() {
	s.depth++
}

// dec decrements the depth counter
func (s *Scoop) dec() {
	s.depth--
}

// FindByPage finds documents matching the filter with pagination support
// Returns paginated results along with total count if ShowTotal is true
func (s *Scoop) FindByPage(opt *core.ListOption, values any) (*core.Paginate, error) {
	if opt == nil {
		opt = &core.ListOption{
			Offset: core.DefaultOffset,
			Limit:  core.DefaultLimit,
		}
	}

	s.Offset(int64(opt.Offset)).Limit(int64(opt.Limit))

	page := &core.Paginate{
		Offset: opt.Offset,
		Limit:  opt.Limit,
	}

	s.inc()
	defer s.dec()

	findResult := s.Find(values)
	if findResult.Error != nil {
		log.Errorf("err:%v", findResult.Error)
		return nil, findResult.Error
	}

	if opt.ShowTotal {
		// Create a new scoop for counting to avoid modifying the current one's state
		countScoop := s.Clone()
		count, err := countScoop.Count()
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}
		page.Total = uint64(count)
	}

	return page, nil
}

// AutoMigrates ensures that all provided models have their corresponding collections in MongoDB
// It delegates to the client's AutoMigrates method
func (s *Scoop) AutoMigrates(models ...interface{}) error {
	return s.client.AutoMigrates(models...)
}

// AutoMigrate ensures that a model has its corresponding collection in MongoDB
// It delegates to the client's AutoMigrate method
func (s *Scoop) AutoMigrate(model interface{}) error {
	return s.client.AutoMigrate(model)
}
