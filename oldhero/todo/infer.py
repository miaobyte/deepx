import sys
import threading
from transformers import AutoModelForCausalLM, AutoTokenizer, TextStreamer
import torch

def init_model():
    model_path = "/home/lipeng/model/deepseek-ai/DeepSeek-R1-Distill-Qwen-7B"
    try:
        tokenizer = AutoTokenizer.from_pretrained(model_path)
        tokenizer.pad_token = tokenizer.eos_token
        
        model = AutoModelForCausalLM.from_pretrained(
            model_path,
            trust_remote_code=True,
            torch_dtype=torch.bfloat16,
            device_map="auto",
            # use_flash_attention_2=True  # 启用Flash Attention
        ).eval()
        
        return model, tokenizer
    except Exception as e:
        raise RuntimeError(f"模型初始化失败: {str(e)}")

class StdoutStreamer(TextStreamer):
    def __init__(self, tokenizer):
        super().__init__(tokenizer, skip_prompt=True, skip_special_tokens=True)
        self.cache = []
        self.first_token = True
    
    def on_finalized_text(self, text: str, stream_end: bool = False):
        self.cache.append(text)
        if stream_end or len(self.cache) >= 2:
            full_text = "".join(self.cache)
            sys.stdout.write(full_text)
            sys.stdout.flush()
            self.cache = []

def generate_stream(model, tokenizer, text, max_length):
    formatted_text = f"<|im_start|>user\n{text}<|im_end|>\n<|im_start|>assistant\n"
    inputs = tokenizer(
        formatted_text, 
        return_tensors='pt', 
        add_special_tokens=False,
        return_attention_mask=True
    ).to(model.device)
    streamer = StdoutStreamer(tokenizer)
    
    generation_kwargs = {
        "input_ids": inputs.input_ids,
        "attention_mask": inputs.attention_mask,
        "max_new_tokens": max_length,
        "pad_token_id": tokenizer.eos_token_id,
        "temperature": 0.3,      # 降低随机性
        "top_p": 0.85,           # 限制采样范围
        "repetition_penalty": 1.2, # 增强重复抑制
        "do_sample": True,
        "streamer": streamer
    }
    
    thread = threading.Thread(target=model.generate, kwargs=generation_kwargs)
    thread.start()
    thread.join()
    print("\n")  # 流式结束换行

def generate_text(model, tokenizer, text, max_length=50):
    formatted_text = f"<|im_start|>user\n{text}<|im_end|>\n<|im_start|>assistant\n"
    inputs = tokenizer(
        formatted_text,
        return_tensors='pt',
        add_special_tokens=False,
        return_attention_mask=True
    ).to(model.device)
    
    with torch.no_grad():
        output = model.generate(
            inputs.input_ids,
            attention_mask=inputs.attention_mask,
            max_new_tokens=max_length,
            pad_token_id=tokenizer.eos_token_id,
            temperature=0.3,
            top_p=0.85,
            repetition_penalty=1.2,
            do_sample=True
        )
    
    return tokenizer.decode(
        output[0][len(inputs.input_ids[0]):],
        skip_special_tokens=True,
        clean_up_tokenization_spaces=True
    )

def main():
    try:
        model, tokenizer = init_model()
        sys.stderr.write("模型加载成功，输入提示开始生成（Ctrl+C退出）\n")
    except Exception as e:
        sys.stderr.write(f"服务启动失败: {e}\n")
        return

    # 单独测试分词器
    text = "<|im_start|>user\n你好<|im_end|>\n<|im_start|>assistant\n"
    tokens = tokenizer.encode(text, add_special_tokens=False)
    decoded = tokenizer.decode(tokens)
    assert decoded == text  # 验证编码解码一致性
    try:
        for line in sys.stdin:
            text = line.strip()
            if not text:
                continue

            # 固定参数设置
            max_length = 2048  # 最大生成长度
            stream = True     # 始终使用流式
            
            if stream:
                generate_stream(model, tokenizer, text, max_length)
            else:
                result = generate_text(model, tokenizer, text, max_length)
                print(result)
                
    except KeyboardInterrupt:
        sys.stderr.write("\n服务已终止\n")
    except Exception as e:
        sys.stderr.write(f"运行时错误: {str(e)}\n")

if __name__ == '__main__':
    main()


