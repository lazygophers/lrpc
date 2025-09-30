package db_test

import (
	"os"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"gotest.tools/v3/assert"
)

// TestProduct is a test model for scoop testing
type TestProduct struct {
	Id          int    `gorm:"primaryKey;autoIncrement"`
	Name        string `gorm:"size:100;not null"`
	Description string `gorm:"size:500"`
	Price       float64 `gorm:"not null"`
	Stock       int    `gorm:"default:0"`
	CategoryId  int    `gorm:"index"`
	IsActive    bool   `gorm:"default:true"`
	CreatedAt   int64  `gorm:"autoCreateTime"`
	UpdatedAt   int64  `gorm:"autoUpdateTime"`
	DeletedAt   int64  `gorm:"index"`
}

func (TestProduct) TableName() string {
	return "test_products"
}

// setupTestDBForScoop creates a test database with TestProduct table
func setupTestDBForScoop(t *testing.T) *db.Client {
	tempDir, err := os.MkdirTemp("", "db_scoop_test_*")
	assert.NilError(t, err)

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	cfg := &db.Config{
		Type:    db.Sqlite,
		Address: tempDir,
		Name:    "test",
	}

	client, err := db.New(cfg, &TestProduct{})
	assert.NilError(t, err)
	assert.Assert(t, client != nil)

	return client
}

// insertTestProducts inserts sample test products
func insertTestProducts(t *testing.T, client *db.Client) []*TestProduct {
	products := []*TestProduct{
		{Name: "Product 1", Description: "Desc 1", Price: 100.0, Stock: 10, CategoryId: 1, IsActive: true},
		{Name: "Product 2", Description: "Desc 2", Price: 200.0, Stock: 20, CategoryId: 1, IsActive: true},
		{Name: "Product 3", Description: "Desc 3", Price: 300.0, Stock: 30, CategoryId: 2, IsActive: false},
		{Name: "Product 4", Description: "Desc 4", Price: 400.0, Stock: 40, CategoryId: 2, IsActive: true},
		{Name: "Product 5", Description: "Desc 5", Price: 500.0, Stock: 50, CategoryId: 3, IsActive: true},
	}

	for _, p := range products {
		result := client.NewScoop().Model(p).Create(p)
		assert.NilError(t, result.Error)
		assert.Assert(t, result.RowsAffected > 0, "Create should affect at least 1 row")
	}

	return products
}

// TestScoopFind tests the Find method of Scoop
func TestScoopFind(t *testing.T) {
	client := setupTestDBForScoop(t)
	insertTestProducts(t, client)

	t.Run("find all", func(t *testing.T) {
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).Find(&products)
		assert.NilError(t, result.Error)
		assert.Equal(t, len(products), 5)
	})

	t.Run("find with equal condition", func(t *testing.T) {
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).Equal("category_id", 1).Find(&products)
		assert.NilError(t, result.Error)
		assert.Equal(t, len(products), 2)
	})

	t.Run("find with limit", func(t *testing.T) {
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).Limit(3).Find(&products)
		assert.NilError(t, result.Error)
		assert.Equal(t, len(products), 3)
	})

	t.Run("find with offset and limit", func(t *testing.T) {
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).Offset(2).Limit(2).Find(&products)
		assert.NilError(t, result.Error)
		assert.Equal(t, len(products), 2)
	})
}

// TestScoopFirst tests the First method of Scoop
func TestScoopFirst(t *testing.T) {
	client := setupTestDBForScoop(t)
	insertTestProducts(t, client)

	t.Run("first record", func(t *testing.T) {
		var product TestProduct
		result := client.NewScoop().Model(&TestProduct{}).First(&product)
		assert.NilError(t, result.Error)
		assert.Assert(t, product.Id > 0)
	})

	t.Run("first with condition", func(t *testing.T) {
		var product TestProduct
		result := client.NewScoop().Model(&TestProduct{}).Equal("name", "Product 3").First(&product)
		assert.NilError(t, result.Error)
		assert.Equal(t, product.Name, "Product 3")
	})

	t.Run("first not found", func(t *testing.T) {
		var product TestProduct
		result := client.NewScoop().Model(&TestProduct{}).Equal("name", "NonExistent").First(&product)
		assert.Assert(t, result.Error != nil)
	})
}

// TestScoopCreate tests the Create method of Scoop
func TestScoopCreate(t *testing.T) {
	client := setupTestDBForScoop(t)

	t.Run("create single product", func(t *testing.T) {
		product := &TestProduct{
			Name:        "New Product",
			Description: "New Description",
			Price:       150.0,
			Stock:       15,
			CategoryId:  1,
			IsActive:    true,
		}

		result := client.NewScoop().Model(&TestProduct{}).Create(product)
		assert.NilError(t, result.Error)
		assert.Assert(t, product.Id > 0)

		// Verify the product can be found
		var found TestProduct
		findResult := client.NewScoop().Model(&TestProduct{}).Where("id", product.Id).First(&found)
		assert.NilError(t, findResult.Error, "Should be able to find the created product")
		assert.Equal(t, found.Name, product.Name)
	})
}

// TestScoopUpdate tests the Updates method of Scoop
func TestScoopUpdate(t *testing.T) {
	client := setupTestDBForScoop(t)
	products := insertTestProducts(t, client)

	t.Run("update single field", func(t *testing.T) {
		result := client.NewScoop().Model(&TestProduct{}).Equal("id", products[0].Id).Updates(map[string]interface{}{
			"price": 150.0,
		})
		assert.NilError(t, result.Error)

		var updated TestProduct
		firstResult := client.NewScoop().Model(&TestProduct{}).Equal("id", products[0].Id).First(&updated)
		assert.NilError(t, firstResult.Error)
		assert.Equal(t, updated.Price, 150.0)
	})

	t.Run("update multiple fields", func(t *testing.T) {
		result := client.NewScoop().Model(&TestProduct{}).Equal("id", products[1].Id).Updates(map[string]interface{}{
			"price": 250.0,
			"stock": 25,
		})
		assert.NilError(t, result.Error)

		var updated TestProduct
		firstResult := client.NewScoop().Model(&TestProduct{}).Equal("id", products[1].Id).First(&updated)
		assert.NilError(t, firstResult.Error)
		assert.Equal(t, updated.Price, 250.0)
		assert.Equal(t, updated.Stock, 25)
	})
}

// TestScoopDelete tests the Delete method of Scoop
func TestScoopDelete(t *testing.T) {
	client := setupTestDBForScoop(t)
	products := insertTestProducts(t, client)

	t.Run("soft delete", func(t *testing.T) {
		result := client.NewScoop().Model(&TestProduct{}).Equal("id", products[0].Id).Delete()
		assert.NilError(t, result.Error)

		var product TestProduct
		firstResult := client.NewScoop().Model(&TestProduct{}).Equal("id", products[0].Id).First(&product)
		assert.Assert(t, firstResult.Error != nil)
	})

	t.Run("unscoped find after delete", func(t *testing.T) {
		result := client.NewScoop().Model(&TestProduct{}).Equal("id", products[1].Id).Delete()
		assert.NilError(t, result.Error)

		var product TestProduct
		firstResult := client.NewScoop().Model(&TestProduct{}).Unscoped().Equal("id", products[1].Id).First(&product)
		assert.NilError(t, firstResult.Error)
		assert.Assert(t, product.DeletedAt != 0)
	})
}

// TestScoopCount tests the Count method of Scoop
func TestScoopCount(t *testing.T) {
	client := setupTestDBForScoop(t)
	insertTestProducts(t, client)

	t.Run("count all", func(t *testing.T) {
		count, err := client.NewScoop().Model(&TestProduct{}).Count()
		assert.NilError(t, err)
		assert.Equal(t, count, uint64(5))
	})

	t.Run("count with condition", func(t *testing.T) {
		count, err := client.NewScoop().Model(&TestProduct{}).Equal("category_id", 1).Count()
		assert.NilError(t, err)
		assert.Equal(t, count, uint64(2))
	})
}

// TestScoopExist tests the Exist method of Scoop
func TestScoopExist(t *testing.T) {
	client := setupTestDBForScoop(t)
	products := insertTestProducts(t, client)

	t.Run("exist true", func(t *testing.T) {
		exists, err := client.NewScoop().Model(&TestProduct{}).Equal("id", products[0].Id).Exist()
		assert.NilError(t, err)
		assert.Assert(t, exists)
	})

	t.Run("exist false", func(t *testing.T) {
		exists, err := client.NewScoop().Model(&TestProduct{}).Equal("id", 99999).Exist()
		assert.NilError(t, err)
		assert.Assert(t, !exists)
	})
}

// TestScoopWhereConditions tests various Where condition methods
func TestScoopWhereConditions(t *testing.T) {
	client := setupTestDBForScoop(t)
	insertTestProducts(t, client)

	t.Run("NotEqual condition", func(t *testing.T) {
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).NotEqual("category_id", 1).Find(&products)
		assert.NilError(t, result.Error)
		assert.Equal(t, len(products), 3)
	})

	t.Run("In condition", func(t *testing.T) {
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).In("category_id", []int{1, 2}).Find(&products)
		assert.NilError(t, result.Error)
		assert.Equal(t, len(products), 4)
	})

	t.Run("In with empty slice", func(t *testing.T) {
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).In("category_id", []int{}).Find(&products)
		assert.NilError(t, result.Error)
		assert.Equal(t, len(products), 0)
	})

	t.Run("NotIn condition", func(t *testing.T) {
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).NotIn("category_id", []int{1}).Find(&products)
		assert.NilError(t, result.Error)
		assert.Equal(t, len(products), 3)
	})

	t.Run("NotIn with empty slice", func(t *testing.T) {
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).NotIn("category_id", []int{}).Find(&products)
		assert.NilError(t, result.Error)
		assert.Equal(t, len(products), 5)
	})

	t.Run("Like condition", func(t *testing.T) {
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).Like("name", "%Product%").Find(&products)
		assert.NilError(t, result.Error)
		assert.Equal(t, len(products), 5)
	})

	t.Run("LeftLike condition", func(t *testing.T) {
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).LeftLike("name", "1").Find(&products)
		assert.NilError(t, result.Error)
		assert.Equal(t, len(products), 1)
	})

	t.Run("RightLike condition", func(t *testing.T) {
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).RightLike("name", "Product").Find(&products)
		assert.NilError(t, result.Error)
		assert.Equal(t, len(products), 5)
	})

	t.Run("NotLike condition", func(t *testing.T) {
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).NotLike("name", "%NonExistent%").Find(&products)
		assert.NilError(t, result.Error)
		assert.Equal(t, len(products), 5)
	})

	t.Run("NotLeftLike condition", func(t *testing.T) {
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).NotLeftLike("name", "X").Find(&products)
		assert.NilError(t, result.Error)
		assert.Equal(t, len(products), 5)
	})

	t.Run("NotRightLike condition", func(t *testing.T) {
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).NotRightLike("name", "NonExistent").Find(&products)
		assert.NilError(t, result.Error)
		assert.Equal(t, len(products), 5)
	})

	t.Run("Between condition", func(t *testing.T) {
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).Between("price", 200, 400).Find(&products)
		assert.NilError(t, result.Error)
		assert.Equal(t, len(products), 3)
	})

	t.Run("NotBetween condition", func(t *testing.T) {
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).NotBetween("price", 200, 400).Find(&products)
		assert.NilError(t, result.Error)
		assert.Equal(t, len(products), 2)
	})

	t.Run("Group condition", func(t *testing.T) {
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).Group("category_id").Find(&products)
		assert.NilError(t, result.Error)
		assert.Assert(t, len(products) > 0)
	})

	t.Run("Order condition", func(t *testing.T) {
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).Order("price").Find(&products)
		assert.NilError(t, result.Error)
		assert.Equal(t, len(products), 5)
		assert.Assert(t, products[0].Price <= products[1].Price)
	})
}

// TestScoopTransaction tests transaction methods
func TestScoopTransaction(t *testing.T) {
	client := setupTestDBForScoop(t)

	t.Run("commit transaction", func(t *testing.T) {
		tx := client.NewScoop().Model(&TestProduct{}).Begin()
		assert.Assert(t, tx != nil)

		product := &TestProduct{
			Name:        "TX Product",
			Description: "TX Description",
			Price:       100.0,
			Stock:       10,
			CategoryId:  1,
			IsActive:    true,
		}

		result := tx.Create(product)
		assert.NilError(t, result.Error)

		committedTx := tx.Commit()
		assert.Assert(t, committedTx != nil)

		count, err := client.NewScoop().Model(&TestProduct{}).Equal("name", "TX Product").Count()
		assert.NilError(t, err)
		assert.Equal(t, count, uint64(1))
	})

	t.Run("rollback transaction", func(t *testing.T) {
		tx := client.NewScoop().Model(&TestProduct{}).Begin()
		assert.Assert(t, tx != nil)

		product := &TestProduct{
			Name:        "RB Product",
			Description: "RB Description",
			Price:       200.0,
			Stock:       20,
			CategoryId:  2,
			IsActive:    true,
		}

		result := tx.Create(product)
		assert.NilError(t, result.Error)

		rolledBackTx := tx.Rollback()
		assert.Assert(t, rolledBackTx != nil)

		count, err := client.NewScoop().Model(&TestProduct{}).Equal("name", "RB Product").Count()
		assert.NilError(t, err)
		assert.Equal(t, count, uint64(0))
	})
}

// TestScoopCreateInBatches tests batch creation
func TestScoopCreateInBatches(t *testing.T) {
	client := setupTestDBForScoop(t)

	t.Run("create in batches", func(t *testing.T) {
		products := []*TestProduct{
			{Name: "Batch 1", Description: "Batch Desc 1", Price: 100.0, Stock: 10, CategoryId: 1, IsActive: true},
			{Name: "Batch 2", Description: "Batch Desc 2", Price: 200.0, Stock: 20, CategoryId: 1, IsActive: true},
			{Name: "Batch 3", Description: "Batch Desc 3", Price: 300.0, Stock: 30, CategoryId: 2, IsActive: true},
		}

		result := client.NewScoop().Model(&TestProduct{}).CreateInBatches(products, 2)
		assert.NilError(t, result.Error)

		count, err := client.NewScoop().Model(&TestProduct{}).Like("name", "Batch%").Count()
		assert.NilError(t, err)
		assert.Equal(t, count, uint64(3))
	})
}

// TestScoopChunk tests chunk processing
func TestScoopChunk(t *testing.T) {
	client := setupTestDBForScoop(t)
	insertTestProducts(t, client)

	t.Run("chunk processing", func(t *testing.T) {
		var totalProcessed int
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).Chunk(&products, 2, func(tx *db.Scoop, offset uint64) error {
			totalProcessed += len(products)
			return nil
		})
		assert.NilError(t, result.Error)
		assert.Equal(t, totalProcessed, 5)
	})
}

// TestScoopHelperMethods tests helper methods
func TestScoopHelperMethods(t *testing.T) {
	client := setupTestDBForScoop(t)

	t.Run("IsNotFound", func(t *testing.T) {
		var product TestProduct
		result := client.NewScoop().Model(&TestProduct{}).Equal("id", 99999).First(&product)
		assert.Assert(t, client.NewScoop().IsNotFound(result.Error))
	})

	t.Run("Model method", func(t *testing.T) {
		scoop := client.NewScoop().Model(&TestProduct{}).Model(&TestProduct{})
		assert.Assert(t, scoop != nil)
	})
}

// TestScoopSelect tests Select method
func TestScoopSelect(t *testing.T) {
	client := setupTestDBForScoop(t)
	insertTestProducts(t, client)

	t.Run("select specific fields", func(t *testing.T) {
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).Select("id", "name").Find(&products)
		assert.NilError(t, result.Error)
		assert.Equal(t, len(products), 5)
	})
}

// TestScoopUnscoped tests Unscoped method
func TestScoopUnscoped(t *testing.T) {
	client := setupTestDBForScoop(t)
	products := insertTestProducts(t, client)

	t.Run("unscoped find after soft delete", func(t *testing.T) {
		result := client.NewScoop().Model(&TestProduct{}).Equal("id", products[0].Id).Delete()
		assert.NilError(t, result.Error)

		var allProducts []*TestProduct
		findResult := client.NewScoop().Model(&TestProduct{}).Unscoped().Find(&allProducts)
		assert.NilError(t, findResult.Error)
		assert.Equal(t, len(allProducts), 5)
	})
}
// TestCategory is a test model for join testing
type TestCategory struct {
	Id        int    `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"size:100;not null"`
	CreatedAt int64  `gorm:"autoCreateTime"`
	DeletedAt int64  `gorm:"index"`
}

func (TestCategory) TableName() string {
	return "test_categories"
}

// TestScoopJoin tests Join methods
func TestScoopJoin(t *testing.T) {
	client := setupTestDBForScoop(t)
	
	// Create test categories table
	err := client.AutoMigrate(&TestCategory{})
	assert.NilError(t, err)
	
	// Insert test categories
	categories := []TestCategory{
		{Name: "Electronics"},
		{Name: "Books"},
		{Name: "Clothing"},
	}
	for i := range categories {
		result := client.NewScoop().Model(&TestCategory{}).Create(&categories[i])
		assert.NilError(t, result.Error)
	}
	
	// Insert test products with categories
	insertTestProducts(t, client)
	
	t.Run("inner join", func(t *testing.T) {
		sql := client.NewScoop().Table("test_products").
			Select("test_products.name", "test_categories.name as category_name").
			InnerJoin("test_categories", "test_products.category_id = test_categories.id").
			Where("test_products.deleted_at", 0)

		sqlStr := sql.ToSQL()
		t.Logf("Generated SQL: %s", sqlStr)
		assert.Assert(t, len(sqlStr) > 0)
	})
	
	t.Run("left join", func(t *testing.T) {
		sql := client.NewScoop().Table("test_products").
			Select("test_products.*", "test_categories.name as category_name").
			LeftJoin("test_categories", "test_products.category_id = test_categories.id").
			Where("test_products.deleted_at", 0)
		
		sqlStr := sql.ToSQL()
		t.Logf("Generated SQL: %s", sqlStr)
		assert.Assert(t, len(sqlStr) > 0)
	})
	
	t.Run("right join", func(t *testing.T) {
		sql := client.NewScoop().Table("test_products").
			Select("test_products.*").
			RightJoin("test_categories", "test_products.category_id = test_categories.id")
		
		sqlStr := sql.ToSQL()
		t.Logf("Generated SQL: %s", sqlStr)
		assert.Assert(t, len(sqlStr) > 0)
	})
	
	t.Run("multiple joins", func(t *testing.T) {
		sql := client.NewScoop().Table("test_products").
			InnerJoin("test_categories", "test_products.category_id = test_categories.id").
			LeftJoin("test_another_table", "test_products.id = test_another_table.product_id")
		
		sqlStr := sql.ToSQL()
		t.Logf("Generated SQL: %s", sqlStr)
		assert.Assert(t, len(sqlStr) > 0)
	})
	
	t.Run("cross join", func(t *testing.T) {
		sql := client.NewScoop().Table("test_products").
			CrossJoin("test_categories")
		
		sqlStr := sql.ToSQL()
		t.Logf("Generated SQL: %s", sqlStr)
		assert.Assert(t, len(sqlStr) > 0)
	})
}

// TestScoopHaving tests Having method
func TestScoopHaving(t *testing.T) {
	client := setupTestDBForScoop(t)
	insertTestProducts(t, client)
	
	t.Run("having with count", func(t *testing.T) {
		sql := client.NewScoop().Table("test_products").
			Select("category_id", "COUNT(*) as count").
			Group("category_id").
			Having("COUNT(*) > ?", 1)
		
		sqlStr := sql.ToSQL()
		t.Logf("Generated SQL: %s", sqlStr)
		assert.Assert(t, len(sqlStr) > 0)
	})
	
	t.Run("having with multiple conditions", func(t *testing.T) {
		sql := client.NewScoop().Table("test_products").
			Select("category_id", "SUM(price) as total_price").
			Group("category_id").
			Having("SUM(price) > ?", 100).
			Having("COUNT(*) > ?", 1)
		
		sqlStr := sql.ToSQL()
		t.Logf("Generated SQL: %s", sqlStr)
		assert.Assert(t, len(sqlStr) > 0)
	})
	
	t.Run("group by with having and order", func(t *testing.T) {
		sql := client.NewScoop().Table("test_products").
			Select("category_id", "AVG(price) as avg_price", "COUNT(*) as count").
			Where("deleted_at", 0).
			Group("category_id").
			Having("AVG(price) > ?", 150).
			Order("avg_price DESC")
		
		sqlStr := sql.ToSQL()
		t.Logf("Generated SQL: %s", sqlStr)
		assert.Assert(t, len(sqlStr) > 0)
	})
}

// TestScoopComplexQuery tests complex queries combining Join and Having
func TestScoopComplexQuery(t *testing.T) {
	client := setupTestDBForScoop(t)
	
	// Create test categories table
	err := client.AutoMigrate(&TestCategory{})
	assert.NilError(t, err)
	
	// Insert test data
	categories := []TestCategory{
		{Name: "Electronics"},
		{Name: "Books"},
		{Name: "Clothing"},
	}
	for i := range categories {
		result := client.NewScoop().Model(&TestCategory{}).Create(&categories[i])
		assert.NilError(t, result.Error)
	}
	
	insertTestProducts(t, client)
	
	t.Run("join with group by and having", func(t *testing.T) {
		sql := client.NewScoop().Table("test_products").
			Select("test_categories.name as category", "COUNT(*) as product_count", "AVG(test_products.price) as avg_price").
			InnerJoin("test_categories", "test_products.category_id = test_categories.id").
			Where("test_products.deleted_at", 0).
			Group("test_categories.name").
			Having("COUNT(*) > ?", 0).
			Order("product_count DESC")
		
		sqlStr := sql.ToSQL()
		t.Logf("Generated SQL: %s", sqlStr)
		assert.Assert(t, len(sqlStr) > 0)
	})
	
	t.Run("multiple joins with having", func(t *testing.T) {
		sql := client.NewScoop().Table("test_products").
			Select("test_categories.name", "SUM(test_products.price) as total").
			InnerJoin("test_categories", "test_products.category_id = test_categories.id").
			Where("test_products.is_active", true).
			Group("test_categories.name").
			Having("SUM(test_products.price) > ?", 200).
			Limit(10)
		
		sqlStr := sql.ToSQL()
		t.Logf("Generated SQL: %s", sqlStr)
		assert.Assert(t, len(sqlStr) > 0)
	})
}

// TestScoopJoinBackwardCompatibility tests that existing code still works
func TestScoopJoinBackwardCompatibility(t *testing.T) {
	client := setupTestDBForScoop(t)
	insertTestProducts(t, client)
	
	t.Run("existing queries without joins", func(t *testing.T) {
		var products []*TestProduct
		result := client.NewScoop().Model(&TestProduct{}).
			Where("price > ?", 100).
			Order("price DESC").
			Limit(5).
			Find(&products)
		
		assert.NilError(t, result.Error)
		assert.Assert(t, len(products) > 0)
	})
	
	t.Run("existing group by without having", func(t *testing.T) {
		sql := client.NewScoop().Table("test_products").
			Select("category_id", "COUNT(*) as count").
			Group("category_id").
			Order("count DESC")
		
		sqlStr := sql.ToSQL()
		t.Logf("Generated SQL: %s", sqlStr)
		assert.Assert(t, len(sqlStr) > 0)
	})
}
