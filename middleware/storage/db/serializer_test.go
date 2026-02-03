package db_test

import (
	"os"
	"testing"
	"time"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/lazygophers/lrpc/middleware/storage/db/testdata"
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

// TestModelWithSimpleProto 测试简单 protobuf 消息的序列化
type TestModelWithSimpleProto struct {
	Id        int                     `gorm:"primaryKey;autoIncrement"`
	Name      string                  `gorm:"size:100;not null"`
	Message   *testdata.SimpleMessage `gorm:"column:message;type:text;serializer:protojson"`
	CreatedAt int64                   `gorm:"autoCreateTime"`
	UpdatedAt int64                   `gorm:"autoUpdateTime"`
}

func (TestModelWithSimpleProto) TableName() string {
	return "test_simple_proto_models"
}

// TestModelWithOneofProto 测试包含 oneof 字段的 protobuf 消息的序列化
type TestModelWithOneofProto struct {
	Id        int                    `gorm:"primaryKey;autoIncrement"`
	Name      string                 `gorm:"size:100;not null"`
	Message   *testdata.ModelMessage `gorm:"column:message;type:text;serializer:protojson"`
	CreatedAt int64                  `gorm:"autoCreateTime"`
	UpdatedAt int64                  `gorm:"autoUpdateTime"`
}

func (TestModelWithOneofProto) TableName() string {
	return "test_oneof_proto_models"
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
// TestProtojsonSerializer 测试 protojson 序列化器
func TestProtojsonSerializer(t *testing.T) {
	t.Run("test simple protobuf message serialization", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_protojson_simple_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_protojson_simple",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithSimpleProto{})
		assert.NilError(t, err)
		assert.Assert(t, client != nil)

		model := db.NewModel[TestModelWithSimpleProto](client)

		// 创建测试数据
		testData := TestModelWithSimpleProto{
			Name: "protojson_test",
			Message: &testdata.SimpleMessage{
				Name: "test_user",
				Age:  25,
				Tags: []string{"tag1", "tag2", "tag3"},
			},
		}

		// 测试 Create
		err = model.NewScoop().Create(&testData)
		assert.NilError(t, err)
		assert.Assert(t, testData.Id > 0)

		// 测试 First
		result, err := model.NewScoop().Equal("id", testData.Id).First()
		assert.NilError(t, err)
		assert.Assert(t, result != nil)
		assert.Assert(t, result.Message != nil)
		assert.Equal(t, "test_user", result.Message.Name)
		assert.Equal(t, int32(25), result.Message.Age)
		assert.Equal(t, 3, len(result.Message.Tags))
		assert.Equal(t, "tag1", result.Message.Tags[0])
	})

	t.Run("test protobuf oneof field with text content", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_protojson_oneof_text_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_protojson_oneof_text",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithOneofProto{})
		assert.NilError(t, err)
		assert.Assert(t, client != nil)

		model := db.NewModel[TestModelWithOneofProto](client)

		// 创建带 TextContent 的测试数据
		testData := TestModelWithOneofProto{
			Name: "oneof_text_test",
			Message: &testdata.ModelMessage{
				Id:   100,
				Name: "test_model",
				Content: &testdata.ModelMessage_TextContent{
					TextContent: &testdata.TextContent{
						Text: "Hello, World!",
					},
				},
			},
		}

		// 测试 Create
		err = model.NewScoop().Create(&testData)
		assert.NilError(t, err)
		assert.Assert(t, testData.Id > 0)

		// 测试 First
		result, err := model.NewScoop().Equal("id", testData.Id).First()
		assert.NilError(t, err)
		assert.Assert(t, result != nil)
		assert.Assert(t, result.Message != nil)
		assert.Equal(t, int64(100), result.Message.Id)
		assert.Equal(t, "test_model", result.Message.Name)

		// 验证 oneof 字段
		textContent := result.Message.GetTextContent()
		assert.Assert(t, textContent != nil, "TextContent should not be nil")
		assert.Equal(t, "Hello, World!", textContent.Text)

		// 确保 ImageContent 为 nil
		imageContent := result.Message.GetImageContent()
		assert.Assert(t, imageContent == nil, "ImageContent should be nil")
	})

	t.Run("test protobuf oneof field with image content", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_protojson_oneof_image_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_protojson_oneof_image",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithOneofProto{})
		assert.NilError(t, err)
		assert.Assert(t, client != nil)

		model := db.NewModel[TestModelWithOneofProto](client)

		// 创建带 ImageContent 的测试数据
		testData := TestModelWithOneofProto{
			Name: "oneof_image_test",
			Message: &testdata.ModelMessage{
				Id:   200,
				Name: "image_model",
				Content: &testdata.ModelMessage_ImageContent{
					ImageContent: &testdata.ImageContent{
						Url:    "https://example.com/image.png",
						Width:  1920,
						Height: 1080,
					},
				},
			},
		}

		// 测试 Create
		err = model.NewScoop().Create(&testData)
		assert.NilError(t, err)
		assert.Assert(t, testData.Id > 0)

		// 测试 First
		result, err := model.NewScoop().Equal("id", testData.Id).First()
		assert.NilError(t, err)
		assert.Assert(t, result != nil)
		assert.Assert(t, result.Message != nil)
		assert.Equal(t, int64(200), result.Message.Id)
		assert.Equal(t, "image_model", result.Message.Name)

		// 验证 oneof 字段
		imageContent := result.Message.GetImageContent()
		assert.Assert(t, imageContent != nil, "ImageContent should not be nil")
		assert.Equal(t, "https://example.com/image.png", imageContent.Url)
		assert.Equal(t, int32(1920), imageContent.Width)
		assert.Equal(t, int32(1080), imageContent.Height)

		// 确保 TextContent 为 nil
		textContent := result.Message.GetTextContent()
		assert.Assert(t, textContent == nil, "TextContent should be nil")
	})

	t.Run("test protobuf message update", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_protojson_update_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_protojson_update",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithOneofProto{})
		assert.NilError(t, err)

		model := db.NewModel[TestModelWithOneofProto](client)

		// 创建初始数据（使用 TextContent）
		testData := TestModelWithOneofProto{
			Name: "update_test",
			Message: &testdata.ModelMessage{
				Id:   300,
				Name: "initial_model",
				Content: &testdata.ModelMessage_TextContent{
					TextContent: &testdata.TextContent{
						Text: "Initial text",
					},
				},
			},
		}

		err = model.NewScoop().Create(&testData)
		assert.NilError(t, err)

		// 更新为 ImageContent
		testData.Message = &testdata.ModelMessage{
			Id:   300,
			Name: "updated_model",
			Content: &testdata.ModelMessage_ImageContent{
				ImageContent: &testdata.ImageContent{
					Url:    "https://example.com/new.png",
					Width:  800,
					Height: 600,
				},
			},
		}

		updateResult := model.NewScoop().Updates(&testData)
		assert.NilError(t, updateResult.Error)

		// 验证更新
		result, err := model.NewScoop().Equal("id", testData.Id).First()
		assert.NilError(t, err)
		assert.Assert(t, result.Message != nil)
		assert.Equal(t, "updated_model", result.Message.Name)

		// 验证 oneof 字段已更新为 ImageContent
		imageContent := result.Message.GetImageContent()
		assert.Assert(t, imageContent != nil, "ImageContent should not be nil after update")
		assert.Equal(t, "https://example.com/new.png", imageContent.Url)

		// 确保 TextContent 为 nil
		textContent := result.Message.GetTextContent()
		assert.Assert(t, textContent == nil, "TextContent should be nil after update")
	})

	t.Run("test protobuf message batch operations", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_protojson_batch_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_protojson_batch",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithOneofProto{})
		assert.NilError(t, err)

		// 创建批量数据
		testData := []TestModelWithOneofProto{
			{
				Name: "batch1",
				Message: &testdata.ModelMessage{
					Id:   1,
					Name: "model1",
					Content: &testdata.ModelMessage_TextContent{
						TextContent: &testdata.TextContent{Text: "text1"},
					},
				},
			},
			{
				Name: "batch2",
				Message: &testdata.ModelMessage{
					Id:   2,
					Name: "model2",
					Content: &testdata.ModelMessage_ImageContent{
						ImageContent: &testdata.ImageContent{Url: "url2", Width: 100, Height: 100},
					},
				},
			},
			{
				Name: "batch3",
				Message: &testdata.ModelMessage{
					Id:   3,
					Name: "model3",
					Content: &testdata.ModelMessage_TextContent{
						TextContent: &testdata.TextContent{Text: "text3"},
					},
				},
			},
		}

		// 批量插入
		result := client.NewScoop().CreateInBatches(&testData, 2)
		assert.NilError(t, result.Error)
		assert.Equal(t, int64(3), result.RowsAffected)

		// 验证数据
		model := db.NewModel[TestModelWithOneofProto](client)
		results, err := model.NewScoop().Find()
		assert.NilError(t, err)
		assert.Equal(t, 3, len(results))

		// 验证第一条是 TextContent
		assert.Assert(t, results[0].Message.GetTextContent() != nil)
		assert.Equal(t, "text1", results[0].Message.GetTextContent().Text)

		// 验证第二条是 ImageContent
		assert.Assert(t, results[1].Message.GetImageContent() != nil)
		assert.Equal(t, "url2", results[1].Message.GetImageContent().Url)

		// 验证第三条是 TextContent
		assert.Assert(t, results[2].Message.GetTextContent() != nil)
		assert.Equal(t, "text3", results[2].Message.GetTextContent().Text)
	})

	t.Run("test nil protobuf message", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_protojson_nil_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_protojson_nil",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithSimpleProto{})
		assert.NilError(t, err)

		model := db.NewModel[TestModelWithSimpleProto](client)

		// 创建 nil message 的数据
		testData := TestModelWithSimpleProto{
			Name:    "nil_message_test",
			Message: nil,
		}

		err = model.NewScoop().Create(&testData)
		assert.NilError(t, err)

		// 查询并验证
		result, err := model.NewScoop().Equal("id", testData.Id).First()
		assert.NilError(t, err)
		assert.Assert(t, result != nil)
		// nil 消息应该保持为 nil 或空消息
	})
}
