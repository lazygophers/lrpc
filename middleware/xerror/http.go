//go:build !httpcode

package xerror

// 内置一下 http 相关的 code

func init() {
	Register()
}
