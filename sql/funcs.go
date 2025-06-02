package sql

type rand struct {
}

func (c rand) ToSQL() (string, []interface{}, error) {
	return "RAND()", nil, nil
}

func Rand() rand {
	return rand{}
}
