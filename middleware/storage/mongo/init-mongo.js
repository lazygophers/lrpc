// Initialize MongoDB test database
db = db.getSiblingDB('test');

// Create collections
db.createCollection('users');
db.createCollection('posts');
db.createCollection('comments');

// Create indexes
db.users.createIndex({ email: 1 }, { unique: true });
db.users.createIndex({ createdAt: 1 });

db.posts.createIndex({ userId: 1 });
db.posts.createIndex({ createdAt: -1 });

db.comments.createIndex({ postId: 1 });
db.comments.createIndex({ userId: 1 });

// Insert sample data for testing
db.users.insertMany([
  {
    _id: ObjectId("000000000000000000000001"),
    email: "user1@example.com",
    name: "User One",
    age: 25,
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    _id: ObjectId("000000000000000000000002"),
    email: "user2@example.com",
    name: "User Two",
    age: 30,
    createdAt: new Date(),
    updatedAt: new Date()
  }
]);

db.posts.insertMany([
  {
    _id: ObjectId("100000000000000000000001"),
    userId: ObjectId("000000000000000000000001"),
    title: "First Post",
    content: "This is the first post",
    createdAt: new Date(),
    updatedAt: new Date()
  },
  {
    _id: ObjectId("100000000000000000000002"),
    userId: ObjectId("000000000000000000000001"),
    title: "Second Post",
    content: "This is the second post",
    createdAt: new Date(),
    updatedAt: new Date()
  }
]);

print("MongoDB initialized successfully");
