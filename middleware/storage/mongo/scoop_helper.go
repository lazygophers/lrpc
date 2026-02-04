package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Scoop) Aggregate(pipeline ...bson.M) *Aggregation {
	return NewAggregation(s.client, s.coll.Name(), s.getContext(), pipeline...)
}

// Clone creates a copy of the scoop with current state
func (s *Scoop) Clone() *Scoop {
	newScoop := &Scoop{
		client:        s.client,
		coll:          s.coll,
		filter:        NewCond(),
		sort:          bson.M{},
		projection:    bson.M{},
		session:       s.session,
		notFoundError: s.notFoundError,
		depth:         s.depth,
		logger:        s.logger,
	}

	// Deep copy filter conditions
	if s.filter != nil && len(s.filter.conds) > 0 {
		newScoop.filter.conds = make([]bson.M, len(s.filter.conds))
		for i, cond := range s.filter.conds {
			// Deep copy each BSON condition
			newScoop.filter.conds[i] = make(bson.M)
			for k, v := range cond {
				newScoop.filter.conds[i][k] = v
			}
		}
		newScoop.filter.isOr = s.filter.isOr
	}

	// Copy limit and offset
	if s.limit != nil {
		newScoop.limit = s.limit
	}
	if s.offset != nil {
		newScoop.offset = s.offset
	}

	// Deep copy sort
	if len(s.sort) > 0 {
		newScoop.sort = make(bson.M)
		for k, v := range s.sort {
			newScoop.sort[k] = v
		}
	}

	// Deep copy projection
	if len(s.projection) > 0 {
		newScoop.projection = make(bson.M)
		for k, v := range s.projection {
			newScoop.projection[k] = v
		}
	}

	return newScoop
}

// Clear resets the scoop
func (s *Scoop) Clear() *Scoop {
	s.filter = NewCond()
	s.limit = nil
	s.offset = nil
	s.sort = bson.M{}
	s.projection = bson.M{}
	return s
}

// GetCollection returns the underlying MongoDB collection
func (s *Scoop) GetCollection() *mongo.Collection {
	return s.coll
}

// SetNotFound sets the not found error for this scoop
func (s *Scoop) SetNotFound(err error) *Scoop {
	s.notFoundError = err
	return s
}

// IsNotFound checks if the error is a not found error
func (s *Scoop) IsNotFound(err error) bool {
	return err == s.notFoundError || err == mongo.ErrNoDocuments
}

// Begin starts a transaction - creates session lazily if needed
