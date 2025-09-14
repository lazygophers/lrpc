# LRPC项目Makefile
# 用于管理测试环境和资源

.PHONY: test test-with-coverage test-setup test-teardown clean clean-all clean-docker clean-files clean-cache clean-test clean-check clean-safe clean-emergency reset
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
		-p ${REDIS_PORT}:6379 redis:7.4.1-alpine redis-server --requirepass test123)

	@docker ps -q -f name=${MYSQL_CONTAINER} | grep -q . && echo "MySQL已运行" || \
		(echo "启动MySQL..." && docker run -d --name ${MYSQL_CONTAINER} \
		-e MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD} \
		-e MYSQL_DATABASE=${MYSQL_DATABASE} \
		-p ${MYSQL_PORT}:3306 mysql:8.2.0 \
		--default-authentication-plugin=mysql_native_password)

	@docker ps -q -f name=${CLICKHOUSE_CONTAINER} | grep -q . && echo "ClickHouse已运行" || \
		(echo "启动ClickHouse..." && docker run -d --name ${CLICKHOUSE_CONTAINER} \
		-p ${CLICKHOUSE_PORT}:8123 \
		-p 19000:9000 \
		clickhouse/clickhouse-server:24.3.3-alpine)

	@echo "等待服务启动完成..."
	@sleep 10

	@echo "检查服务状态..."
	@docker exec ${REDIS_CONTAINER} redis-cli -a test123 ping 2>/dev/null || true
	@MYSQL_PWD=${MYSQL_ROOT_PASSWORD} docker exec ${MYSQL_CONTAINER} mysqladmin ping -h localhost -u root --silent 2>/dev/null || true
	@curl -s http://localhost:${CLICKHOUSE_PORT}/ping | grep -q "Ok" && echo "ClickHouse: 就绪" || echo "ClickHouse: 启动中..."

	@echo "测试环境启动完成!"

test-teardown: ## 停止并删除测试服务
	@echo "清理测试环境..."
	@docker stop ${REDIS_CONTAINER} ${MYSQL_CONTAINER} ${CLICKHOUSE_CONTAINER} 2>/dev/null || true
	@docker rm ${REDIS_CONTAINER} ${MYSQL_CONTAINER} ${CLICKHOUSE_CONTAINER} 2>/dev/null || true
	@echo "测试环境已清理"

clean-test: test-teardown ## 别名: 清理测试环境 (同 test-teardown)

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
	 go test -v -timeout 30s ./... || (echo "测试失败!" && exit 1)

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

clean-files: ## 清理所有生成的文件 (覆盖率报告、日志等)
	@echo "🧹 清理生成的文件..."
	@echo "  清理覆盖率报告..."
	@find . -name "*.out" -type f -exec rm -f {} \; 2>/dev/null || true
	@find . -name "coverage*.html" -type f -exec rm -f {} \; 2>/dev/null || true
	@echo "  清理日志文件..."
	@find . -name "*.log" -type f -exec rm -f {} \; 2>/dev/null || true
	@find . -name "*.tmp" -type f -exec rm -f {} \; 2>/dev/null || true
	@find . -name "*.temp" -type f -exec rm -f {} \; 2>/dev/null || true
	@echo "  清理临时文件..."
	@rm -rf .tmp/ tmp/ temp/ 2>/dev/null || true
	@rm -f nohup.out 2>/dev/null || true
	@echo "  文件清理完成 ✅"

clean-cache: ## 清理Go测试缓存和模块缓存
	@echo "🧹 清理Go缓存..."
	@echo "  清理测试缓存..."
	@go clean -testcache
	@echo "  清理构建缓存..."
	@go clean -cache
	@echo "  清理模块缓存..."
	@go clean -modcache -x 2>/dev/null || echo "  (需要管理员权限跳过模块缓存清理)"
	@echo "  缓存清理完成 ✅"

clean-docker: ## 强制清理所有测试相关Docker资源
	@echo "🧹 清理Docker测试资源..."
	@echo "  停止测试容器..."
	@docker stop ${REDIS_CONTAINER} ${MYSQL_CONTAINER} ${CLICKHOUSE_CONTAINER} 2>/dev/null || true
	@echo "  删除测试容器..."
	@docker rm ${REDIS_CONTAINER} ${MYSQL_CONTAINER} ${CLICKHOUSE_CONTAINER} 2>/dev/null || true
	@echo "  检查并清理孤立容器..."
	@docker ps -a --filter "name=lrpc-test-" --format "{{.Names}}" | xargs -r docker rm -f 2>/dev/null || true
	@echo "  清理未使用的测试镜像..."
	@docker images --filter "dangling=true" -q | xargs -r docker rmi 2>/dev/null || true
	@echo "  清理Docker网络 (如果存在)..."
	@docker network ls --filter "name=lrpc" -q | xargs -r docker network rm 2>/dev/null || true
	@echo "  清理Docker卷 (如果存在)..."
	@docker volume ls --filter "name=lrpc" -q | xargs -r docker volume rm 2>/dev/null || true
	@echo "  Docker资源清理完成 ✅"

clean: clean-files ## 标准清理 (文件 + 测试容器)
	@echo "🧹 执行标准清理..."
	@$(MAKE) clean-docker --no-print-directory
	@echo "标准清理完成 ✅"

clean-all: clean-files clean-cache clean-docker ## 完全清理 (所有文件、缓存、Docker资源)
	@echo "🧹 执行完全清理..."
	@echo "  检查端口占用..."
	@lsof -ti:${REDIS_PORT} | xargs -r kill -9 2>/dev/null || true
	@lsof -ti:${MYSQL_PORT} | xargs -r kill -9 2>/dev/null || true  
	@lsof -ti:${CLICKHOUSE_PORT} | xargs -r kill -9 2>/dev/null || true
	@echo "  清理Git临时文件..."
	@git clean -fd 2>/dev/null || true
	@echo "完全清理完成 ✅"

clean-check: ## 检查清理效果 (显示剩余的测试相关资源)
	@echo "🔍 检查清理效果..."
	@echo ""
	@echo "📁 文件检查:"
	@echo "  覆盖率文件:" && find . -name "*.out" -o -name "coverage*.html" | head -5 || echo "    ✅ 无覆盖率文件"
	@echo "  日志文件:" && find . -name "*.log" | head -5 || echo "    ✅ 无日志文件"  
	@echo "  临时文件:" && find . -name "*.tmp" -o -name "*.temp" | head -5 || echo "    ✅ 无临时文件"
	@echo ""
	@echo "🐳 Docker检查:"
	@echo "  测试容器:" && docker ps -a --filter "name=lrpc-test-" --format "table {{.Names}}\t{{.Status}}" || echo "    ✅ 无测试容器"
	@echo "  孤立镜像:" && docker images --filter "dangling=true" --format "table {{.Repository}}\t{{.Tag}}\t{{.ID}}" | head -3 || echo "    ✅ 无孤立镜像"
	@echo ""
	@echo "🌐 端口检查:"
	@echo "  Redis端口 ${REDIS_PORT}:" && lsof -ti:${REDIS_PORT} | head -1 && echo "    ⚠️  端口被占用" || echo "    ✅ 端口空闲"
	@echo "  MySQL端口 ${MYSQL_PORT}:" && lsof -ti:${MYSQL_PORT} | head -1 && echo "    ⚠️  端口被占用" || echo "    ✅ 端口空闲"
	@echo "  ClickHouse端口 ${CLICKHOUSE_PORT}:" && lsof -ti:${CLICKHOUSE_PORT} | head -1 && echo "    ⚠️  端口被占用" || echo "    ✅ 端口空闲"
	@echo ""
	@echo "检查完成!"

clean-safe: ## 安全清理 (会先询问确认)
	@echo "🔐 安全清理模式"
	@echo ""
	@echo "将要清理以下内容:"
	@echo "  📁 所有生成的文件 (*.out, *.html, *.log, *.tmp)"
	@echo "  🐳 所有测试Docker容器和相关资源"  
	@echo "  🌐 占用的测试端口 (${REDIS_PORT}, ${MYSQL_PORT}, ${CLICKHOUSE_PORT})"
	@echo ""
	@read -p "确认要继续吗? (y/N): " -n 1 -r; echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "开始清理..."; \
		$(MAKE) clean-all --no-print-directory; \
	else \
		echo "已取消清理操作"; \
	fi

clean-emergency: ## 紧急清理 (强制清理所有资源，用于解决资源占用问题)
	@echo "🚨 紧急清理模式 - 强制清理所有资源"
	@echo "  强制停止所有相关进程..."
	@pkill -f "redis-server.*${REDIS_PORT}" 2>/dev/null || true
	@pkill -f "mysqld.*${MYSQL_PORT}" 2>/dev/null || true
	@pkill -f "clickhouse.*${CLICKHOUSE_PORT}" 2>/dev/null || true
	@echo "  强制清理端口占用..."
	@lsof -ti:${REDIS_PORT} | xargs -r kill -9 2>/dev/null || true
	@lsof -ti:${MYSQL_PORT} | xargs -r kill -9 2>/dev/null || true
	@lsof -ti:${CLICKHOUSE_PORT} | xargs -r kill -9 2>/dev/null || true
	@echo "  强制清理Docker资源..."
	@docker kill $$(docker ps -q --filter "name=lrpc-test-") 2>/dev/null || true
	@docker rm -f $$(docker ps -aq --filter "name=lrpc-test-") 2>/dev/null || true
	@docker system prune -f --filter "label=lrpc-test" 2>/dev/null || true
	@echo "  清理文件系统..."
	@$(MAKE) clean-files --no-print-directory
	@echo "🚨 紧急清理完成"

reset: clean-all deps ## 完全重置环境 (清理 + 重新下载依赖)
	@echo "🔄 重置开发环境完成"

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
