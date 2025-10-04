package db

import (
	"context"
	"reflect"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/json"
	"gorm.io/gorm/schema"
)

func init() {
	schema.RegisterSerializer("json", &JsonSerializer{})
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
