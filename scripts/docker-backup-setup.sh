#!/bin/bash

# Docker 备份环境设置脚本
# 作者: Timelocker Protocol
# 版本: 1.0.0
# 描述: 快速设置Docker环境的备份功能

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 颜色输出
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')] ✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[$(date '+%Y-%m-%d %H:%M:%S')] ⚠${NC} $1"
}

show_help() {
    cat << EOF
Docker Backup Setup Script

Usage:
  $0 [Options]

Options:
  -h, --help              show help
  --enable-scheduler      enable backup scheduler service
  --disable-scheduler     disable backup scheduler service
  --rebuild               rebuild Docker images
  --test                  test backup functionality

Examples:
  # basic setup
  $0
  
  # enable automatic backup scheduler
  $0 --enable-scheduler
  
  # rebuild and setup
  $0 --rebuild

EOF
}

# 检查Docker环境
check_docker() {
    log "checking Docker environment..."
    
    if ! command -v docker &> /dev/null; then
        echo "Error: Docker is not installed"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        echo "Error: docker-compose is not installed"
        exit 1
    fi
    
    log_success "Docker environment check passed"
}

# 创建必要的目录
create_directories() {
    log "creating necessary directories..."
    
    mkdir -p "$PROJECT_ROOT/backups"
    chmod 755 "$PROJECT_ROOT/backups"
    
    log_success "directories created"
}

# 设置脚本权限
set_permissions() {
    log "setting script permissions..."
    
    chmod +x "$PROJECT_ROOT/scripts/backup.sh"
    chmod +x "$PROJECT_ROOT/scripts/auto-backup.sh"
    
    log_success "script permissions set"
}

# 启用备份调度器
enable_scheduler() {
    log "enabling backup scheduler..."
    
    # 取消注释备份调度器配置
    sed -i.bak 's/^  # backup-scheduler:/  backup-scheduler:/' "$PROJECT_ROOT/docker-compose.yml"
    sed -i.bak 's/^  #   /    /' "$PROJECT_ROOT/docker-compose.yml"
    
    log_success "backup scheduler enabled"
    log_warning "please restart docker-compose to apply changes: docker-compose up -d"
}

# 禁用备份调度器
disable_scheduler() {
    log "disabling backup scheduler..."
    
    # 注释掉备份调度器配置
    sed -i.bak 's/^  backup-scheduler:/  # backup-scheduler:/' "$PROJECT_ROOT/docker-compose.yml"
    sed -i.bak 's/^    /  #   /' "$PROJECT_ROOT/docker-compose.yml"
    
    log_success "backup scheduler disabled"
    log_warning "please restart docker-compose to apply changes: docker-compose up -d"
}

# 重建Docker镜像
rebuild_images() {
    log "rebuilding Docker images..."
    
    cd "$PROJECT_ROOT"
    docker-compose build --no-cache timelocker-backend
    
    log_success "Docker images rebuilt"
}

# 测试备份功能
test_backup() {
    log "testing backup functionality..."
    
    cd "$PROJECT_ROOT"
    
    # 检查服务是否运行
    if ! docker-compose ps | grep -q "timelocker-backend.*Up"; then
        log_warning "starting services..."
        docker-compose up -d
        sleep 10
    fi
    
    # 测试备份命令
    log "testing backup creation..."
    if ./scripts/backup.sh --action backup --file test_backup.json; then
        log_success "backup test passed"
        
        # 清理测试文件
        rm -f backups/test_backup.json
    else
        echo "Error: backup test failed"
        exit 1
    fi
}

# 显示配置信息
show_config() {
    log "current backup configuration:"
    echo ""
    echo "Project Root: $PROJECT_ROOT"
    echo "Backup Directory: $PROJECT_ROOT/backups"
    echo "Scripts Directory: $PROJECT_ROOT/scripts"
    echo ""
    echo "Available commands:"
    echo "  ./scripts/backup.sh --action backup"
    echo "  ./scripts/backup.sh --action restore --file backup.json"
    echo "  ./scripts/auto-backup.sh"
    echo ""
}

# 主函数
main() {
    local enable_scheduler=false
    local disable_scheduler=false
    local rebuild=false
    local test_mode=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            --enable-scheduler)
                enable_scheduler=true
                shift
                ;;
            --disable-scheduler)
                disable_scheduler=true
                shift
                ;;
            --rebuild)
                rebuild=true
                shift
                ;;
            --test)
                test_mode=true
                shift
                ;;
            *)
                echo "Error: unknown parameter: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    log "TimeLocker Docker Backup Setup"
    log "project root directory: $PROJECT_ROOT"
    
    # 检查Docker环境
    check_docker
    
    # 创建目录
    create_directories
    
    # 设置权限
    set_permissions
    
    # 重建镜像（如果需要）
    if [ "$rebuild" = true ]; then
        rebuild_images
    fi
    
    # 启用/禁用调度器
    if [ "$enable_scheduler" = true ]; then
        enable_scheduler
    elif [ "$disable_scheduler" = true ]; then
        disable_scheduler
    fi
    
    # 测试功能
    if [ "$test_mode" = true ]; then
        test_backup
    fi
    
    # 显示配置
    show_config
    
    log_success "Docker backup setup completed"
}

# 执行主函数
main "$@"
