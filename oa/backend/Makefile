# 权限管理系统后端常用快捷指令
# 使用方式: make <target>

# 变量定义
APP_NAME := permission-system
SERVER_ENTRY := ./cmd/server
BUILD_OUTPUT := server
DOCKER_IMAGE := $(APP_NAME):go1.23
CONFIG_FILE := config/config.yaml

# 默认目标
.PHONY: help
help: ## 显示所有可用指令
	@echo "可用的 make 指令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 开发运行
.PHONY: run
run: ## 本地运行后端服务（开发模式）
	go run $(SERVER_ENTRY)

# 编译构建
.PHONY: build
build: ## 编译生成可执行文件（./server）
	go build -o $(BUILD_OUTPUT) $(SERVER_ENTRY)

.PHONY: build-linux
build-linux: ## 交叉编译 Linux 可执行文件
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_OUTPUT)-linux $(SERVER_ENTRY)

# 代码质量
.PHONY: fmt
fmt: ## 格式化所有 Go 代码
	go fmt ./...

.PHONY: vet
vet: ## 执行 go vet 静态检查
	go vet ./...

.PHONY: lint
lint: fmt vet ## 格式化并执行静态检查

.PHONY: test
test: ## 运行所有单元测试
	go test -v ./...

# 依赖管理
.PHONY: tidy
tidy: ## 整理并下载 Go 模块依赖
	go mod tidy

.PHONY: download
download: ## 下载 Go 模块依赖
	go mod download

# Swagger 文档生成
.PHONY: swag
swag: ## 生成 Swagger 接口文档
	go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g cmd/server/main.go

# Docker 构建与运行
.PHONY: docker-build
docker-build: ## 构建 Docker 镜像（标签: permission-system:go1.23）
	docker build \
	  --build-arg GIT_USER=$(GIT_USER) \
	  --build-arg GIT_TOKEN=$(GIT_TOKEN) \
	  -t $(DOCKER_IMAGE) .

.PHONY: docker-run
docker-run: ## 使用默认配置运行 Docker 容器（端口 8080）
	docker run --rm -p 8080:8080 $(DOCKER_IMAGE)

.PHONY: docker-run-with-config
docker-run-with-config: ## 挂载本地 config.yaml 运行 Docker 容器
	docker run --rm -p 8080:8080 -v $(PWD)/$(CONFIG_FILE):/app/config/config.yaml $(DOCKER_IMAGE)

.PHONY: docker-test
docker-test: ## 运行容器化集成测试（自动清理数据库容器）
	docker compose -f test/docker-compose.test.yml up --build --abort-on-container-exit && \
	docker compose -f test/docker-compose.test.yml down

.PHONY: docker-push
docker-push: ## 推送 Docker 镜像（需提前设置仓库地址）
	@echo "请修改 Makefile 中的 DOCKER_REGISTRY 后再使用此命令"
	# docker tag $(DOCKER_IMAGE) your-registry/$(DOCKER_IMAGE)
	# docker push your-registry/$(DOCKER_IMAGE)

# 清理
.PHONY: clean
clean: ## 清理编译产物和临时文件
	rm -f $(BUILD_OUTPUT) $(BUILD_OUTPUT)-linux
	go clean -cache

# 完整 CI 检查
.PHONY: ci
ci: fmt vet test build ## 执行完整的 CI 流程：格式化、检查、测试、编译
