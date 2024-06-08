package db

func Expr(expression string, args ...interface{}) map[string]any {
	return map[string]any{
		"expr": expression,
		"args": args,
	}
}
