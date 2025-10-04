package db

import (
	"context"
	"reflect"
	"sync"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/json"
	"gorm.io/gorm/schema"
)

var (
	serializerOnce sync.Once
)

func init() {
	ensureSerializerRegistered()
}

// ensureSerializerRegistered 确保 JSON 序列化器已注册（可安全多次调用）
func ensureSerializerRegistered() {
	serializerOnce.Do(func() {
		schema.RegisterSerializer("json", &JsonSerializer{})
		log.Debugf("JSON serializer registered")
	})
}

type JsonSerializer struct {
}

func (p *JsonSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	log.Debugf("JsonSerializer.Scan called for field: %s, dbValue type: %T", field.Name, dbValue)

	fieldValue := reflect.New(field.FieldType)

	if dbValue != nil {
		var bytes []byte
		switch v := dbValue.(type) {
		case []byte:
			bytes = v
		case string:
			bytes = []byte(v)
		default:
			bytes, err = json.Marshal(v)
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}
		}

		if len(bytes) > 0 {
			err = json.Unmarshal(bytes, fieldValue.Interface())
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}
		}
	}

	field.ReflectValueOf(ctx, dst).Set(fieldValue.Elem())
	return nil
}

func (p *JsonSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	log.Debugf("JsonSerializer.Value called for field: %s, fieldValue type: %T", field.Name, fieldValue)

	if fieldValue == nil {
		return "", nil
	}

	buffer, err := json.Marshal(fieldValue)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	result := string(buffer)
	log.Debugf("JsonSerializer.Value returning: %s", result)
	return result, nil
}
