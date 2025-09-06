# TimeLocker Makefile
# 完整的Docker化部署和运维管理工具

.PHONY: help build up down restart logs backup restore validate info reset setup-backup test-backup clean \
        status health check-env setup-env dev-setup prod-setup \
        db-connect db-size db-backup db-restore \
        update rebuild pull push \
        monitor tail-logs clear-logs \
        backup-auto backup-manual backup-list backup-cleanup \
        security-check permissions lint test \
        install uninstall docs

# 版本和配置
VERSION ?= latest
ENV_FILE ?= .env
COMPOSE_FILE ?= docker-compose.yml
BACKUP_PREFIX ?= timelocker

# 颜色定义
GREEN = \033[0;32m
YELLOW = \033[1;33m
BLUE = \033[0;34m
RED = \033[0;31m
NC = \033[0m # No Color

# 默认目标
help:
	@echo ""
	@echo "$(BLUE)════════════════════════════════════════════════════════════════$(NC)"
	@echo "$(BLUE)                    TimeLocker Docker Management                   $(NC)"
	@echo "$(BLUE)════════════════════════════════════════════════════════════════$(NC)"
	@echo ""
	@echo "$(GREEN)🚀 部署管理$(NC)"
	@echo "  make setup-env        - 设置环境变量文件"
	@echo "  make dev-setup        - 开发环境一键设置"
	@echo "  make prod-setup       - 生产环境一键设置"
	@echo "  make build            - 构建Docker镜像"
	@echo "  make up               - 启动所有服务"
	@echo "  make down             - 停止所有服务"
	@echo "  make restart          - 重启所有服务"
	@echo ""
	@echo "$(GREEN)📊 监控管理$(NC)"
	@echo "  make status           - 显示服务状态"
	@echo "  make health           - 健康检查"
	@echo "  make logs             - 显示所有日志"
	@echo "  make tail-logs        - 实时查看日志"
	@echo "  make monitor          - 系统监控面板"
	@echo ""
	@echo "$(GREEN)💾 备份管理$(NC)"
	@echo "  make backup           - 创建数据库备份"
	@echo "  make backup-auto      - 自动备份"
	@echo "  make backup-manual    - 手动备份(带时间戳)"
	@echo "  make backup-list      - 列出所有备份"
	@echo "  make backup-cleanup   - 清理旧备份"
	@echo "  make backup-status    - 检查备份系统状态"
	@echo "  make restore FILE=    - 从备份完全恢复(删除现有数据)"
	@echo "  make restore-safe FILE= - 安全恢复(跳过冲突)"
	@echo "  make validate FILE=   - 验证备份文件"
	@echo "  make info FILE=       - 显示备份信息"
	@echo ""
	@echo "$(GREEN)🗄️  数据库管理$(NC)"
	@echo "  make db-connect       - 连接数据库"
	@echo "  make db-size          - 查看数据库大小"
	@echo "  make db-backup        - 数据库备份"
	@echo "  make db-restore FILE= - 数据库恢复"
	@echo "  make reset            - 重置数据库(危险)"
	@echo ""
	@echo "$(GREEN)🔧 维护管理$(NC)"
	@echo "  make update           - 更新服务"
	@echo "  make rebuild          - 重新构建"
	@echo "  make clean            - 清理资源"
	@echo "  make clear-logs       - 清理日志文件"
	@echo "  make permissions      - 设置文件权限"
	@echo ""
	@echo "$(GREEN)🛠️  开发工具$(NC)"
	@echo "  make check-env        - 检查环境配置"
	@echo "  make setup-backup     - 设置备份环境"
	@echo "  make test-backup      - 测试备份功能"
	@echo "  make security-check   - 安全检查"
	@echo "  make lint             - 代码检查"
	@echo ""
	@echo "$(YELLOW)📖 使用示例:$(NC)"
	@echo "  make dev-setup                    # 开发环境一键部署"
	@echo "  make backup                       # 创建备份"
	@echo "  make restore FILE=backup.json     # 完全恢复数据(删除现有数据)"
	@echo "  make restore-safe FILE=backup.json # 安全恢复(跳过冲突)"
	@echo "  make monitor                      # 查看系统状态"
	@echo ""

# ================================
# 环境设置和检查
# ================================

check-env:
	@echo "$(BLUE)🔍 检查环境配置...$(NC)"
	@if [ ! -f "$(ENV_FILE)" ]; then \
		echo "$(RED)❌ 环境文件 $(ENV_FILE) 不存在$(NC)"; \
		echo "$(YELLOW)💡 运行 'make setup-env' 创建环境文件$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✅ 环境文件检查通过$(NC)"
	@docker --version
	@docker-compose --version

setup-env:
	@echo "$(BLUE)🔧 设置环境变量文件...$(NC)"
	@if [ ! -f "$(ENV_FILE)" ]; then \
		cp env.example $(ENV_FILE); \
		echo "$(GREEN)✅ 已创建 $(ENV_FILE) 文件$(NC)"; \
		echo "$(YELLOW)⚠️  请编辑 $(ENV_FILE) 文件配置必要的环境变量$(NC)"; \
	else \
		echo "$(YELLOW)⚠️  $(ENV_FILE) 已存在，跳过创建$(NC)"; \
	fi

# ================================
# 一键部署
# ================================

dev-setup: setup-env setup-backup
	@echo "$(BLUE)🚀 开发环境一键设置...$(NC)"
	@make build
	@make up
	@echo "$(GREEN)✅ 开发环境设置完成$(NC)"
	@make status

prod-setup: setup-env setup-backup security-check
	@echo "$(BLUE)🚀 生产环境一键设置...$(NC)"
	@make build
	@make up
	@echo "$(GREEN)✅ 生产环境设置完成$(NC)"
	@make status

# ================================
# Docker 基础操作
# ================================

build:
	@echo "$(BLUE)🔨 构建Docker镜像...$(NC)"
	@docker-compose build --no-cache

up:
	@echo "$(BLUE)🚀 启动所有服务...$(NC)"
	@docker-compose up -d
	@echo "$(GREEN)✅ 服务启动完成$(NC)"

down:
	@echo "$(BLUE)🛑 停止所有服务...$(NC)"
	@docker-compose down
	@echo "$(GREEN)✅ 服务停止完成$(NC)"

restart:
	@echo "$(BLUE)🔄 重启所有服务...$(NC)"
	@make down
	@sleep 3
	@make up

# ================================
# 监控和日志
# ================================

status:
	@echo "$(BLUE)📊 服务状态概览$(NC)"
	@echo "$(YELLOW)═══════════════════════════════════════$(NC)"
	@docker-compose ps
	@echo ""
	@echo "$(BLUE)💾 备份文件状态$(NC)"
	@echo "$(YELLOW)═══════════════════════════════════════$(NC)"
	@if [ -d "backups" ]; then \
		echo "备份目录: $$(ls -la backups/ | wc -l) 个文件"; \
		echo "最新备份: $$(ls -t backups/*.json 2>/dev/null | head -1 | xargs basename 2>/dev/null || echo '无')"; \
		echo "目录大小: $$(du -sh backups/ 2>/dev/null | cut -f1 || echo '0B')"; \
	else \
		echo "$(RED)❌ 备份目录不存在$(NC)"; \
	fi

health:
	@echo "$(BLUE)🏥 健康检查...$(NC)"
	@docker-compose exec timelocker-backend wget --spider -q http://localhost:8080/api/v1/health && \
		echo "$(GREEN)✅ 后端服务健康$(NC)" || \
		echo "$(RED)❌ 后端服务异常$(NC)"
	@docker-compose exec postgres pg_isready -U timelocker -d timelocker_db && \
		echo "$(GREEN)✅ 数据库服务健康$(NC)" || \
		echo "$(RED)❌ 数据库服务异常$(NC)"
	@docker-compose exec redis redis-cli ping | grep -q PONG && \
		echo "$(GREEN)✅ Redis服务健康$(NC)" || \
		echo "$(RED)❌ Redis服务异常$(NC)"

logs:
	@echo "$(BLUE)📋 显示所有服务日志...$(NC)"
	@docker-compose logs --tail=100

tail-logs:
	@echo "$(BLUE)📋 实时查看日志...$(NC)"
	@docker-compose logs -f

monitor:
	@echo "$(BLUE)📊 系统监控面板$(NC)"
	@echo "$(YELLOW)═══════════════════════════════════════$(NC)"
	@make status
	@echo ""
	@echo "$(BLUE)💻 系统资源使用情况$(NC)"
	@echo "$(YELLOW)═══════════════════════════════════════$(NC)"
	@docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}"

# ================================
# 备份管理
# ================================

backup:
	@echo "$(BLUE)💾 创建数据库备份...$(NC)"
	@./scripts/backup.sh --action backup

backup-auto:
	@echo "$(BLUE)🤖 执行自动备份...$(NC)"
	@./scripts/auto-backup.sh

backup-manual:
	@echo "$(BLUE)📝 创建手动备份...$(NC)"
	@./scripts/backup.sh --action backup --file "manual_backup_$$(date +%Y%m%d_%H%M%S).json"

backup-list:
	@echo "$(BLUE)📋 备份文件列表$(NC)"
	@echo "$(YELLOW)═══════════════════════════════════════$(NC)"
	@if [ -d "backups" ]; then \
		ls -lah backups/*.json 2>/dev/null | head -20 || echo "$(YELLOW)📁 暂无备份文件$(NC)"; \
	else \
		echo "$(RED)❌ 备份目录不存在$(NC)"; \
	fi

backup-cleanup:
	@echo "$(BLUE)🧹 清理旧备份文件...$(NC)"
	@./scripts/auto-backup.sh --cleanup-only

restore:
	@if [ -z "$(FILE)" ]; then \
		echo "$(RED)❌ 请指定备份文件: make restore FILE=backup.json$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)🔄 从备份完全恢复(删除现有数据): $(FILE)$(NC)"
	@echo "$(RED)⚠️  警告: 这将删除所有现有数据并从备份恢复$(NC)"
	@./scripts/backup.sh --action restore --file $(FILE)

restore-safe:
	@if [ -z "$(FILE)" ]; then \
		echo "$(RED)❌ 请指定备份文件: make restore-safe FILE=backup.json$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)🔄 安全恢复(跳过冲突): $(FILE)$(NC)"
	@./scripts/backup.sh --action restore --file $(FILE) --conflict skip

validate:
	@if [ -z "$(FILE)" ]; then \
		echo "$(RED)❌ 请指定备份文件: make validate FILE=backup.json$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)✅ 验证备份文件: $(FILE)$(NC)"
	@./scripts/backup.sh --action validate --file $(FILE)

info:
	@if [ -z "$(FILE)" ]; then \
		echo "$(RED)❌ 请指定备份文件: make info FILE=backup.json$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)📄 备份文件信息: $(FILE)$(NC)"
	@./scripts/backup.sh --action info --file $(FILE)

# ================================
# 数据库管理
# ================================

db-connect:
	@echo "$(BLUE)🔗 连接数据库...$(NC)"
	@docker-compose exec postgres psql -U timelocker -d timelocker_db

db-size:
	@echo "$(BLUE)📏 数据库大小信息$(NC)"
	@echo "$(YELLOW)═══════════════════════════════════════$(NC)"
	@docker-compose exec postgres psql -U timelocker -d timelocker_db -c "SELECT pg_size_pretty(pg_database_size('timelocker_db')) as database_size;"

db-backup: backup

db-restore: restore

reset:
	@echo "$(RED)⚠️  警告: 这将删除所有数据库数据!$(NC)"
	@read -p "确认删除所有数据? 输入 'RESET' 继续: " confirm && [ "$$confirm" = "RESET" ]
	@./scripts/backup.sh --action reset

# ================================
# 维护管理
# ================================

update:
	@echo "$(BLUE)🔄 更新服务...$(NC)"
	@docker-compose pull
	@make restart

rebuild:
	@echo "$(BLUE)🔨 重新构建所有服务...$(NC)"
	@docker-compose build --no-cache
	@make restart

clean:
	@echo "$(BLUE)🧹 清理Docker资源...$(NC)"
	@docker system prune -f
	@docker volume prune -f
	@echo "$(GREEN)✅ 清理完成$(NC)"

clear-logs:
	@echo "$(BLUE)🗑️  清理日志文件...$(NC)"
	@docker-compose exec timelocker-backend sh -c "rm -f /var/log/timelocker/*.log" 2>/dev/null || true
	@echo "$(GREEN)✅ 日志清理完成$(NC)"

permissions:
	@echo "$(BLUE)🔐 设置文件权限...$(NC)"
	@chmod +x scripts/*.sh
	@chmod 755 backups/ 2>/dev/null || mkdir -p backups/ && chmod 755 backups/
	@echo "$(GREEN)✅ 权限设置完成$(NC)"

# ================================
# 开发和测试工具
# ================================

setup-backup:
	@echo "$(BLUE)🛠️  设置备份环境...$(NC)"
	@./scripts/docker-backup-setup.sh

test-backup:
	@echo "$(BLUE)🧪 测试备份功能...$(NC)"
	@./scripts/docker-backup-setup.sh --test

security-check:
	@echo "$(BLUE)🔒 安全检查...$(NC)"
	@echo "$(YELLOW)═══════════════════════════════════════$(NC)"
	@if [ -f "$(ENV_FILE)" ]; then \
		echo "✅ 环境文件权限: $$(stat -f %A $(ENV_FILE) 2>/dev/null || stat -c %a $(ENV_FILE) 2>/dev/null || echo 'unknown')"; \
		if grep -q "your_.*_here" $(ENV_FILE); then \
			echo "$(RED)❌ 检测到默认密码，请修改$(NC)"; \
		else \
			echo "$(GREEN)✅ 未检测到默认密码$(NC)"; \
		fi \
	fi
	@echo "✅ Docker容器安全扫描..."
	@docker-compose config --quiet && echo "$(GREEN)✅ Docker配置验证通过$(NC)" || echo "$(RED)❌ Docker配置有误$(NC)"

lint:
	@echo "$(BLUE)🔍 代码检查...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)⚠️  golangci-lint 未安装，跳过代码检查$(NC)"; \
	fi

# ================================
# 文档和帮助
# ================================

docs:
	@echo "$(BLUE)📚 可用文档:$(NC)"
	@echo "  📄 README.md                     - 项目说明"
	@echo "  📄 env.example                   - 环境变量示例"
	@echo "  📄 backups/README.md             - 备份说明"
	@echo "  📄 docker-compose.yml            - Docker配置"
	@echo ""
	@echo "$(BLUE)🔗 有用链接:$(NC)"
	@echo "  🌐 应用地址: http://localhost:8080"
	@echo "  🌐 健康检查: http://localhost:8080/api/v1/health"
	@echo "  📊 Caddy管理: http://localhost:80"

# ================================
# 生产环境特殊操作
# ================================

prod-backup:
	@echo "$(BLUE)🏭 生产环境备份...$(NC)"
	@./scripts/backup.sh --action backup --file "prod_backup_$$(date +%Y%m%d_%H%M%S).json" --auto
	@echo "$(GREEN)✅ 生产备份完成$(NC)"

emergency-backup:
	@echo "$(RED)🚨 紧急备份...$(NC)"
	@./scripts/backup.sh --action backup --file "emergency_backup_$$(date +%Y%m%d_%H%M%S).json" --auto
	@echo "$(GREEN)✅ 紧急备份完成$(NC)"

# 备份状态检查
backup-status:
	@echo "$(BLUE)📊 备份系统状态$(NC)"
	@echo "$(YELLOW)═══════════════════════════════════════$(NC)"
	@if docker ps | grep -q "timelocker-backup-scheduler"; then \
		echo "$(GREEN)✅ 自动备份调度器: 运行中$(NC)"; \
	else \
		echo "$(RED)❌ 自动备份调度器: 未运行$(NC)"; \
	fi
	@if [ -d "backups" ]; then \
		echo "$(GREEN)✅ 备份目录: 存在$(NC)"; \
		echo "备份文件数量: $$(ls -1 backups/*.json 2>/dev/null | wc -l || echo 0)"; \
		echo "最新备份: $$(ls -t backups/*.json 2>/dev/null | head -1 | xargs basename 2>/dev/null || echo '无')"; \
		echo "目录大小: $$(du -sh backups/ 2>/dev/null | cut -f1 || echo '0B')"; \
	else \
		echo "$(RED)❌ 备份目录: 不存在$(NC)"; \
	fi

# ================================
# 快速操作别名
# ================================

start: up
stop: down
ps: status
shell:
	@docker-compose exec timelocker-backend sh

# 显示版本信息
version:
	@echo "TimeLocker Docker Management v1.0.0"
	@echo "Docker: $$(docker --version)"
	@echo "Docker Compose: $$(docker-compose --version)"