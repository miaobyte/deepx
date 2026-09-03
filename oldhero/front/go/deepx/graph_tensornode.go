package deepx

type TensorNode struct {
	name   string
	ntype  NodeType
	inputs map[string]Node
	tensor *Tensor // 对于 NodeTensor 类型，存储实际的张量数据
}

func NewTensorNode(name string) *TensorNode {
	return &TensorNode{
		name:   name,
		ntype:  NodeTensor,
		inputs: make(map[string]Node),
	}
}
func (n *TensorNode) Ntype() NodeType {
	return n.ntype
}
func (n *TensorNode) Name() string {
	return n.name
}
func (n *TensorNode) Input(name string) Node {
	return n.inputs[name]
}
func (n *TensorNode) Inputs() map[string]Node {
	return n.inputs
}
func (n *TensorNode) AddInput(name string, input Node) {
	n.inputs[name] = input
}
func (n *TensorNode) RemoveInput(name string) {
	delete(n.inputs, name)
}
func (n *TensorNode) Tensor() *Tensor {
	return n.tensor
}
func (n *TensorNode) SetTensor(tensor *Tensor) {
	n.tensor = tensor
}
