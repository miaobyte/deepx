package transformer

// Config 定义模型配置
type Config struct {
	// 模型基本配置
	HiddenSize       int
	NumLayers        int
	NumHeads         int
	MLPRatio         int
	VocabSize        int
	MaxSeqLength     int
	InitializerRange float32

	// 注意力相关配置
	AttentionImpl string
	SlidingWindow int
	UseFlashAttn  bool

	// 生成相关配置
	UseCache    bool
	BeamSize    int
	TopK        int
	TopP        float32
	Temperature float32
}
