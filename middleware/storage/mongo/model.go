package mongo

import "go.mongodb.org/mongo-driver/mongo"

// Collectioner interface for models that can provide their collection name
type Collectioner interface {
	Collection() string
}

// Model[M] is a lightweight model wrapper managing model metadata and client reference
type Model[M Collectioner] struct {
	client         *Client
	model          M
	collectionName string
	notFoundError  error
}

// NewModel creates a new model wrapper
func NewModel[M Collectioner](client *Client, model M) *Model[M] {
	return &Model[M]{
		client:         client,
		model:          model,
		collectionName: model.Collection(),
		notFoundError:  mongo.ErrNoDocuments,
	}
}

// NewScoop creates a type-safe query builder for this model, optionally accepting a transaction scoop
func (m *Model[M]) NewScoop(tx ...*Scoop) *ModelScoop[M] {
	var baseScoop *Scoop
	if len(tx) > 0 && tx[0] != nil {
		baseScoop = m.client.NewScoop(tx[0]).CollectionName(m.collectionName)
	} else {
		baseScoop = m.client.NewScoop().CollectionName(m.collectionName)
	}

	baseScoop.SetNotFound(m.notFoundError)

	return &ModelScoop[M]{
		Scoop: baseScoop,
		m:     m.model,
	}
}

// CollectionName returns the collection name for this model
func (m *Model[M]) CollectionName() string {
	return m.collectionName
}

// SetNotFound sets the not found error for this model
func (m *Model[M]) SetNotFound(err error) *Model[M] {
	m.notFoundError = err
	return m
}

// IsNotFound checks if the error is a not found error
func (m *Model[M]) IsNotFound(err error) bool {
	return err == m.notFoundError || err == mongo.ErrNoDocuments
}
