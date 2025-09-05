#!/bin/bash

# TimeLocker 自动备份脚本
# 作者: Timelocker Protocol
# 版本: 1.0.0
# 描述: 自动执行数据库备份，支持定期清理旧备份

set -e

# 配置变量
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BACKUP_DIR="${PROJECT_ROOT}/backups"
BACKUP_SCRIPT="${SCRIPT_DIR}/backup.sh"

# 备份配置
BACKUP_RETENTION_DAYS=${BACKUP_RETENTION_DAYS:-30}  # 保留备份的天数
BACKUP_PREFIX=${BACKUP_PREFIX:-"timelocker_auto"}    # 备份文件前缀
MAX_BACKUPS=${MAX_BACKUPS:-50}                       # 最大备份文件数

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 日志函数
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')] ✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[$(date '+%Y-%m-%d %H:%M:%S')] ⚠${NC} $1"
}

log_error() {
    echo -e "${RED}[$(date '+%Y-%m-%d %H:%M:%S')] ✗${NC} $1"
}

# 显示帮助信息
show_help() {
    cat << EOF
TimeLocker auto backup script

Usage:
  $0 [Options]

Options:
  -h, --help              show help
  --retention-days DAYS   backup retention days (default: 30)
  --max-backups COUNT     max backup files count (default: 50)
  --prefix PREFIX         backup file prefix (default: timelocker_auto)
  --cleanup-only          only execute cleanup, not create new backup
  --dry-run               simulate run, not actual execute operation

Environment Variables:
  BACKUP_RETENTION_DAYS   backup retention days
  MAX_BACKUPS             max backup files count
  BACKUP_PREFIX           backup file prefix

Examples:
  # execute auto backup
  $0
  
  # custom retention period
  $0 --retention-days 7
  
  # only cleanup old backups
  $0 --cleanup-only
  
  # simulate run
  $0 --dry-run

Cron Job Examples:
  # execute backup every day at 2:00
  0 2 * * * /path/to/auto-backup.sh
  
  # execute backup every 6 hours
  0 */6 * * * /path/to/auto-backup.sh

EOF
}

# 清理旧备份
cleanup_old_backups() {
    local dry_run=$1
    
    log "start cleaning old backup files..."
    log "retention days: $BACKUP_RETENTION_DAYS"
    log "max backup files count: $MAX_BACKUPS"
    log "backup directory: $BACKUP_DIR"
    
    if [ ! -d "$BACKUP_DIR" ]; then
        log_warning "backup directory does not exist: $BACKUP_DIR"
        return 0
    fi
    
    cd "$BACKUP_DIR"
    
    # 按时间清理：删除超过指定天数的备份
    local old_files
    old_files=$(find . -name "${BACKUP_PREFIX}_*.json" -type f -mtime +$BACKUP_RETENTION_DAYS 2>/dev/null || true)
    
    if [ -n "$old_files" ]; then
        log "found $(echo "$old_files" | wc -l) backup files over $BACKUP_RETENTION_DAYS days"
        while IFS= read -r file; do
            if [ "$dry_run" = "true" ]; then
                log "[simulate] delete old backup: $file"
            else
                log "delete old backup: $file"
                rm -f "$file"
            fi
        done <<< "$old_files"
    else
        log "no backup files need to be cleaned up by time"
    fi
    
    # 按数量清理：保留最新的N个备份
    local all_backups
    all_backups=$(ls -t ${BACKUP_PREFIX}_*.json 2>/dev/null || true)
    
    if [ -n "$all_backups" ]; then
        local backup_count
        backup_count=$(echo "$all_backups" | wc -l)
        log "current backup files count: $backup_count"
        
        if [ "$backup_count" -gt "$MAX_BACKUPS" ]; then
            local excess_count=$((backup_count - MAX_BACKUPS))
            log "exceeds the maximum backup count limit, need to delete $excess_count oldest backups"
            
            echo "$all_backups" | tail -n $excess_count | while IFS= read -r file; do
                if [ "$dry_run" = "true" ]; then
                    log "[simulate] delete excess backups: $file"
                else
                    log "delete excess backups: $file"
                    rm -f "$file"
                fi
            done
        else
            log "backup files count within the limit"
        fi
    else
        log "no existing backup files found"
    fi
    
    log_success "backup cleanup completed"
}

# 创建备份
create_backup() {
    local dry_run=$1
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    local backup_filename="${BACKUP_PREFIX}_${timestamp}.json"
    
    log "start creating auto backup..."
    log "backup file name: $backup_filename"
    
    if [ "$dry_run" = "true" ]; then
        log "[simulate] create backup: $backup_filename"
        log_success "[simulate] backup created successfully"
        return 0
    fi
    
    # 检查备份脚本是否存在
    if [ ! -f "$BACKUP_SCRIPT" ]; then
        log_error "backup script does not exist: $BACKUP_SCRIPT"
        return 1
    fi
    
    # 执行备份
    if "$BACKUP_SCRIPT" --action backup --file "$backup_filename"; then
        log_success "auto backup created successfully: $backup_filename"
        
        # 验证备份文件
        if [ -f "$BACKUP_DIR/$backup_filename" ]; then
            local file_size=$(du -h "$BACKUP_DIR/$backup_filename" | cut -f1)
            log_success "backup file size: $file_size"
            
            # 验证备份文件完整性
            if "$BACKUP_SCRIPT" --action validate --file "$backup_filename" >/dev/null 2>&1; then
                log_success "backup file validation passed"
            else
                log_warning "backup file validation failed, but file is created"
            fi
        else
            log_error "backup file not found: $BACKUP_DIR/$backup_filename"
            return 1
        fi
    else
        log_error "auto backup creation failed"
        return 1
    fi
}

# 发送通知（可扩展）
send_notification() {
    local status=$1
    local message=$2
    
    # 这里可以集成邮件、Slack、钉钉等通知方式
    # 例如：
    # curl -X POST -H 'Content-type: application/json' \
    #      --data '{"text":"TimeLocker Backup: '"$message"'"}' \
    #      YOUR_WEBHOOK_URL
    
    log "notification: $status - $message"
}

# 主函数
main() {
    local cleanup_only=false
    local dry_run=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            --retention-days)
                BACKUP_RETENTION_DAYS="$2"
                shift 2
                ;;
            --max-backups)
                MAX_BACKUPS="$2"
                shift 2
                ;;
            --prefix)
                BACKUP_PREFIX="$2"
                shift 2
                ;;
            --cleanup-only)
                cleanup_only=true
                shift
                ;;
            --dry-run)
                dry_run=true
                shift
                ;;
            *)
                log_error "unknown parameter: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    log "TimeLocker auto backup started"
    log "project root directory: $PROJECT_ROOT"
    log "backup directory: $BACKUP_DIR"
    
    if [ "$dry_run" = "true" ]; then
        log_warning "simulate run mode, will not execute actual operation"
    fi
    
    # 创建备份目录
    if [ ! -d "$BACKUP_DIR" ]; then
        if [ "$dry_run" = "true" ]; then
            log "[simulate] create backup directory: $BACKUP_DIR"
        else
            log "create backup directory: $BACKUP_DIR"
            mkdir -p "$BACKUP_DIR"
        fi
    fi
    
    local exit_code=0
    
    # 执行备份（除非仅清理模式）
    if [ "$cleanup_only" = "false" ]; then
        if create_backup "$dry_run"; then
            send_notification "SUCCESS" "auto backup created successfully"
        else
            send_notification "FAILED" "auto backup created failed"
            exit_code=1
        fi
    fi
    
    # 清理旧备份
    cleanup_old_backups "$dry_run"
    
    if [ $exit_code -eq 0 ]; then
        log_success "auto backup task completed"
    else
        log_error "auto backup task failed"
    fi
    
    exit $exit_code
}

# 捕获信号，优雅退出
trap 'log_error "auto backup interrupted"; exit 130' INT TERM

# 执行主函数
main "$@"
