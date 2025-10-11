// Package xerror 提供了结构化的错误处理机制，支持：
// - 错误码管理和注册
// - 国际化错误消息
// - HTTP 错误码集成
// - 错误链支持（Go 1.13+）
//
// # 基本用法
//
//	err := xerror.NewInvalidParam("user_id is required")
//	if xerror.CheckCode(err, xerror.ErrInvalidParam) {
//	    // 处理参数错误
//	}
//
// # 多语言支持
//
//	xerror.SetI18n(myI18nImpl)
//	err := xerror.New(1001, "en", "zh")
//
// # 错误链支持
//
//	err := xerror.WrapError(dbErr, xerror.ErrSystemError, "database query failed")
//	if errors.Is(err, dbErr) {
//	    // 可以追溯到底层错误
//	}
package xerror

import (
	"errors"
	"fmt"
)

var _ error = (*Error)(nil)

// 错误码常量定义
const (
	// ErrSystemError 系统级错误（< 0 均为系统错误）
	ErrSystemError = -1

	// 框架错误码范围：1～10000

	// ErrInvalidParam 表示请求参数无效或缺失
	// 使用场景：参数校验失败、必填参数缺失、参数格式错误
	ErrInvalidParam = 1001

	// ErrNoAuth 表示未授权或认证失败
	// 使用场景：token 无效、未登录、权限不足
	ErrNoAuth = 1002

	// ErrNoData 表示请求的数据不存在
	// 使用场景：查询记录不存在、资源未找到
	ErrNoData = 1003

	// ErrConflict 表示数据冲突
	// 使用场景：并发更新冲突、唯一键冲突、版本不匹配
	ErrConflict = 1004
)

// errMap 存储已注册的错误码映射
var errMap = map[int32]*Error{}

// I18n 国际化接口，用于实现多语言错误消息
// 实现此接口可以根据错误码和语言返回对应的错误消息，并支持动态注册翻译
//
// 示例实现：
//
//	type MyI18n struct {
//	    mu       sync.RWMutex
//	    messages map[int32]map[string]string
//	}
//
//	func (m *MyI18n) Localize(key int32, langs ...string) (string, bool) {
//	    m.mu.RLock()
//	    defer m.mu.RUnlock()
//	    for _, lang := range langs {
//	        if msg, ok := m.messages[key][lang]; ok {
//	            return msg, true
//	        }
//	    }
//	    return "", false
//	}
//
//	func (m *MyI18n) Register(lang string, code int32, msg string) {
//	    m.mu.Lock()
//	    defer m.mu.Unlock()
//	    if m.messages[code] == nil {
//	        m.messages[code] = make(map[string]string)
//	    }
//	    m.messages[code][lang] = msg
//	}
//
//	func (m *MyI18n) RegisterBatch(lang string, data map[int32]string) {
//	    m.mu.Lock()
//	    defer m.mu.Unlock()
//	    for code, msg := range data {
//	        if m.messages[code] == nil {
//	            m.messages[code] = make(map[string]string)
//	        }
//	        m.messages[code][lang] = msg
//	    }
//	}
type I18n interface {
	// Localize 根据错误码和语言返回对应的错误消息
	// key: 错误码
	// langs: 语言列表，按优先级排序
	// 返回: (消息文本, 是否找到)
	Localize(key int32, langs ...string) (string, bool)

	// Register 追加注册指定语言的错误码翻译
	// lang: 语言代码（如 "en", "zh"）
	// code: 错误码
	// msg: 翻译文本
	Register(lang string, code int32, msg string)

	// RegisterBatch 批量追加注册指定语言的多个错误码翻译
	// lang: 语言代码（如 "en", "zh"）
	// data: 错误码到翻译文本的映射
	RegisterBatch(lang string, data map[int32]string)
}

// i18n 全局国际化实例
var i18n I18n

// SetI18n 设置全局国际化实例
// 设置后，New 函数将自动使用此实例获取多语言错误消息
func SetI18n(i I18n) {
	i18n = i
}

// Register 注册错误码到全局错误映射表
// 已注册的错误码可以通过 New 函数快速创建错误实例
//
// 示例：
//
//	xerror.Register(
//	    &xerror.Error{Code: 10001, Msg: "User not found"},
//	    &xerror.Error{Code: 10002, Msg: "Invalid email"},
//	)
func Register(errs ...*Error) {
	for _, err := range errs {
		errMap[err.Code] = err
	}
}

// Error 结构化错误类型，包含错误码和错误消息
// 实现了 error 接口和 Go 1.13+ 的错误链接口
type Error struct {
	Code  int32  `json:"code,omitempty" yaml:"code,omitempty"` // 错误码
	Msg   string `json:"msg,omitempty" yaml:"msg,omitempty"`   // 错误消息
	Cause error  `json:"-" yaml:"-"`                           // 底层错误（支持错误链）
}

// Error 实现 error 接口，返回错误消息
func (p *Error) Error() string {
	if p.Cause != nil {
		return fmt.Sprintf("%s: %v", p.Msg, p.Cause)
	}
	return p.Msg
}

// Unwrap 实现 Go 1.13+ 错误链接口，返回底层错误
// 支持使用 errors.Is 和 errors.As 追溯错误链
func (p *Error) Unwrap() error {
	return p.Cause
}

// Clone 创建错误的副本，避免共享状态
// 注意：Cause 字段为浅拷贝
func (p *Error) Clone() *Error {
	return &Error{
		Code:  p.Code,
		Msg:   p.Msg,
		Cause: p.Cause,
	}
}

// Is 实现错误比较接口，比较错误码是否相同
// 仅比较 Code 字段，忽略 Msg 和 Cause
func (p *Error) Is(err error) bool {
	if x, ok := err.(*Error); ok {
		return x.Code == p.Code
	}
	return false
}

// CheckCode 检查错误码是否匹配指定值
func (p *Error) CheckCode(code int32) bool {
	return p.Code == code
}

// Is 判断两个错误是否相等
// 对于 *Error 类型，比较错误码；否则使用标准库 errors.Is
func Is(err1, err2 error) bool {
	if x, ok := err1.(*Error); ok {
		return x.Is(err2)
	}
	return errors.Is(err1, err2)
}

// CheckCode 检查错误的错误码是否匹配
// 如果错误不是 *Error 类型，返回 false
func CheckCode(err1 error, code int32) bool {
	return GetCode(err1) == code
}

// GetCode 获取错误的错误码
// 如果错误不是 *Error 类型，返回 -1
func GetCode(err error) int32 {
	if x, ok := err.(*Error); ok {
		return x.Code
	}
	return -1
}

// GetMsg 获取错误消息
// 支持 *Error 类型和标准 error 类型
func GetMsg(err error) string {
	if err == nil {
		return ""
	}
	if x, ok := err.(*Error); ok {
		return x.Msg
	}
	return err.Error()
}

// New 创建新的错误实例，支持可选的多语言参数
//
// 创建过程：
// 1. 如果错误码已注册，返回注册错误的克隆
// 2. 如果提供了语言参数且设置了 i18n，尝试获取国际化消息
// 3. 否则创建仅包含错误码的错误实例
//
// 参数：
//   - code: 错误码
//   - lang: 可选的语言代码（如 "en", "zh"），支持多个语言作为回退
//
// 示例：
//
//	err1 := xerror.New(1001)                    // 使用已注册的错误消息
//	err2 := xerror.New(1001, "en")              // 使用英文消息
//	err3 := xerror.New(1001, "fr", "en", "zh")  // 尝试法语，回退到英语或中文
func New(code int32, args ...interface{}) *Error {
	// 优先返回已注册的错误
	if err, ok := errMap[code]; ok {
		return err.Clone()
	}

	if len(args) > 0 {
		return NewErrorWithMsg(code, fmt.Sprint(args...))
	}

	if i18n != nil {
		if msg, ok := i18n.Localize(code); ok {
			return NewErrorWithMsg(code, msg)
		}
	}

	return NewErrorWithMsg(code, "")
}

// NewErrorWithMsg 创建带有自定义消息的错误
//
// 直接创建错误实例，不查询 errMap 和 i18n
func NewErrorWithMsg(code int32, msg string) *Error {
	return &Error{
		Code: code,
		Msg:  msg,
	}
}

// NewSystemError 创建系统错误（错误码 -1）
//
// 使用场景：系统级故障、意外错误、未处理的异常
func NewSystemError(msg string) *Error {
	return &Error{
		Code: ErrSystemError,
		Msg:  msg,
	}
}

// NewInvalidParam 创建参数无效错误（错误码 1001）
//
// 使用场景：参数校验失败、必填参数缺失、参数格式错误
//
// 示例：
//
//	err := xerror.NewInvalidParam("user_id is required")
//	err := xerror.NewInvalidParam("invalid email format: ", email)
func NewInvalidParam(a ...any) *Error {
	return &Error{
		Code: ErrInvalidParam,
		Msg:  fmt.Sprint(a...),
	}
}

// NewNoAuth 创建未授权错误（错误码 1002）
//
// 使用场景：token 无效、未登录、权限不足
//
// 示例：
//
//	err := xerror.NewNoAuth("token expired")
//	err := xerror.NewNoAuth("permission denied")
func NewNoAuth(msg string) *Error {
	return &Error{
		Code: ErrNoAuth,
		Msg:  msg,
	}
}

// NewNoData 创建数据不存在错误（错误码 1003）
//
// 使用场景：查询记录不存在、资源未找到
//
// 示例：
//
//	err := xerror.NewNoData("user not found")
//	err := xerror.NewNoData("record id:", id, " not found")
func NewNoData(a ...any) *Error {
	return &Error{
		Code: ErrNoData,
		Msg:  fmt.Sprint(a...),
	}
}

// NewConflict 创建数据冲突错误（错误码 1004）
//
// 使用场景：并发更新冲突、唯一键冲突、版本不匹配
//
// 示例：
//
//	err := xerror.NewConflict("email already exists")
//	err := xerror.NewConflict("version mismatch")
func NewConflict(msg string) *Error {
	return &Error{
		Code: ErrConflict,
		Msg:  msg,
	}
}

// Wrap 将标准 error 包装为 *Error
//
// 如果 err 已经是 *Error 类型，直接返回；
// 如果 err 为 nil，返回 nil；
// 否则创建新的 *Error 实例，将原错误作为消息
//
// 示例：
//
//	dbErr := db.QueryRow(...)
//	return xerror.Wrap(dbErr, xerror.ErrSystemError)
func Wrap(err error, code int32) *Error {
	if err == nil {
		return nil
	}
	if xerr, ok := err.(*Error); ok {
		return xerr
	}
	return &Error{
		Code: code,
		Msg:  err.Error(),
	}
}

// WrapError 将错误包装为 *Error 并保留错误链
//
// 与 Wrap 的区别：
// - WrapError 保留原始错误作为 Cause，支持 errors.Is/As
// - Wrap 将原始错误转换为字符串作为消息
//
// 示例：
//
//	dbErr := db.QueryRow(...)
//	err := xerror.WrapError(dbErr, xerror.ErrSystemError, "database query failed")
//	if errors.Is(err, dbErr) { // 可以追溯到原始错误
//	    // 处理数据库错误
//	}
func WrapError(err error, code int32, msg string) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Code:  code,
		Msg:   msg,
		Cause: err,
	}
}
