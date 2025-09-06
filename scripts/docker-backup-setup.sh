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
    
    # 检查备份调度器是否已经启用
    if docker-compose ps | grep -q "timelocker-backup-scheduler"; then
        log_success "backup scheduler is already running"
        return 0
    fi
    
    # 重启服务以启用备份调度器
    log "restarting services to enable backup scheduler..."
    docker-compose up -d backup-scheduler
    
    log_success "backup scheduler enabled and started"
    log "backup schedule: daily at 2:00 AM UTC, cleanup weekly at 3:00 AM UTC on Sunday"
}

# 禁用备份调度器
disable_scheduler() {
    log "disabling backup scheduler..."
    
    # 停止备份调度器容器
    if docker-compose ps | grep -q "timelocker-backup-scheduler"; then
        docker-compose stop backup-scheduler
        docker-compose rm -f backup-scheduler
        log_success "backup scheduler stopped and removed"
    else
        log_warning "backup scheduler is not running"
    fi
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
        sleep 30  # 给更多时间启动
        
        # 等待服务健康检查通过
        log "waiting for services to be healthy..."
        local max_attempts=30
        local attempt=0
        while [ $attempt -lt $max_attempts ]; do
            if docker-compose ps | grep -q "timelocker-backend.*Up.*healthy"; then
                break
            fi
            sleep 5
            attempt=$((attempt + 1))
            log "waiting... ($attempt/$max_attempts)"
        done
        
        if [ $attempt -eq $max_attempts ]; then
            log_warning "services may not be fully ready, but continuing with test"
        fi
    fi
    
    # 测试备份命令
    log "testing backup creation..."
    local test_file="test_backup_$(date +%Y%m%d_%H%M%S).json"
    if ./scripts/backup.sh --action backup --file "$test_file" --auto; then
        log_success "backup creation test passed"
        
        # 测试备份验证
        log "testing backup validation..."
        if ./scripts/backup.sh --action validate --file "$test_file" --auto; then
            log_success "backup validation test passed"
        else
            log_warning "backup validation test failed"
        fi
        
        # 测试备份信息
        log "testing backup info..."
        if ./scripts/backup.sh --action info --file "$test_file" --auto; then
            log_success "backup info test passed"
        else
            log_warning "backup info test failed"
        fi
        
        # 清理测试文件
        rm -f "backups/$test_file"
        log_success "test backup file cleaned up"
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
    echo "Backup Schedule:"
    echo "  Daily backup: 2:00 AM UTC"
    echo "  Weekly cleanup: 3:00 AM UTC on Sunday"
    echo ""
    echo "Available commands:"
    echo "  make backup                       # 手动创建备份"
    echo "  make backup-status               # 检查备份系统状态"
    echo "  make restore FILE=backup.json    # 完全恢复数据"
    echo "  make restore-safe FILE=backup.json # 安全恢复数据"
    echo "  make test-backup-restore         # 测试备份恢复流程"
    echo ""
    echo "Direct script commands:"
    echo "  ./scripts/backup.sh --action backup --file mybackup.json"
    echo "  ./scripts/backup.sh --action restore --file mybackup.json"
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
