package deepx

func init() {
	RegistOpType("add", "+")
	RegistOpType("sub", "-")
	RegistOpType("mul", "✖")
	RegistOpType("div", "÷")
	RegistOpType("scale", "×1/√d")
}

func (t *Tensor) Add(other *Tensor) *Tensor {
	// 创建新Tensor（继承graph）
	result := t.graph.AddTensor("", t.Dtype, t.Shape.shape, t.requiresGrad)
	// 添加操作节点
	op := t.graph.AddOp("add", t.node, other.node)
	result.AddInput(op.name, op)
	return result.tensor
}
func (t *Tensor) AddScalar(other float32) *Tensor {
	result := t.graph.AddTensor("", t.Dtype, t.Shape.shape, t.requiresGrad)
	// 添加操作节点
	op := t.graph.AddOp("add_scalar", t.node)
	scalar_node := t.graph.AddConstArg("")
	scalar_node.SetFloat(float64(other))
	op.AddInput(scalar_node.Name(), scalar_node)
	result.AddInput(op.name, op)
	return result.tensor
}
func (t *Tensor) Sub(other *Tensor) *Tensor {
	result := t.graph.AddTensor("", t.Dtype, t.Shape.shape, t.requiresGrad)
	// 添加操作节点
	op := t.graph.AddOp("sub", t.node, other.node)
	result.AddInput(op.name, op)
	return result.tensor
}
func (t *Tensor) Mul(other *Tensor) *Tensor {
	result := t.graph.AddTensor("", t.Dtype, t.Shape.shape, t.requiresGrad)
	// 添加操作节点
	op := t.graph.AddOp("mul", t.node, other.node)
	result.AddInput(op.name, op)
	return result.tensor
}
func (t *Tensor) MulScalar(other float32) *Tensor {
	result := t.graph.AddTensor("", t.Dtype, t.Shape.shape, t.requiresGrad)
	// 添加操作节点
	op := t.graph.AddOp("mul_scalar", t.node)
	scalar_node := t.graph.AddConstArg("")
	scalar_node.SetFloat(float64(other))
	op.AddInput(scalar_node.Name(), scalar_node)
	result.AddInput(op.name, op)
	return result.tensor
}
func (t *Tensor) Div(other *Tensor) *Tensor {
	result := t.graph.AddTensor("", t.Dtype, t.Shape.shape, t.requiresGrad)
	// 添加操作节点
	op := t.graph.AddOp("div", t.node, other.node)
	result.AddInput(op.name, op)
	return result.tensor
}
func (t *Tensor) DivScalar(other float32) *Tensor {
	result := t.graph.AddTensor("", t.Dtype, t.Shape.shape, t.requiresGrad)
	// 添加操作节点
	op := t.graph.AddOp("div_scalar", t.node)
	scalar_node := t.graph.AddConstArg("")
	scalar_node.SetFloat(float64(other))
	op.AddInput(scalar_node.Name(), scalar_node)
	result.AddInput(op.name, op)
	return result.tensor
}
func (t *Tensor) Exp() *Tensor {
	result := t.graph.AddTensor("", t.Dtype, t.Shape.shape, t.requiresGrad)
	// 添加操作节点
	op := t.graph.AddOp("exp", t.node)
	result.AddInput(op.name, op)
	return result.tensor
}
func (t *Tensor) Log() *Tensor {
	result := t.graph.AddTensor("", t.Dtype, t.Shape.shape, t.requiresGrad)
	// 添加操作节点
	op := t.graph.AddOp("log", t.node)
	result.AddInput(op.name, op)
	return result.tensor
}
func (t *Tensor) Pow(other float32) *Tensor {
	result := t.graph.AddTensor("", t.Dtype, t.Shape.shape, t.requiresGrad)
	// 添加操作节点
	op := t.graph.AddOp("pow", t.node)
	scalar_node := t.graph.AddConstArg("")
	scalar_node.SetFloat(float64(other))
	op.AddInput(scalar_node.Name(), scalar_node)
	result.AddInput(op.name, op)
	return result.tensor
}
func (t *Tensor) Sqrt() *Tensor {
	result := t.graph.AddTensor("", t.Dtype, t.Shape.shape, t.requiresGrad)
	// 添加操作节点
	op := t.graph.AddOp("sqrt", t.node)
	result.AddInput(op.name, op)
	return result.tensor
}
