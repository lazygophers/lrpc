package db_test

// TestUser model for testing database operations
// NOTE: This is a duplicate definition for db_test package
// The canonical definition is in db package (test_models.go)
type TestUser struct {
	Id        int64  `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"size:100;not null"`
	Email     string `gorm:"size:100;unique"`
	Age       int    `gorm:"default:0"`
	CreatedAt int64  `gorm:"autoCreateTime"`
	UpdatedAt int64  `gorm:"autoUpdateTime"`
	DeletedAt int64  `gorm:"index;default:0"`
}

func (TestUser) TableName() string {
	return "test_users"
}

// TestProfile model for testing relationships
type TestProfile struct {
	Id       int64  `gorm:"primaryKey;autoIncrement"`
	UserId   int64  `gorm:"index"`
	Bio      string `gorm:"type:text"`
	Website  string `gorm:"size:255"`
}

func (TestProfile) TableName() string {
	return "test_profiles"
}
