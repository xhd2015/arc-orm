package expr

type Expr interface {
	ToSQL() (string, []interface{}, error)
}
