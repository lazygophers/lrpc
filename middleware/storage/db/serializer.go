package db

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/json"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/proto"
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm/schema"
)

var (
	serializerOnce sync.Once
)

func init() {
	ensureSerializerRegistered()
}

// ensureSerializerRegistered 确保所有序列化器已注册（可安全多次调用）
func ensureSerializerRegistered() {
	serializerOnce.Do(func() {
		schema.RegisterSerializer("json", &JsonSerializer{})
		schema.RegisterSerializer("yaml", &YamlSerializer{})
		schema.RegisterSerializer("ini", &IniSerializer{})
		schema.RegisterSerializer("bson", &BsonSerializer{})
		schema.RegisterSerializer("toml", &TomlSerializer{})
		schema.RegisterSerializer("protobuf", &ProtobufSerializer{})
	})
}

type JsonSerializer struct {
}

func (p *JsonSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
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
	if fieldValue == nil {
		return "", nil
	}

	buffer, err := json.Marshal(fieldValue)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return string(buffer), nil
}

// YamlSerializer YAML 序列化器
type YamlSerializer struct {
}

func (p *YamlSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	fieldValue := reflect.New(field.FieldType)

	if dbValue != nil {
		var bytes []byte
		switch v := dbValue.(type) {
		case []byte:
			bytes = v
		case string:
			bytes = []byte(v)
		default:
			bytes, err = yaml.Marshal(v)
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}
		}

		if len(bytes) > 0 {
			err = yaml.Unmarshal(bytes, fieldValue.Interface())
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}
		}
	}

	field.ReflectValueOf(ctx, dst).Set(fieldValue.Elem())
	return nil
}

func (p *YamlSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	if fieldValue == nil {
		return "", nil
	}

	buffer, err := yaml.Marshal(fieldValue)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return string(buffer), nil
}

// IniSerializer INI 序列化器
type IniSerializer struct {
}

func (p *IniSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	fieldValue := reflect.New(field.FieldType)

	if dbValue != nil {
		var bytes []byte
		switch v := dbValue.(type) {
		case []byte:
			bytes = v
		case string:
			bytes = []byte(v)
		default:
			// INI format doesn't support direct marshaling of complex types
			// Fall back to JSON for storage
			bytes, err = json.Marshal(v)
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}
		}

		if len(bytes) > 0 {
			cfg, err := ini.Load(bytes)
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}

			err = cfg.MapTo(fieldValue.Interface())
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}
		}
	}

	field.ReflectValueOf(ctx, dst).Set(fieldValue.Elem())
	return nil
}

func (p *IniSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	if fieldValue == nil {
		return "", nil
	}

	cfg := ini.Empty()

	// INI ReflectFrom requires a pointer to struct
	// If fieldValue is not a pointer, create one
	val := reflect.ValueOf(fieldValue)
	if val.Kind() != reflect.Ptr {
		ptrVal := reflect.New(val.Type())
		ptrVal.Elem().Set(val)
		fieldValue = ptrVal.Interface()
	}

	err := ini.ReflectFrom(cfg, fieldValue)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	buf := &bytes.Buffer{}
	_, err = cfg.WriteTo(buf)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return buf.String(), nil
}

// BsonSerializer BSON 序列化器
type BsonSerializer struct {
}

func (p *BsonSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	fieldValue := reflect.New(field.FieldType)

	if dbValue != nil {
		var bytes []byte
		switch v := dbValue.(type) {
		case []byte:
			bytes = v
		case string:
			bytes = []byte(v)
		default:
			bytes, err = bson.Marshal(v)
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}
		}

		if len(bytes) > 0 {
			err = bson.Unmarshal(bytes, fieldValue.Interface())
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}
		}
	}

	field.ReflectValueOf(ctx, dst).Set(fieldValue.Elem())
	return nil
}

func (p *BsonSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	if fieldValue == nil {
		return "", nil
	}

	buffer, err := bson.Marshal(fieldValue)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return string(buffer), nil
}

// TomlSerializer TOML 序列化器
type TomlSerializer struct {
}

func (p *TomlSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	fieldValue := reflect.New(field.FieldType)

	if dbValue != nil {
		var data []byte
		switch v := dbValue.(type) {
		case []byte:
			data = v
		case string:
			data = []byte(v)
		default:
			buf := &bytes.Buffer{}
			err = toml.NewEncoder(buf).Encode(v)
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}
			data = buf.Bytes()
		}

		if len(data) > 0 {
			err = toml.Unmarshal(data, fieldValue.Interface())
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}
		}
	}

	field.ReflectValueOf(ctx, dst).Set(fieldValue.Elem())
	return nil
}

func (p *TomlSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	if fieldValue == nil {
		return "", nil
	}

	buf := &bytes.Buffer{}
	err := toml.NewEncoder(buf).Encode(fieldValue)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return buf.String(), nil
}

// ProtobufSerializer Protobuf 序列化器
type ProtobufSerializer struct {
}

func (p *ProtobufSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	fieldValue := reflect.New(field.FieldType)

	if dbValue != nil {
		var bytes []byte
		switch v := dbValue.(type) {
		case []byte:
			bytes = v
		case string:
			bytes = []byte(v)
		default:
			// Check if it's a proto.Message
			if msg, ok := v.(proto.Message); ok {
				bytes, err = proto.Marshal(msg)
				if err != nil {
					log.Errorf("err:%v", err)
					return err
				}
			} else {
				return errors.New("value is not a proto.Message")
			}
		}

		if len(bytes) > 0 {
			// Check if fieldValue is a proto.Message
			if msg, ok := fieldValue.Interface().(proto.Message); ok {
				err = proto.Unmarshal(bytes, msg)
				if err != nil {
					log.Errorf("err:%v", err)
					return err
				}
			} else {
				return errors.New("field is not a proto.Message")
			}
		}
	}

	field.ReflectValueOf(ctx, dst).Set(fieldValue.Elem())
	return nil
}

func (p *ProtobufSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	if fieldValue == nil {
		return "", nil
	}

	// Check if it's a proto.Message
	msg, ok := fieldValue.(proto.Message)
	if !ok {
		return nil, errors.New("value is not a proto.Message")
	}

	buffer, err := proto.Marshal(msg)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return string(buffer), nil
}
