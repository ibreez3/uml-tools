.PHONY: build clean test help

# 项目名称
PROJECT := uml-tools

# Go 参数
GO := go
GOFLAGS := -v

# 输出目录
OUTPUT_DIR := bin

# 工具列表
TOOLS := ulm-class ulm-pkg

all: build

# 编译所有工具
build: $(TOOLS)

ulm-class:
	@echo "🔨 编译类图工具..."
	@mkdir -p $(OUTPUT_DIR)
	$(GO) build $(GOFLAGS) -o $(OUTPUT_DIR)/ulm-class ./cmd/class-diagram

ulm-pkg:
	@echo "🔨 编译包图工具..."
	@mkdir -p $(OUTPUT_DIR)
	$(GO) build $(GOFLAGS) -o $(OUTPUT_DIR)/ulm-pkg ./cmd/package-diagram

# 清理
clean:
	@echo "🧹 清理构建文件..."
	@rm -rf $(OUTPUT_DIR)

# 测试
test:
	@echo "🧪 运行测试..."
	$(GO) test $(GOFLAGS) ./...

# 格式化代码
fmt:
	@echo "📝 格式化代码..."
	$(GO) fmt ./...

# 安装到 GOPATH
install:
	@echo "📦 安装工具..."
	$(GO) install ./cmd/class-diagram
	$(GO) install ./cmd/package-diagram

# 帮助
help:
	@echo "ulmutil - Go UML Diagram Generator"
	@echo ""
	@echo "可用命令:"
	@echo "  make build    - 编译所有工具"
	@echo "  make clean    - 清理构建文件"
	@echo "  make test     - 运行测试"
	@echo "  make fmt      - 格式化代码"
	@echo "  make install  - 安装到 GOPATH"
	@echo "  make help     - 显示帮助"
	@echo ""
	@echo "使用方法:"
	@echo "  ./bin/ulm-class -o output.puml /path/to/project"
	@echo "  ./bin/ulm-pkg   -o output.puml /path/to/project"
