# 🧹 LRPC 清理工具使用指南

本项目提供了全面的自动清理功能来管理测试环境和中间文件。

## 🚀 快速清理

```bash
# 标准清理 - 清理文件和测试容器
make clean

# 完全清理 - 清理所有内容(文件+缓存+Docker)
make clean-all

# 安全清理 - 会先询问确认
make clean-safe
```

## 📋 清理类别

### 🗂️ 文件清理
```bash
make clean-files    # 清理生成的文件(*.out, *.html, *.log, *.tmp)
```
自动清理：
- 覆盖率报告 (`*.out`, `coverage*.html`)
- 日志文件 (`*.log`)  
- 临时文件 (`*.tmp`, `*.temp`)
- 临时目录 (`tmp/`, `temp/`, `.tmp/`)

### 🐳 Docker资源清理
```bash
make clean-docker   # 强制清理所有Docker测试资源
```
清理内容：
- 测试容器 (lrpc-test-*)
- 孤立镜像和网络
- 相关Docker卷
- 僵尸容器进程

### 🗄️ 缓存清理
```bash
make clean-cache    # 清理Go测试和构建缓存
```
清理内容：
- Go测试缓存 (`go clean -testcache`)
- Go构建缓存 (`go clean -cache`) 
- Go模块缓存 (`go clean -modcache`)

### 🧪 测试环境清理
```bash
make clean-test     # 清理测试环境(同 test-teardown)
make test-teardown  # 停止并删除测试服务
```

## 🛡️ 安全模式

### 交互式安全清理
```bash
make clean-safe
```
- 会列出要清理的内容
- 需要用户确认后才执行
- 适合日常开发使用

### 紧急清理模式
```bash  
make clean-emergency
```
- 强制清理所有资源
- 杀死占用端口的进程
- 用于解决资源占用问题
- 谨慎使用！

## 🔍 检查工具

### 清理状态检查
```bash
make clean-check
```
显示：
- 📁 剩余的生成文件
- 🐳 Docker容器状态  
- 🌐 端口占用情况
- 详细的清理建议

### 完全重置环境
```bash
make reset
```
= `clean-all` + `deps`
- 完全清理环境
- 重新下载依赖
- 适合解决环境问题

## 💡 使用建议

### 日常开发
```bash
make clean          # 标准清理，保留缓存
make dev-test       # 快速测试
```

### 彻底清理
```bash
make clean-all      # 完全清理
make test           # 重新测试
```

### 问题排查
```bash
make clean-check    # 检查状态
make clean-emergency # 强制清理(如有必要)
make reset          # 重置环境
```

### CI/CD环境
```bash
make clean-all      # 每次构建前完全清理
make test           # 干净环境测试
```

## ⚠️ 注意事项

1. **`clean-emergency`**: 会强制杀死进程，谨慎使用
2. **`clean-cache`**: 清理模块缓存可能需要管理员权限
3. **端口占用**: 清理后端口可能需要几秒钟才完全释放
4. **Docker权限**: 确保有Docker操作权限

## 🔧 自定义清理

可以组合使用不同的清理命令：

```bash
# 只清理文件，保留Docker环境
make clean-files

# 只重启Docker环境，保留文件  
make clean-docker test-setup

# 检查 -> 清理 -> 验证流程
make clean-check && make clean-all && make clean-check
```

## 📞 获得帮助

```bash
make help           # 查看所有可用命令
make clean-check    # 检查当前状态
```