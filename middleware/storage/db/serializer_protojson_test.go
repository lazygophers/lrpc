package db_test

import (
	"os"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/lazygophers/lrpc/middleware/storage/db/testdata"
	"gotest.tools/v3/assert"
)

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
