package deepx

import (
	"fmt"
	"strings"
)

func (g *Graph) ToDOT() string {
	var builder strings.Builder
	builder.WriteString("digraph computational_graph {\n")
	builder.WriteString("  rankdir=TB;\n")
	builder.WriteString("  node [shape=record];\n")

	// 遍历所有节点
	for _, node := range g.nodes {
		builder.WriteString(fmt.Sprintf(`"%p" `, node))
		switch node.Ntype() {
		case NodeTensor:
			// 张量节点：显示形状和梯度信息
			builder.WriteString(fmt.Sprintf(`[label= "%s \n %v`, node.Name(), node.(*TensorNode).Tensor().Shape))
			// if node.(*TensorNode).Tensor().requiresGrad {
			// 	builder.WriteString(`\n require_grad`)
			// }
			builder.WriteString(`"`)
			builder.WriteString(",shape=box")
			builder.WriteString(",labeljust=l")
			builder.WriteString(",color=skyblue")
			builder.WriteString(",style=filled")
			builder.WriteString(",fillcolor=aliceblue")
			builder.WriteString(",fontname=\"Sans-Serif\"")

		case NodeOp:
			// 操作节点：突出显示操作类型
			opNode := node.(*OpNode)
			builder.WriteString(fmt.Sprintf(`[label="%s"`, opNode.Shortchar()))
			builder.WriteString(",shape=box")
			builder.WriteString(",style=filled")
			builder.WriteString(",fillcolor=lightgray")
			builder.WriteString(",color=darkslategray")
			builder.WriteString(",fontname=\"Courier Bold\"")

		case NodeConstArg:
			// 常量参数节点：显示参数值
			constNode := node.(*ConstArgNode)
			var valueStr string
			switch constNode.argType {
			case ArgTypeInt:
				valueStr = fmt.Sprintf("%d", constNode.Int())
			case ArgTypeFloat:
				valueStr = fmt.Sprintf("%.2f", constNode.Float()) + "f"
			case ArgTypeString:
				valueStr = constNode.String()
			}
			builder.WriteString(fmt.Sprintf(`[label="%s"`, valueStr))
			builder.WriteString(",shape=diamond")
			builder.WriteString(",style=filled")
			builder.WriteString(",fillcolor=lightyellow")
			builder.WriteString(",color=goldenrod")
			builder.WriteString(",fontname=\"Sans-Serif\"")
		}
		builder.WriteString("];\n")
	}

	// 添加边连接，为边也添加样式
	for _, node := range g.nodes {
		for _, input := range node.Inputs() {
			builder.WriteString(fmt.Sprintf(`  "%p" -> "%p"`, input, node))
			builder.WriteString(`[color=gray40,penwidth=1.2,arrowsize=0.8]`)
			builder.WriteString(";\n")
		}
	}

	builder.WriteString("}\n")
	return builder.String()
}
