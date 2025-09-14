# LRPC项目Makefile
# 用于管理测试环境和资源

.PHONY: test test-with-coverage test-setup test-teardown clean
.DEFAULT_GOAL := help

# 测试端口配置
REDIS_PORT := 16379
MYSQL_PORT := 13306
CLICKHOUSE_PORT := 18123

# Docker容器名称
REDIS_CONTAINER := lrpc-test-redis
MYSQL_CONTAINER := lrpc-test-mysql
CLICKHOUSE_CONTAINER := lrpc-test-clickhouse

# 测试数据库配置
MYSQL_ROOT_PASSWORD := test123
MYSQL_DATABASE := lrpc_test

help: ## 显示帮助信息
	@echo "LRPC测试工具"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test-setup: ## 启动测试所需的外部服务 (Redis, MySQL, ClickHouse)
	@echo "启动测试环境..."
	@docker ps -q -f name=${REDIS_CONTAINER} | grep -q . && echo "Redis已运行" || \
		(echo "启动Redis..." && docker run -d --name ${REDIS_CONTAINER} \
		-p ${REDIS_PORT}:6379 redis:7-alpine redis-server --requirepass test123)
	
	@docker ps -q -f name=${MYSQL_CONTAINER} | grep -q . && echo "MySQL已运行" || \
		(echo "启动MySQL..." && docker run -d --name ${MYSQL_CONTAINER} \
		-e MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD} \
		-e MYSQL_DATABASE=${MYSQL_DATABASE} \
		-p ${MYSQL_PORT}:3306 mysql:8.0 \
		--default-authentication-plugin=mysql_native_password)
	
	@docker ps -q -f name=${CLICKHOUSE_CONTAINER} | grep -q . && echo "ClickHouse已运行" || \
		(echo "启动ClickHouse..." && docker run -d --name ${CLICKHOUSE_CONTAINER} \
		-p ${CLICKHOUSE_PORT}:8123 \
		-p 19000:9000 \
		clickhouse/clickhouse-server:latest)
	
	@echo "等待服务启动完成..."
	@sleep 10
	
	@echo "检查服务状态..."
	@docker exec ${REDIS_CONTAINER} redis-cli -a test123 ping || true
	@docker exec ${MYSQL_CONTAINER} mysqladmin ping -h localhost -u root -p${MYSQL_ROOT_PASSWORD} || true
	@curl -s http://localhost:${CLICKHOUSE_PORT}/ping | grep -q "Ok" && echo "ClickHouse: 就绪" || echo "ClickHouse: 启动中..."
	
	@echo "测试环境启动完成!"

test-teardown: ## 停止并删除测试服务
	@echo "清理测试环境..."
	@docker stop ${REDIS_CONTAINER} ${MYSQL_CONTAINER} ${CLICKHOUSE_CONTAINER} 2>/dev/null || true
	@docker rm ${REDIS_CONTAINER} ${MYSQL_CONTAINER} ${CLICKHOUSE_CONTAINER} 2>/dev/null || true
	@echo "测试环境已清理"

test-status: ## 检查测试服务状态
	@echo "检查测试服务状态..."
	@echo "Redis:"
	@docker ps -f name=${REDIS_CONTAINER} --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" || echo "  未运行"
	@echo "MySQL:"  
	@docker ps -f name=${MYSQL_CONTAINER} --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" || echo "  未运行"
	@echo "ClickHouse:"
	@docker ps -f name=${CLICKHOUSE_CONTAINER} --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" || echo "  未运行"

test-logs: ## 查看测试服务日志
	@echo "Redis日志:"
	@docker logs --tail 20 ${REDIS_CONTAINER} 2>/dev/null || echo "Redis容器未运行"
	@echo -e "\nMySQL日志:"
	@docker logs --tail 20 ${MYSQL_CONTAINER} 2>/dev/null || echo "MySQL容器未运行"  
	@echo -e "\nClickHouse日志:"
	@docker logs --tail 20 ${CLICKHOUSE_CONTAINER} 2>/dev/null || echo "ClickHouse容器未运行"

test: test-setup ## 运行所有测试 (自动管理测试环境)
	@echo "运行测试..."
	@export REDIS_URL="redis://localhost:${REDIS_PORT}" && \
	 export MYSQL_URL="root:${MYSQL_ROOT_PASSWORD}@tcp(localhost:${MYSQL_PORT})/${MYSQL_DATABASE}?charset=utf8mb4&parseTime=True&loc=Local" && \
	 export CLICKHOUSE_URL="tcp://localhost:19000?database=default" && \
	 go test -v -timeout 30s ./...

test-with-coverage: test-setup ## 运行测试并生成覆盖率报告
	@echo "运行测试并生成覆盖率报告..."
	@export REDIS_URL="redis://localhost:${REDIS_PORT}" && \
	 export MYSQL_URL="root:${MYSQL_ROOT_PASSWORD}@tcp(localhost:${MYSQL_PORT})/${MYSQL_DATABASE}?charset=utf8mb4&parseTime=True&loc=Local" && \
	 export CLICKHOUSE_URL="tcp://localhost:19000?database=default" && \
	 go test -v -timeout 30s -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"
	@go tool cover -func=coverage.out | tail -1

test-unit: ## 仅运行单元测试 (不需要外部服务)
	@echo "运行单元测试..."
	@go test -v -short ./...

test-integration: test-setup ## 仅运行集成测试
	@echo "运行集成测试..."  
	@export REDIS_URL="redis://localhost:${REDIS_PORT}" && \
	 export MYSQL_URL="root:${MYSQL_ROOT_PASSWORD}@tcp(localhost:${MYSQL_PORT})/${MYSQL_DATABASE}?charset=utf8mb4&parseTime=True&loc=Local" && \
	 export CLICKHOUSE_URL="tcp://localhost:19000?database=default" && \
	 go test -v -run Integration ./...

benchmark: test-setup ## 运行性能测试
	@echo "运行性能测试..."
	@export REDIS_URL="redis://localhost:${REDIS_PORT}" && \
	 go test -v -bench=. -benchmem ./...

clean: test-teardown ## 清理所有生成的文件和测试环境
	@echo "清理生成的文件..."
	@rm -f coverage.out coverage.html
	@go clean -testcache
	@echo "清理完成"

deps: ## 安装/更新依赖
	@echo "更新Go依赖..."
	@go mod tidy
	@go mod download

build: ## 构建项目
	@echo "构建项目..."
	@go build ./...

lint: ## 运行代码检查
	@echo "运行代码检查..."
	@go vet ./...
	@go fmt ./...

# 快速开发循环
dev-test: ## 开发模式: 快速运行测试 (使用现有环境)
	@go test -v -short -timeout 10s ./...

# 检查Docker是否可用
check-docker:
	@docker --version > /dev/null 2>&1 || (echo "错误: 需要安装Docker来运行集成测试" && exit 1)

# 所有测试相关的目标都依赖Docker检查
test test-with-coverage test-setup benchmark: check-docker