package lrpc

type disableLogger struct{}

func (*disableLogger) Printf(_ string, _ ...any) {
	// fmt.Println(fmt.Sprintf(format, args...))
}
