package deepx

type MLP struct {
	ModuleBase
	fc1 *Linear
	fc2 *Linear
}

func NewMLP(name string, in_features, hidden_features int, dtype Dtype, g *Graph) *MLP {
	return &MLP{
		ModuleBase: ModuleBase{
			g:    g,
			name: name,
		},
		fc1: NewLinear(name+".fc1", in_features, hidden_features, dtype, g),
		fc2: NewLinear(name+".fc2", hidden_features, in_features, dtype, g),
	}
}

func (m *MLP) Forward(x *Tensor) *Tensor {
	x = m.fc1.Linear(x)
	x = x.Relu()
	x = m.fc2.Linear(x)
	return x
}
