package db_test

import (
	"os"
	"testing"
	"time"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"gotest.tools/v3/assert"
)

// TestModelWithYAML tests a model with YAML serialized field
type TestModelWithYAML struct {
	Id        int            `gorm:"primaryKey;autoIncrement"`
	Name      string         `gorm:"size:100;not null"`
	Config    map[string]any `gorm:"column:config;type:text;serializer:yaml;not null" yaml:"config,omitempty"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
}

func (TestModelWithYAML) TableName() string {
	return "test_yaml_models"
}

// TestModelWithTOML tests a model with TOML serialized field
type TestModelWithTOML struct {
	Id        int            `gorm:"primaryKey;autoIncrement"`
	Name      string         `gorm:"size:100;not null"`
	Settings  map[string]any `gorm:"column:settings;type:text;serializer:toml;not null"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
}

func (TestModelWithTOML) TableName() string {
	return "test_toml_models"
}

// TestModelWithBSON tests a model with BSON serialized field
type TestModelWithBSON struct {
	Id        int            `gorm:"primaryKey;autoIncrement"`
	Name      string         `gorm:"size:100;not null"`
	Data      map[string]any `gorm:"column:data;type:blob;serializer:bson;not null"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
}

func (TestModelWithBSON) TableName() string {
	return "test_bson_models"
}

// TestConfigStruct for INI serialization testing
type TestConfigStruct struct {
	Server   ServerConfig   `ini:"server"`
	Database DatabaseConfig `ini:"database"`
}

type ServerConfig struct {
	Host string `ini:"host"`
	Port int    `ini:"port"`
}

type DatabaseConfig struct {
	Name     string `ini:"name"`
	User     string `ini:"user"`
	Password string `ini:"password"`
}

// TestModelWithINI tests a model with INI serialized field
type TestModelWithINI struct {
	Id        int              `gorm:"primaryKey;autoIncrement"`
	Name      string           `gorm:"size:100;not null"`
	Config    TestConfigStruct `gorm:"column:config;type:text;serializer:ini;not null"`
	CreatedAt time.Time        `gorm:"autoCreateTime"`
	UpdatedAt time.Time        `gorm:"autoUpdateTime"`
}

func (TestModelWithINI) TableName() string {
	return "test_ini_models"
}

// TestYAMLSerializer tests the YAML serializer
func TestYAMLSerializer(t *testing.T) {
	t.Run("test YAML serialization with Scoop", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_yaml_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_yaml",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithYAML{})
		assert.NilError(t, err)
		assert.Assert(t, client != nil)

		model := db.NewModel[TestModelWithYAML](client)

		testData := TestModelWithYAML{
			Name: "yaml_test",
			Config: map[string]any{
				"enabled": true,
				"timeout": 30,
				"servers": []any{"server1", "server2"},
			},
		}

		err = model.NewScoop().Create(&testData)
		assert.NilError(t, err)
		assert.Assert(t, testData.Id > 0)

		result, err := model.NewScoop().Equal("id", testData.Id).First()
		assert.NilError(t, err)
		assert.Assert(t, result != nil)

		assert.Equal(t, testData.Name, result.Name)
		assert.Assert(t, result.Config != nil)
		assert.Equal(t, true, result.Config["enabled"])
		assert.Equal(t, 30, result.Config["timeout"])
	})

	t.Run("test YAML update with Scoop", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_yaml_update_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_yaml_update",
		}

		client, err := db.New(config, TestModelWithYAML{})
		assert.NilError(t, err)

		model := db.NewModel[TestModelWithYAML](client)

		testData := TestModelWithYAML{
			Name: "yaml_update",
			Config: map[string]any{
				"initial": "value",
			},
		}

		err = model.NewScoop().Create(&testData)
		assert.NilError(t, err)

		testData.Config = map[string]any{
			"updated": "new_value",
			"count":   100,
		}

		updateResult := model.NewScoop().Updates(&testData)
		assert.NilError(t, updateResult.Error)

		result, err := model.NewScoop().Equal("id", testData.Id).First()
		assert.NilError(t, err)

		assert.Equal(t, "new_value", result.Config["updated"])
		assert.Equal(t, 100, result.Config["count"])
	})
}

// TestTOMLSerializer tests the TOML serializer
func TestTOMLSerializer(t *testing.T) {
	t.Run("test TOML serialization with Scoop", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_toml_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_toml",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithTOML{})
		assert.NilError(t, err)
		assert.Assert(t, client != nil)

		model := db.NewModel[TestModelWithTOML](client)

		testData := TestModelWithTOML{
			Name: "toml_test",
			Settings: map[string]any{
				"debug":   true,
				"workers": int64(10),
				"version": "1.0.0",
			},
		}

		err = model.NewScoop().Create(&testData)
		assert.NilError(t, err)
		assert.Assert(t, testData.Id > 0)

		result, err := model.NewScoop().Equal("id", testData.Id).First()
		assert.NilError(t, err)
		assert.Assert(t, result != nil)

		assert.Equal(t, testData.Name, result.Name)
		assert.Assert(t, result.Settings != nil)
		assert.Equal(t, true, result.Settings["debug"])
		assert.Equal(t, int64(10), result.Settings["workers"])
		assert.Equal(t, "1.0.0", result.Settings["version"])
	})

	t.Run("test TOML batch operations", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_toml_batch_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_toml_batch",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithTOML{})
		assert.NilError(t, err)

		testData := []TestModelWithTOML{
			{
				Name: "batch1",
				Settings: map[string]any{
					"id":   int64(1),
					"type": "test",
				},
			},
			{
				Name: "batch2",
				Settings: map[string]any{
					"id":   int64(2),
					"type": "test",
				},
			},
		}

		result := client.NewScoop().CreateInBatches(&testData, 2)
		assert.NilError(t, result.Error)
		assert.Equal(t, int64(2), result.RowsAffected)

		model := db.NewModel[TestModelWithTOML](client)
		results, err := model.NewScoop().Find()
		assert.NilError(t, err)
		assert.Equal(t, 2, len(results))

		for i, r := range results {
			assert.Equal(t, int64(i+1), r.Settings["id"])
			assert.Equal(t, "test", r.Settings["type"])
		}
	})
}

// TestBSONSerializer tests the BSON serializer
func TestBSONSerializer(t *testing.T) {
	t.Run("test BSON serialization with Scoop", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_bson_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_bson",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithBSON{})
		assert.NilError(t, err)
		assert.Assert(t, client != nil)

		model := db.NewModel[TestModelWithBSON](client)

		testData := TestModelWithBSON{
			Name: "bson_test",
			Data: map[string]any{
				"field1": "value1",
				"field2": int32(42),
				"field3": true,
				"nested": map[string]any{
					"key": "nested_value",
				},
			},
		}

		err = model.NewScoop().Create(&testData)
		assert.NilError(t, err)
		assert.Assert(t, testData.Id > 0)

		result, err := model.NewScoop().Equal("id", testData.Id).First()
		assert.NilError(t, err)
		assert.Assert(t, result != nil)

		assert.Equal(t, testData.Name, result.Name)
		assert.Assert(t, result.Data != nil)
		assert.Equal(t, "value1", result.Data["field1"])
		assert.Equal(t, int32(42), result.Data["field2"])
		assert.Equal(t, true, result.Data["field3"])
	})

	t.Run("test BSON Find operations", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_bson_find_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_bson_find",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithBSON{})
		assert.NilError(t, err)

		model := db.NewModel[TestModelWithBSON](client)

		testData := []TestModelWithBSON{
			{
				Name: "user1",
				Data: map[string]any{
					"role":  "admin",
					"level": int32(10),
				},
			},
			{
				Name: "user2",
				Data: map[string]any{
					"role":  "user",
					"level": int32(5),
				},
			},
		}

		for i := range testData {
			err = model.NewScoop().Create(&testData[i])
			assert.NilError(t, err)
		}

		results, err := model.NewScoop().Find()
		assert.NilError(t, err)
		assert.Equal(t, 2, len(results))

		assert.Equal(t, "admin", results[0].Data["role"])
		assert.Equal(t, int32(10), results[0].Data["level"])
		assert.Equal(t, "user", results[1].Data["role"])
		assert.Equal(t, int32(5), results[1].Data["level"])
	})
}

// TestINISerializer tests the INI serializer
func TestINISerializer(t *testing.T) {
	t.Run("test INI serialization with Scoop", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_ini_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_ini",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithINI{})
		assert.NilError(t, err)
		assert.Assert(t, client != nil)

		model := db.NewModel[TestModelWithINI](client)

		testData := TestModelWithINI{
			Name: "ini_test",
			Config: TestConfigStruct{
				Server: ServerConfig{
					Host: "localhost",
					Port: 8080,
				},
				Database: DatabaseConfig{
					Name:     "testdb",
					User:     "admin",
					Password: "secret",
				},
			},
		}

		err = model.NewScoop().Create(&testData)
		assert.NilError(t, err)
		assert.Assert(t, testData.Id > 0)

		result, err := model.NewScoop().Equal("id", testData.Id).First()
		assert.NilError(t, err)
		assert.Assert(t, result != nil)

		assert.Equal(t, testData.Name, result.Name)
		assert.Equal(t, "localhost", result.Config.Server.Host)
		assert.Equal(t, 8080, result.Config.Server.Port)
		assert.Equal(t, "testdb", result.Config.Database.Name)
		assert.Equal(t, "admin", result.Config.Database.User)
		assert.Equal(t, "secret", result.Config.Database.Password)
	})

	t.Run("test INI update operations", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_ini_update_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_ini_update",
		}

		client, err := db.New(config, TestModelWithINI{})
		assert.NilError(t, err)

		model := db.NewModel[TestModelWithINI](client)

		testData := TestModelWithINI{
			Name: "ini_update",
			Config: TestConfigStruct{
				Server: ServerConfig{
					Host: "localhost",
					Port: 3000,
				},
				Database: DatabaseConfig{
					Name: "olddb",
					User: "user",
				},
			},
		}

		err = model.NewScoop().Create(&testData)
		assert.NilError(t, err)

		testData.Config.Server.Port = 9090
		testData.Config.Database.Name = "newdb"

		updateResult := model.NewScoop().Updates(&testData)
		assert.NilError(t, updateResult.Error)

		result, err := model.NewScoop().Equal("id", testData.Id).First()
		assert.NilError(t, err)

		assert.Equal(t, 9090, result.Config.Server.Port)
		assert.Equal(t, "newdb", result.Config.Database.Name)
	})
}
