package mock

import (
	"context"

	"github.com/lazygophers/log"
	gomongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BulkWrite executes multiple write operations in bulk
// Supports all 6 MongoDB write models:
//   - InsertOneModel: inserts a document
//   - UpdateOneModel: updates one document
//   - UpdateManyModel: updates multiple documents
//   - ReplaceOneModel: replaces one document
//   - DeleteOneModel: deletes one document
//   - DeleteManyModel: deletes multiple documents
//
// Returns BulkWriteResult with statistics about the operations
func (m *MockCollection) BulkWrite(ctx context.Context, models []gomongo.WriteModel, opts ...*options.BulkWriteOptions) (*gomongo.BulkWriteResult, error) {
	if len(models) == 0 {
		return &gomongo.BulkWriteResult{
			InsertedCount: 0,
			MatchedCount:  0,
			ModifiedCount: 0,
			DeletedCount:  0,
			UpsertedCount: 0,
			UpsertedIDs:   map[int64]interface{}{},
		}, nil
	}

	// Parse options
	ordered := true
	if len(opts) > 0 && opts[0] != nil && opts[0].Ordered != nil {
		ordered = *opts[0].Ordered
	}

	// Initialize result
	result := &gomongo.BulkWriteResult{
		InsertedCount: 0,
		MatchedCount:  0,
		ModifiedCount: 0,
		DeletedCount:  0,
		UpsertedCount: 0,
		UpsertedIDs:   map[int64]interface{}{},
	}

	// Execute each model
	for i, model := range models {
		err := m.executeBulkWriteModel(ctx, model, result, int64(i))
		if err != nil {
			log.Errorf("err:%v", err)
			// If ordered mode, stop on first error
			if ordered {
				return result, err
			}
			// Otherwise, continue processing
		}
	}

	return result, nil
}

// executeBulkWriteModel executes a single write model and updates the result
func (m *MockCollection) executeBulkWriteModel(ctx context.Context, model gomongo.WriteModel, result *gomongo.BulkWriteResult, index int64) error {
	switch writeModel := model.(type) {
	case *gomongo.InsertOneModel:
		return m.executeInsertOne(ctx, writeModel, result)

	case *gomongo.UpdateOneModel:
		return m.executeUpdateOne(ctx, writeModel, result, index)

	case *gomongo.UpdateManyModel:
		return m.executeUpdateMany(ctx, writeModel, result, index)

	case *gomongo.ReplaceOneModel:
		return m.executeReplaceOne(ctx, writeModel, result, index)

	case *gomongo.DeleteOneModel:
		return m.executeDeleteOne(ctx, writeModel, result)

	case *gomongo.DeleteManyModel:
		return m.executeDeleteMany(ctx, writeModel, result)

	default:
		err := ErrInvalidArgument
		log.Errorf("err:%v", err)
		return err
	}
}

// executeInsertOne executes an InsertOneModel
func (m *MockCollection) executeInsertOne(ctx context.Context, model *gomongo.InsertOneModel, result *gomongo.BulkWriteResult) error {
	_, err := m.InsertOne(ctx, model.Document)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	result.InsertedCount++
	return nil
}

// executeUpdateOne executes an UpdateOneModel
func (m *MockCollection) executeUpdateOne(ctx context.Context, model *gomongo.UpdateOneModel, result *gomongo.BulkWriteResult, index int64) error {
	updateOpts := options.Update()
	if model.Upsert != nil {
		updateOpts.SetUpsert(*model.Upsert)
	}

	updateResult, err := m.UpdateOne(ctx, model.Filter, model.Update, updateOpts)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	result.MatchedCount += updateResult.MatchedCount
	result.ModifiedCount += updateResult.ModifiedCount
	result.UpsertedCount += updateResult.UpsertedCount

	if updateResult.UpsertedID != nil {
		result.UpsertedIDs[index] = updateResult.UpsertedID
	}

	return nil
}

// executeUpdateMany executes an UpdateManyModel
func (m *MockCollection) executeUpdateMany(ctx context.Context, model *gomongo.UpdateManyModel, result *gomongo.BulkWriteResult, index int64) error {
	updateOpts := options.Update()
	if model.Upsert != nil {
		updateOpts.SetUpsert(*model.Upsert)
	}

	updateResult, err := m.UpdateMany(ctx, model.Filter, model.Update, updateOpts)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	result.MatchedCount += updateResult.MatchedCount
	result.ModifiedCount += updateResult.ModifiedCount
	result.UpsertedCount += updateResult.UpsertedCount

	if updateResult.UpsertedID != nil {
		result.UpsertedIDs[index] = updateResult.UpsertedID
	}

	return nil
}

// executeReplaceOne executes a ReplaceOneModel
func (m *MockCollection) executeReplaceOne(ctx context.Context, model *gomongo.ReplaceOneModel, result *gomongo.BulkWriteResult, index int64) error {
	replaceOpts := options.Replace()
	if model.Upsert != nil {
		replaceOpts.SetUpsert(*model.Upsert)
	}

	replaceResult, err := m.ReplaceOne(ctx, model.Filter, model.Replacement, replaceOpts)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	result.MatchedCount += replaceResult.MatchedCount
	result.ModifiedCount += replaceResult.ModifiedCount
	result.UpsertedCount += replaceResult.UpsertedCount

	if replaceResult.UpsertedID != nil {
		result.UpsertedIDs[index] = replaceResult.UpsertedID
	}

	return nil
}

// executeDeleteOne executes a DeleteOneModel
func (m *MockCollection) executeDeleteOne(ctx context.Context, model *gomongo.DeleteOneModel, result *gomongo.BulkWriteResult) error {
	deleteResult, err := m.DeleteOne(ctx, model.Filter)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	result.DeletedCount += deleteResult.DeletedCount
	return nil
}

// executeDeleteMany executes a DeleteManyModel
func (m *MockCollection) executeDeleteMany(ctx context.Context, model *gomongo.DeleteManyModel, result *gomongo.BulkWriteResult) error {
	deleteResult, err := m.DeleteMany(ctx, model.Filter)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	result.DeletedCount += deleteResult.DeletedCount
	return nil
}
