# 设置变量
BINARY=runshell
VERSION=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-X main.Version=${VERSION} -X main.GitCommit=${COMMIT} -X main.BuildTime=${BUILD_TIME}

# Go 相关变量
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Docker 相关变量
DOCKER_IMAGE=runshell
DOCKER_TAG=$(VERSION)

# 默认目标
.PHONY: all
all: clean test build

# 清理构建产物
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@$(GOCLEAN)

# 更新依赖
.PHONY: deps
deps:
	@echo "Updating dependencies..."
	@$(GOGET) -v -t -d ./...
	@$(GOMOD) tidy

# 运行测试
.PHONY: test
test:
	@echo "Running tests..."
	@./script/test.sh

# 运行测试（跳过集成测试）
.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	@SKIP_DOCKER_TESTS=1 $(GOTEST) -v -short ./...

# 构建所有平台
.PHONY: build
build:
	@echo "Building for all platforms..."
	@./script/build.sh

# 仅构建当前平台
.PHONY: build-local
build-local:
	@echo "Building for local platform..."
	@mkdir -p bin
	@$(GOBUILD) -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/runshell

# 构建 Docker 镜像
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	@./script/docker-build.sh

# 运行 Docker 容器
.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	@docker-compose -f docker/docker-compose.yml up -d

# 停止 Docker 容器
.PHONY: docker-stop
docker-stop:
	@echo "Stopping Docker container..."
	@docker-compose -f docker/docker-compose.yml down

# 运行本地服务器
.PHONY: run
run: build-local
	@echo "Running local server..."
	@./bin/$(BINARY) server --http :8080

# 生成代码覆盖率报告
.PHONY: coverage
coverage: test
	@echo "Generating coverage report..."
	@go tool cover -html=coverage.out

# 代码格式化
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# 代码检查
.PHONY: lint
lint:
	@echo "Linting code..."
	@go vet ./...
	@test -z "$$(gofmt -l .)"

# 创建新的 Git 标签
.PHONY: tag
tag:
	@echo "Current version: $(VERSION)"
	@read -p "Enter new version: " version; \
	git tag -a $$version -m "Release $$version"
	@echo "Run 'git push --tags' to push the new tag."

# 帮助信息
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make              - Clean, test, and build"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make deps         - Update dependencies"
	@echo "  make test         - Run all tests"
	@echo "  make test-unit    - Run unit tests only"
	@echo "  make build        - Build for all platforms"
	@echo "  make build-local  - Build for local platform"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run Docker container"
	@echo "  make docker-stop  - Stop Docker container"
	@echo "  make run          - Run local server"
	@echo "  make coverage     - Generate coverage report"
	@echo "  make fmt          - Format code"
	@echo "  make lint         - Check code style"
	@echo "  make tag          - Create a new Git tag"
	@echo "  make help         - Show this help" 