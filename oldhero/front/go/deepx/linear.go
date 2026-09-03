package deepx

type Linear struct {
	ModuleBase
	W *Tensor
	b *Tensor
}

func NewLinear(name string, in_features, out_features int, dtype Dtype, g *Graph) (m *Linear) {
	if g == nil {
		g = NewGraph()
	}
	if name == "" {
		name = "linear"
	}
	m = &Linear{
		ModuleBase: ModuleBase{
			g:    g,
			name: name,
		},
	}

	in_features_node := g.AddConstArg(name + ".in_features")
	in_features_node.SetInt(in_features)
	out_features_node := g.AddConstArg(name + ".out_features")
	out_features_node.SetInt(out_features)

	//如果利用矩阵grad时的取巧运算，则需要将W的shape设置为[out_features,in_features]来实现提前转置
	m.W = g.AddTensor(name+".W", dtype, []int{in_features, out_features}, true, in_features_node, out_features_node).Tensor()
	m.b = g.AddTensor(name+".bias", dtype, []int{out_features}, true, out_features_node).Tensor()
	return m
}
func (m *Linear) Linear(input *Tensor) *Tensor {
	y := input.Matmul(m.W)
	z := y.Add(m.b)
	return z
}
