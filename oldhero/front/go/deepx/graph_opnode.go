package deepx

type OpType struct {
	name      string
	shortchar string
}

var (
	opmaps = make(map[string]OpType)
)

func RegistOpType(name string, shortchar string) {
	opmaps[name] = OpType{name, shortchar}
}

type OpNode struct {
	OpType
	ntype NodeType

	inputs map[string]Node
}

func NewOpNode(name string) *OpNode {
	return &OpNode{
		OpType: opmaps[name],
		ntype:  NodeOp,
		inputs: make(map[string]Node),
	}
}
func (n *OpNode) Ntype() NodeType {
	return n.ntype
}
func (n *OpNode) Name() string {
	return n.name
}
func (n *OpNode) Input(name string) Node {
	return n.inputs[name]
}
func (n *OpNode) Inputs() map[string]Node {
	return n.inputs
}
func (n *OpNode) AddInput(name string, input Node) {
	n.inputs[name] = input
}
func (n *OpNode) RemoveInput(name string) {
	delete(n.inputs, name)
}
func (n *OpNode) Shortchar() string {
	return n.shortchar
}
