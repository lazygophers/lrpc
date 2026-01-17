package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Aggregation represents a MongoDB aggregation pipeline
type Aggregation struct {
	coll     *mongo.Collection
	ctx      context.Context
	pipeline bson.A
	opts     *options.AggregateOptions
}

// NewAggregation creates a new aggregation pipeline
func NewAggregation(coll *mongo.Collection, ctx context.Context, stages ...bson.M) *Aggregation {
	pipeline := bson.A{}
	for _, stage := range stages {
		pipeline = append(pipeline, stage)
	}

	return &Aggregation{
		coll:     coll,
		ctx:      ctx,
		pipeline: pipeline,
		opts:     options.Aggregate(),
	}
}

// Match adds a $match stage
func (a *Aggregation) Match(filter bson.M) *Aggregation {
	a.pipeline = append(a.pipeline, bson.M{"$match": filter})
	return a
}

// Project adds a $project stage
func (a *Aggregation) Project(projection bson.M) *Aggregation {
	a.pipeline = append(a.pipeline, bson.M{"$project": projection})
	return a
}

// Group adds a $group stage
func (a *Aggregation) Group(groupBy bson.M) *Aggregation {
	a.pipeline = append(a.pipeline, bson.M{"$group": groupBy})
	return a
}

// Sort adds a $sort stage
func (a *Aggregation) Sort(sort bson.M) *Aggregation {
	a.pipeline = append(a.pipeline, bson.M{"$sort": sort})
	return a
}

// Skip adds a $skip stage
func (a *Aggregation) Skip(skip int64) *Aggregation {
	a.pipeline = append(a.pipeline, bson.M{"$skip": skip})
	return a
}

// Limit adds a $limit stage
func (a *Aggregation) Limit(limit int64) *Aggregation {
	a.pipeline = append(a.pipeline, bson.M{"$limit": limit})
	return a
}

// Lookup adds a $lookup stage (join)
func (a *Aggregation) Lookup(from string, localField string, foreignField string, as string) *Aggregation {
	a.pipeline = append(a.pipeline, bson.M{
		"$lookup": bson.M{
			"from":         from,
			"localField":   localField,
			"foreignField": foreignField,
			"as":           as,
		},
	})
	return a
}

// Unwind adds an $unwind stage
func (a *Aggregation) Unwind(path string, preserveNullAndEmptyArrays ...bool) *Aggregation {
	unwindStage := bson.M{"$unwind": bson.M{"path": path}}
	if len(preserveNullAndEmptyArrays) > 0 {
		unwindMap := unwindStage["$unwind"].(bson.M)
		unwindMap["preserveNullAndEmptyArrays"] = preserveNullAndEmptyArrays[0]
	}
	a.pipeline = append(a.pipeline, unwindStage)
	return a
}

// AddFields adds an $addFields stage
func (a *Aggregation) AddFields(fields bson.M) *Aggregation {
	a.pipeline = append(a.pipeline, bson.M{"$addFields": fields})
	return a
}

// Count adds a $count stage
func (a *Aggregation) Count(field string) *Aggregation {
	a.pipeline = append(a.pipeline, bson.M{"$count": field})
	return a
}

// Facet adds a $facet stage
func (a *Aggregation) Facet(facets bson.M) *Aggregation {
	a.pipeline = append(a.pipeline, bson.M{"$facet": facets})
	return a
}

// AddStage adds a custom stage to the pipeline
func (a *Aggregation) AddStage(stage bson.M) *Aggregation {
	a.pipeline = append(a.pipeline, stage)
	return a
}

// Execute executes the aggregation pipeline
func (a *Aggregation) Execute(result interface{}) error {
	cursor, err := a.coll.Aggregate(a.ctx, a.pipeline, a.opts)
	if err != nil {
		return err
	}
	defer cursor.Close(a.ctx)

	err = cursor.All(a.ctx, result)
	if err != nil {
		return err
	}

	return nil
}

// ExecuteOne executes the aggregation pipeline and returns a single result
func (a *Aggregation) ExecuteOne(result interface{}) error {
	cursor, err := a.coll.Aggregate(a.ctx, a.pipeline, a.opts)
	if err != nil {
		return err
	}
	defer cursor.Close(a.ctx)

	if cursor.Next(a.ctx) {
		err := cursor.Decode(result)
		if err != nil {
			return err
		}
	}

	return cursor.Err()
}

// GetPipeline returns the aggregation pipeline
func (a *Aggregation) GetPipeline() bson.A {
	return a.pipeline
}

// Clear clears the pipeline
func (a *Aggregation) Clear() *Aggregation {
	a.pipeline = bson.A{}
	return a
}
