#!/bin/bash
# 运行 MongoDB 副本集测试脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "🚀 启动 MongoDB 副本集环境..."
cd "$PROJECT_ROOT"

# 启动 docker-compose
docker-compose -f docker-compose.test.yml up -d
echo "⏳ 等待 MongoDB 副本集初始化..."
sleep 5

# 运行测试
echo "📝 运行测试..."
go test -v -coverprofile=/tmp/coverage_mongo_rs.out ./middleware/storage/mongo/... -timeout 120s || TEST_FAILED=1

# 显示覆盖率
echo ""
echo "📊 覆盖率统计:"
go tool cover -func=/tmp/coverage_mongo_rs.out | tail -5

# 清理
echo ""
echo "🧹 清理 Docker 环境..."
docker-compose -f docker-compose.test.yml down

if [ "$TEST_FAILED" = "1" ]; then
    echo "❌ 测试失败"
    exit 1
else
    echo "✅ 测试完成"
    exit 0
fi
