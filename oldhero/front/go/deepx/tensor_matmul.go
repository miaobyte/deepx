package deepx

func init() {
	RegistOpType("matmul", "⊗")
}

func (t *Tensor) Matmul(other *Tensor) *Tensor {
	result := t.graph.AddTensor("", t.Dtype, t.Shape.shape, t.requiresGrad)
	op := t.graph.AddOp("matmul", t.node, other.node)
	result.AddInput(op.name, op)
	return result.tensor
}
