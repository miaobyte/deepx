# cursor 根据deepseek-qwen2 1.5b到处的deepx模型目录名，反推导出的pytorch模型
# DeepSeek-R1-Distill-Qwen-1.5B

import os
import yaml
import math
import numpy as np
import torch
import torch.nn as nn
from transformers import AutoTokenizer

class ModelConfig:
    def __init__(self):
        # 根据实际配置文件更新参数
        self.hidden_size = 1536          # 与配置文件中hidden_size一致
        self.num_hidden_layers = 28      # 对应num_hidden_layers
        self.num_attention_heads = 12    # 对应num_attention_heads
        self.num_key_value_heads = 2     # 保持GQA结构
        self.intermediate_size = 8960    # 对应intermediate_size
        self.rms_norm_eps = 1e-6        
        self.vocab_size = 151936        # 与vocab_size一致
        self.rope_theta = 10000.0       
        self.hidden_act = "silu"        
        self.max_position_embeddings = 131072  # 新增参数
        self.sliding_window = 4096      # 滑动窗口参数
        # 确保与DeepX模型配置一致

class RMSNorm(nn.Module):
    def __init__(self, dim, eps=1e-6):
        super().__init__()
        self.eps = eps
        self.weight = nn.Parameter(torch.ones(dim))

    def _norm(self, x):
        return x * torch.rsqrt(x.pow(2).mean(-1, keepdim=True) + self.eps)

    def forward(self, x):
        return self.weight * self._norm(x.float()).type_as(x)

def apply_rotary_pos_emb(q, k, theta=10000.0):
    """修正后的旋转位置编码实现"""
    dim = q.size(-1)
    seq_len = q.size(1)
    
    # 生成频率矩阵
    inv_freq = 1.0 / (theta ** (torch.arange(0, dim, 2).float() / dim))
    position = torch.arange(seq_len, device=q.device).type_as(inv_freq)
    sinusoid = torch.einsum("i,j->ij", position, inv_freq)
    
    # 应用旋转矩阵
    sin = torch.sin(sinusoid).unsqueeze(0).unsqueeze(2)  # [1, seq, 1, dim//2]
    cos = torch.cos(sinusoid).unsqueeze(0).unsqueeze(2)  # [1, seq, 1, dim//2]
    
    q_rot = q[..., :dim//2]
    q_pass = q[..., dim//2:]
    q = torch.cat((q_rot * cos - q_pass * sin, q_rot * sin + q_pass * cos), dim=-1)
    
    k_rot = k[..., :dim//2]
    k_pass = k[..., dim//2:]
    k = torch.cat((k_rot * cos - k_pass * sin, k_rot * sin + k_pass * cos), dim=-1)
    
    return q, k

class Qwen2Attention(nn.Module):
    def __init__(self, config):
        super().__init__()
        self.config = config
        self.hidden_size = config.hidden_size
        self.num_heads = config.num_attention_heads
        self.head_dim = self.hidden_size // self.num_heads
        self.num_key_value_heads = config.num_key_value_heads
        self.rope_theta = getattr(config, 'rope_theta', 10000.0)
        
        # 修正投影层维度（关键修改）
        self.q_proj = nn.Linear(self.hidden_size, self.num_heads * self.head_dim, bias=True)
        self.k_proj = nn.Linear(self.hidden_size, self.num_key_value_heads * self.head_dim, bias=True)
        self.v_proj = nn.Linear(self.hidden_size, self.num_key_value_heads * self.head_dim, bias=True)
        self.o_proj = nn.Linear(self.num_heads * self.head_dim, self.hidden_size, bias=False)

        # 新增缓存机制
        self.kv_cache = None

    def forward(self, hidden_states, attention_mask=None, use_cache=False):
        bs, seq_len, _ = hidden_states.size()
        
        # 投影操作（修正视图维度）
        q = self.q_proj(hidden_states).view(bs, seq_len, self.num_heads, self.head_dim)
        k = self.k_proj(hidden_states).view(bs, seq_len, self.num_key_value_heads, self.head_dim)
        v = self.v_proj(hidden_states).view(bs, seq_len, self.num_key_value_heads, self.head_dim)

        # 处理KV缓存时扩展key/value头数
        if use_cache and self.kv_cache is not None:
            k = torch.cat([self.kv_cache[0], k], dim=1)
            v = torch.cat([self.kv_cache[1], v], dim=1)
        self.kv_cache = (k, v) if use_cache else None

        # 应用RoPE时需要扩展key/value头数
        q, k = apply_rotary_pos_emb(q, k, theta=self.rope_theta)
        
        # 扩展key/value头数以匹配query头数（GQA关键）
        k = k.repeat_interleave(self.num_heads // self.num_key_value_heads, dim=2)
        v = v.repeat_interleave(self.num_heads // self.num_key_value_heads, dim=2)

        # 注意力掩码处理（改进空值情况）
        if attention_mask is not None:
            if attention_mask.dim() == 2:
                attention_mask = attention_mask.unsqueeze(1).unsqueeze(1)
            attention_mask = attention_mask.to(dtype=q.dtype) * -1e4
        else:
            # 当没有提供mask时创建基础mask
            attention_mask = torch.zeros_like(q[:, :, 0, 0], dtype=q.dtype)  # [bs, seq_len]

        # 添加滑动窗口掩码（修正空值处理）
        if self.config.sliding_window is not None:
            seq_len = hidden_states.size(1)
            sliding_window = self.config.sliding_window
            diagonal = seq_len - sliding_window - 1
            
            # 基于当前设备创建mask
            window_mask = torch.full((seq_len, seq_len), float('-inf'), 
                                   device=hidden_states.device,
                                   dtype=hidden_states.dtype)
            window_mask = torch.triu(window_mask, diagonal=diagonal)
            window_mask = window_mask.unsqueeze(0).unsqueeze(0)  # [1, 1, seq, seq]
            
            attention_mask = attention_mask + window_mask

        # 注意力计算
        attn_weights = torch.einsum("bqhd,bkhd->bhqk", q, k) / math.sqrt(self.head_dim)
        attn_weights = torch.softmax(attn_weights, dim=-1)
        attn_output = torch.einsum("bhqk,bkhd->bqhd", attn_weights, v)
        
        # 合并头并输出
        attn_output = attn_output.contiguous().view(bs, seq_len, -1)
        return self.o_proj(attn_output)

class Qwen2MLP(nn.Module):
    def __init__(self, config):
        super().__init__()
        self.hidden_size = config.hidden_size
        self.intermediate_size = config.intermediate_size
        
        # 修正名称与权重文件对齐
        self.gate_proj = nn.Linear(self.hidden_size, self.intermediate_size, bias=False)
        self.up_proj = nn.Linear(self.hidden_size, self.intermediate_size, bias=False)
        self.down_proj = nn.Linear(self.intermediate_size, self.hidden_size, bias=False)
        self.act_fn = nn.SiLU()

    def forward(self, x):
        return self.down_proj(self.act_fn(self.gate_proj(x)) * self.up_proj(x))

class Qwen2DecoderLayer(nn.Module):
    def __init__(self, config):
        super().__init__()
        self.hidden_size = config.hidden_size
        self.self_attn = Qwen2Attention(config)
        
        # 根据目录结构中的layernorm定义
        self.input_layernorm = RMSNorm(self.hidden_size, eps=config.rms_norm_eps)
        self.post_attention_layernorm = RMSNorm(self.hidden_size, eps=config.rms_norm_eps)
        
        self.mlp = Qwen2MLP(config)

    def forward(self, hidden_states, attention_mask=None):
        # 实现残差连接
        residual = hidden_states
        hidden_states = self.input_layernorm(hidden_states)
        hidden_states = self.self_attn(hidden_states, attention_mask)
        hidden_states = residual + hidden_states

        residual = hidden_states
        hidden_states = self.post_attention_layernorm(hidden_states)
        hidden_states = self.mlp(hidden_states)
        hidden_states = residual + hidden_states
        return hidden_states

class Qwen2Model(nn.Module):
    def __init__(self, config):
        super().__init__()
        self.config = config
        self.embed_tokens = nn.Embedding(config.vocab_size, config.hidden_size)
        self.layers = nn.ModuleList(
            [Qwen2DecoderLayer(config) for _ in range(config.num_hidden_layers)]
        )
        self.norm = RMSNorm(config.hidden_size, eps=config.rms_norm_eps)

    def forward(self, input_ids, attention_mask=None, position_ids=None, use_cache=False):
        hidden_states = self.embed_tokens(input_ids)
        for layer in self.layers:
            hidden_states = layer(hidden_states, attention_mask=attention_mask)
        hidden_states = self.norm(hidden_states)
        return hidden_states

# deepx目录结构
# tokenizer.json
# tokenizer_config.json
# config.yaml
# lm_head_weight.data
# lm_head_weight.shape
# model_embed_tokens_weight.data
# model_embed_tokens_weight.shape
# model_layers_0_input_layernorm_weight.data
# model_layers_0_input_layernorm_weight.shape
# model_layers_0_mlp_down_proj_weight.data
# model_layers_0_mlp_down_proj_weight.shape
# model_layers_0_mlp_gate_proj_weight.data
# model_layers_0_mlp_gate_proj_weight.shape
# model_layers_0_mlp_up_proj_weight.data
# model_layers_0_mlp_up_proj_weight.shape
# model_layers_0_post_attention_layernorm_weight.data
# model_layers_0_post_attention_layernorm_weight.shape
# model_layers_0_self_attn_k_proj_bias.data
# model_layers_0_self_attn_k_proj_bias.shape
# model_layers_0_self_attn_k_proj_weight.data
# model_layers_0_self_attn_k_proj_weight.shape
# model_layers_0_self_attn_o_proj_weight.data
# model_layers_0_self_attn_o_proj_weight.shape
# model_layers_0_self_attn_q_proj_bias.data
# model_layers_0_self_attn_q_proj_bias.shape
# model_layers_0_self_attn_q_proj_weight.data
# model_layers_0_self_attn_q_proj_weight.shape
# model_layers_0_self_attn_v_proj_bias.data
# model_layers_0_self_attn_v_proj_bias.shape
# model_layers_0_self_attn_v_proj_weight.data
# model_layers_0_self_attn_v_proj_weight.shape
# model_layers_1到27层
# model_norm_weight.data
# model_norm_weight.shape

class DeviceConfig:
    def __init__(self, device_type="auto"):
        self.device = self._get_device(device_type)
        self.is_cuda = self.device.type == 'cuda'
    
    def _get_device(self, device_type):
        if device_type == "auto":
            return torch.device("cuda" if torch.cuda.is_available() else "cpu")
        return torch.device(device_type)

class DeepXModelLoader:
    def __init__(self, model_path):
        self.model_path = model_path
        self.config = self._load_config()
        self.tokenizer = AutoTokenizer.from_pretrained(model_path)
        
    def _load_config(self):
        config_path = os.path.join(self.model_path, "config.yaml")
        with open(config_path, 'r') as f:
            return yaml.safe_load(f)
    
    def load_tensor(self, tensor_name):
        """加载DeepX格式的权重"""
        # 直接使用原始名称构建路径
        base_path = os.path.join(self.model_path, "tensors", tensor_name)
        
        # 加载形状信息
        with open(f"{base_path}.shape", 'r') as f:
            shape_info = yaml.safe_load(f)
            shape = tuple(shape_info['shape'])
            
        # 加载二进制数据
        dtype_map = {'float32': np.float32, 'bfloat16': np.uint16}  # 根据实际类型调整
        np_dtype = dtype_map.get(shape_info['dtype'], np.float32)
        
        with open(f"{base_path}.data", 'rb') as f:
            data = np.frombuffer(f.read(), dtype=np_dtype)
            data = data.copy()  # 创建可写副本
            data.setflags(write=True)
        
        return torch.from_numpy(data.reshape(shape))

class Qwen2Inference(nn.Module):
    def __init__(self, config, device_config):
        super().__init__()
        self.device_config = device_config
        self.model = Qwen2Model(config).to(device_config.device)
        self.lm_head = nn.Linear(config.hidden_size, config.vocab_size, bias=False).to(device_config.device)

    def forward(self, input_ids, attention_mask=None):
        hidden_states = self.model(input_ids, attention_mask)
        return self.lm_head(hidden_states)

def load_model(model_path, device_type="auto"):
    config = ModelConfig()
    device_config = DeviceConfig(device_type)
    
    model = Qwen2Inference(config, device_config)
    
    # 加载权重到指定设备
    state_dict = {}
    loader = DeepXModelLoader(model_path)
    tensor_dir = os.path.join(model_path, "tensors")
    for filename in os.listdir(tensor_dir):
        if filename.endswith(".shape"):
            param_name = filename[:-6]
            torch_name = param_name.replace('model_layers_', 'model.layers.')
            torch_name = torch_name.replace('_self_attn_', '.self_attn.')
            param = loader.load_tensor(param_name).to(device_config.device)
            state_dict[torch_name] = param
    
    model.load_state_dict(state_dict, strict=True)
    
    # 应用动态量化
    if device_config.device.type == 'cuda':
        model = torch.quantization.quantize_dynamic(
            model,
            {torch.nn.Linear},
            dtype=torch.qint8
        )
    
    if device_config.is_cuda:
        torch.cuda.empty_cache()
    
    return model

def run_inference(text, model, loader):
    inputs = loader.tokenizer(text, return_tensors="pt")
    
    with torch.no_grad():
        outputs = model(**inputs)
    
    predicted_id = torch.argmax(outputs[0, -1]).item()
    return loader.tokenizer.decode(predicted_id)

# 新增生成函数
def generate(
    model, 
    tokenizer,
    prompt: str,
    max_length: int = 512,
    temperature: float = 0.7,
    top_p: float = 0.9
):
    device = model.device_config.device
    inputs = tokenizer(prompt, return_tensors="pt")
    input_ids = inputs["input_ids"].to(device)
    attention_mask = inputs.get("attention_mask", None)
    if attention_mask is not None:
        attention_mask = attention_mask.to(device)
    
    generated = input_ids
    
    penalty_alpha = 0.6  # 重复惩罚系数
    for _ in range(max_length):
        with torch.autocast(device_type='cuda', dtype=torch.bfloat16):
            outputs = model(input_ids=generated, attention_mask=attention_mask)
            
        logits = outputs[:, -1, :] / temperature
        
        probs = torch.softmax(logits, dim=-1)
        
        # 修正Top-p采样逻辑（关键修改）
        sorted_probs, sorted_indices = torch.sort(probs, dim=-1, descending=True)
        cumulative_probs = torch.cumsum(sorted_probs, dim=-1)
        
        # 创建要移除的索引掩码
        sorted_indices_to_remove = cumulative_probs > top_p
        # 保留第一个超过阈值的token
        sorted_indices_to_remove[..., 1:] = sorted_indices_to_remove[..., :-1].clone()
        sorted_indices_to_remove[..., 0] = 0

        # 安全地应用掩码（修正索引越界问题）
        indices_to_remove = sorted_indices_to_remove.scatter(
            dim=-1, 
            index=sorted_indices, 
            src=sorted_indices_to_remove
        )
        probs = probs.masked_fill(indices_to_remove, 0.0)
        
        # 重新归一化概率
        probs = probs / probs.sum(dim=-1, keepdim=True)
        
        next_token = torch.multinomial(probs, num_samples=1)
        print(tokenizer.decode(next_token[0].cpu().numpy()), end="", flush=True)
        generated = torch.cat([generated, next_token.to(device)], dim=-1)
        
        if next_token.item() == tokenizer.eos_token_id:
            break
            
        # 更新attention_mask
        if attention_mask is not None:
            attention_mask = torch.cat([
                attention_mask,
                torch.ones((1, 1), device=device)
            ], dim=1)
            
    return tokenizer.decode(generated[0].cpu().numpy(), skip_special_tokens=True)

def preprocess_input(text: str, tokenizer):
    inputs = tokenizer(
        text,
        padding=True,
        truncation=True,
        max_length=512,
        return_tensors="pt"
    )
    # 创建因果注意力掩码
    seq_len = inputs["input_ids"].size(1)
    mask = torch.tril(torch.ones(seq_len, seq_len))
    inputs["attention_mask"] = inputs["attention_mask"] * mask
    return inputs

class InferenceContext:
    def __enter__(self):
        torch.set_grad_enabled(False)
        torch.set_num_threads(os.cpu_count())  # 优化CPU并行
        
    def __exit__(self, *args):
        torch.set_grad_enabled(True)
        # 清空缓存
        for module in model.modules():
            if hasattr(module, 'kv_cache'):
                module.kv_cache = None

# 使用示例
if __name__ == "__main__":
    model_path = "/home/lipeng/model/deepseek-ai/deepx"
    device_type = "cpu"  # 可选： "cpu" 或 "auto"
    
    model = load_model(model_path, device_type)
    loader = DeepXModelLoader(model_path)
    
    with InferenceContext():
        while True:
            text = input("Input: ")
            result = generate(model, loader.tokenizer, text)
            print(f"Response: {result}")