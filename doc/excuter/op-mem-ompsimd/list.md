## op-mem-ompsimd 支持算子列表 

本页面由 `excuter/op-mem-ompsimd 生成，请勿手动修改 

| Operation | Author | Func Def | Math Formula | IR Instruction |
|-----------|--------|------------|--------------|----------------|
| concat |  none  | concat()->() | Tresult = concat([T1, T2...], axis=3) | concat()->() |
| addscalar | miaobyte | addscalar(tensor<any> a, var<any> scalar)->(tensor<any> c) | T3=T1+scalar | addscalar(tensor<any> a, var<any> scalar)->(tensor<any> c) |
| add | cblas | add(tensor<float64|float32> a, tensor<float64|float32> b)->(tensor<float64|float32> c) | T3=T1+T2 | add(tensor<float64|float32> a, tensor<float64|float32> b)->(tensor<float64|float32> c) |
| add | miaobyte | add(tensor<any> a, tensor<any> b)->(tensor<any> c) | T3=T1+T2 | add(tensor<any> a, tensor<any> b)->(tensor<any> c) |
| uniform | miaobyte | uniform(tensor<any> t, var<any> low, var<any> high, var<int32> seed)->() | uniform(T1,low,high,seed) | uniform(tensor<any> t, var<any> low, var<any> high, var<int32> seed)->() |
| subscalar | miaobyte | subscalar(tensor<any> a, var<any> scalar)->(tensor<any> c) | T3=T1-scalar | subscalar(tensor<any> a, var<any> scalar)->(tensor<any> c) |
| arange | miaobyte | arange(tensor<any> t, var<any> start, var<any> step)->() | arange(T1,start,step) | arange(tensor<any> t, var<any> start, var<any> step)->() |
| constant | miaobyte | constant(tensor<any> t, var<any> value)->() | constant(T1,value) | constant(tensor<any> t, var<any> value)->() |
| print | miaobyte | print(tensor<any> )->() | print(T1) | print(tensor<any> )->() |
| print | miaobyte | print(tensor<any> , var<string> )->() | print(T1) | print(tensor<any> , var<string> )->() |
| newtensor |  none  | newtensor(vector<int32> shape)->(tensor<any> tensor1) | T1 =Tensor(shape=[...]) | newtensor(vector<int32> shape)->(tensor<any> tensor1) |
| newtensor |  none  | newtensor(var<string> shape)->(tensor<any> tensor1) | T1 =Tensor(shape=[...]) | newtensor(var<string> shape)->(tensor<any> tensor1) |
| vecset |  none  | vecset(vector<any> value)->(vector<any> name) | shape = [3  4  5] | vecset(vector<any> value)->(vector<any> name) |
| sub | miaobyte | sub(tensor<any> a, tensor<any> b)->(tensor<any> c) | T3=T1-T2 | sub(tensor<any> a, tensor<any> b)->(tensor<any> c) |
| argset |  none  | argset(var<any> value)->(var<any> name) | var argname = argvalue | argset(var<any> value)->(var<any> name) |
