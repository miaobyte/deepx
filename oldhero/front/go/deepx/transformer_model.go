package deepx

import "fmt"

type Transformer struct {
	ModuleBase
	embedding *Linear
	layers    []*TransformerLayer
	ln_final  *LayerNorm
}

func NewTransformer(name string, num_layers, hidden_size, num_heads, mlp_ratio int, dtype Dtype, g *Graph) *Transformer {
	if name == "" {
		name = "transformer"
	}

	m := &Transformer{
		ModuleBase: ModuleBase{
			g:    g,
			name: name,
		},
		embedding: NewLinear(name+".embedding", hidden_size, hidden_size, dtype, g),
		layers:    make([]*TransformerLayer, num_layers),
		ln_final:  NewLayerNorm(name+".ln_final", hidden_size, dtype, g),
	}

	for i := 0; i < num_layers; i++ {
		m.layers[i] = NewTransformerLayer(
			fmt.Sprintf("%s.layer_%d", name, i),
			hidden_size,
			num_heads,
			mlp_ratio,
			dtype,
			g,
		)
	}

	return m
}

func (m *Transformer) Forward(x *Tensor) *Tensor {
	x = m.embedding.Linear(x)

	for _, layer := range m.layers {
		x = layer.Forward(x)
	}

	x = m.ln_final.LayerNorm(x)
	return x
}
