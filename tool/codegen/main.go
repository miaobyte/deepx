// deepx key codegen — 从 spec/keys.yaml 生成各语言 Redis key SDK
//
// 用法:
//   go run . <project_root>
//
// 输入:  <project_root>/spec/keys.yaml
// 输出:
//   <project_root>/executor/vm/internal/keys/keys.go
//   <project_root>/executor/deepx-core/include/deepx/key_defs.h
//   <project_root>/tool/dashboard/frontend/src/api/keys.ts
//
// 设计原则:
//   单一真源 (spec/keys.yaml) → 多语言 SDK
//   go generate 兼容
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

// ============================================================================
// Spec 数据结构 — 对应 spec/keys.yaml
// ============================================================================

type Spec struct {
	Version   string             `json:"spec_version"`
	Ref       string             `json:"spec_ref"`
	Prefixes  map[string]Prefix  `json:"prefixes"`
	Keys      []KeyDef           `json:"keys"`
	Enums     []EnumDef          `json:"enums"`
	Constants []ConstantDef      `json:"constants"`
}

type Prefix struct {
	Value string `json:"value"`
	Doc   string `json:"doc"`
}

type Param struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Doc  string `json:"doc"`
}

type KeyDef struct {
	Name    string  `json:"name"`
	Section string  `json:"section"`
	Params  []Param `json:"params"`
	Pattern string  `json:"pattern"`
	Doc     string  `json:"doc"`
}

type EnumValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Doc   string `json:"doc"`
}

type EnumDef struct {
	Name     string      `json:"name"`
	Section  string      `json:"section"`
	Doc      string      `json:"doc"`
	Values   []EnumValue `json:"values"`
	Terminal []string    `json:"terminal"`
}

type ConstantDef struct {
	Name      string   `json:"name"`
	Value     string   `json:"value"`
	Type      string   `json:"type"`
	ListValue []string `json:"list_value"`
	Section  string   `json:"section"`
	Doc      string   `json:"doc"`
}

// ============================================================================
// 模板数据 — 传给各语言模板
// ============================================================================

type TemplateData struct {
	Spec        *Spec
	Prefixes    map[string]Prefix
	Keys        []KeyDef
	Enums       []EnumDef
	Constants   []ConstantDef
	HasParams   func(KeyDef) bool
	ParamList   func(KeyDef) string          // Go: name string, n int
	ParamListCPP func(KeyDef) string         // C++: const std::string& name, int n
	ParamListTS  func(KeyDef) string         // TS:  name: string, n: number
	ArgList     func(KeyDef) string          // Go: name, n
	ArgListCPP  func(KeyDef) string          // C++: name, n
	ArgListTS   func(KeyDef) string          // TS:  name, n
	GoType      func(string) string
	CPPType     func(string) string
	TSType      func(string) string
	Expand      func(string) string          // {prefix}name → value + name
	PatternExpr func(KeyDef) string          // Generate format string expression
	PatternExprCPP func(KeyDef) string       // C++ string concat expression
	PatternExprTS   func(KeyDef) string      // TS template literal
}

func goType(t string) string {
	switch t {
	case "string": return "string"
	case "int": return "int"
	default: return "string"
	}
}

func cppType(t string) string {
	switch t {
	case "string": return "const std::string&"
	case "int": return "int"
	default: return "const std::string&"
	}
}

func tsType(t string) string {
	switch t {
	case "string": return "string"
	case "int": return "number"
	default: return "string"
	}
}

func expandPattern(pattern string, prefixes map[string]Prefix) string {
	result := pattern
	for name, p := range prefixes {
		result = strings.ReplaceAll(result, "{"+name+"}", p.Value)
	}
	return result
}

func goParamList(k KeyDef) string {
	parts := make([]string, len(k.Params))
	for i, p := range k.Params {
		parts[i] = safeGoIdent(p.Name) + " " + goType(p.Type)
	}
	return strings.Join(parts, ", ")
}

func cppParamList(k KeyDef) string {
	parts := make([]string, len(k.Params))
	for i, p := range k.Params {
		parts[i] = cppType(p.Type) + " " + p.Name  // C++: no reserved word issue
	}
	return strings.Join(parts, ", ")
}

func tsParamList(k KeyDef) string {
	parts := make([]string, len(k.Params))
	for i, p := range k.Params {
		parts[i] = safeGoIdent(p.Name) + ": " + tsType(p.Type)
	}
	return strings.Join(parts, ", ")
}

func goArgList(k KeyDef) string {
	parts := make([]string, len(k.Params))
	for i, p := range k.Params {
		parts[i] = safeGoIdent(p.Name)
	}
	return strings.Join(parts, ", ")
}

func goPatternExpr(k KeyDef, prefixes map[string]Prefix) string {
	pattern := k.Pattern
	// Replace {prefix} with prefix value + %s or %d
	for name, p := range prefixes {
		pattern = strings.ReplaceAll(pattern, "{"+name+"}", p.Value)
	}
	// Replace {param} with format verbs
	args := make([]string, 0)
	for _, p := range k.Params {
		switch p.Type {
		case "int":
			pattern = strings.Replace(pattern, "{"+p.Name+"}", "%d", 1)
			args = append(args, safeGoIdent(p.Name))
		default:
			pattern = strings.Replace(pattern, "{"+p.Name+"}", "%s", 1)
			args = append(args, safeGoIdent(p.Name))
		}
	}
	if strings.Contains(pattern, "%") {
		return fmt.Sprintf("fmt.Sprintf(%q, %s)", pattern, strings.Join(args, ", "))
	}
	// No format verbs needed — just simple concatenation or literal
	// Check if it's fully expanded (no {} remaining)
	if strings.Contains(pattern, "{") {
		// Has unresolved placeholders, use concat
		return buildConcatExpr(pattern, prefixes, k)
	}
	return fmt.Sprintf("%q", pattern)
}

func buildConcatExpr(pattern string, prefixes map[string]Prefix, k KeyDef) string {
	// Build a concatenation expression for patterns that mix prefix + params
	// Replace known prefixes and params, what remains becomes string literal
	parts := tokenize(pattern, prefixes, k)
	var expr strings.Builder
	for _, part := range parts {
		if expr.Len() > 0 {
			expr.WriteString(" + ")
		}
		expr.WriteString(part)
	}
	return expr.String()
}

func tokenize(pattern string, prefixes map[string]Prefix, k KeyDef) []string {
	type token struct {
		isParam bool
		value   string
	}
	var tokens []token

	// Replace prefixes first
	for name, p := range prefixes {
		if strings.Contains(pattern, "{"+name+"}") {
			tokens = append(tokens, token{false, fmt.Sprintf("%q", p.Value)})
			pattern = strings.Replace(pattern, "{"+name+"}", "\x00", 1)
		}
	}

	// Now handle params in the remaining pattern parts
	remaining := strings.Split(pattern, "\x00")
	for i, part := range remaining {
		if i > 0 && i <= len(k.Params) {
			p := k.Params[i-1]
			tokens = append(tokens, token{true, safeGoIdent(p.Name)})
		}
		// Handle any remaining {} in the part
		for _, p := range k.Params {
			placeholder := "{" + p.Name + "}"
			if strings.Contains(part, placeholder) {
				before, after, _ := strings.Cut(part, placeholder)
				if before != "" {
					tokens = append(tokens, token{false, fmt.Sprintf("%q", before)})
				}
				tokens = append(tokens, token{true, safeGoIdent(p.Name)})
				part = after
			}
		}
		if part != "" && !strings.Contains(part, "{") {
			tokens = append(tokens, token{false, fmt.Sprintf("%q", part)})
		}
	}
	if len(tokens) == 0 {
		return []string{fmt.Sprintf("%q", pattern)}
	}
	var result []string
	for _, t := range tokens {
		result = append(result, t.value)
	}
	return result
}

func cppPatternExpr(k KeyDef, prefixes map[string]Prefix) string {
	// Tokenize pattern: split into literal chunks and param placeholders
	// preserving the original order.
	pattern := k.Pattern

	// Build param lookup
	paramMap := make(map[string]Param)
	for _, p := range k.Params {
		paramMap["{"+p.Name+"}"] = p
	}

	// Replace prefix placeholders with their values
	for name, p := range prefixes {
		pattern = strings.ReplaceAll(pattern, "{"+name+"}", p.Value)
	}

	if len(k.Params) == 0 {
		return fmt.Sprintf("%q", pattern)
	}

	// Parse remaining pattern into ordered tokens
	var parts []string
	remaining := pattern
	for remaining != "" {
		// Find next placeholder
		start := strings.Index(remaining, "{")
		if start == -1 {
			if remaining != "" {
				parts = append(parts, fmt.Sprintf("%q", remaining))
			}
			break
		}
		if start > 0 {
			parts = append(parts, fmt.Sprintf("%q", remaining[:start]))
			remaining = remaining[start:]
		}
		end := strings.Index(remaining, "}")
		if end == -1 {
			if remaining != "" {
				parts = append(parts, fmt.Sprintf("%q", remaining))
			}
			break
		}
		placeholder := remaining[:end+1]
		if p, ok := paramMap[placeholder]; ok {
			if p.Type == "int" {
				parts = append(parts, "std::to_string("+p.Name+")")
			} else {
				parts = append(parts, p.Name)
			}
		} else {
			// Unknown placeholder — treat as literal
			parts = append(parts, fmt.Sprintf("%q", placeholder))
		}
		remaining = remaining[end+1:]
	}

	if len(parts) == 0 {
		return fmt.Sprintf("%q", pattern)
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return strings.Join(parts, " + ")
}

func tsPatternExpr(k KeyDef, prefixes map[string]Prefix) string {
	result := k.Pattern
	for name, p := range prefixes {
		result = strings.ReplaceAll(result, "{"+name+"}", p.Value)
	}
	for _, p := range k.Params {
		result = strings.ReplaceAll(result, "{"+p.Name+"}", "${"+p.Name+"}")
	}
	return "`" + result + "`"
}

// ============================================================================
// Go 模板
// ============================================================================

const goTemplate = `// Code generated by tool/codegen from spec/keys.yaml. DO NOT EDIT.
// 所有 Redis key 规范参见: {{.Spec.Ref}}

package keys

import "fmt"

// ============================================================================
// §2.1 保留路径前缀
// ============================================================================
const ({{range $name, $p := .Prefixes}}
	Prefix{{camel $name}} = {{printf "%q" $p.Value}} // {{$p.Doc}}{{end}}
)

// ============================================================================
// Key 构造函数
// ============================================================================
{{range .Keys}}
// {{.Name}} {{.Doc}}{{if .Params}}{{range .Params}}
//   {{.Name}}: {{.Doc}}{{end}}{{end}}
func {{.Name}}({{call $.ParamList .}}) string {
	{{- if .Params}}
	return {{call $.PatternExpr .}}
	{{- else}}
	return {{call $.PatternExpr .}}
	{{- end}}
}
{{end}}

// ============================================================================
// {{(index .Enums 0).Section}} 状态
// ============================================================================
{{range .Enums}}{{$enum := .}}
type {{$enum.Name}} string
const ({{range $enum.Values}}
	{{$enum.Name}}{{.Name}} {{$enum.Name}} = {{printf "%q" .Value}} // {{.Doc}}{{end}}
)

func (s {{$enum.Name}}) IsTerminal() bool {
	return {{range $i, $t := $enum.Terminal}}{{if $i}} || {{end}}s == {{$enum.Name}}{{$t}}{{end}}
}
{{end}}

// ============================================================================
// 常量
// ============================================================================
{{range .Constants}}
{{- if eq .Type "list"}}
var {{.Name}} = []string{ {{range .ListValue}}{{printf "%q" .}},{{end}} } // {{.Doc}}
{{- else}}
const {{.Name}} = {{printf "%q" .Value}} // {{.Doc}}
{{- end}}
{{end}}

// ============================================================================
// 工具函数
// ============================================================================

func IsRelative(param string) bool {
	return len(param) >= 2 && param[:2] == "./"
}

func ResolveRelative(vtid, param string) string {
	if IsRelative(param) {
		return Metathread(vtid) + "/" + param[2:]
	}
	return param
}

func FuncMain() string { return UsrFunc("main") }
`

// ============================================================================
// C++ 模板
// ============================================================================

const cppTemplate = `// Code generated by tool/codegen from spec/keys.yaml. DO NOT EDIT.
// 所有 Redis key 规范参见: {{.Spec.Ref}}

#pragma once

#include <string>
#include <string_view>

namespace deepx::keys {

// ============================================================================
// §2.1 保留路径前缀
// ============================================================================
{{range $name, $p := .Prefixes}}
inline constexpr std::string_view kPrefix{{camel $name}} = {{printf "%q" $p.Value}}; // {{$p.Doc}}{{end}}

// ============================================================================
// Key 构造函数
// ============================================================================
{{range .Keys}}
// {{.Name}}: {{.Doc}}
{{- if .Params}}
inline std::string {{.Name}}({{call $.ParamListCPP .}}) {
{{- if gt (len .Params) 0}}
    return {{call $.PatternExprCPP .}};
{{- else}}
    return {{call $.PatternExprCPP .}};
{{- end}}
}
{{- else}}
inline constexpr std::string_view {{.Name}} = {{call $.PatternExprCPP .}};
{{- end}}
{{end}}

// ============================================================================
// {{(index .Enums 0).Section}} 状态
// ============================================================================
{{range .Enums}}
namespace {{snake .Name}} {
{{- range .Values}}
    inline constexpr std::string_view k{{.Name}} = {{printf "%q" .Value}}; // {{.Doc}}{{end}}
} // namespace {{snake .Name}}
{{end}}

// ============================================================================
// 常量
// ============================================================================
{{range .Constants}}
{{- if eq .Type "list"}}
// {{.Doc}}
// (C++ 中列表定义为 constexpr array)
{{- else}}
inline constexpr std::string_view k{{.Name}} = {{printf "%q" .Value}}; // {{.Doc}}
{{- end}}
{{end}}

} // namespace deepx::keys
`

// ============================================================================
// TypeScript 模板
// ============================================================================

const tsTemplate = `// Code generated by tool/codegen from spec/keys.yaml. DO NOT EDIT.
// 所有 Redis key 规范参见: {{.Spec.Ref}}

// ============================================================================
// Redis Key — 服务端 key 路径
// ============================================================================

export const KEY = {
{{range .Keys}}  /** {{.Doc}} */
  {{.Name}}({{call $.ParamListTS .}}): string { return {{call $.PatternExprTS .}} },
{{end}}
} as const;

// ============================================================================
// {{(index .Enums 0).Section}} 状态枚举
// ============================================================================
{{range .Enums}}{{$enum := .}}
export enum {{$enum.Name}} {
{{- range $enum.Values}}
  {{.Name}} = {{printf "%q" .Value}}, // {{.Doc}}
{{- end}}
}

export function isTerminal{{$enum.Name}}(s: string): boolean {
  return {{range $i, $t := $enum.Terminal}}{{if $i}} || {{end}}s === {{$enum.Name}}.{{$t}}{{end}};
}

export const {{$enum.Name}}_LABELS: Record<{{$enum.Name}}, string> = {
{{- range $enum.Values}}
  [{{$enum.Name}}.{{.Name}}]: '{{.Doc}}',
{{- end}}
};
{{end}}

// ============================================================================
// 常量
// ============================================================================
{{range .Constants}}
{{- if eq .Type "list"}}
export const {{.Name}} = [{{range .ListValue}}{{printf "%q" .}},{{end}}] as const;
{{- else}}
export const {{.Name}} = {{printf "%q" .Value}}; // {{.Doc}}
{{- end}}
{{end}}

// ============================================================================
// 默认值 (前后端共享语义)
// ============================================================================

export const DEFAULTS = {
  /** 默认入口函数名 */
  ENTRY_FUNC: 'main',
  /** 代码执行默认超时 (秒) */
  TIMEOUT_SEC: 60,
} as const;
`

// ============================================================================
// 模板辅助函数
// ============================================================================

func newTemplateData(spec *Spec) TemplateData {
	return TemplateData{
		Spec:      spec,
		Prefixes:  spec.Prefixes,
		Keys:      spec.Keys,
		Enums:     spec.Enums,
		Constants: spec.Constants,
		HasParams: func(k KeyDef) bool { return len(k.Params) > 0 },
		ParamList:    goParamList,
		ParamListCPP: cppParamList,
		ParamListTS:  tsParamList,
		ArgList:      goArgList,
		ArgListCPP:   func(k KeyDef) string { return goArgList(k) },
		ArgListTS:    func(k KeyDef) string { return goArgList(k) },
		GoType:       goType,
		CPPType:      cppType,
		TSType:       tsType,
		PatternExpr: func(k KeyDef) string { return goPatternExpr(k, spec.Prefixes) },
		PatternExprCPP: func(k KeyDef) string { return cppPatternExpr(k, spec.Prefixes) },
		PatternExprTS:  func(k KeyDef) string { return tsPatternExpr(k, spec.Prefixes) },
	}
}

// ============================================================================
// 模板函数映射
// ============================================================================

// safeGoIdent 将 Go 保留字参数名转为合法标识符 (加 _ 后缀)。
func safeGoIdent(name string) string {
	keywords := map[string]bool{
		"type": true, "func": true, "map": true, "chan": true,
		"interface": true, "select": true, "range": true,
		"go": true, "defer": true, "fallthrough": true,
		"break": true, "case": true, "const": true, "continue": true,
		"default": true, "else": true, "for": true, "goto": true,
		"if": true, "import": true, "package": true, "return": true,
		"struct": true, "switch": true, "var": true,
	}
	if keywords[name] {
		return name + "_"
	}
	return name
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"safeGoIdent": safeGoIdent,
		"camel": func(s string) string {
			parts := strings.Split(s, "_")
			for i, p := range parts {
				if len(p) > 0 {
					parts[i] = strings.ToUpper(p[:1]) + p[1:]
				}
			}
			return strings.Join(parts, "")
		},
		"snake": func(s string) string {
			var result strings.Builder
			for i, r := range s {
				if i > 0 && r >= 'A' && r <= 'Z' {
					result.WriteByte('_')
				}
				result.WriteRune(r)
			}
			return strings.ToLower(result.String())
		},
		"EnumConstName": func(enums []EnumDef, enumName, valueName string) string {
			return enumName + valueName
		},
	}
}

// ============================================================================
// 生成逻辑
// ============================================================================

func loadSpec(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read spec: %w", err)
	}
	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parse spec: %w", err)
	}
	return &spec, nil
}

func generate(spec *Spec, root string) error {
	td := newTemplateData(spec)
	funcs := templateFuncs()

	outputs := []struct {
		path     string
		tmpl     string
		language string
	}{
		{filepath.Join(root, "executor/vm/internal/keys/keys.go"), goTemplate, "Go"},
		{filepath.Join(root, "executor/deepx-core/include/deepx/key_defs.h"), cppTemplate, "C++"},
		{filepath.Join(root, "tool/dashboard/frontend/src/api/keys.ts"), tsTemplate, "TypeScript"},
	}

	for _, out := range outputs {
		tmpl, err := template.New(out.language).Funcs(funcs).Parse(out.tmpl)
		if err != nil {
			return fmt.Errorf("parse %s template: %w", out.language, err)
		}

		dir := filepath.Dir(out.path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}

		f, err := os.Create(out.path)
		if err != nil {
			return fmt.Errorf("create %s: %w", out.path, err)
		}

		if err := tmpl.Execute(f, td); err != nil {
			f.Close()
			return fmt.Errorf("execute %s: %w", out.language, err)
		}
		f.Close()
		fmt.Printf("  ✓ %s (%s)\n", out.language, out.path)
	}
	return nil
}

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	specPath := filepath.Join(root, "spec", "keys.yaml")
	fmt.Printf("deepx key codegen\n")
	fmt.Printf("  spec: %s\n", specPath)

	spec, err := loadSpec(specPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	if err := generate(spec, root); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("done.")
}
