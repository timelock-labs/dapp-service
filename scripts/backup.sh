#!/bin/bash

# TimeLocker Database Backup Script for Docker Environment
# 作者: Timelocker Protocol
# 版本: 1.0.0
# 描述: 在Docker环境中执行数据库备份操作的脚本

set -e  # 遇到错误立即退出

# 配置变量
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BACKUP_DIR="${PROJECT_ROOT}/backups"
CONTAINER_NAME="timelocker-backend"
LOG_FILE="${BACKUP_DIR}/backup.log"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')] ✓${NC} $1" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[$(date '+%Y-%m-%d %H:%M:%S')] ⚠${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[$(date '+%Y-%m-%d %H:%M:%S')] ✗${NC} $1" | tee -a "$LOG_FILE"
}

# 显示帮助信息
show_help() {
    cat << EOF
TimeLocker Database Backup Script

Usage:
  $0 [Options]

Options:
  -h, --help              show help
  -a, --action ACTION     action type (backup|restore|validate|info|reset)
  -f, --file FILE         backup file path
  -c, --clear              clear existing data when restore
  --conflict STRATEGY      conflict resolution strategy (skip|replace|error)
  --auto                   auto mode, use default settings

Examples:
  # create backup
  $0 --action backup
  
  # create backup to specified file
  $0 --action backup --file ./my_backup.json
  
  # restore from backup (skip conflicts)
  $0 --action restore --file ./my_backup.json --conflict skip
  
  # restore from backup (clear existing data)
  $0 --action restore --file ./my_backup.json --clear --conflict replace
  
  # validate backup file
  $0 --action validate --file ./my_backup.json
  
  # view backup file info
  $0 --action info --file ./my_backup.json
  
  # reset database (dangerous operation)
  $0 --action reset

EOF
}

# 检查Docker环境
check_docker_environment() {
    log "check Docker environment..."
    
    # 检查Docker是否安装
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    # 检查docker-compose是否安装
    if ! command -v docker-compose &> /dev/null; then
        log_error "docker-compose is not installed or not in PATH"
        exit 1
    fi
    
    # 检查容器是否运行
    if ! docker ps | grep -q "$CONTAINER_NAME"; then
        log_error "TimeLocker backend container ($CONTAINER_NAME) is not running"
        log "please start the service: cd $PROJECT_ROOT && docker-compose up -d"
        exit 1
    fi
    
    log_success "Docker environment check passed"
}

# 创建备份目录
ensure_backup_directory() {
    if [ ! -d "$BACKUP_DIR" ]; then
        log "create backup directory: $BACKUP_DIR"
        mkdir -p "$BACKUP_DIR"
    fi
}

# 执行备份命令
execute_backup_command() {
    local cmd="$1"
    log "execute backup command: $cmd"
    
    if docker exec "$CONTAINER_NAME" $cmd; then
        log_success "command executed successfully"
        return 0
    else
        log_error "command executed failed"
        return 1
    fi
}

# 主函数
main() {
    local action=""
    local file=""
    local clear_flag=""
    local conflict="skip"
    local auto_mode=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -a|--action)
                action="$2"
                shift 2
                ;;
            -f|--file)
                file="$2"
                shift 2
                ;;
            -c|--clear)
                clear_flag="--clear"
                shift
                ;;
            --conflict)
                conflict="$2"
                shift 2
                ;;
            --auto)
                auto_mode=true
                shift
                ;;
            *)
                log_error "unknown parameter: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 检查必需参数
    if [ -z "$action" ]; then
        log_error "please specify the action type (--action)"
        show_help
        exit 1
    fi
    
    # 初始化
    ensure_backup_directory
    log "start executing backup operation..."
    log "action type: $action"
    log "project root directory: $PROJECT_ROOT"
    log "backup directory: $BACKUP_DIR"
    
    # 检查Docker环境
    check_docker_environment
    
    # 构建备份命令
    local backup_cmd="/app/backup -action=$action"
    
    # 处理文件路径
    if [ -n "$file" ]; then
        # 如果是相对路径，转换为容器内的绝对路径
        if [[ "$file" != /* ]]; then
            file="/app/backups/$(basename "$file")"
        fi
        backup_cmd="$backup_cmd -file=$file"
        log "backup file: $file"
    fi
    
    # 添加其他参数
    if [ -n "$clear_flag" ]; then
        backup_cmd="$backup_cmd $clear_flag"
        log "clear existing data: yes"
    fi
    
    if [ "$action" = "restore" ]; then
        backup_cmd="$backup_cmd -conflict=$conflict"
        log "conflict strategy: $conflict"
    fi
    
    # 执行备份操作
    case $action in
        backup)
            log "start creating data backup..."
            if execute_backup_command "$backup_cmd"; then
                log_success "backup created successfully"
                if [ -n "$file" ]; then
                    # 将备份文件从容器复制到宿主机
                    local host_file="$BACKUP_DIR/$(basename "$file")"
                    if docker cp "$CONTAINER_NAME:$file" "$host_file"; then
                        log_success "backup file saved to: $host_file"
                    else
                        log_warning "cannot copy backup file to host"
                    fi
                fi
            else
                exit 1
            fi
            ;;
        restore)
            if [ -z "$file" ]; then
                log_error "restore operation needs to specify the backup file (--file)"
                exit 1
            fi
            
            # 将备份文件从宿主机复制到容器
            local host_file="$file"
            if [[ "$file" != /* ]]; then
                host_file="$BACKUP_DIR/$file"
            fi
            
            if [ ! -f "$host_file" ]; then
                log_error "backup file does not exist: $host_file"
                exit 1
            fi
            
            log "copy backup file to container..."
            if docker cp "$host_file" "$CONTAINER_NAME:/app/backups/$(basename "$file")"; then
                log_success "backup file copied successfully"
            else
                log_error "backup file copied failed"
                exit 1
            fi
            
            log "start restoring data..."
            if [ "$auto_mode" = false ]; then
                log_warning "this operation will modify the database data"
                read -p "confirm to continue? (y/N): " -n 1 -r
                echo
                if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                    log "operation cancelled"
                    exit 0
                fi
            fi
            
            if execute_backup_command "$backup_cmd"; then
                log_success "data restored successfully"
            else
                exit 1
            fi
            ;;
        validate|info)
            if [ -z "$file" ]; then
                log_error "$action operation needs to specify the backup file (--file)"
                exit 1
            fi
            
            # 将备份文件从宿主机复制到容器（如果需要）
            local host_file="$file"
            if [[ "$file" != /* ]]; then
                host_file="$BACKUP_DIR/$file"
            fi
            
            if [ ! -f "$host_file" ]; then
                log_error "backup file does not exist: $host_file"
                exit 1
            fi
            
            log "copy backup file to container..."
            if docker cp "$host_file" "$CONTAINER_NAME:/app/backups/$(basename "$file")"; then
                log_success "backup file copied successfully"
            else
                log_error "backup file copied failed"
                exit 1
            fi
            
            if execute_backup_command "$backup_cmd"; then
                log_success "$action operation completed"
            else
                exit 1
            fi
            ;;
        reset)
            log_warning "dangerous operation: will delete all database data!"
            if [ "$auto_mode" = false ]; then
                read -p "please enter 'RESET' to confirm deleting all data: " -r
                if [ "$REPLY" != "RESET" ]; then
                    log "operation cancelled"
                    exit 0
                fi
            fi
            
            if execute_backup_command "$backup_cmd"; then
                log_success "database reset successfully"
            else
                exit 1
            fi
            ;;
        *)
            log_error "unsupported operation type: $action"
            exit 1
            ;;
    esac
    
    log_success "all operations completed"
}

# 捕获信号，优雅退出
trap 'log_error "operation interrupted"; exit 130' INT TERM

# 执行主函数
main "$@"
