############-------PyTorch-------################

import torch
torch_t1 = torch.full((2,3,4, ), 10, dtype=torch.float32)
torch_t2 = torch.full((2,3,4, ), 10, dtype=torch.float32)
torch_t3 = torch_t1 + torch_t2
torch_t3.add_(0.5)
print()
print(torch_t3)

############-------DEEPX-------################

from deepx import Tensor,full
import os

print()

t1 = full(2,3,4, value=10,dtype="float32",name="t1")
t2 = full(2,3,4, value=10,dtype="float32",name="t2")
t3 = t1.add(t2,"t3")
t3.add_(0.5)
print(t3)

script_name = os.path.splitext(os.path.basename( os.path.abspath(__file__)))[0]
str=t3.graph.to_dot()
str.render(script_name+".dot", format='svg')
