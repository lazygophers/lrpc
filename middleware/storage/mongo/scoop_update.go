package mongo

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lazygophers/log"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *Scoop) Updates(update interface{}) *UpdateResult {
	begin := time.Now()

	if s.coll == nil {
		err := fmt.Errorf("collection not set, call Collection(model) or use Find/Create first")
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db collection not set"), 0
		}, err)
		return &UpdateResult{Error: err}
	}

	updateDoc := bson.M{}

	// If update is a map, wrap it in $set
	switch v := update.(type) {
	case bson.M:
		if _, ok := v["$set"]; !ok && len(v) > 0 {
			updateDoc = bson.M{"$set": v}
		} else {
			updateDoc = v
		}
	case map[string]interface{}:
		// Check if it has update operators
		hasOperator := false
		for key := range v {
			if len(key) > 0 && key[0] == '$' {
				hasOperator = true
				break
			}
		}
		if !hasOperator {
			updateDoc = bson.M{"$set": v}
		} else {
			updateDoc = bson.M(v)
		}
	default:
		// Convert to map via JSON
		data, err := json.Marshal(v)
		if err != nil {
			log.Errorf("err:%v", err)
			s.logger.Log(s.depth, begin, func() (string, int64) {
				return fmt.Sprintf("db.%s marshal update", s.coll.Name()), 0
			}, err)
			return &UpdateResult{Error: err}
		}

		var m map[string]interface{}
		err = json.Unmarshal(data, &m)
		if err != nil {
			log.Errorf("err:%v", err)
			s.logger.Log(s.depth, begin, func() (string, int64) {
				return fmt.Sprintf("db.%s unmarshal update", s.coll.Name()), 0
			}, err)
			return &UpdateResult{Error: err}
		}

		updateDoc = bson.M{"$set": m}
	}

	// Auto fill updated_at field
	autoFillUpdateFields(updateDoc)

	result, err := s.coll.UpdateMany(s.getContext(), s.filter.ToBson(), updateDoc)
	if err != nil {
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.updateMany(%v, %v)", s.coll.Name(), s.filter.ToBson(), updateDoc), 0
		}, err)
		return &UpdateResult{Error: err}
	}

	s.logger.Log(s.depth, begin, func() (string, int64) {
		return fmt.Sprintf("db.%s.updateMany(%v, %v)", s.coll.Name(), s.filter.ToBson(), updateDoc), result.ModifiedCount
	}, nil)

	return &UpdateResult{
		DocsAffected: result.ModifiedCount,
		Error:        nil,
	}
}

// Delete deletes documents matching the filter and returns a DeleteResult
