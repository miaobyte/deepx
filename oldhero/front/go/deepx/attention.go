package deepx

import (
	"math"
)

// MultiHeadAttention 模块
type MultiHeadAttention struct {
	ModuleBase
	q_proj      *Linear
	k_proj      *Linear
	v_proj      *Linear
	o_proj      *Linear
	num_heads   int
	hidden_size int
}

func NewMultiHeadAttention(name string, hidden_size, num_heads int, dtype Dtype, g *Graph) *MultiHeadAttention {
	if g == nil {
		g = NewGraph()
	}
	if name == "" {
		name = "attention"
	}

	m := &MultiHeadAttention{
		ModuleBase: ModuleBase{
			g:    g,
			name: name,
		},
		q_proj:      NewLinear(name+".q_proj", hidden_size, hidden_size, dtype, g),
		k_proj:      NewLinear(name+".k_proj", hidden_size, hidden_size, dtype, g),
		v_proj:      NewLinear(name+".v_proj", hidden_size, hidden_size, dtype, g),
		o_proj:      NewLinear(name+".o_proj", hidden_size, hidden_size, dtype, g),
		num_heads:   num_heads,
		hidden_size: hidden_size,
	}
	return m
}

func (m *MultiHeadAttention) Forward(q, k, v *Tensor) *Tensor {
	batch_size := q.Shape.shape[0]
	seq_len := q.Shape.shape[1]
	head_dim := q.Shape.shape[2] / m.num_heads

	// 1. 线性变换
	query := m.q_proj.Linear(q)
	key := m.k_proj.Linear(k)
	value := m.v_proj.Linear(v)

	// 2. 重塑为多头形式
	// [batch_size, seq_len, hidden_size] -> [batch_size, seq_len, num_heads, head_dim]
	query = query.Reshape([]int{batch_size, seq_len, m.num_heads, head_dim})
	key = key.Reshape([]int{batch_size, seq_len, m.num_heads, head_dim})
	value = value.Reshape([]int{batch_size, seq_len, m.num_heads, head_dim})

	// 3. 转置以便进行批量矩阵乘法
	// [batch_size, seq_len, num_heads, head_dim] -> [batch_size, num_heads, seq_len, head_dim]
	query = query.Transpose([]int{0, 2, 1, 3})
	key = key.Transpose([]int{0, 2, 1, 3})
	value = value.Transpose([]int{0, 2, 1, 3})

	// 4. 计算注意力分数
	scores := query.Matmul(key)

	// 5. Scale
	d_k := float32(head_dim)
	scores = scores.MulScalar(1.0 / float32(math.Sqrt(float64(d_k))))

	// 6. Softmax
	attn := scores.Softmax(1)

	// 7. 加权求和
	out := attn.Matmul(value)

	// 8. 转置回原始形状
	// [batch_size, num_heads, seq_len, head_dim] -> [batch_size, seq_len, num_heads, head_dim]
	out = out.Transpose([]int{0, 2, 1, 3})

	// 9. 重塑回原始维度
	// [batch_size, seq_len, num_heads, head_dim] -> [batch_size, seq_len, hidden_size]
	out = out.Reshape([]int{batch_size, seq_len, m.hidden_size})

	// 10. 输出投影
	out = m.o_proj.Linear(out)

	return out
}
