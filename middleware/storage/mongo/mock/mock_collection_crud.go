package mock

import (
	"context"

	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/storage/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	gomongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Find finds all documents that match the filter
// Returns a cursor for iterating through the results
func (m *MockCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (mongo.MongoCursor, error) {
	filterDoc, err := toBsonM(filter)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	// Convert FindOptions to FindOptions
	findOpts := &FindOptions{}
	if len(opts) > 0 && opts[0] != nil {
		opt := opts[0]
		if opt.Limit != nil {
			findOpts.Limit = opt.Limit
		}
		if opt.Skip != nil {
			findOpts.Skip = opt.Skip
		}
		if opt.Sort != nil {
			sortDoc, err := toBsonM(opt.Sort)
			if err != nil {
				log.Errorf("err:%v", err)
				return nil, err
			}
			findOpts.Sort = sortDoc
		}
		if opt.Projection != nil {
			projDoc, err := toBsonM(opt.Projection)
			if err != nil {
				log.Errorf("err:%v", err)
				return nil, err
			}
			findOpts.Projection = projDoc
		}
	}

	documents := m.storage.Find(m.name, filterDoc, findOpts)

	return NewMockCursor(documents), nil
}

// FindOne finds a single document that matches the filter
// Returns a SingleResult that can be decoded
func (m *MockCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *gomongo.SingleResult {
	filterDoc, err := toBsonM(filter)
	if err != nil {
		log.Errorf("err:%v", err)
		emptyDoc := bson.D{}
		return gomongo.NewSingleResultFromDocument(emptyDoc, err, nil)
	}

	// Use Find with limit 1
	findOpts := &FindOptions{
		Limit: ptrInt64(1),
	}

	// Apply projection if provided
	if len(opts) > 0 && opts[0] != nil && opts[0].Projection != nil {
		projDoc, err := toBsonM(opts[0].Projection)
		if err != nil {
			log.Errorf("err:%v", err)
			emptyDoc := bson.D{}
			return gomongo.NewSingleResultFromDocument(emptyDoc, err, nil)
		}
		findOpts.Projection = projDoc
	}

	documents := m.storage.Find(m.name, filterDoc, findOpts)

	if len(documents) == 0 {
		emptyDoc := bson.D{}
		return gomongo.NewSingleResultFromDocument(emptyDoc, gomongo.ErrNoDocuments, nil)
	}

	// Convert bson.M to raw BSON
	rawBytes, err := bson.Marshal(documents[0])
	if err != nil {
		log.Errorf("err:%v", err)
		emptyDoc := bson.D{}
		return gomongo.NewSingleResultFromDocument(emptyDoc, err, nil)
	}

	return gomongo.NewSingleResultFromDocument(rawBytes, nil, nil)
}

// FindOneAndDelete finds a single document and deletes it
// Returns a SingleResult containing the deleted document
func (m *MockCollection) FindOneAndDelete(ctx context.Context, filter interface{}, opts ...*options.FindOneAndDeleteOptions) *gomongo.SingleResult {
	filterDoc, err := toBsonM(filter)
	if err != nil {
		log.Errorf("err:%v", err)
		emptyDoc := bson.D{}
		return gomongo.NewSingleResultFromDocument(emptyDoc, err, nil)
	}

	// Find the document first
	findOpts := &FindOptions{
		Limit: ptrInt64(1),
	}

	documents := m.storage.Find(m.name, filterDoc, findOpts)

	if len(documents) == 0 {
		emptyDoc := bson.D{}
		return gomongo.NewSingleResultFromDocument(emptyDoc, gomongo.ErrNoDocuments, nil)
	}

	// Delete the document
	m.storage.DeleteOne(m.name, filterDoc)

	// Convert bson.M to raw BSON
	rawBytes, err := bson.Marshal(documents[0])
	if err != nil {
		log.Errorf("err:%v", err)
		emptyDoc := bson.D{}
		return gomongo.NewSingleResultFromDocument(emptyDoc, err, nil)
	}

	return gomongo.NewSingleResultFromDocument(rawBytes, nil, nil)
}

// FindOneAndReplace finds a single document and replaces it
// Returns a SingleResult containing the original or replaced document based on options
func (m *MockCollection) FindOneAndReplace(ctx context.Context, filter interface{}, replacement interface{}, opts ...*options.FindOneAndReplaceOptions) *gomongo.SingleResult {
	filterDoc, err := toBsonM(filter)
	if err != nil {
		log.Errorf("err:%v", err)
		emptyDoc := bson.D{}
		return gomongo.NewSingleResultFromDocument(emptyDoc, err, nil)
	}

	replacementDoc, err := toBsonM(replacement)
	if err != nil {
		log.Errorf("err:%v", err)
		emptyDoc := bson.D{}
		return gomongo.NewSingleResultFromDocument(emptyDoc, err, nil)
	}

	// Find the document first
	findOpts := &FindOptions{
		Limit: ptrInt64(1),
	}

	documents := m.storage.Find(m.name, filterDoc, findOpts)

	returnNew := false
	upsert := false
	if len(opts) > 0 && opts[0] != nil {
		if opts[0].ReturnDocument != nil && *opts[0].ReturnDocument == options.After {
			returnNew = true
		}
		if opts[0].Upsert != nil {
			upsert = *opts[0].Upsert
		}
	}

	// Handle upsert case
	if len(documents) == 0 {
		if upsert {
			// Generate new ID if not present
			if _, hasID := replacementDoc["_id"]; !hasID {
				replacementDoc["_id"] = primitive.NewObjectID()
			}

			err := m.storage.Insert(m.name, replacementDoc)
			if err != nil {
				log.Errorf("err:%v", err)
				emptyDoc := bson.D{}
				return gomongo.NewSingleResultFromDocument(emptyDoc, err, nil)
			}

			if returnNew {
				rawBytes, err := bson.Marshal(replacementDoc)
				if err != nil {
					log.Errorf("err:%v", err)
					emptyDoc := bson.D{}
					return gomongo.NewSingleResultFromDocument(emptyDoc, err, nil)
				}
				return gomongo.NewSingleResultFromDocument(rawBytes, nil, nil)
			}
			emptyDoc := bson.D{}
			return gomongo.NewSingleResultFromDocument(emptyDoc, gomongo.ErrNoDocuments, nil)
		}
		emptyDoc := bson.D{}
		return gomongo.NewSingleResultFromDocument(emptyDoc, gomongo.ErrNoDocuments, nil)
	}

	originalDoc := documents[0]

	// Delete old document and insert new one (replace)
	m.storage.DeleteOne(m.name, filterDoc)

	// Preserve _id if present in original
	if id, hasID := originalDoc["_id"]; hasID {
		replacementDoc["_id"] = id
	}

	err = m.storage.Insert(m.name, replacementDoc)
	if err != nil {
		log.Errorf("err:%v", err)
		emptyDoc := bson.D{}
		return gomongo.NewSingleResultFromDocument(emptyDoc, err, nil)
	}

	// Return the appropriate document
	var resultDoc bson.M
	if returnNew {
		resultDoc = replacementDoc
	} else {
		resultDoc = originalDoc
	}

	rawBytes, err := bson.Marshal(resultDoc)
	if err != nil {
		log.Errorf("err:%v", err)
		emptyDoc := bson.D{}
		return gomongo.NewSingleResultFromDocument(emptyDoc, err, nil)
	}

	return gomongo.NewSingleResultFromDocument(rawBytes, nil, nil)
}

// FindOneAndUpdate finds a single document and updates it
// Returns a SingleResult containing the original or updated document based on options
func (m *MockCollection) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) *gomongo.SingleResult {
	filterDoc, err := toBsonM(filter)
	if err != nil {
		log.Errorf("err:%v", err)
		emptyDoc := bson.D{}
		return gomongo.NewSingleResultFromDocument(emptyDoc, err, nil)
	}

	updateDoc, err := toBsonM(update)
	if err != nil {
		log.Errorf("err:%v", err)
		emptyDoc := bson.D{}
		return gomongo.NewSingleResultFromDocument(emptyDoc, err, nil)
	}

	// Find the document first
	findOpts := &FindOptions{
		Limit: ptrInt64(1),
	}

	documents := m.storage.Find(m.name, filterDoc, findOpts)

	returnNew := false
	upsert := false
	if len(opts) > 0 && opts[0] != nil {
		if opts[0].ReturnDocument != nil && *opts[0].ReturnDocument == options.After {
			returnNew = true
		}
		if opts[0].Upsert != nil {
			upsert = *opts[0].Upsert
		}
	}

	// Handle upsert case
	if len(documents) == 0 {
		if upsert {
			// Create new document from update operations
			newDoc := bson.M{}
			if setFields, ok := updateDoc["$set"].(bson.M); ok {
				for k, v := range setFields {
					newDoc[k] = v
				}
			}

			// Generate new ID if not present
			if _, hasID := newDoc["_id"]; !hasID {
				newDoc["_id"] = primitive.NewObjectID()
			}

			err := m.storage.Insert(m.name, newDoc)
			if err != nil {
				log.Errorf("err:%v", err)
				emptyDoc := bson.D{}
				return gomongo.NewSingleResultFromDocument(emptyDoc, err, nil)
			}

			if returnNew {
				rawBytes, err := bson.Marshal(newDoc)
				if err != nil {
					log.Errorf("err:%v", err)
					emptyDoc := bson.D{}
					return gomongo.NewSingleResultFromDocument(emptyDoc, err, nil)
				}
				return gomongo.NewSingleResultFromDocument(rawBytes, nil, nil)
			}
			emptyDoc := bson.D{}
			return gomongo.NewSingleResultFromDocument(emptyDoc, gomongo.ErrNoDocuments, nil)
		}
		emptyDoc := bson.D{}
		return gomongo.NewSingleResultFromDocument(emptyDoc, gomongo.ErrNoDocuments, nil)
	}

	originalDoc := documents[0]

	// Update the document
	m.storage.UpdateOne(m.name, filterDoc, updateDoc)

	// Return the appropriate document
	var resultDoc bson.M
	if returnNew {
		// Fetch the updated document
		updatedDocs := m.storage.Find(m.name, filterDoc, findOpts)
		if len(updatedDocs) == 0 {
			emptyDoc := bson.D{}
			return gomongo.NewSingleResultFromDocument(emptyDoc, gomongo.ErrNoDocuments, nil)
		}
		resultDoc = updatedDocs[0]
	} else {
		resultDoc = originalDoc
	}

	rawBytes, err := bson.Marshal(resultDoc)
	if err != nil {
		log.Errorf("err:%v", err)
		emptyDoc := bson.D{}
		return gomongo.NewSingleResultFromDocument(emptyDoc, err, nil)
	}

	return gomongo.NewSingleResultFromDocument(rawBytes, nil, nil)
}

// InsertMany inserts multiple documents into the collection
// Returns the IDs of the inserted documents
func (m *MockCollection) InsertMany(ctx context.Context, documents []interface{}, opts ...*options.InsertManyOptions) (*gomongo.InsertManyResult, error) {
	if len(documents) == 0 {
		return &gomongo.InsertManyResult{
			InsertedIDs: []interface{}{},
		}, nil
	}

	// Convert documents to bson.M
	bsonDocs := make([]bson.M, len(documents))
	insertedIDs := make([]interface{}, len(documents))

	for i, doc := range documents {
		bsonDoc, err := toBsonM(doc)
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}

		// Generate ID if not present
		if _, hasID := bsonDoc["_id"]; !hasID {
			bsonDoc["_id"] = primitive.NewObjectID()
		}

		bsonDocs[i] = bsonDoc
		insertedIDs[i] = bsonDoc["_id"]
	}

	err := m.storage.InsertMany(m.name, bsonDocs)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return &gomongo.InsertManyResult{
		InsertedIDs: insertedIDs,
	}, nil
}

// InsertOne inserts a single document into the collection
// Returns the ID of the inserted document
func (m *MockCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*gomongo.InsertOneResult, error) {
	bsonDoc, err := toBsonM(document)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	// Generate ID if not present
	if _, hasID := bsonDoc["_id"]; !hasID {
		bsonDoc["_id"] = primitive.NewObjectID()
	}

	err = m.storage.Insert(m.name, bsonDoc)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return &gomongo.InsertOneResult{
		InsertedID: bsonDoc["_id"],
	}, nil
}

// UpdateOne updates a single document that matches the filter
// Returns the number of matched and modified documents
func (m *MockCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*gomongo.UpdateResult, error) {
	filterDoc, err := toBsonM(filter)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	updateDoc, err := toBsonM(update)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	// Find the document first to check if it exists
	findOpts := &FindOptions{
		Limit: ptrInt64(1),
	}

	documents := m.storage.Find(m.name, filterDoc, findOpts)

	// Handle upsert
	upsert := false
	if len(opts) > 0 && opts[0] != nil && opts[0].Upsert != nil {
		upsert = *opts[0].Upsert
	}

	if len(documents) == 0 {
		if upsert {
			// Create new document from update operations
			newDoc := bson.M{}
			if setFields, ok := updateDoc["$set"].(bson.M); ok {
				for k, v := range setFields {
					newDoc[k] = v
				}
			}

			// Generate new ID if not present
			if _, hasID := newDoc["_id"]; !hasID {
				newDoc["_id"] = primitive.NewObjectID()
			}

			err := m.storage.Insert(m.name, newDoc)
			if err != nil {
				log.Errorf("err:%v", err)
				return nil, err
			}

			return &gomongo.UpdateResult{
				MatchedCount:  0,
				ModifiedCount: 0,
				UpsertedCount: 1,
				UpsertedID:    newDoc["_id"],
			}, nil
		}

		return &gomongo.UpdateResult{
			MatchedCount:  0,
			ModifiedCount: 0,
			UpsertedCount: 0,
		}, nil
	}

	modifiedCount := m.storage.UpdateOne(m.name, filterDoc, updateDoc)

	return &gomongo.UpdateResult{
		MatchedCount:  1,
		ModifiedCount: modifiedCount,
		UpsertedCount: 0,
	}, nil
}

// UpdateMany updates all documents that match the filter
// Returns the number of matched and modified documents
func (m *MockCollection) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*gomongo.UpdateResult, error) {
	filterDoc, err := toBsonM(filter)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	updateDoc, err := toBsonM(update)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	// Count matched documents before update
	matchedCount := m.storage.Count(m.name, filterDoc)

	// Handle upsert
	upsert := false
	if len(opts) > 0 && opts[0] != nil && opts[0].Upsert != nil {
		upsert = *opts[0].Upsert
	}

	if matchedCount == 0 && upsert {
		// Create new document from update operations
		newDoc := bson.M{}
		if setFields, ok := updateDoc["$set"].(bson.M); ok {
			for k, v := range setFields {
				newDoc[k] = v
			}
		}

		// Generate new ID if not present
		if _, hasID := newDoc["_id"]; !hasID {
			newDoc["_id"] = primitive.NewObjectID()
		}

		err := m.storage.Insert(m.name, newDoc)
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}

		return &gomongo.UpdateResult{
			MatchedCount:  0,
			ModifiedCount: 0,
			UpsertedCount: 1,
			UpsertedID:    newDoc["_id"],
		}, nil
	}

	modifiedCount := m.storage.Update(m.name, filterDoc, updateDoc)

	return &gomongo.UpdateResult{
		MatchedCount:  matchedCount,
		ModifiedCount: modifiedCount,
		UpsertedCount: 0,
	}, nil
}

// UpdateByID updates a single document by its ID
// Returns the number of matched and modified documents
func (m *MockCollection) UpdateByID(ctx context.Context, id interface{}, update interface{}, opts ...*options.UpdateOptions) (*gomongo.UpdateResult, error) {
	filter := bson.M{"_id": id}
	return m.UpdateOne(ctx, filter, update, opts...)
}

// ReplaceOne replaces a single document that matches the filter
// Returns the number of matched and modified documents
func (m *MockCollection) ReplaceOne(ctx context.Context, filter interface{}, replacement interface{}, opts ...*options.ReplaceOptions) (*gomongo.UpdateResult, error) {
	filterDoc, err := toBsonM(filter)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	replacementDoc, err := toBsonM(replacement)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	// Find the document first
	findOpts := &FindOptions{
		Limit: ptrInt64(1),
	}

	documents := m.storage.Find(m.name, filterDoc, findOpts)

	// Handle upsert
	upsert := false
	if len(opts) > 0 && opts[0] != nil && opts[0].Upsert != nil {
		upsert = *opts[0].Upsert
	}

	if len(documents) == 0 {
		if upsert {
			// Generate new ID if not present
			if _, hasID := replacementDoc["_id"]; !hasID {
				replacementDoc["_id"] = primitive.NewObjectID()
			}

			err := m.storage.Insert(m.name, replacementDoc)
			if err != nil {
				log.Errorf("err:%v", err)
				return nil, err
			}

			return &gomongo.UpdateResult{
				MatchedCount:  0,
				ModifiedCount: 0,
				UpsertedCount: 1,
				UpsertedID:    replacementDoc["_id"],
			}, nil
		}

		return &gomongo.UpdateResult{
			MatchedCount:  0,
			ModifiedCount: 0,
			UpsertedCount: 0,
		}, nil
	}

	// Preserve _id from original document
	originalDoc := documents[0]
	if id, hasID := originalDoc["_id"]; hasID {
		replacementDoc["_id"] = id
	}

	// Use ReplaceOne to publish correct event
	replaced := m.storage.ReplaceOne(m.name, filterDoc, replacementDoc)

	return &gomongo.UpdateResult{
		MatchedCount:  replaced,
		ModifiedCount: replaced,
		UpsertedCount: 0,
	}, nil
}

// DeleteOne deletes a single document that matches the filter
// Returns the number of deleted documents (0 or 1)
func (m *MockCollection) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*gomongo.DeleteResult, error) {
	filterDoc, err := toBsonM(filter)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	deletedCount := m.storage.DeleteOne(m.name, filterDoc)

	return &gomongo.DeleteResult{
		DeletedCount: deletedCount,
	}, nil
}

// DeleteMany deletes all documents that match the filter
// Returns the number of deleted documents
func (m *MockCollection) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*gomongo.DeleteResult, error) {
	filterDoc, err := toBsonM(filter)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	deletedCount := m.storage.Delete(m.name, filterDoc)

	return &gomongo.DeleteResult{
		DeletedCount: deletedCount,
	}, nil
}
