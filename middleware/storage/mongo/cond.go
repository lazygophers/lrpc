package mongo

import (
	"fmt"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Cond is a MongoDB condition builder for constructing complex queries
// Similar to db/cond.go but generates BSON instead of SQL
type Cond struct {
	conds []bson.M
	isOr  bool

	// skip marks conditions that should not execute queries
	skip bool
}

// NewCond creates a new Cond for building conditions
func NewCond() *Cond {
	return &Cond{}
}

// addCond adds a simple field condition (field: {$op: value})
func (p *Cond) addCond(fieldName, op string, val interface{}) {
	if fieldName == "" {
		panic("fieldName empty")
	}
	if op == "" {
		panic(fmt.Sprintf("empty op for field %s", fieldName))
	}

	// Special handling for id/_id equality queries
	if (op == "=" || op == "$eq") && (fieldName == "id" || fieldName == "_id") {
		orCond, skip := p.buildIdOrCondition(fieldName, val)
		if skip {
			// Skip empty/zero values
			return
		}
		if orCond != nil {
			p.conds = append(p.conds, orCond)
			return
		}
	}

	// Simple equality condition: field: value
	if op == "=" || op == "$eq" {
		p.conds = append(p.conds, bson.M{fieldName: val})
		return
	}

	// Other operators: field: {$op: value}
	mongoOp := op
	if !strings.HasPrefix(op, "$") {
		mongoOp = "$" + op
	}
	p.conds = append(p.conds, bson.M{fieldName: bson.M{mongoOp: val}})
}

// buildIdOrCondition builds an $or condition for id/_id equivalence
// When querying id, also query _id; when querying _id, also query id
// Returns (condition, skip). If skip is true, the condition should be skipped entirely
func (p *Cond) buildIdOrCondition(fieldName string, val interface{}) (bson.M, bool) {
	var orConditions []bson.M

	if fieldName == "id" {
		// Querying id (string type)
		// Add id condition if value is string type
		if strVal, ok := val.(string); ok {
			if strVal == "" {
				// Skip empty string
				return nil, true
			}
			orConditions = append(orConditions, bson.M{"id": strVal})

			// Try to convert to ObjectID and add _id condition
			if objectID, err := primitive.ObjectIDFromHex(strVal); err == nil {
				orConditions = append(orConditions, bson.M{"_id": objectID})
			}
		}
	} else if fieldName == "_id" {
		// Querying _id (ObjectID type)
		// Add _id condition if value is ObjectID type
		if objectID, ok := val.(primitive.ObjectID); ok {
			if objectID.IsZero() {
				// Skip zero ObjectID
				return nil, true
			}
			orConditions = append(orConditions, bson.M{"_id": objectID})

			// Convert to Hex string and add id condition
			orConditions = append(orConditions, bson.M{"id": objectID.Hex()})
		} else if strVal, ok := val.(string); ok {
			if strVal == "" {
				// Skip empty string
				return nil, true
			}
			// If value is string, try to convert to ObjectID
			if objectID, err := primitive.ObjectIDFromHex(strVal); err == nil {
				orConditions = append(orConditions, bson.M{"_id": objectID})
				orConditions = append(orConditions, bson.M{"id": strVal})
			}
		}
	}

	if len(orConditions) == 0 {
		return nil, false
	}

	if len(orConditions) == 1 {
		return orConditions[0], false
	}

	return bson.M{"$or": orConditions}, false
}

// addSubWhere adds a sub-condition (nested OR/AND group)
func (p *Cond) addSubWhere(isOr bool, args ...interface{}) {
	subCond := &Cond{
		isOr: isOr,
	}
	subCond.where(args...)
	if len(subCond.conds) == 0 {
		return
	}

	subBson := subCond.ToBson()
	if subBson == nil {
		return
	}

	p.conds = append(p.conds, subBson)
}

// where supports multiple calling forms:
//   - map[string]interface{} with field names as keys
//   - []interface{} with mixed elements
//   - field name, value pairs (assumed equality)
//   - *Cond for nested conditions
func (p *Cond) where(args ...interface{}) {
	if len(args) == 0 {
		return
	}

	// Handle *Cond as argument
	if cond, ok := args[0].(*Cond); ok {
		bsonCond := cond.ToBson()
		if bsonCond != nil {
			p.conds = append(p.conds, bsonCond)
		}
		p.where(args[1:]...)
		return
	}

	arg0 := reflect.ValueOf(args[0])
	for arg0.Kind() == reflect.Interface || arg0.Kind() == reflect.Ptr {
		arg0 = arg0.Elem()
	}

	switch arg0.Kind() {
	case reflect.Bool:
		v := arg0.Bool()
		if !v {
			p.skip = true
		}

	case reflect.String:
		fieldName := arg0.String()
		var op string
		var val interface{}

		if len(args) == 2 {
			// fieldName, value -> equality
			fieldName, op = getOp(fieldName)
			val = args[1]
			p.addCond(fieldName, op, val)
		} else if len(args) == 3 {
			// fieldName, op, value
			op = reflect.ValueOf(args[1]).String()
			val = args[2]
			p.addCond(fieldName, op, val)
		} else if len(args) == 1 {
			// Just a field name string - skip
			// This might be a raw MongoDB condition string
		} else {
			panic(fmt.Sprintf("invalid number of where args %d by `string` prefix", len(args)))
		}

	case reflect.Map:
		typ := arg0.Type()
		if typ.Key().Kind() != reflect.String {
			panic(fmt.Sprintf("map key type required string, but got %v", typ.Key()))
		}
		for _, k := range arg0.MapKeys() {
			fieldName := k.String()
			val := arg0.MapIndex(k)
			if !val.IsValid() || !val.CanInterface() {
				panic(fmt.Sprintf("invalid map val for field %s", fieldName))
			}
			var op string
			fieldName, op = getOp(fieldName)
			p.addCond(fieldName, op, val.Interface())
		}
		if len(args) > 1 {
			p.where(args[1:]...)
		}

	case reflect.Slice, reflect.Array:
		n := arg0.Len()
		if n == 0 {
			break
		}
		// Check if first element is string (then it's a condition)
		{
			v := arg0.Index(0)
			if v.Kind() == reflect.String {
				list := make([]interface{}, 0, n)
				for i := 0; i < n; i++ {
					vv := arg0.Index(i)
					if !vv.CanInterface() {
						panic("slice element can't convert to interface")
					}
					list = append(list, vv.Interface())
				}
				p.where(list...)
				if len(args) > 1 {
					p.where(args[1:]...)
				}
				break
			}
		}
		// Process each element in the slice
		for i := 0; i < n; i++ {
			v := arg0.Index(i)
			for v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			vk := v.Kind()
			if vk == reflect.Map {
				p.addSubWhere(false, v.Interface())
			} else {
				var list []interface{}
				if vk == reflect.Slice || vk == reflect.Array {
					vLen := v.Len()
					list = make([]interface{}, 0, vLen)
					for ii := 0; ii < vLen; ii++ {
						vv := v.Index(ii)
						if !vv.CanInterface() {
							panic("slice element can't convert to interface")
						}
						list = append(list, vv.Interface())
					}
				} else {
					if !v.CanInterface() {
						panic("slice element can't convert to interface")
					}
					list = make([]interface{}, 1)
					list[0] = v.Interface()
				}
				p.where(list...)
			}
		}

	default:
		panic(fmt.Sprintf("unhandled type: %v", arg0.Type()))
	}
}

// ToBson converts the condition to a BSON query
// Returns nil if no conditions were added
func (p *Cond) ToBson() bson.M {
	if len(p.conds) == 0 {
		return nil
	}

	if len(p.conds) == 1 {
		return p.conds[0]
	}

	if p.isOr {
		return bson.M{"$or": p.conds}
	}
	// AND: merge all conditions
	result := bson.M{}
	for _, cond := range p.conds {
		for k, v := range cond {
			result[k] = v
		}
	}
	return result
}

// Where adds AND conditions
func (p *Cond) Where(args ...interface{}) *Cond {
	p.addSubWhere(false, args...)
	return p
}

// OrWhere adds OR conditions
func (p *Cond) OrWhere(args ...interface{}) *Cond {
	p.addSubWhere(true, args...)
	return p
}

// Or is an alias for OrWhere
func (p *Cond) Or(args ...interface{}) *Cond {
	p.addSubWhere(true, args...)
	return p
}

// Equal adds an equality condition
func (p *Cond) Equal(column string, value interface{}) *Cond {
	p.where(column, "=", value)
	return p
}

// Ne adds a != condition using $ne operator
func (p *Cond) Ne(column string, value interface{}) *Cond {
	p.where(column, "$ne", value)
	return p
}

// Gt adds a > condition using $gt operator
func (p *Cond) Gt(column string, value interface{}) *Cond {
	p.where(column, "$gt", value)
	return p
}

// Lt adds a < condition using $lt operator
func (p *Cond) Lt(column string, value interface{}) *Cond {
	p.where(column, "$lt", value)
	return p
}

// Gte adds a >= condition using $gte operator
func (p *Cond) Gte(column string, value interface{}) *Cond {
	p.where(column, "$gte", value)
	return p
}

// Lte adds a <= condition using $lte operator
func (p *Cond) Lte(column string, value interface{}) *Cond {
	p.where(column, "$lte", value)
	return p
}

// In adds an $in condition
func (p *Cond) In(column string, values ...interface{}) *Cond {
	p.where(column, "$in", values)
	return p
}

// NotIn adds a $nin condition
func (p *Cond) NotIn(column string, values ...interface{}) *Cond {
	p.where(column, "$nin", values)
	return p
}

// Like adds a $regex condition (case-insensitive)
func (p *Cond) Like(column string, pattern string) *Cond {
	if pattern == "" {
		return p
	}
	p.conds = append(p.conds, bson.M{
		column: bson.M{
			"$regex":   pattern,
			"$options": "i",
		},
	})
	return p
}

// LeftLike adds a $regex condition with pattern%
func (p *Cond) LeftLike(column string, pattern string) *Cond {
	if pattern == "" {
		return p
	}
	p.conds = append(p.conds, bson.M{
		column: bson.M{
			"$regex":   pattern + ".*",
			"$options": "i",
		},
	})
	return p
}

// RightLike adds a $regex condition with %pattern
func (p *Cond) RightLike(column string, pattern string) *Cond {
	if pattern == "" {
		return p
	}
	p.conds = append(p.conds, bson.M{
		column: bson.M{
			"$regex":   ".*" + pattern,
			"$options": "i",
		},
	})
	return p
}

// NotLike adds a $not $regex condition
func (p *Cond) NotLike(column string, pattern string) *Cond {
	if pattern == "" {
		return p
	}
	p.conds = append(p.conds, bson.M{
		column: bson.M{
			"$not": bson.M{
				"$regex":   pattern,
				"$options": "i",
			},
		},
	})
	return p
}

// NotLeftLike adds a $not $regex condition with pattern%
func (p *Cond) NotLeftLike(column string, pattern string) *Cond {
	if pattern == "" {
		return p
	}
	p.conds = append(p.conds, bson.M{
		column: bson.M{
			"$not": bson.M{
				"$regex":   pattern + ".*",
				"$options": "i",
			},
		},
	})
	return p
}

// NotRightLike adds a $not $regex condition with %pattern
func (p *Cond) NotRightLike(column string, pattern string) *Cond {
	if pattern == "" {
		return p
	}
	p.conds = append(p.conds, bson.M{
		column: bson.M{
			"$not": bson.M{
				"$regex":   ".*" + pattern,
				"$options": "i",
			},
		},
	})
	return p
}

// Between adds a condition for values between min and max (inclusive)
func (p *Cond) Between(column string, min, max interface{}) *Cond {
	p.conds = append(p.conds, bson.M{
		column: bson.M{
			"$gte": min,
			"$lte": max,
		},
	})
	return p
}

// NotBetween adds a condition for values not between min and max
func (p *Cond) NotBetween(column string, min, max interface{}) *Cond {
	p.conds = append(p.conds, bson.M{
		column: bson.M{
			"$not": bson.M{
				"$gte": min,
				"$lte": max,
			},
		},
	})
	return p
}

// Reset clears all conditions and resets the Cond to its initial state
func (p *Cond) Reset() *Cond {
	p.conds = p.conds[:0]
	p.isOr = false
	p.skip = false
	return p
}

// String returns a string representation of the condition for debugging
// Converts the condition to BSON and returns its string representation
func (p *Cond) String() string {
	bsonCond := p.ToBson()
	if bsonCond == nil {
		return "{}"
	}
	return fmt.Sprintf("%v", bsonCond)
}

// getOp extracts operator from field name
// Examples: "age >" -> ("age", ">"), "name LIKE" -> ("name", "LIKE")
func getOp(fieldName string) (newFieldName, op string) {
	op = "="
	newFieldName = fieldName

	// Find the first invalid character (space, operator char)
	idx := getFirstInvalidFieldNameCharIndex(fieldName)
	if idx > 0 {
		o := strings.TrimSpace(fieldName[idx:])
		newFieldName = fieldName[:idx]
		if o != "" {
			op = o
		}
	}
	return
}

// getFirstInvalidFieldNameCharIndex finds the first non-field-name character
func getFirstInvalidFieldNameCharIndex(s string) int {
	for i := 0; i < len(s); i++ {
		c := s[i]
		// Valid characters: alphanumeric, underscore, dot
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
			c == '_' || c == '.' {
			continue
		}
		return i
	}
	return -1
}
