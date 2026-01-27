package queue

import (
	"fmt"

	"github.com/lazygophers/utils/runtime"
)

func Handle[T any](handler Handler[T], msg *Message[T]) (rsp ProcessRsp, err error) {
	defer runtime.CachePanicWithHandle(func(err interface{}) {
		err = fmt.Errorf("PANIC:%v", err)
	})

	rsp, err = handler(msg)
	return
}
