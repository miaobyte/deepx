
from typing import Union

from deepx.tensor import Tensor,tensor_method

@tensor_method
def reducemax(self, dim:tuple[int,...],keepdim:bool=False,out:Union[Tensor,str]='')->Tensor:
    assert isinstance(dim,tuple)
    for i in dim:
        assert isinstance(i,int)
    from deepx.nn.functional import reducemax as reduce_max_func
    return reduce_max_func(self,dim,keepdim,out)

@tensor_method
def reducemin(self, dim:tuple[int,...],keepdim:bool=False,out:Union[Tensor,str]='')->Tensor:
    assert isinstance(dim,tuple)
    for i in dim:
        assert isinstance(i,int)
    from deepx.nn.functional import reducemin as reduce_min_func
    return reduce_min_func(self,dim,keepdim,out)


@tensor_method
def sum(self, dim:tuple[int,...],keepdim:bool=False,out:Union[Tensor,str]='')->Tensor:
    assert isinstance(dim,tuple)
    for i in dim:
        assert isinstance(i,int)
    from deepx.nn.functional import  sum as sum_func
    return  sum_func(self,dim,keepdim,out)

@tensor_method
def prod(self, dim:tuple[int,...],keepdim:bool=False,out:Union[Tensor,str]='')->Tensor:
    assert isinstance(dim,tuple)
    for i in dim:
        assert isinstance(i,int)
    from deepx.nn.functional import prod as prod_func
    return prod_func(self,dim,keepdim,out)   

@tensor_method
def mean(self,dim:tuple[int,...],keepdim:bool=False)->Tensor:
    assert isinstance(dim,tuple)
    for i in dim:
        assert isinstance(i,int)
    from deepx.nn.functional import mean as mean_func
    return mean_func(self,dim,keepdim)
 