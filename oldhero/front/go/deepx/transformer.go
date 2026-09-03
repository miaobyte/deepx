package deepx

type TransformerLayer struct {
	ModuleBase
	attention *MultiHeadAttention
	mlp       *MLP
	ln1       *LayerNorm
	ln2       *LayerNorm
}

func NewTransformerLayer(name string, hidden_size, num_heads, mlp_ratio int, dtype Dtype, g *Graph) *TransformerLayer {
	if name == "" {
		name = "transformer_layer"
	}

	return &TransformerLayer{
		ModuleBase: ModuleBase{
			g:    g,
			name: name,
		},
		attention: NewMultiHeadAttention(name+".attn", hidden_size, num_heads, dtype, g),
		mlp:       NewMLP(name+".mlp", hidden_size, mlp_ratio*hidden_size, dtype, g),
		ln1:       NewLayerNorm(name+".ln1", hidden_size, dtype, g),
		ln2:       NewLayerNorm(name+".ln2", hidden_size, dtype, g),
	}
}

func (m *TransformerLayer) Forward(x *Tensor) *Tensor {
	// 1. Self Attention
	h := m.ln1.LayerNorm(x)
	h = m.attention.Forward(h, h, h)
	x = x.Add(h) // residual

	// 2. MLP
	h = m.ln2.LayerNorm(x)
	h = m.mlp.Forward(h)
	x = x.Add(h) // residual

	return x
}
