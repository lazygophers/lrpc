// Package core 提供列表查询选项的处理框架
//
// # 主要功能
//
//   - 分页参数管理（Offset/Limit）
//   - 动态查询选项处理
//   - 类型安全的参数解析
//   - 链式调用 API
//
// # 基本用法
//
//	option := core.NewListOption().SetLimit(50).SetOffset(100)
//	processor := option.Processor()
//	processor.Int32("status", func(status int32) error {
//	    // 处理状态过滤
//	    return nil
//	}).String("keyword", func(keyword string) error {
//	    // 处理关键字搜索
//	    return nil
//	}).Process()
//
// # Slice 处理
//
//	processor.Int32Slice("ids", func(ids []int32) error {
//	    // 处理多个ID：支持 "1,2,3" 或 " 1 , 2 , 3 "
//	    return nil
//	})
package core

import (
	"errors"
	"strconv"
	"strings"

	"github.com/lazygophers/log"
)

const (
	defaultOffset = 0  // 默认偏移量
	defaultLimit  = 20 // 默认每页条数

	maxLimit = 1000 // 最大每页条数限制
)

// splitAndTrim 分割逗号分隔的字符串并清理每个部分的空白字符
// 空字符串会被过滤掉
//
// 示例：
//
//	splitAndTrim("1, 2, , 3") // 返回 ["1", "2", "3"]
//	splitAndTrim("  ") // 返回 []
func splitAndTrim(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return []string{}
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// NewListOption 创建一个新的 ListOption 实例，使用默认的分页参数
func NewListOption() *ListOption {
	return &ListOption{
		Offset: defaultOffset,
		Limit:  defaultLimit,
	}
}

// SetOffset 设置查询偏移量（用于分页）
// 返回自身以支持链式调用
func (p *ListOption) SetOffset(offset uint64) *ListOption {
	p.Offset = offset
	return p
}

// SetLimit 设置每页条数限制
// 返回自身以支持链式调用
func (p *ListOption) SetLimit(limit uint64) *ListOption {
	p.Limit = limit
	return p
}

// SetShowTotal 设置是否显示总数
// 如果不提供参数，默认为 true
// 返回自身以支持链式调用
func (p *ListOption) SetShowTotal(showTotal ...bool) *ListOption {
	if len(showTotal) > 0 {
		p.ShowTotal = showTotal[0]
	} else {
		p.ShowTotal = true
	}
	return p
}

// SetOptions 设置查询选项（覆盖现有选项）
// 返回自身以支持链式调用
func (p *ListOption) SetOptions(options ...*ListOption_Option) *ListOption {
	p.Options = options
	return p
}

// AddOptions 添加查询选项（追加到现有选项）
// 返回自身以支持链式调用
func (p *ListOption) AddOptions(options ...*ListOption_Option) *ListOption {
	p.Options = append(p.Options, options...)
	return p
}

// AddOption 添加单个查询选项的便捷方法
// 返回自身以支持链式调用
func (p *ListOption) AddOption(key int32, value string) *ListOption {
	return p.AddOptions(&ListOption_Option{
		Key:   key,
		Value: value,
	})
}

// Clone 创建 ListOption 的深拷贝
// 如果 p 为 nil，返回使用默认值的新实例
func (p *ListOption) Clone() *ListOption {
	if p == nil {
		return &ListOption{
			Offset: defaultOffset,
			Limit:  defaultLimit,
		}
	}

	// Deep copy Options slice to avoid shared state
	opts := make([]*ListOption_Option, len(p.Options))
	for i, opt := range p.Options {
		if opt != nil {
			opts[i] = &ListOption_Option{
				Key:   opt.Key,
				Value: opt.Value,
			}
		}
	}

	return &ListOption{
		Offset:    p.Offset,
		Limit:     p.Limit,
		Options:   opts,
		ShowTotal: p.ShowTotal,
	}
}

// Processor 创建一个 ListOptionProcessor 来处理查询选项
// 内部会克隆当前 ListOption 以避免副作用
func (p *ListOption) Processor() *ListOptionProcessor {
	return NewListOptionProcessor(p.Clone())
}

// Paginate 转换为 Paginate 对象（用于分页结果）
// 如果 p 为 nil，返回使用默认值的新实例
func (p *ListOption) Paginate() *Paginate {
	if p == nil {
		return &Paginate{
			Offset: defaultOffset,
			Limit:  defaultLimit,
		}
	}

	return &Paginate{
		Offset: p.Offset,
		Limit:  p.Limit,
	}
}

// ListOptionProcessor 提供类型安全的查询选项处理器
// 使用链式调用注册处理函数，然后调用 Process() 执行
//
// 示例：
//
//	processor := option.Processor()
//	err := processor.
//	    Int32("status", func(status int32) error {
//	        query = query.Where("status = ?", status)
//	        return nil
//	    }).
//	    StringSlice("tags", func(tags []string) error {
//	        query = query.Where("tag IN ?", tags)
//	        return nil
//	    }).
//	    Process()
type ListOptionProcessor struct {
	ListOption *ListOption

	Handler map[int32]func(string) error
}

// NewListOptionProcessor 创建新的选项处理器
func NewListOptionProcessor(option *ListOption) *ListOptionProcessor {
	return &ListOptionProcessor{
		ListOption: option,
		Handler:    make(map[int32]func(string) error, len(option.Options)),
	}
}

// String 注册字符串类型的选项处理器
// 返回自身以支持链式调用
func (p *ListOptionProcessor) String(key int32, logic func(value string) error) *ListOptionProcessor {
	p.Handler[key] = logic
	return p
}

// Int 注册 int 类型的选项处理器，自动解析字符串为 int
func (p *ListOptionProcessor) Int(key int32, logic func(value int) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		val, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		return logic(val)
	})
}

// Int8 注册 int8 类型的选项处理器，自动解析字符串为 int8
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

// Bool 注册布尔类型的选项处理器
// 支持多种表示：true/false, 1/0, yes/no, y/n, on/off, enable/disable, enabled/disabled, ok/cancel
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

// Timestamp 注册时间戳类型（int64）的选项处理器
func (p *ListOptionProcessor) Timestamp(key int32, logic func(value int64) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		return logic(val)
	})
}

// TimestampRange 注册时间范围处理器，解析逗号分隔的两个时间戳
// 格式：「start,end」
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

// BetweenTimestamp 是 TimestampRange 的别名，语义更清晰
func (p *ListOptionProcessor) BetweenTimestamp(key int32, logic func(start, end int64) error) *ListOptionProcessor {
	return p.TimestampRange(key, logic)
}

// StringSlice 注册字符串切片处理器，解析逗号分隔的字符串
// 自动清理空白字符，过滤空字符串
// 示例："a, b, , c" -> ["a", "b", "c"]
func (p *ListOptionProcessor) StringSlice(key int32, logic func(value []string) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		return logic(splitAndTrim(value))
	})
}

// IntSlice 注册 int 切片处理器，解析逗号分隔的整数
// 自动清理空白字符，过滤空值
// 示例："1, 2, , 3" -> [1, 2, 3]
func (p *ListOptionProcessor) IntSlice(key int32, logic func(value []int) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		parts := splitAndTrim(value)
		values := make([]int, 0, len(parts))
		for _, part := range parts {
			val, err := strconv.Atoi(part)
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
		parts := splitAndTrim(value)
		values := make([]int8, 0, len(parts))
		for _, part := range parts {
			val, err := strconv.ParseInt(part, 10, 8)
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
		parts := splitAndTrim(value)
		values := make([]int16, 0, len(parts))
		for _, part := range parts {
			val, err := strconv.ParseInt(part, 10, 16)
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
		parts := splitAndTrim(value)
		values := make([]int32, 0, len(parts))
		for _, part := range parts {
			val, err := strconv.ParseInt(part, 10, 32)
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
		parts := splitAndTrim(value)
		values := make([]int64, 0, len(parts))
		for _, part := range parts {
			val, err := strconv.ParseInt(part, 10, 64)
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
		parts := splitAndTrim(value)
		values := make([]uint, 0, len(parts))
		for _, part := range parts {
			val, err := strconv.ParseUint(part, 10, 0)
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
		parts := splitAndTrim(value)
		values := make([]uint8, 0, len(parts))
		for _, part := range parts {
			val, err := strconv.ParseUint(part, 10, 8)
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
		parts := splitAndTrim(value)
		values := make([]uint16, 0, len(parts))
		for _, part := range parts {
			val, err := strconv.ParseUint(part, 10, 16)
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
		parts := splitAndTrim(value)
		values := make([]uint32, 0, len(parts))
		for _, part := range parts {
			val, err := strconv.ParseUint(part, 10, 32)
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
		parts := splitAndTrim(value)
		values := make([]uint64, 0, len(parts))
		for _, part := range parts {
			val, err := strconv.ParseUint(part, 10, 64)
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
		parts := splitAndTrim(value)
		values := make([]float32, 0, len(parts))
		for _, part := range parts {
			val, err := strconv.ParseFloat(part, 32)
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
		parts := splitAndTrim(value)
		values := make([]float64, 0, len(parts))
		for _, part := range parts {
			val, err := strconv.ParseFloat(part, 64)
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
		parts := splitAndTrim(value)
		values := make([]bool, 0, len(parts))
		for _, part := range parts {
			switch strings.ToLower(part) {
			case "true", "1", "yes", "y", "on", "enable", "enabled", "ok":
				values = append(values, true)
			case "false", "0", "no", "n", "off", "disable", "disabled", "cancel":
				values = append(values, false)
			default:
				return errors.New("invalid bool value: " + part)
			}
		}
		return logic(values)
	})
}

// Has 注册存在性检查处理器
// 只要选项存在就会调用 logic，不关心选项的值
// 适用于 flag 类型的选项
func (p *ListOptionProcessor) Has(key int32, logic func() error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		return logic()
	})
}

// Order 注册排序方向处理器
// 解析排序方向：desc/descend/descending -> true (降序)
// asc/ascend/ascending/"" -> false (升序，默认)
// 返回布尔值而非字符串，更易于使用
func (p *ListOptionProcessor) Order(key int32, logic func(isDesc bool) error) *ListOptionProcessor {
	return p.String(key, func(value string) error {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "desc", "descend", "descending":
			return logic(true)
		case "asc", "ascend", "ascending", "":
			return logic(false)
		default:
			return logic(false) // default to ascending
		}
	})
}

// Process 执行所有已注册的选项处理器
// 1. 自动校正并限制 Offset 和 Limit 的值
// 2. 按注册顺序执行各选项的处理逻辑
// 3. 遇到错误立即返回
//
// 限制规则：
//   - Offset < 0 -> 设为默认值 0
//   - Limit < 0 -> 设为默认值 20
//   - Limit > 1000 -> 限制为 1000
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
