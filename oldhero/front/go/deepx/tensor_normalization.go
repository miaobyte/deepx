package deepx

func init() {
	RegistOpType("softmax", "softmax")
	RegistOpType("layernorm", "LN")
	RegistOpType("batchnorm", "BN")
	RegistOpType("instancenorm", "IN")
	RegistOpType("groupnorm", "GN")
	RegistOpType("rmsnorm", "RMS")
}
func (t *Tensor) Softmax(axis int) *Tensor {
	// 1. 计算最大值
	x_max := t.Max([]int{axis})
	// 2. 减去最大值
	shifted := t.Sub(x_max)
	// 3. 计算指数
	exp_x := shifted.Exp()
	// 4. 计算和
	sum_exp := exp_x.Sum([]int{axis})
	// 5. 归一化
	result := exp_x.Div(sum_exp)
	return result
}

func (t *Tensor) MinMax(axis int) *Tensor {
	// 1. 计算最大值
	x_max := t.Max([]int{axis})
	// 2. 计算最小值
	x_min := t.Min([]int{axis})
	// 3. 计算范围
	ranged := x_max.Sub(x_min)
	// 4. 归一化
	result := t.Sub(x_min).Div(ranged)
	return result
}
