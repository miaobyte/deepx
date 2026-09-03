package deepx

type ArgType int

const (
	ArgTypeInt ArgType = iota
	ArgTypeFloat
	ArgTypeString
	ArgTypeIntVector
)

type ConstArgNode struct {
	name    string
	ntype   NodeType
	inputs  map[string]Node
	value   any
	argType ArgType
}

func NewConstArgNode(name string) *ConstArgNode {
	return &ConstArgNode{
		name:   name,
		ntype:  NodeConstArg,
		inputs: make(map[string]Node),
	}
}
func (n *ConstArgNode) Ntype() NodeType {
	return n.ntype
}
func (n *ConstArgNode) Name() string {
	return n.name
}
func (n *ConstArgNode) Input(name string) Node {
	return n.inputs[name]
}
func (n *ConstArgNode) Inputs() map[string]Node {
	return n.inputs
}
func (n *ConstArgNode) AddInput(name string, input Node) {
	n.inputs[name] = input
}
func (n *ConstArgNode) RemoveInput(name string) {
	delete(n.inputs, name)
}
func (n *ConstArgNode) Int() int {
	if n.argType != ArgTypeInt {
		panic("ConstArgNode is not an integer")
	}
	return n.value.(int)
}
func (n *ConstArgNode) Float() float64 {
	if n.argType != ArgTypeFloat {
		panic("ConstArgNode is not a float")
	}
	return n.value.(float64)
}
func (n *ConstArgNode) String() string {
	if n.argType != ArgTypeString {
		panic("ConstArgNode is not a string")
	}
	return n.value.(string)
}
func (n *ConstArgNode) SetInt(value int) {
	n.value = value
	n.argType = ArgTypeInt
}
func (n *ConstArgNode) SetInts(value []int) {
	n.value = value
	n.argType = ArgTypeIntVector
}
func (n *ConstArgNode) SetFloat(value float64) {
	n.value = value
	n.argType = ArgTypeFloat
}
func (n *ConstArgNode) SetString(value string) {
	n.value = value
	n.argType = ArgTypeString
}
