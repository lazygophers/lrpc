package mongo

import (
	"context"
	"fmt"

	"github.com/kamva/mgm/v3"
	"github.com/lazygophers/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ChangeEvent represents a MongoDB change stream event
type ChangeEvent struct {
	OperationType            string                 `bson:"operationType"`
	FullDocument             map[string]interface{} `bson:"fullDocument"`
	DocumentKey              map[string]interface{} `bson:"documentKey"`
	UpdateDescription        map[string]interface{} `bson:"updateDescription"`
	Ns                       map[string]string      `bson:"ns"`
	ClusterTime              interface{}            `bson:"clusterTime"`
	TxnNumber                interface{}            `bson:"txnNumber"`
	Lsid                     interface{}            `bson:"lsid"`
	FullDocumentBeforeChange map[string]interface{} `bson:"fullDocumentBeforeChange"`
}

// ChangeStreamHandler is a function type for handling change stream events
type ChangeStreamHandler func(*ChangeEvent) error

// ChangeStream represents a MongoDB change stream watcher
type ChangeStream struct {
	coll   *mongo.Collection
	client *Client
}

// NewChangeStream creates a new change stream for a collection
func (s *Scoop) WatchChanges(pipeline ...bson.M) (*ChangeStream, error) {
	return &ChangeStream{
		coll:   s.coll,
		client: s.client,
	}, nil
}

// Watch watches for changes on the collection
func (cs *ChangeStream) Watch(pipeline ...bson.M) (*mongo.ChangeStream, error) {
	pipelineA := bson.A{}
	for _, stage := range pipeline {
		pipelineA = append(pipelineA, stage)
	}

	opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)

	stream, err := cs.coll.Watch(context.Background(), pipelineA, opts)
	if err != nil {
		cs.client.cfg.Logger.Errorf("err:%v", err)
		return nil, err
	}

	return stream, nil
}

// Listen listens for changes and calls the handler for each event
func (cs *ChangeStream) Listen(handler ChangeStreamHandler, pipeline ...bson.M) error {
	stream, err := cs.Watch(pipeline...)
	if err != nil {
		return err
	}
	defer stream.Close(context.Background())

	for stream.Next(context.Background()) {
		var event ChangeEvent
		err := stream.Decode(&event)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		// Call the handler
		err = handler(&event)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}

	return stream.Err()
}

// ListenWithFilters listens for changes matching the filter
func (cs *ChangeStream) ListenWithFilters(handler ChangeStreamHandler, operationTypes ...string) error {
	var pipelines []bson.M

	if len(operationTypes) > 0 {
		// Filter by operation types
		filterOps := bson.A{}
		for _, op := range operationTypes {
			filterOps = append(filterOps, op)
		}

		pipelines = append(pipelines, bson.M{
			"$match": bson.M{
				"operationType": bson.M{
					"$in": filterOps,
				},
			},
		})
	}

	return cs.Listen(handler, pipelines...)
}

// Close closes the change stream
func (cs *ChangeStream) Close() {
	// No-op: context is background
}

// WatchAllCollections watches for changes across all collections in a database
func (c *Client) WatchAllCollections() (*DatabaseChangeStream, error) {
	return &DatabaseChangeStream{
		client: c,
	}, nil
}

// DatabaseChangeStream represents a change stream for an entire database
type DatabaseChangeStream struct {
	client *Client
}

// Watch watches for changes across all collections
func (dcs *DatabaseChangeStream) Watch(pipeline ...bson.M) (*mongo.ChangeStream, error) {
	pipelineA := bson.A{}
	for _, stage := range pipeline {
		pipelineA = append(pipelineA, stage)
	}

	opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)

	// Get database from MGM
	_, _, db, err := mgm.DefaultConfigs()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	stream, err := db.Watch(context.Background(), pipelineA, opts)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return stream, nil
}

// Listen listens for database-wide changes
func (dcs *DatabaseChangeStream) Listen(handler ChangeStreamHandler, pipeline ...bson.M) error {
	stream, err := dcs.Watch(pipeline...)
	if err != nil {
		return err
	}
	defer stream.Close(context.Background())

	for stream.Next(context.Background()) {
		var event ChangeEvent
		err := stream.Decode(&event)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		// Call the handler
		err = handler(&event)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}

	return stream.Err()
}

// Close closes the change stream
func (dcs *DatabaseChangeStream) Close() {
	// No-op: context is background
}

// PrintEvent is a simple event printer for debugging
func PrintEvent(event *ChangeEvent) error {
	fmt.Printf("Change Event: %+v\n", event)
	return nil
}
