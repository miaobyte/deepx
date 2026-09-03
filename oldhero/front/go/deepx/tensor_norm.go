package deepx

func init() {
	RegistOpType("L2Norm", "l2norm")
	RegistOpType("L1Norm", "l1norm")
}

// L2Norm 计算L2范数
// ||x||₂ = sqrt(Σx²)
func (t *Tensor) L2Norm() *Tensor {
	result := t.graph.AddTensor("", t.Dtype, t.Shape.shape, t.requiresGrad)
	op := t.graph.AddOp("l2norm", t.node)
	result.AddInput(op.name, op)
	return result.tensor
}

// L1Norm 计算L1范数
// ||x||₁ = Σ|x|
func (t *Tensor) L1Norm() *Tensor {
	result := t.graph.AddTensor("", t.Dtype, t.Shape.shape, t.requiresGrad)
	op := t.graph.AddOp("l1norm", t.node)
	result.AddInput(op.name, op)
	return result.tensor
}
