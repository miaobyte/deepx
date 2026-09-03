package deepx

func init() {
	RegistOpType("reshape", "reshape")
	RegistOpType("transpose", "T")
}

func (t *Tensor) Reshape(shape []int) *Tensor {
	result := t.graph.AddTensor("", t.Dtype, shape, t.requiresGrad)
	op := t.graph.AddOp("reshape", t.node)
	result.AddInput(op.name, op)
	return result.tensor
}

func (t *Tensor) Transpose(axes []int) *Tensor {
	result := t.graph.AddTensor("", t.Dtype, t.Shape.shape, t.requiresGrad)
	op := t.graph.AddOp("transpose", t.node)
	result.AddInput(op.name, op)
	return result.tensor
}
