package mongo

import (
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
func (s *Scoop) GetCollection() MongoCollection {
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

// autoFillCreateFields automatically fills id, _id, created_at, updated_at fields for Create operations
func autoFillCreateFields(doc interface{}) {
	if doc == nil {
		return
	}

	now := time.Now().Unix()
	objectID := primitive.NewObjectID()

	// Handle bson.M directly
	if m, ok := doc.(bson.M); ok {
		autoFillBsonMCreate(m, objectID, now)
		return
	}

	// Handle map[string]interface{} directly
	if m, ok := doc.(map[string]interface{}); ok {
		autoFillMapCreate(m, objectID, now)
		return
	}

	// Handle struct via reflection
	autoFillStructCreate(doc, objectID, now)
}

// autoFillBsonMCreate fills fields for bson.M type
func autoFillBsonMCreate(m bson.M, objectID primitive.ObjectID, now int64) {
	// Fill _id if not exists or is zero
	if _, exists := m["_id"]; !exists {
		m["_id"] = objectID
	} else if isZeroValue(m["_id"]) {
		m["_id"] = objectID
	}

	// Fill id if not exists or is zero (string type)
	if _, exists := m["id"]; !exists {
		m["id"] = objectID.Hex()
	} else if v, ok := m["id"].(string); ok && v == "" {
		m["id"] = objectID.Hex()
	}

	// Fill created_at if not exists or is zero (int64 type)
	if _, exists := m["created_at"]; !exists {
		m["created_at"] = now
	} else if v, ok := m["created_at"].(int64); ok && v == 0 {
		m["created_at"] = now
	}

	// Fill updated_at if not exists or is zero (int64 type)
	if _, exists := m["updated_at"]; !exists {
		m["updated_at"] = now
	} else if v, ok := m["updated_at"].(int64); ok && v == 0 {
		m["updated_at"] = now
	}
}

// autoFillMapCreate fills fields for map[string]interface{} type
func autoFillMapCreate(m map[string]interface{}, objectID primitive.ObjectID, now int64) {
	// Fill _id if not exists or is zero
	if _, exists := m["_id"]; !exists {
		m["_id"] = objectID
	} else if isZeroValue(m["_id"]) {
		m["_id"] = objectID
	}

	// Fill id if not exists or is zero (string type)
	if _, exists := m["id"]; !exists {
		m["id"] = objectID.Hex()
	} else if v, ok := m["id"].(string); ok && v == "" {
		m["id"] = objectID.Hex()
	}

	// Fill created_at if not exists or is zero (int64 type)
	if _, exists := m["created_at"]; !exists {
		m["created_at"] = now
	} else if v, ok := m["created_at"].(int64); ok && v == 0 {
		m["created_at"] = now
	}

	// Fill updated_at if not exists or is zero (int64 type)
	if _, exists := m["updated_at"]; !exists {
		m["updated_at"] = now
	} else if v, ok := m["updated_at"].(int64); ok && v == 0 {
		m["updated_at"] = now
	}
}

// autoFillStructCreate fills fields for struct type via reflection
func autoFillStructCreate(doc interface{}, objectID primitive.ObjectID, now int64) {
	val := reflect.ValueOf(doc)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !field.CanSet() {
			continue
		}

		// Get bson tag name
		bsonTag := fieldType.Tag.Get("bson")
		if bsonTag == "" || bsonTag == "-" {
			continue
		}

		// Parse bson tag to get field name
		fieldName := parseBsonTag(bsonTag)

		// Handle _id field (primitive.ObjectID type)
		if fieldName == "_id" && field.Type() == reflect.TypeOf(primitive.ObjectID{}) {
			if field.IsZero() {
				field.Set(reflect.ValueOf(objectID))
			}
			continue
		}

		// Handle id field (string type)
		if fieldName == "id" && field.Kind() == reflect.String {
			if field.IsZero() {
				field.SetString(objectID.Hex())
			}
			continue
		}

		// Handle created_at field (int64 type)
		if fieldName == "created_at" && field.Kind() == reflect.Int64 {
			if field.IsZero() {
				field.SetInt(now)
			}
			continue
		}

		// Handle updated_at field (int64 type)
		if fieldName == "updated_at" && field.Kind() == reflect.Int64 {
			if field.IsZero() {
				field.SetInt(now)
			}
			continue
		}
	}
}

// autoFillUpdateFields automatically adds updated_at field to update operations
func autoFillUpdateFields(updateDoc bson.M) {
	now := time.Now().Unix()

	// If updateDoc has $set operator, add updated_at to it
	if setOp, exists := updateDoc["$set"]; exists {
		if setMap, ok := setOp.(bson.M); ok {
			// Only add if not already exists or is zero and type matches
			if _, hasField := setMap["updated_at"]; !hasField {
				setMap["updated_at"] = now
			} else if v, ok := setMap["updated_at"].(int64); ok && v == 0 {
				setMap["updated_at"] = now
			}
		} else if setMap, ok := setOp.(map[string]interface{}); ok {
			// Only add if not already exists or is zero and type matches
			if _, hasField := setMap["updated_at"]; !hasField {
				setMap["updated_at"] = now
			} else if v, ok := setMap["updated_at"].(int64); ok && v == 0 {
				setMap["updated_at"] = now
			}
		}
	} else {
		// If no $set operator, create one with updated_at
		updateDoc["$set"] = bson.M{"updated_at": now}
	}
}

// parseBsonTag parses bson tag to extract field name
func parseBsonTag(tag string) string {
	// bson tag format: "fieldname,omitempty" or just "fieldname"
	for i, c := range tag {
		if c == ',' {
			return tag[:i]
		}
	}
	return tag
}

// isZeroValue checks if a value is zero value
func isZeroValue(v interface{}) bool {
	if v == nil {
		return true
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.String:
		return val.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Bool:
		return !val.Bool()
	case reflect.Ptr, reflect.Interface:
		return val.IsNil()
	case reflect.Slice, reflect.Map:
		return val.Len() == 0
	default:
		return val.IsZero()
	}
}
