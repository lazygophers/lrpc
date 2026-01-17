package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// newTestConfig provides a test MongoDB configuration
func newTestConfig() *Config {
	return &Config{
		Address:        "localhost",
		Port:           27017,
		Database:       "test",
		Debug:          true,
		ConnectTimeout: 10 * time.Second,
		ContextTimeout: 30 * time.Second,
		MaxPoolSize:    10,
		MinPoolSize:    1,
		Logger:         NewLogger(),
	}
}

// newTestClient creates a test MongoDB client
func newTestClient(t *testing.T) *Client {
	cfg := newTestConfig()
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	return client
}

// CleanupTestCollections cleans up test collections after tests
func CleanupTestCollections(t *testing.T, client *Client, collections ...string) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, _, db, err := mgm.DefaultConfigs()
	if err != nil {
		t.Logf("warning: failed to get MGM config: %v", err)
		return
	}

	for _, collName := range collections {
		err := db.Collection(collName).Drop(ctx)
		if err != nil && err != mongo.ErrNilDocument {
			t.Logf("warning: failed to drop collection %s: %v", collName, err)
		}
	}
}

// InsertTestData inserts test data into a collection
func InsertTestData(t *testing.T, client *Client, collName string, docs ...interface{}) []interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, _, db, err := mgm.DefaultConfigs()
	if err != nil {
		t.Fatalf("failed to get MGM config: %v", err)
	}
	coll := db.Collection(collName)
	result, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}

	return result.InsertedIDs
}

// GetTestDocument retrieves a test document
func GetTestDocument(t *testing.T, client *Client, collName string, filter bson.M) map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, _, db, err := mgm.DefaultConfigs()
	if err != nil {
		t.Fatalf("failed to get MGM config: %v", err)
	}
	var result map[string]interface{}
	err = db.Collection(collName).FindOne(ctx, filter).Decode(&result)
	if err != nil && err != mongo.ErrNoDocuments {
		t.Fatalf("failed to get test document: %v", err)
	}

	return result
}

// CountTestDocuments counts documents in a collection
func CountTestDocuments(t *testing.T, client *Client, collName string, filter bson.M) int64 {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, _, db, err := mgm.DefaultConfigs()
	if err != nil {
		t.Fatalf("failed to get MGM config: %v", err)
	}
	count, err := db.Collection(collName).CountDocuments(ctx, filter)
	if err != nil {
		t.Fatalf("failed to count documents: %v", err)
	}

	return count
}

// AssertDocumentExists asserts that a document exists
func AssertDocumentExists(t *testing.T, client *Client, collName string, filter bson.M) {
	count := CountTestDocuments(t, client, collName, filter)
	if count == 0 {
		t.Errorf("expected document to exist, but it was not found")
	}
}

// AssertDocumentNotExists asserts that a document does not exist
func AssertDocumentNotExists(t *testing.T, client *Client, collName string, filter bson.M) {
	count := CountTestDocuments(t, client, collName, filter)
	if count > 0 {
		t.Errorf("expected document to not exist, but found %d documents", count)
	}
}

// AssertCount asserts the count of documents
func AssertCount(t *testing.T, expected int64, client *Client, collName string, filter bson.M) {
	actual := CountTestDocuments(t, client, collName, filter)
	if actual != expected {
		t.Errorf("expected %d documents, but got %d", expected, actual)
	}
}

// WaitForCondition waits for a condition to be true
func WaitForCondition(t *testing.T, timeout time.Duration, condition func() bool) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if condition() {
			return nil
		}

		select {
		case <-ticker.C:
			if time.Now().After(deadline) {
				return ErrWaitTimeout
			}
		}
	}
}

// ErrWaitTimeout is returned when waiting for a condition times out
var ErrWaitTimeout = &timeoutError{}

type timeoutError struct{}

func (e *timeoutError) Error() string {
	return "wait condition timeout"
}

// User is a test user model
type User struct {
	ID        interface{} `bson:"_id,omitempty"`
	Email     string      `bson:"email"`
	Name      string      `bson:"name"`
	Age       int         `bson:"age"`
	CreatedAt time.Time   `bson:"createdAt"`
	UpdatedAt time.Time   `bson:"updatedAt"`
}

// Post is a test post model
type Post struct {
	ID        interface{} `bson:"_id,omitempty"`
	UserID    interface{} `bson:"userId"`
	Title     string      `bson:"title"`
	Content   string      `bson:"content"`
	CreatedAt time.Time   `bson:"createdAt"`
	UpdatedAt time.Time   `bson:"updatedAt"`
}

// Comment is a test comment model
type Comment struct {
	ID        interface{} `bson:"_id,omitempty"`
	PostID    interface{} `bson:"postId"`
	UserID    interface{} `bson:"userId"`
	Content   string      `bson:"content"`
	CreatedAt time.Time   `bson:"createdAt"`
	UpdatedAt time.Time   `bson:"updatedAt"`
}

// Collection returns the collection name for User
func (u User) Collection() string {
	return "users"
}

// Collection returns the collection name for Post
func (p Post) Collection() string {
	return "posts"
}

// Collection returns the collection name for Comment
func (c Comment) Collection() string {
	return "comments"
}
