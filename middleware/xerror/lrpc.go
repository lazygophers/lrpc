package xerror

import (
	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/core"
)

func ToBaseResponse(err error) *core.BaseResponse {
	if err == nil {
		return &core.BaseResponse{
			Hint: log.GetTrace(),
		}
	}

	if e, ok := err.(*Error); ok {
		return &core.BaseResponse{
			Code:    e.Code,
			Message: e.Msg,
			Hint:    log.GetTrace(),
		}
	}

	return &core.BaseResponse{
		Code:    ErrSystemError,
		Message: err.Error(),
		Hint:    log.GetTrace(),
	}
}
