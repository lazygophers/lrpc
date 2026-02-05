package mock

import (
	"fmt"
	"reflect"

	"github.com/lazygophers/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// convertPipelineToBsonM converts pipeline interface to []bson.M
func convertPipelineToBsonM(pipeline interface{}) ([]bson.M, error) {
	if pipeline == nil {
		return []bson.M{}, nil
	}

	// Check if already []bson.M
	if bsonPipeline, ok := pipeline.([]bson.M); ok {
		return bsonPipeline, nil
	}

	// Try to convert using reflection
	pipelineVal := reflect.ValueOf(pipeline)
	if pipelineVal.Kind() != reflect.Slice {
		err := fmt.Errorf("pipeline must be a slice, got %T", pipeline)
		log.Errorf("err:%v", err)
		return nil, err
	}

	result := make([]bson.M, pipelineVal.Len())
	for i := 0; i < pipelineVal.Len(); i++ {
		stageVal := pipelineVal.Index(i).Interface()

		// Try direct conversion to bson.M
		if stage, ok := stageVal.(bson.M); ok {
			result[i] = stage
			continue
		}

		// Try using toBsonM helper
		stage, err := toBsonM(stageVal)
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}
		result[i] = stage
	}

	return result, nil
}

// processAggregationPipeline processes an aggregation pipeline on a set of documents
// Supports common MongoDB aggregation stages: $match, $project, $sort, $limit, $skip, $group, $unwind
// Returns the processed documents and any error encountered
func processAggregationPipeline(docs []bson.M, pipeline []bson.M) ([]bson.M, error) {
	if len(pipeline) == 0 {
		return docs, nil
	}

	result := docs
	var err error

	for i, stage := range pipeline {
		// Each stage should have exactly one operator
		if len(stage) == 0 {
			err := fmt.Errorf("aggregation stage %d is empty", i)
			log.Errorf("err:%v", err)
			return nil, err
		}

		if len(stage) > 1 {
			err := fmt.Errorf("aggregation stage %d has multiple operators", i)
			log.Errorf("err:%v", err)
			return nil, err
		}

		// Process each stage
		for operator, operatorValue := range stage {
			result, err = processAggregationStage(result, operator, operatorValue)
			if err != nil {
				log.Errorf("err:%v", err)
				return nil, err
			}
		}
	}

	return result, nil
}

// processAggregationStage processes a single aggregation stage
func processAggregationStage(docs []bson.M, operator string, value interface{}) ([]bson.M, error) {
	switch operator {
	case "$match":
		return processMatch(docs, value)

	case "$project":
		return processProject(docs, value)

	case "$sort":
		return processSort(docs, value)

	case "$limit":
		return processLimit(docs, value)

	case "$skip":
		return processSkip(docs, value)

	case "$group":
		return processGroup(docs, value)

	case "$unwind":
		return processUnwind(docs, value)

	default:
		err := fmt.Errorf("unsupported aggregation operator: %s", operator)
		log.Errorf("err:%v", err)
		return nil, err
	}
}

// processMatch applies $match stage - filters documents
func processMatch(docs []bson.M, value interface{}) ([]bson.M, error) {
	filterMap, ok := value.(bson.M)
	if !ok {
		err := fmt.Errorf("$match value must be bson.M")
		log.Errorf("err:%v", err)
		return nil, err
	}

	storage := NewMemoryStorage()
	var matched []bson.M

	for _, doc := range docs {
		if storage.matchFilter(doc, filterMap) {
			matched = append(matched, doc)
		}
	}

	return matched, nil
}

// processProject applies $project stage - selects/transforms fields
func processProject(docs []bson.M, value interface{}) ([]bson.M, error) {
	projectionMap, ok := value.(bson.M)
	if !ok {
		err := fmt.Errorf("$project value must be bson.M")
		log.Errorf("err:%v", err)
		return nil, err
	}

	storage := NewMemoryStorage()
	return storage.applyProjection(docs, projectionMap), nil
}

// processSort applies $sort stage - sorts documents
func processSort(docs []bson.M, value interface{}) ([]bson.M, error) {
	sortMap, ok := value.(bson.M)
	if !ok {
		err := fmt.Errorf("$sort value must be bson.M")
		log.Errorf("err:%v", err)
		return nil, err
	}

	storage := NewMemoryStorage()
	return storage.applySorting(docs, sortMap), nil
}

// processLimit applies $limit stage - limits number of documents
func processLimit(docs []bson.M, value interface{}) ([]bson.M, error) {
	var limit int64

	switch v := value.(type) {
	case int:
		limit = int64(v)
	case int32:
		limit = int64(v)
	case int64:
		limit = v
	case float64:
		limit = int64(v)
	default:
		err := fmt.Errorf("$limit value must be numeric, got %T", value)
		log.Errorf("err:%v", err)
		return nil, err
	}

	if limit < 0 {
		err := fmt.Errorf("$limit value must be non-negative")
		log.Errorf("err:%v", err)
		return nil, err
	}

	if limit == 0 || int64(len(docs)) <= limit {
		return docs, nil
	}

	return docs[:limit], nil
}

// processSkip applies $skip stage - skips number of documents
func processSkip(docs []bson.M, value interface{}) ([]bson.M, error) {
	var skip int64

	switch v := value.(type) {
	case int:
		skip = int64(v)
	case int32:
		skip = int64(v)
	case int64:
		skip = v
	case float64:
		skip = int64(v)
	default:
		err := fmt.Errorf("$skip value must be numeric, got %T", value)
		log.Errorf("err:%v", err)
		return nil, err
	}

	if skip < 0 {
		err := fmt.Errorf("$skip value must be non-negative")
		log.Errorf("err:%v", err)
		return nil, err
	}

	if skip == 0 || int64(len(docs)) <= skip {
		if skip >= int64(len(docs)) {
			return []bson.M{}, nil
		}
		return docs, nil
	}

	return docs[skip:], nil
}

// processGroup applies $group stage - groups documents and performs aggregations
func processGroup(docs []bson.M, value interface{}) ([]bson.M, error) {
	groupSpec, ok := value.(bson.M)
	if !ok {
		err := fmt.Errorf("$group value must be bson.M")
		log.Errorf("err:%v", err)
		return nil, err
	}

	// Get _id field (grouping key)
	idSpec, hasId := groupSpec["_id"]
	if !hasId {
		err := fmt.Errorf("$group requires _id field")
		log.Errorf("err:%v", err)
		return nil, err
	}

	// Group documents by key
	groups := make(map[string]*groupState)

	for _, doc := range docs {
		// Calculate group key
		groupKey, err := evaluateGroupKey(doc, idSpec)
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}

		// Get or create group state
		keyStr := formatGroupKey(groupKey)
		state, exists := groups[keyStr]
		if !exists {
			state = &groupState{
				key:        groupKey,
				docs:       []bson.M{},
				fieldStats: make(map[string]*fieldStat),
			}
			groups[keyStr] = state
		}

		// Add document to group
		state.docs = append(state.docs, doc)

		// Process accumulator fields
		for field, accSpec := range groupSpec {
			if field == "_id" {
				continue
			}

			err := processAccumulator(state, doc, field, accSpec)
			if err != nil {
				log.Errorf("err:%v", err)
				return nil, err
			}
		}
	}

	// Build result documents
	var result []bson.M
	for _, state := range groups {
		resultDoc := bson.M{
			"_id": state.key,
		}

		// Add accumulated fields
		for field, stat := range state.fieldStats {
			resultDoc[field] = stat.result
		}

		result = append(result, resultDoc)
	}

	return result, nil
}

// groupState holds state for a group during aggregation
type groupState struct {
	key        interface{}
	docs       []bson.M
	fieldStats map[string]*fieldStat
}

// fieldStat holds statistics for a field during group aggregation
type fieldStat struct {
	operator string
	result   interface{}
	count    int64
	sum      float64
	values   []interface{}
}

// evaluateGroupKey evaluates the _id expression for grouping
func evaluateGroupKey(doc bson.M, idSpec interface{}) (interface{}, error) {
	// Handle null grouping (all documents in one group)
	if idSpec == nil {
		return nil, nil
	}

	// Handle simple field reference like "$fieldName"
	if fieldRef, ok := idSpec.(string); ok && len(fieldRef) > 0 && fieldRef[0] == '$' {
		fieldName := fieldRef[1:]
		storage := NewMemoryStorage()
		value, exists := storage.getNestedValue(doc, fieldName)
		if !exists {
			return nil, nil
		}
		return value, nil
	}

	// Handle complex grouping with multiple fields
	if groupDoc, ok := idSpec.(bson.M); ok {
		result := make(bson.M)
		for key, fieldSpec := range groupDoc {
			if fieldRef, ok := fieldSpec.(string); ok && len(fieldRef) > 0 && fieldRef[0] == '$' {
				fieldName := fieldRef[1:]
				storage := NewMemoryStorage()
				value, exists := storage.getNestedValue(doc, fieldName)
				if exists {
					result[key] = value
				} else {
					result[key] = nil
				}
			} else {
				result[key] = fieldSpec
			}
		}
		return result, nil
	}

	// Handle literal values
	return idSpec, nil
}

// formatGroupKey converts a group key to a string for map indexing
func formatGroupKey(key interface{}) string {
	if key == nil {
		return "<null>"
	}

	// Use BSON marshaling for stable key representation
	bytes, err := bson.Marshal(bson.M{"k": key})
	if err != nil {
		log.Errorf("err:%v", err)
		return fmt.Sprintf("%v", key)
	}

	return string(bytes)
}

// processAccumulator processes an accumulator expression for a field
func processAccumulator(state *groupState, doc bson.M, field string, accSpec interface{}) error {
	accDoc, ok := accSpec.(bson.M)
	if !ok {
		err := fmt.Errorf("accumulator for field %s must be bson.M", field)
		log.Errorf("err:%v", err)
		return err
	}

	// Get field stat or create new one
	stat, exists := state.fieldStats[field]
	if !exists {
		stat = &fieldStat{
			values: []interface{}{},
		}
		state.fieldStats[field] = stat
	}

	// Process each accumulator operator
	for operator, operatorValue := range accDoc {
		switch operator {
		case "$sum":
			return processSum(state, doc, stat, operatorValue)

		case "$avg":
			return processAvg(state, doc, stat, operatorValue)

		case "$min":
			return processMin(state, doc, stat, operatorValue)

		case "$max":
			return processMax(state, doc, stat, operatorValue)

		case "$count":
			return processCount(state, stat)

		case "$first":
			return processFirst(state, doc, stat, operatorValue)

		case "$last":
			return processLast(state, doc, stat, operatorValue)

		default:
			err := fmt.Errorf("unsupported accumulator operator: %s", operator)
			log.Errorf("err:%v", err)
			return err
		}
	}

	return nil
}

// processSum processes $sum accumulator
func processSum(state *groupState, doc bson.M, stat *fieldStat, value interface{}) error {
	stat.operator = "$sum"

	// Handle constant value (like $sum: 1)
	if numVal, ok := toNumericValue(value); ok {
		stat.sum += numVal
		stat.result = stat.sum
		return nil
	}

	// Handle field reference (like $sum: "$fieldName")
	if fieldRef, ok := value.(string); ok && len(fieldRef) > 0 && fieldRef[0] == '$' {
		fieldName := fieldRef[1:]
		storage := NewMemoryStorage()
		fieldValue, exists := storage.getNestedValue(doc, fieldName)
		if exists {
			if numVal, ok := toNumericValue(fieldValue); ok {
				stat.sum += numVal
			}
		}
		stat.result = stat.sum
		return nil
	}

	err := fmt.Errorf("$sum value must be numeric or field reference")
	log.Errorf("err:%v", err)
	return err
}

// processAvg processes $avg accumulator
func processAvg(state *groupState, doc bson.M, stat *fieldStat, value interface{}) error {
	stat.operator = "$avg"

	// Only support field references for avg
	if fieldRef, ok := value.(string); ok && len(fieldRef) > 0 && fieldRef[0] == '$' {
		fieldName := fieldRef[1:]
		storage := NewMemoryStorage()
		fieldValue, exists := storage.getNestedValue(doc, fieldName)
		if exists {
			if numVal, ok := toNumericValue(fieldValue); ok {
				stat.sum += numVal
				stat.count++
				stat.result = stat.sum / float64(stat.count)
			}
		}
		return nil
	}

	err := fmt.Errorf("$avg value must be field reference")
	log.Errorf("err:%v", err)
	return err
}

// processMin processes $min accumulator
func processMin(state *groupState, doc bson.M, stat *fieldStat, value interface{}) error {
	stat.operator = "$min"

	// Only support field references
	if fieldRef, ok := value.(string); ok && len(fieldRef) > 0 && fieldRef[0] == '$' {
		fieldName := fieldRef[1:]
		storage := NewMemoryStorage()
		fieldValue, exists := storage.getNestedValue(doc, fieldName)
		if !exists {
			return nil
		}

		if stat.result == nil {
			stat.result = fieldValue
		} else {
			if compareValues(fieldValue, stat.result) < 0 {
				stat.result = fieldValue
			}
		}
		return nil
	}

	err := fmt.Errorf("$min value must be field reference")
	log.Errorf("err:%v", err)
	return err
}

// processMax processes $max accumulator
func processMax(state *groupState, doc bson.M, stat *fieldStat, value interface{}) error {
	stat.operator = "$max"

	// Only support field references
	if fieldRef, ok := value.(string); ok && len(fieldRef) > 0 && fieldRef[0] == '$' {
		fieldName := fieldRef[1:]
		storage := NewMemoryStorage()
		fieldValue, exists := storage.getNestedValue(doc, fieldName)
		if !exists {
			return nil
		}

		if stat.result == nil {
			stat.result = fieldValue
		} else {
			if compareValues(fieldValue, stat.result) > 0 {
				stat.result = fieldValue
			}
		}
		return nil
	}

	err := fmt.Errorf("$max value must be field reference")
	log.Errorf("err:%v", err)
	return err
}

// processCount processes $count accumulator
func processCount(state *groupState, stat *fieldStat) error {
	stat.operator = "$count"
	stat.count++
	stat.result = stat.count
	return nil
}

// processFirst processes $first accumulator
func processFirst(state *groupState, doc bson.M, stat *fieldStat, value interface{}) error {
	stat.operator = "$first"

	// Only set if this is the first document
	if stat.result != nil {
		return nil
	}

	// Only support field references
	if fieldRef, ok := value.(string); ok && len(fieldRef) > 0 && fieldRef[0] == '$' {
		fieldName := fieldRef[1:]
		storage := NewMemoryStorage()
		fieldValue, exists := storage.getNestedValue(doc, fieldName)
		if exists {
			stat.result = fieldValue
		}
		return nil
	}

	err := fmt.Errorf("$first value must be field reference")
	log.Errorf("err:%v", err)
	return err
}

// processLast processes $last accumulator
func processLast(state *groupState, doc bson.M, stat *fieldStat, value interface{}) error {
	stat.operator = "$last"

	// Only support field references
	if fieldRef, ok := value.(string); ok && len(fieldRef) > 0 && fieldRef[0] == '$' {
		fieldName := fieldRef[1:]
		storage := NewMemoryStorage()
		fieldValue, exists := storage.getNestedValue(doc, fieldName)
		if exists {
			stat.result = fieldValue
		}
		return nil
	}

	err := fmt.Errorf("$last value must be field reference")
	log.Errorf("err:%v", err)
	return err
}

// processUnwind applies $unwind stage - deconstructs array fields
func processUnwind(docs []bson.M, value interface{}) ([]bson.M, error) {
	var fieldPath string

	// Handle simple string format: "$arrayField"
	if fieldRef, ok := value.(string); ok {
		if len(fieldRef) == 0 || fieldRef[0] != '$' {
			err := fmt.Errorf("$unwind field must start with $")
			log.Errorf("err:%v", err)
			return nil, err
		}
		fieldPath = fieldRef[1:]
	} else if unwindDoc, ok := value.(bson.M); ok {
		// Handle document format: { path: "$arrayField", preserveNullAndEmptyArrays: true }
		pathValue, hasPath := unwindDoc["path"]
		if !hasPath {
			err := fmt.Errorf("$unwind document must have path field")
			log.Errorf("err:%v", err)
			return nil, err
		}

		fieldRef, ok := pathValue.(string)
		if !ok || len(fieldRef) == 0 || fieldRef[0] != '$' {
			err := fmt.Errorf("$unwind path must be string starting with $")
			log.Errorf("err:%v", err)
			return nil, err
		}
		fieldPath = fieldRef[1:]

		// TODO: Support preserveNullAndEmptyArrays option
	} else {
		err := fmt.Errorf("$unwind value must be string or document")
		log.Errorf("err:%v", err)
		return nil, err
	}

	var result []bson.M
	storage := NewMemoryStorage()

	for _, doc := range docs {
		// Get array field value
		arrayValue, exists := storage.getNestedValue(doc, fieldPath)
		if !exists {
			// Skip documents without the field
			continue
		}

		// Check if value is an array
		arrayVal := reflect.ValueOf(arrayValue)
		if arrayVal.Kind() != reflect.Slice && arrayVal.Kind() != reflect.Array {
			// If not array, treat as single value
			newDoc := copyDoc(doc)
			result = append(result, newDoc)
			continue
		}

		// Unwind array - create one document per array element
		arrayLen := arrayVal.Len()
		if arrayLen == 0 {
			// Empty array - skip document (unless preserveNullAndEmptyArrays is true)
			continue
		}

		for i := 0; i < arrayLen; i++ {
			newDoc := copyDoc(doc)
			// Replace array field with single element
			setNestedValue(newDoc, fieldPath, arrayVal.Index(i).Interface())
			result = append(result, newDoc)
		}
	}

	return result, nil
}

// toNumericValue converts a value to float64 if possible
func toNumericValue(v interface{}) (float64, bool) {
	if v == nil {
		return 0, false
	}

	switch val := v.(type) {
	case int:
		return float64(val), true
	case int8:
		return float64(val), true
	case int16:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint8:
		return float64(val), true
	case uint16:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return val, true
	default:
		return 0, false
	}
}

// compareValues compares two values for min/max operations
// Returns -1 if a < b, 0 if a == b, 1 if a > b
func compareValues(a, b interface{}) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	// Try numeric comparison
	aNum, aOk := toNumericValue(a)
	bNum, bOk := toNumericValue(b)
	if aOk && bOk {
		if aNum < bNum {
			return -1
		}
		if aNum > bNum {
			return 1
		}
		return 0
	}

	// Try string comparison
	aStr, aOk := a.(string)
	bStr, bOk := b.(string)
	if aOk && bOk {
		if aStr < bStr {
			return -1
		}
		if aStr > bStr {
			return 1
		}
		return 0
	}

	// Try ObjectID comparison
	aID, aOk := a.(primitive.ObjectID)
	bID, bOk := b.(primitive.ObjectID)
	if aOk && bOk {
		aHex := aID.Hex()
		bHex := bID.Hex()
		if aHex < bHex {
			return -1
		}
		if aHex > bHex {
			return 1
		}
		return 0
	}

	// Default: consider equal
	return 0
}

// copyDoc creates a deep copy of a document
func copyDoc(doc bson.M) bson.M {
	newDoc := make(bson.M, len(doc))
	for k, v := range doc {
		newDoc[k] = v
	}
	return newDoc
}

// setNestedValue sets a value in a document using dot notation
func setNestedValue(doc bson.M, path string, value interface{}) {
	// For simplicity, only support top-level fields for now
	// Full dot notation support can be added later
	doc[path] = value
}
