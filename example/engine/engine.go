package engine

import "github.com/xhd2015/ormx/engine"

type ExampleEngine struct {
	engine.Engine
}

func GetEngine() engine.Factory {
	return &ExampleEngine{}
}

func (e *ExampleEngine) GetEngine() engine.Engine {
	return e.Engine
}
