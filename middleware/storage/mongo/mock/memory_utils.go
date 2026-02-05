package mock

import (
	"reflect"
	"regexp"
	"time"

	"github.com/lazygophers/log"
	"go.mongodb.org/mongo-driver/bson"
)

// toBsonM converts any value to bson.M
// If the value is already bson.M, it returns directly
// Otherwise, it uses bson.Marshal/Unmarshal for conversion
func toBsonM(v interface{}) (bson.M, error) {
	if v == nil {
		return bson.M{}, nil
	}

	// If already bson.M, return directly
	if m, ok := v.(bson.M); ok {
		return m, nil
	}

	// Marshal to BSON bytes
	bytes, err := bson.Marshal(v)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	// Unmarshal to bson.M
	var result bson.M
	err = bson.Unmarshal(bytes, &result)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return result, nil
}

// bsonArrayToStruct converts []bson.M to target type
// result must be a pointer to a slice
// Uses bson.Marshal/Unmarshal for conversion
func bsonArrayToStruct(docs []bson.M, result interface{}) error {
	if result == nil {
		log.Errorf("err:result is nil")
		return ErrInvalidArgument
	}

	resultVal := reflect.ValueOf(result)
	if resultVal.Kind() != reflect.Ptr {
		log.Errorf("err:result must be a pointer")
		return ErrInvalidArgument
	}

	resultVal = resultVal.Elem()
	if resultVal.Kind() != reflect.Slice {
		log.Errorf("err:result must be a pointer to slice")
		return ErrInvalidArgument
	}

	// Get the element type of the slice
	elemType := resultVal.Type().Elem()

	// Create a new slice with the same length as docs
	newSlice := reflect.MakeSlice(resultVal.Type(), len(docs), len(docs))

	// Convert each bson.M to the target type
	for i, doc := range docs {
		// Create a new element of the target type
		elem := reflect.New(elemType)

		// Convert bson.M to the element
		err := bsonToStruct(doc, elem.Interface())
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		// Set the element in the slice
		newSlice.Index(i).Set(elem.Elem())
	}

	// Set the result to the new slice
	resultVal.Set(newSlice)

	return nil
}

// bsonToStruct converts a single bson.M to target type
// result must be a pointer
// Uses bson.Marshal/Unmarshal for conversion
func bsonToStruct(doc bson.M, result interface{}) error {
	if result == nil {
		log.Errorf("err:result is nil")
		return ErrInvalidArgument
	}

	resultVal := reflect.ValueOf(result)
	if resultVal.Kind() != reflect.Ptr {
		log.Errorf("err:result must be a pointer")
		return ErrInvalidArgument
	}

	// Marshal bson.M to BSON bytes
	bytes, err := bson.Marshal(doc)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	// Unmarshal to target type
	err = bson.Unmarshal(bytes, result)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

// compare compares two values and returns -1, 0, or 1
// Returns -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
// Supports int, float, string, time.Time types
func compare(v1, v2 interface{}) int {
	if v1 == nil && v2 == nil {
		return 0
	}
	if v1 == nil {
		return -1
	}
	if v2 == nil {
		return 1
	}

	// Try time.Time first
	t1, ok1 := v1.(time.Time)
	t2, ok2 := v2.(time.Time)
	if ok1 && ok2 {
		if t1.Before(t2) {
			return -1
		}
		if t1.After(t2) {
			return 1
		}
		return 0
	}

	// Try numeric comparison
	f1 := toFloat64(v1)
	f2 := toFloat64(v2)
	if f1 < f2 {
		return -1
	}
	if f1 > f2 {
		return 1
	}

	// Try string comparison
	s1, ok1 := v1.(string)
	s2, ok2 := v2.(string)
	if ok1 && ok2 {
		if s1 < s2 {
			return -1
		}
		if s1 > s2 {
			return 1
		}
		return 0
	}

	return 0
}

// toInt64 converts a value to int64
// Supports int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64
func toInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}

	switch val := v.(type) {
	case int:
		return int64(val)
	case int8:
		return int64(val)
	case int16:
		return int64(val)
	case int32:
		return int64(val)
	case int64:
		return val
	case uint:
		return int64(val)
	case uint8:
		return int64(val)
	case uint16:
		return int64(val)
	case uint32:
		return int64(val)
	case uint64:
		return int64(val)
	case float32:
		return int64(val)
	case float64:
		return int64(val)
	default:
		return 0
	}
}

// toFloat64 converts a value to float64
// Supports int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64
func toFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}

	switch val := v.(type) {
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	default:
		return 0
	}
}

// inSlice checks if value is in the slice
// Uses reflection to handle any slice type
func inSlice(value interface{}, sliceVal interface{}) bool {
	if sliceVal == nil {
		return false
	}

	sliceValue := reflect.ValueOf(sliceVal)
	if sliceValue.Kind() != reflect.Slice && sliceValue.Kind() != reflect.Array {
		return false
	}

	for i := 0; i < sliceValue.Len(); i++ {
		item := sliceValue.Index(i).Interface()
		if reflect.DeepEqual(value, item) {
			return true
		}
	}

	return false
}

// matchRegex checks if value matches the regex pattern
// Both value and pattern are converted to string
// Uses regexp.MatchString for matching
func matchRegex(value interface{}, pattern interface{}) bool {
	if value == nil || pattern == nil {
		return false
	}

	// Convert value to string
	var valueStr string
	switch v := value.(type) {
	case string:
		valueStr = v
	default:
		valueStr = ""
	}

	// Convert pattern to string
	var patternStr string
	switch p := pattern.(type) {
	case string:
		patternStr = p
	default:
		return false
	}

	// Compile and match regex
	matched, err := regexp.MatchString(patternStr, valueStr)
	if err != nil {
		log.Errorf("err:%v", err)
		return false
	}

	return matched
}
