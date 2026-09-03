package deepx

type Module interface {
	Graph() *Graph
	Name() string
}
type ModuleBase struct {
	name string
	g    *Graph
}
