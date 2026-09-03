package deepx

func (t *Tensor) Sum(dims []int) *Tensor {
	result := t.graph.AddTensor("", t.Dtype, t.Shape.shape, t.requiresGrad)

	axis_node := t.graph.AddConstArg("")
	axis_node.SetInts(dims)

	op := t.graph.AddOp("sum", t.node, axis_node)
	result.AddInput(op.name, op)

	return result.tensor
}

func (t *Tensor) Mean(dims []int) *Tensor {
	if len(dims) == 0 {
		return t.DivScalar(float32(t.Shape.size))
	}
	// 1. 先计算 sum
	sum := t.Sum(dims)

	// 2. 计算需要除的维度大小
	divisor := 1
	for _, dim := range dims {
		divisor *= t.Shape.shape[dim]
	}

	// 3. 用 sum 除以维度大小
	result := sum.DivScalar(float32(divisor))

	return result
}

func (t *Tensor) Prod(dims []int) *Tensor {
	result := t.graph.AddTensor("", t.Dtype, t.Shape.shape, t.requiresGrad)

	axis_node := t.graph.AddConstArg("")
	axis_node.SetInts(dims)

	op := t.graph.AddOp("prod", t.node, axis_node)
	result.AddInput(op.name, op)

	return result.tensor
}

func (t *Tensor) Max(dims []int) *Tensor {
	result := t.graph.AddTensor("", t.Dtype, t.Shape.shape, t.requiresGrad)

	axis_node := t.graph.AddConstArg("")
	axis_node.SetInts(dims)

	op := t.graph.AddOp("max", t.node, axis_node)
	result.AddInput(op.name, op)

	return result.tensor
}

func (t *Tensor) Min(dims []int) *Tensor {
	result := t.graph.AddTensor("", t.Dtype, t.Shape.shape, t.requiresGrad)

	axis_node := t.graph.AddConstArg("")
	axis_node.SetInts(dims)

	op := t.graph.AddOp("min", t.node, axis_node)
	result.AddInput(op.name, op)

	return result.tensor
}

// Var 计算方差
func (t *Tensor) Var(dims []int) *Tensor {
	// 1. 计算均值 mean(x)
	mean := t.Mean(dims)

	// 2. 计算差值的平方 (x - mean(x))²
	diff := t.Sub(mean)
	square := diff.Mul(diff) // 或使用 diff.Square()

	// 3. 计算平方后的均值 mean((x - mean(x))²)
	variance := square.Mean(dims)

	return variance
}
