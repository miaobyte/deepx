package deepx

func init() {
	RegistOpType("relu", "ReLU")
}

func (t *Tensor) Relu() *Tensor {
	result := t.graph.AddTensor("", t.Dtype, t.Shape.shape, t.requiresGrad)
	op := t.graph.AddOp("relu", t.node)
	result.AddInput(op.name, op)
	return result.tensor
}
