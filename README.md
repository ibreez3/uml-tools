# uml-tools - Go UML Diagram Generator

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

**uml-tools** 是一个 Go 语言编写的 UML 图表生成工具，可以从 Go 项目源代码自动生成 **类图 (Class Diagram)** 和 **包图 (Package Diagram)**。

## ✨ 特性

- 🎯 **类图生成** - 自动分析 Go 代码中的 struct 和 interface
- 📦 **包图生成** - 分析包之间的依赖关系
- 🎨 **三格式支持** - PlantUML、Mermaid、draw.io (diagrams.net)
- 🚀 **零依赖** - 仅使用 Go 标准库
- 📊 **按包分类** - 类图使用 namespace 按包名分组显示

## 📦 安装

```bash
git clone git@github.com:ibreez3/uml-tools.git
cd uml-tools

# 编译
go build -o bin/uml-tools ./cmd/ulm-tools
```

## 🚀 使用方法

### 统一命令格式

```bash
uml-tools <command> [options] <project-path>
```

### 生成类图

```bash
# PlantUML 格式
uml-tools class -o classDiagram.puml /path/to/project

# Mermaid 格式
uml-tools class -format mermaid -o classDiagram.mmd /path/to/project

# draw.io 格式
uml-tools class -format drawio -o classDiagram.drawio /path/to/project

# 自定义标题
uml-tools class -title "My Class Diagram" -format drawio -o output.drawio /path/to/project
```

### 生成包图

```bash
# PlantUML 格式
uml-tools pkg -o packageDiagram.puml /path/to/project

# Mermaid 格式
uml-tools pkg -format mermaid -o packageDiagram.mmd /path/to/project

# draw.io 格式
uml-tools pkg -format drawio -o packageDiagram.drawio /path/to/project

# 自定义标题
uml-tools pkg -title "My Package Diagram" -format drawio -o output.drawio /path/to/project
```

## 📋 命令行参数

### 类图 (class)

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-o` | 输出文件路径 | classDiagram.puml |
| `-title` | 图表标题 | Go Project Class Diagram |
| `-format` | 输出格式：plantuml / mermaid / drawio | plantuml |

### 包图 (pkg)

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-o` | 输出文件路径 | packageDiagram.puml |
| `-title` | 图表标题 | Go Project Package Diagram |
| `-format` | 输出格式：plantuml / mermaid / drawio | plantuml |

## 📊 输出格式

### PlantUML
- 文件扩展名：`.puml`
- 在线查看：https://www.planttext.com/
- VS Code 插件：PlantUML

### Mermaid
- 文件扩展名：`.mmd`
- 在线查看：https://mermaid.live/
- GitHub/Notion 原生支持

### draw.io
- 文件扩展名：`.drawio`
- 在线查看：https://app.diagrams.net/
- 可编辑的 XML 格式

## 📁 项目结构

```
uml-tools/
├── cmd/
│   └── ulm-tools/       # 统一命令行工具
│       └── main.go
├── internal/
│   └── diagram/         # 核心生成逻辑
│       ├── class.go     # 类图生成
│       └── package.go   # 包图生成
├── go.mod
├── README.md
└── LICENSE
```

## ⚠️ 注意事项

1. **跳过的文件/目录**:
   - `*_test.go` 测试文件
   - `vendor/` 依赖目录
   - `.git/` Git 目录
   - `node_modules/` Node 依赖

2. **类图关系**: 自动生成的关系基于字段类型，复杂关系需手动补充

3. **包图依赖**: 仅显示项目内部包之间的依赖

## 🛠️ 开发

```bash
# 格式化代码
go fmt ./...

# 编译
go build -o bin/uml-tools ./cmd/ulm-tools

# 安装到 GOPATH
go install ./cmd/ulm-tools
```

## 📝 License

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📧 联系方式

- GitHub: [@ibreez3](https://github.com/ibreez3)
- Project: [uml-tools](https://github.com/ibreez3/uml-tools)
