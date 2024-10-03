package xerror

import (
	"errors"
	"fmt"
)

var _ error = (*Error)(nil)

// NOTE: 内置一些错误码
const (
	// < 0 均为类的错误

	ErrSystemError = -1

	//	1～10000 框架错误
	ErrInvalidParam = 1001 // 入参有问题
	ErrNoAuth       = 1002 // 没有授权
	ErrNoData       = 1003 // 没有数据
	ErrConflict     = 1004 // 更新冲突
)

var errMap = map[int32]*Error{
	//ErrSystemError: {
	//	Code: ErrSystemError,
	//	Msg:  "System error",
	//},
	//ErrInvalidParam: {
	//	Code: ErrInvalidParam,
	//	Msg:  "Invalid param",
	//},
	//ErrNoAuth: {
	//	Code: ErrNoAuth,
	//	Msg:  "No auth",
	//},
	//ErrNoData: {
	//	Code: ErrNoData,
	//	Msg:  "No data",
	//},
}

type I18n interface {
	Localize(key int32, langs ...string) (string, bool)
}

// 对于想要多语言场景下，可以通过本方法按照 errode 自动返回对应语言的错误信息
var i18n I18n

func SetI18n(i I18n) {
	i18n = i
}

func Register(errs ...*Error) {
	for _, err := range errs {
		errMap[err.Code] = err
	}
}

type Error struct {
	Code int32  `json:"code,omitempty" yaml:"code,omitempty"`
	Msg  string `json:"msg,omitempty" yaml:"msg,omitempty"`
}

func (p *Error) Error() string {
	return fmt.Sprintf("code:%d,msg:%s", p.Code, p.Msg)
}

func (p *Error) Clone() *Error {
	return &Error{
		Code: p.Code,
		Msg:  p.Msg,
	}
}

func (p *Error) Is(err error) bool {
	if x, ok := err.(*Error); ok {
		return x.Code == p.Code
	}

	return errors.Is(p, err)
}

func (p *Error) CheckCode(code int32) bool {
	return p.Code == code
}

func Is(err1, err2 error) bool {
	if x, ok := err1.(*Error); ok {
		return x.Is(err2)
	}

	return errors.Is(err1, err2)
}

func CheckCode(err1 error, code int32) bool {
	if x, ok := err1.(*Error); ok {
		return x.Code == code
	}

	return false
}

func GetCode(err error) int32 {
	if x, ok := err.(*Error); ok {
		return x.Code
	}

	return -1
}

func New(code int32) *Error {
	if err, ok := errMap[code]; ok {
		return err.Clone()
	}

	return &Error{
		Code: code,
	}
}

func NewError(code int32, lang ...string) *Error {
	if err, ok := errMap[code]; ok {
		return err.Clone()
	}

	if i18n != nil {
		msg, ok := i18n.Localize(code, lang...)
		if ok {
			return NewErrorWithMsg(code, msg)
		}
	}

	return &Error{
		Code: code,
	}
}

func NewErrorWithMsg(code int32, msg string) *Error {
	return &Error{
		Code: code,
		Msg:  msg,
	}
}

func NewSystemError(msg string) *Error {
	return &Error{
		Code: ErrSystemError,
		Msg:  msg,
	}
}

func NewInvalidParam(a ...any) *Error {
	return &Error{
		Code: ErrInvalidParam,
		Msg:  fmt.Sprint(a...),
	}
}

func NewNoData(a ...any) *Error {
	return &Error{
		Code: ErrNoData,
		Msg:  fmt.Sprint(a...),
	}
}