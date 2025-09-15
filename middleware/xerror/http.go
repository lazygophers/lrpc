//go:build !httpcode

package xerror

// 内置一下 http 相关的 code

func init() {
	// Register HTTP-related error codes
	Register(
		&Error{Code: 400, Msg: "Bad Request"},
		&Error{Code: 401, Msg: "Unauthorized"},
		&Error{Code: 403, Msg: "Forbidden"},
		&Error{Code: 404, Msg: "Not Found"},
		&Error{Code: 500, Msg: "Internal Server Error"},
	)
}
