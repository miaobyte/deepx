from typing import Union
from deepx.tensor import Tensor,Shape

from .leaffunc_life import newtensor
from .authormap import defaultauthor

def reshape(t:Tensor,shape:tuple[int,...],out:Union[Tensor,str]='')->Tensor:
    assert isinstance(shape,tuple)
    for i in shape:
        assert isinstance(i,int) and i>0

    if isinstance(out,str) or out is None:
        outshape=shape
        outtensor=newtensor(outshape,dtype=t.dtype,name=out)
    else:
        outtensor=out
        outtensor._shape=Shape(shape)
    from .rtf_changeshape import rtf_reshape
    rtf_reshape(t,shape,outtensor,defaultauthor['reshape'])
    return outtensor
 

def permute(t:Tensor,
            dimorder:tuple[int,...],
            out:Union[Tensor,str]='')->Tensor:
    assert isinstance(dimorder,tuple)
    for i in dimorder:
        assert isinstance(i,int)

    if t.ndim!=len(dimorder):
        raise ValueError(f"shape参数不合法,当前输入维度数：{len(dimorder)}，张量维度数：{t.ndim}")
    dimorder = [d % t.ndim for d in dimorder]
    outtensor=out
    if isinstance(out,str) or out is None:
        outshape = [t.shape[dim] for dim in dimorder]
        outtensor=newtensor(tuple(outshape),dtype=t.dtype,name=out)

    from .rtf_changeshape import rtf_transpose
    rtf_transpose(t,dimorder,outtensor,defaultauthor['transpose'])
    return outtensor

 

def concat(tensors:Union[list[Tensor],tuple[Tensor,...]],dim:int,out:Union[Tensor,str]='')->Tensor:
    assert isinstance(dim,int)
    assert isinstance(tensors,list) or isinstance(tensors,tuple)
    for t in tensors:
        assert isinstance(t,Tensor)
    dim=dim%tensors[0].ndim
    outtensor=out
    if isinstance(out,str) or out is None:
        outshape=Shape.concat(tuple(t.shape for t in tensors),dim)
        outtensor=newtensor(tuple(outshape),dtype=tensors[0].dtype,name=out)
    from .rtf_changeshape import rtf_concat
    rtf_concat(tensors,dim,outtensor,defaultauthor['concat'])
    return outtensor

def broadcastTo(t:Tensor,newshape:tuple[int,...],out:Union[Tensor,str]='')->Tensor:
    assert isinstance(newshape,tuple)
    assert len(newshape)==t.ndim
    new_shape=[]
    for idx,i in enumerate(newshape):
        assert isinstance(i,int)
        if i==-1:
            new_shape.append(t.shape[idx])
        elif i>0:
            new_shape.append(i)
        else:
            raise ValueError(f"新形状参数不合法，维度 {idx} 的值{i}")
    new_shape=tuple(new_shape)
    if t.shape==new_shape:
        return t

    bshape=Shape.broadcast_shape(t.shape,new_shape)
    if bshape!=new_shape:
        raise ValueError(f"广播失败：{t.shape} 无法广播为 {new_shape}，请先reshape")
    outtensor=out
    if isinstance(out,str) or out is None:
        outshape=new_shape
        outtensor=newtensor(outshape,dtype=t.dtype,name=out)
    from .rtf_changeshape import rtf_broadcastTo
    rtf_broadcastTo(t,new_shape,outtensor,defaultauthor['broadcastTo'])
    return outtensor

broadcast_to = broadcastTo


def indexselect(input:Tensor,indices:Tensor,dim:int,out:Union[Tensor,str]='')->Tensor:
    assert dim>=0 and dim<input.ndim
    outtensor=out
    if isinstance(out,str) or out is None:
        outshape=Shape.indexselectshape(input.shape,indices.shape,dim)
        outtensor=newtensor(outshape,dtype=input.dtype,name=out)
    assert outtensor.shape==outshape
    
    from .rtf_changeshape import rtf_indexselect
    rtf_indexselect(input,indices,dim,outtensor,defaultauthor['indexselect'])
    return outtensor

def repeat(input:Tensor,repeats:tuple[int,...],out:Union[Tensor,str]=''):
    assert isinstance(repeats,tuple)
    assert input.Shape.ndim==len(repeats)
    for i in repeats:
        assert isinstance(i,int) and i>0
    outtensor=out
    if isinstance(out,str) or out is None:
        outshape=Shape.repeatshape(input.shape,repeats)
        outtensor=newtensor(outshape,dtype=input.dtype,name=out)
    from .rtf_changeshape import rtf_repeat
    rtf_repeat(input,repeats,outtensor,defaultauthor['repeat'])
    return outtensor
