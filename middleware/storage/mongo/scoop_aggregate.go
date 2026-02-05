package mongo

import (
	"github.com/lazygophers/log"
	"go.mongodb.org/mongo-driver/bson"
)

// Sum calculates the sum of a field
// Returns the sum value or 0 if collection is empty
func (s *Scoop) Sum(field string) (float64, error) {
	err := s.ensureCollection(nil)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	pipeline := bson.A{
		bson.M{"$match": s.filter.ToBson()},
		bson.M{"$group": bson.M{
			"_id": nil,
			"sum": bson.M{"$sum": "$" + field},
		}},
	}

	var result []bson.M
	coll := s.coll
	cursor, err := coll.Aggregate(s.getContext(), pipeline)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	defer cursor.Close(s.getContext())

	err = cursor.All(s.getContext(), &result)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	if len(result) == 0 {
		return 0, nil
	}

	sum, ok := result[0]["sum"]
	if !ok {
		return 0, nil
	}

	// Handle different numeric types
	switch v := sum.(type) {
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return 0, nil
	}
}

// Avg calculates the average of a field
// Returns the average value or 0 if collection is empty
func (s *Scoop) Avg(field string) (float64, error) {
	err := s.ensureCollection(nil)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	pipeline := bson.A{
		bson.M{"$match": s.filter.ToBson()},
		bson.M{"$group": bson.M{
			"_id": nil,
			"avg": bson.M{"$avg": "$" + field},
		}},
	}

	var result []bson.M
	coll := s.coll
	cursor, err := coll.Aggregate(s.getContext(), pipeline)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	defer cursor.Close(s.getContext())

	err = cursor.All(s.getContext(), &result)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	if len(result) == 0 {
		return 0, nil
	}

	avg, ok := result[0]["avg"]
	if !ok {
		return 0, nil
	}

	// Handle different numeric types
	switch v := avg.(type) {
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return 0, nil
	}
}

// Max finds the maximum value of a field
// Returns the max value or 0 if collection is empty
func (s *Scoop) Max(field string) (float64, error) {
	err := s.ensureCollection(nil)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	pipeline := bson.A{
		bson.M{"$match": s.filter.ToBson()},
		bson.M{"$group": bson.M{
			"_id": nil,
			"max": bson.M{"$max": "$" + field},
		}},
	}

	var result []bson.M
	coll := s.coll
	cursor, err := coll.Aggregate(s.getContext(), pipeline)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	defer cursor.Close(s.getContext())

	err = cursor.All(s.getContext(), &result)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	if len(result) == 0 {
		return 0, nil
	}

	max, ok := result[0]["max"]
	if !ok {
		return 0, nil
	}

	// Handle different numeric types
	switch v := max.(type) {
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return 0, nil
	}
}

// Min finds the minimum value of a field
// Returns the min value or 0 if collection is empty
func (s *Scoop) Min(field string) (float64, error) {
	err := s.ensureCollection(nil)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	pipeline := bson.A{
		bson.M{"$match": s.filter.ToBson()},
		bson.M{"$group": bson.M{
			"_id": nil,
			"min": bson.M{"$min": "$" + field},
		}},
	}

	var result []bson.M
	coll := s.coll
	cursor, err := coll.Aggregate(s.getContext(), pipeline)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	defer cursor.Close(s.getContext())

	err = cursor.All(s.getContext(), &result)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	if len(result) == 0 {
		return 0, nil
	}

	min, ok := result[0]["min"]
	if !ok {
		return 0, nil
	}

	// Handle different numeric types
	switch v := min.(type) {
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return 0, nil
	}
}
