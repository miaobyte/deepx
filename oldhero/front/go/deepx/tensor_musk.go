package deepx

func init() {
	RegistOpType("mask", "mask")
}

func (t *Tensor) ApplyMask(mask *Tensor) *Tensor {
	result := t.graph.AddTensor("", t.Dtype, t.Shape.shape, t.requiresGrad)
	op := t.graph.AddOp("mask", t.node, mask.node)
	result.AddInput(op.name, op)
	return result.tensor
}
