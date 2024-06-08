package db

func Where(args ...interface{}) *Cond {
	var cond Cond

	return cond.Where(args...)
}

func OrWhere(args ...interface{}) *Cond {
	var cond Cond

	return cond.OrWhere(args...)
}
