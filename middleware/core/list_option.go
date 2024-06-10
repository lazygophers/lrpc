package core

import (
	"errors"
	"github.com/lazygophers/log"
	"strconv"
	"strings"
)

const (
	defaultOffset = 0
	defaultLimit  = 20

	maxLimit = 1000
)

func NewListOption() *ListOption {
	return &ListOption{
		Offset: defaultOffset,
		Limit:  defaultLimit,
	}
}

func (p *ListOption) SetOffset(offset uint64) *ListOption {
	p.Offset = offset
	return p
}

func (p *ListOption) SetLimit(limit uint64) *ListOption {
	p.Limit = limit
	return p
}

func (p *ListOption) SetShowTotal(showTotal ...bool) *ListOption {
	if len(showTotal) > 0 {
		p.ShowTotal = showTotal[0]
	} else {
		p.ShowTotal = true
	}
	return p
}

func (p *ListOption) SetOptions(options ...*ListOption_Option) *ListOption {
	p.Options = options
	return p
}

func (p *ListOption) AddOptions(options ...*ListOption_Option) *ListOption {
	p.Options = append(p.Options, options...)
	return p
}

func (p *ListOption) AddOption(key int32, value string) *ListOption {
	return p.AddOptions(&ListOption_Option{
		Key:   key,
		Value: value,
	})
}

func (p *ListOption) Clone() *ListOption {
	return &ListOption{
		Offset:    p.Offset,
		Limit:     p.Limit,
		Options:   p.Options,
		ShowTotal: p.ShowTotal,
	}
}

func (p *ListOption) Processor() *ListOptionProcessor {
	return NewListOptionProcessor(p.Clone())
}

func (p *ListOption) Paginate() *Paginate {
	return &Paginate{
		Offset: p.Offset,
		Limit:  p.Limit,
	}
}

type ListOptionProcessor struct {
	ListOption *ListOption

	Handler map[int32]func(string) error
}

func NewListOptionProcessor(option *ListOption) *ListOptionProcessor {
	return &ListOptionProcessor{
		ListOption: option,
		Handler:    make(map[int32]func(string) error, len(option.Options)),
	}
}

func (p *ListOptionProcessor) String(key int32, logic func(value string) error) *ListOptionProcessor {
	p.Handler[key] = logic
	return p
}

func (p *ListOptionProcessor) Int(key int32, logic func(value int) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		val, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		return logic(val)
	})
}

func (p *ListOptionProcessor) Int8(key int32, logic func(value int8) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		val, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return err
		}
		return logic(int8(val))
	})
}

func (p *ListOptionProcessor) Int16(key int32, logic func(value int16) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		val, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return err
		}
		return logic(int16(val))
	})
}

func (p *ListOptionProcessor) Int32(key int32, logic func(value int32) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		val, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return err
		}
		return logic(int32(val))
	})
}

func (p *ListOptionProcessor) Int64(key int32, logic func(value int64) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		return logic(val)
	})
}

func (p *ListOptionProcessor) Uint(key int32, logic func(value uint) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		val, err := strconv.ParseUint(value, 10, 0)
		if err != nil {
			return err
		}
		return logic(uint(val))
	})
}

func (p *ListOptionProcessor) Uint8(key int32, logic func(value uint8) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		val, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return err
		}
		return logic(uint8(val))
	})
}

func (p *ListOptionProcessor) Uint16(key int32, logic func(value uint16) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		val, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return err
		}
		return logic(uint16(val))
	})
}

func (p *ListOptionProcessor) Uint32(key int32, logic func(value uint32) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		val, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return err
		}
		return logic(uint32(val))
	})
}

func (p *ListOptionProcessor) Uint64(key int32, logic func(value uint64) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		val, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		return logic(val)
	})
}

func (p *ListOptionProcessor) Float32(key int32, logic func(value float32) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		val, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return err
		}
		return logic(float32(val))
	})
}

func (p *ListOptionProcessor) Float64(key int32, logic func(value float64) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		return logic(val)
	})
}

func (p *ListOptionProcessor) Bool(key int32, logic func(value bool) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		switch strings.ToLower(value) {
		case "true", "1", "yes", "y", "on", "enable", "enabled", "ok":
			return logic(true)
		case "false", "0", "no", "n", "off", "disable", "disabled", "cancel":
			return logic(false)
		default:
			return errors.New("invalid bool value: " + value)
		}
	})
}

func (p *ListOptionProcessor) Timestamp(key int32, logic func(value int64) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		return logic(val)
	})
}

func (p *ListOptionProcessor) TimestampRange(key int32, logic func(start, end int64) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		timestamps := strings.Split(value, ",")
		if len(timestamps) != 2 {
			return errors.New("invalid timestamp value: " + value)
		}

		start, err := strconv.ParseInt(timestamps[0], 10, 64)
		if err != nil {
			return err
		}

		end, err := strconv.ParseInt(timestamps[1], 10, 64)
		if err != nil {
			return err
		}

		return logic(start, end)
	})
}

func (p *ListOptionProcessor) BetweenTimestamp(key int32, logic func(start, end int64) error) *ListOptionProcessor {
	return p.TimestampRange(key, logic)
}

func (p *ListOptionProcessor) StringSlice(key int32, logic func(value []string) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		return logic(strings.Split(value, ","))
	})
}

func (p *ListOptionProcessor) IntSlice(key int32, logic func(value []int) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		var values []int
		for _, val := range strings.Split(value, ",") {
			val, err := strconv.Atoi(val)
			if err != nil {
				return err
			}
			values = append(values, val)
		}
		return logic(values)
	})
}

func (p *ListOptionProcessor) Int8Slice(key int32, logic func(value []int8) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		var values []int8
		for _, val := range strings.Split(value, ",") {
			val, err := strconv.ParseInt(val, 10, 8)
			if err != nil {
				return err
			}
			values = append(values, int8(val))
		}
		return logic(values)
	})
}

func (p *ListOptionProcessor) Int16Slice(key int32, logic func(value []int16) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		var values []int16
		for _, val := range strings.Split(value, ",") {
			val, err := strconv.ParseInt(val, 10, 16)
			if err != nil {
				return err
			}
			values = append(values, int16(val))
		}
		return logic(values)
	})
}

func (p *ListOptionProcessor) Int32Slice(key int32, logic func(value []int32) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		var values []int32
		for _, val := range strings.Split(value, ",") {
			val, err := strconv.ParseInt(val, 10, 32)
			if err != nil {
				return err
			}
			values = append(values, int32(val))
		}
		return logic(values)
	})
}

func (p *ListOptionProcessor) Int64Slice(key int32, logic func(value []int64) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		var values []int64
		for _, val := range strings.Split(value, ",") {
			val, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return err
			}
			values = append(values, val)
		}
		return logic(values)
	})
}

func (p *ListOptionProcessor) UintSlice(key int32, logic func(value []uint) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		var values []uint
		for _, val := range strings.Split(value, ",") {
			val, err := strconv.ParseUint(val, 10, 0)
			if err != nil {
				return err
			}
			values = append(values, uint(val))
		}
		return logic(values)
	})
}

func (p *ListOptionProcessor) Uint8Slice(key int32, logic func(value []uint8) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		var values []uint8
		for _, val := range strings.Split(value, ",") {
			val, err := strconv.ParseUint(val, 10, 8)
			if err != nil {
				return err
			}
			values = append(values, uint8(val))
		}
		return logic(values)
	})
}

func (p *ListOptionProcessor) Uint16Slice(key int32, logic func(value []uint16) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		var values []uint16
		for _, val := range strings.Split(value, ",") {
			val, err := strconv.ParseUint(val, 10, 16)
			if err != nil {
				return err
			}
			values = append(values, uint16(val))
		}
		return logic(values)
	})
}

func (p *ListOptionProcessor) Uint32Slice(key int32, logic func(value []uint32) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		var values []uint32
		for _, val := range strings.Split(value, ",") {
			val, err := strconv.ParseUint(val, 10, 32)
			if err != nil {
				return err
			}
			values = append(values, uint32(val))
		}
		return logic(values)
	})
}

func (p *ListOptionProcessor) Uint64Slice(key int32, logic func(value []uint64) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		var values []uint64
		for _, val := range strings.Split(value, ",") {
			val, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				return err
			}
			values = append(values, val)
		}
		return logic(values)
	})
}

func (p *ListOptionProcessor) Float32Slice(key int32, logic func(value []float32) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		var values []float32
		for _, val := range strings.Split(value, ",") {
			val, err := strconv.ParseFloat(val, 32)
			if err != nil {
				return err
			}
			values = append(values, float32(val))
		}
		return logic(values)
	})
}

func (p *ListOptionProcessor) Float64Slice(key int32, logic func(value []float64) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		var values []float64
		for _, val := range strings.Split(value, ",") {
			val, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return err
			}
			values = append(values, val)
		}
		return logic(values)
	})
}

func (p *ListOptionProcessor) BoolSlice(key int32, logic func(value []bool) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		var values []bool
		for _, val := range strings.Split(value, ",") {
			switch strings.ToLower(val) {
			case "true", "1", "yes", "y", "on", "enable", "enabled", "ok":
				values = append(values, true)
			case "false", "0", "no", "n", "off", "disable", "disabled", "cancel":
				values = append(values, false)
			default:
				return errors.New("invalid bool value: " + val)
			}
		}
		return logic(values)
	})
}

func (p *ListOptionProcessor) Has(key int32, logic func() error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		return logic()
	})
}

func (p *ListOptionProcessor) Order(key int32, logic func(value string) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		switch strings.ToLower(value) {
		case "desc", "descend", "descending":
			return logic(" desc")
		case "asc", "ascend", "ascending", "":
			fallthrough
		default:
			return logic(" asc")
		}
	})
}

func (p *ListOptionProcessor) Process() error {
	if p.ListOption.Offset < 0 {
		p.ListOption.Offset = defaultOffset
	}

	if p.ListOption.Limit < 0 {
		p.ListOption.Limit = defaultLimit
	} else if p.ListOption.Limit > maxLimit {
		p.ListOption.Limit = maxLimit
	}

	for _, option := range p.ListOption.Options {
		if handler, ok := p.Handler[option.Key]; ok {
			err := handler(option.Value)
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}
		}
	}

	return nil
}
