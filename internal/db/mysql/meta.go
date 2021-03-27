package mysql

type selectQuery struct {
	selectSQL string
	selectArgs []interface{}
	countSQL string
	countArgs []interface{}
}

type meta struct {
	Total int `db:"total"`
}
