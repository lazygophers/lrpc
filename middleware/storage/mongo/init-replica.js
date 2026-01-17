// Initialize MongoDB Replica Set
sleep(2000); // Wait for MongoDB to start
var status = rs.status();

if (status.ok === 0) {
  // Replica set not initialized
  rs.initiate(
    {
      _id: "rs0",
      members: [
        {
          _id: 0,
          host: "mongo:27017"
        }
      ]
    },
    { force: true }
  );
} else if (status.members.length === 0) {
  // Replica set initialized but no members
  rs.add("mongo:27017");
}

// Wait for replica set to be ready
var count = 0;
while (rs.status().members[0].state !== 1 && count < 100) {
  sleep(500);
  count++;
}

print("Replica set initialized successfully");
