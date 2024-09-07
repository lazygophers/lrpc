package db

func Where(args ...interface{}) *Cond {
	cond := &Cond{}
	return cond.Where(args...)
}

func OrWhere(args ...interface{}) *Cond {
	cond := &Cond{}
	return cond.OrWhere(args...)
}

func Or(args ...interface{}) *Cond {
	cond := &Cond{}
	return cond.OrWhere(args...)
}

func Like(column string, value string) *Cond {
	cond := &Cond{}
	return cond.Like(column, value)
}

func LeftLike(column string, value string) *Cond {
	cond := &Cond{}
	return cond.LeftLike(column, value)
}

func RightLike(column string, value string) *Cond {
	cond := &Cond{}
	return cond.RightLike(column, value)
}

func NotLike(column string, value string) *Cond {
	cond := &Cond{}
	return cond.NotLike(column, value)
}

func Between(column string, min, max interface{}) *Cond {
	cond := &Cond{}
	return cond.Between(column, min, max)
}
