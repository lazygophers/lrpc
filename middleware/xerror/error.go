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
	ErrInvalidParam = 1001
	ErrNoAuth       = 1002
	ErrNoData       = 1003
)

var errMap = map[int32]*Error{
	ErrSystemError: {
		Code: ErrSystemError,
		Msg:  "System error",
	},
	ErrInvalidParam: {
		Code: ErrInvalidParam,
		Msg:  "Invalid param",
	},
	ErrNoAuth: {
		Code: ErrNoAuth,
		Msg:  "No auth",
	},
	ErrNoData: {
		Code: ErrNoData,
		Msg:  "No data",
	},
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

func NewError(code int32) *Error {
	if err, ok := errMap[code]; ok {
		return err.Clone()
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
