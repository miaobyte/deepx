package deepx

// LayerNorm 层归一化
type LayerNorm struct {
	ModuleBase
	weight *Tensor
	bias   *Tensor
}

func NewLayerNorm(name string, normalized_shape int, dtype Dtype, g *Graph) *LayerNorm {
	m := &LayerNorm{
		ModuleBase: ModuleBase{
			g:    g,
			name: name,
		},
	}

	m.weight = g.AddTensor(name+".weight", dtype, []int{normalized_shape}, true).Tensor()
	m.bias = g.AddTensor(name+".bias", dtype, []int{normalized_shape}, true).Tensor()

	return m
}

func (m *LayerNorm) LayerNorm(x *Tensor) *Tensor {

	return nil
}

// BatchNorm 批归一化
type BatchNorm struct {
	ModuleBase
	weight       *Tensor
	bias         *Tensor
	running_mean *Tensor
	running_var  *Tensor
}

func NewBatchNorm(name string, num_features int, dtype Dtype, g *Graph) *BatchNorm {
	m := &BatchNorm{
		ModuleBase: ModuleBase{
			g:    g,
			name: name,
		},
	}

	m.weight = g.AddTensor(name+".weight", dtype, []int{num_features}, true).Tensor()
	m.bias = g.AddTensor(name+".bias", dtype, []int{num_features}, true).Tensor()
	m.running_mean = g.AddTensor(name+".running_mean", dtype, []int{num_features}, false).Tensor()
	m.running_var = g.AddTensor(name+".running_var", dtype, []int{num_features}, false).Tensor()

	return m
}

func (m *BatchNorm) BatchNorm(x *Tensor) *Tensor {
	return nil
}

// InstanceNorm 实例归一化
type InstanceNorm struct {
	ModuleBase
	weight *Tensor
	bias   *Tensor
}

func NewInstanceNorm(name string, num_features int, dtype Dtype, g *Graph) *InstanceNorm {
	m := &InstanceNorm{
		ModuleBase: ModuleBase{
			g:    g,
			name: name,
		},
	}

	m.weight = g.AddTensor(name+".weight", dtype, []int{num_features}, true).Tensor()
	m.bias = g.AddTensor(name+".bias", dtype, []int{num_features}, true).Tensor()

	return m
}

func (m *InstanceNorm) InstanceNorm(x *Tensor) *Tensor {
	return nil
}

// GroupNorm 组归一化
type GroupNorm struct {
	ModuleBase
	weight     *Tensor
	bias       *Tensor
	num_groups int
}

func NewGroupNorm(name string, num_groups, num_channels int, dtype Dtype, g *Graph) *GroupNorm {
	m := &GroupNorm{
		ModuleBase: ModuleBase{
			g:    g,
			name: name,
		},
		num_groups: num_groups,
	}

	m.weight = g.AddTensor(name+".weight", dtype, []int{num_channels}, true).Tensor()
	m.bias = g.AddTensor(name+".bias", dtype, []int{num_channels}, true).Tensor()

	// 添加num_groups作为常量参数
	num_groups_node := g.AddConstArg(name + ".num_groups")
	num_groups_node.SetInt(num_groups)

	return m
}

func (m *GroupNorm) GroupNorm(x *Tensor) *Tensor {
	return nil
}

// RMSNorm Root Mean Square Layer Normalization
type RMSNorm struct {
	ModuleBase
	weight *Tensor // 缩放参数
	eps    float32 // 数值稳定性参数
}

func NewRMSNorm(name string, normalized_shape int, dtype Dtype, g *Graph) *RMSNorm {
	m := &RMSNorm{
		ModuleBase: ModuleBase{
			g:    g,
			name: name,
		},
		eps: 1e-6, // 默认值
	}

	// 只需要缩放参数，不需要偏置项
	m.weight = g.AddTensor(name+".weight", dtype, []int{normalized_shape}, true).Tensor()

	return m
}

func (m *RMSNorm) RMSNorm(x *Tensor) *Tensor {
	return nil
}
