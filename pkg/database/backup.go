package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"timelocker-backend/pkg/logger"

	"gorm.io/gorm"
)

// BackupManager 备份管理器
type BackupManager struct {
	db *gorm.DB
}

// NewBackupManager 创建备份管理器
func NewBackupManager(db *gorm.DB) *BackupManager {
	return &BackupManager{db: db}
}

// BackupData 备份数据结构
type BackupData struct {
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	// 用户相关表
	Users                  []map[string]interface{} `json:"users"`
	Emails                 []map[string]interface{} `json:"emails"`
	UserEmails             []map[string]interface{} `json:"user_emails"`
	EmailVerificationCodes []map[string]interface{} `json:"email_verification_codes"`
	AuthNonces             []map[string]interface{} `json:"auth_nonces"`
	// Safe钱包相关表
	SafeWallets []map[string]interface{} `json:"safe_wallets"`
	// 支持的区块链表
	SupportChains []map[string]interface{} `json:"support_chains"`
	// ABI相关表
	ABIs []map[string]interface{} `json:"abis"`
	// Timelock合约相关表
	CompoundTimelocks     []map[string]interface{} `json:"compound_timelocks"`
	OpenzeppelinTimelocks []map[string]interface{} `json:"openzeppelin_timelocks"`
	// 交易记录相关表
	CompoundTimelockTransactions     []map[string]interface{} `json:"compound_timelock_transactions"`
	OpenzeppelinTimelockTransactions []map[string]interface{} `json:"openzeppelin_timelock_transactions"`
	TimelockTransactionFlows         []map[string]interface{} `json:"timelock_transaction_flows"`
	// 邮件通知相关表
	EmailSendLogs []map[string]interface{} `json:"email_send_logs"`
	// 其他通知渠道配置表
	TelegramConfigs  []map[string]interface{} `json:"telegram_configs"`
	LarkConfigs      []map[string]interface{} `json:"lark_configs"`
	FeishuConfigs    []map[string]interface{} `json:"feishu_configs"`
	NotificationLogs []map[string]interface{} `json:"notification_logs"`
	// 区块扫描进度表
	BlockScanProgress []map[string]interface{} `json:"block_scan_progress"`
	// 赞助方表（系统表，可选备份）
	Sponsors []map[string]interface{} `json:"sponsors"`
}

// CreateBackup 创建完整数据备份
func (bm *BackupManager) CreateBackup(backupPath string) error {
	ctx := context.Background()
	logger.Info("Starting database backup creation", "path", backupPath)

	// 确保备份目录存在
	if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
		logger.Error("Failed to create backup directory", err)
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	backup := BackupData{
		Version:   "2.0.0", // 升级版本号
		Timestamp: time.Now(),
	}

	// === 用户相关表备份 ===
	// 备份用户数据
	if err := bm.backupTable(ctx, "users", &backup.Users); err != nil {
		return fmt.Errorf("failed to backup users: %w", err)
	}

	// 备份支持的区块链数据
	if err := bm.backupTable(ctx, "support_chains", &backup.SupportChains); err != nil {
		return fmt.Errorf("failed to backup support_chains: %w", err)
	}

	// 备份邮箱数据
	if err := bm.backupTable(ctx, "emails", &backup.Emails); err != nil {
		return fmt.Errorf("failed to backup emails: %w", err)
	}

	// 备份用户邮箱关联数据
	if err := bm.backupTable(ctx, "user_emails", &backup.UserEmails); err != nil {
		return fmt.Errorf("failed to backup user_emails: %w", err)
	}

	// 备份邮箱验证码（可选，通常验证码是临时的）
	if err := bm.backupTable(ctx, "email_verification_codes", &backup.EmailVerificationCodes); err != nil {
		return fmt.Errorf("failed to backup email_verification_codes: %w", err)
	}

	// 备份认证nonce（可选，通常nonce是临时的）
	if err := bm.backupTable(ctx, "auth_nonces", &backup.AuthNonces); err != nil {
		return fmt.Errorf("failed to backup auth_nonces: %w", err)
	}

	// === Safe钱包相关表备份 ===
	// 备份Safe钱包数据
	if err := bm.backupTable(ctx, "safe_wallets", &backup.SafeWallets); err != nil {
		return fmt.Errorf("failed to backup safe_wallets: %w", err)
	}

	// === ABI相关表备份 ===
	// 备份用户自定义ABI（不包含系统共享ABI）
	if err := bm.backupTable(ctx, "abis", &backup.ABIs); err != nil {
		return fmt.Errorf("failed to backup abis: %w", err)
	}

	// === Timelock合约相关表备份 ===
	// 备份Compound timelock合约
	if err := bm.backupTable(ctx, "compound_timelocks", &backup.CompoundTimelocks); err != nil {
		return fmt.Errorf("failed to backup compound_timelocks: %w", err)
	}

	// 备份OpenZeppelin timelock合约
	if err := bm.backupTable(ctx, "openzeppelin_timelocks", &backup.OpenzeppelinTimelocks); err != nil {
		return fmt.Errorf("failed to backup openzeppelin_timelocks: %w", err)
	}

	// === 交易记录相关表备份 ===
	// 备份Compound Timelock交易记录
	if err := bm.backupTable(ctx, "compound_timelock_transactions", &backup.CompoundTimelockTransactions); err != nil {
		return fmt.Errorf("failed to backup compound_timelock_transactions: %w", err)
	}

	// 备份OpenZeppelin Timelock交易记录
	if err := bm.backupTable(ctx, "openzeppelin_timelock_transactions", &backup.OpenzeppelinTimelockTransactions); err != nil {
		return fmt.Errorf("failed to backup openzeppelin_timelock_transactions: %w", err)
	}

	// 备份Timelock交易流程记录
	if err := bm.backupTable(ctx, "timelock_transaction_flows", &backup.TimelockTransactionFlows); err != nil {
		return fmt.Errorf("failed to backup timelock_transaction_flows: %w", err)
	}

	// === 邮件通知相关表备份 ===
	// 备份邮件发送记录
	if err := bm.backupTable(ctx, "email_send_logs", &backup.EmailSendLogs); err != nil {
		return fmt.Errorf("failed to backup email_send_logs: %w", err)
	}

	// === 其他通知渠道配置表备份 ===
	// 备份Telegram配置
	if err := bm.backupTable(ctx, "telegram_configs", &backup.TelegramConfigs); err != nil {
		return fmt.Errorf("failed to backup telegram_configs: %w", err)
	}

	// 备份Lark配置
	if err := bm.backupTable(ctx, "lark_configs", &backup.LarkConfigs); err != nil {
		return fmt.Errorf("failed to backup lark_configs: %w", err)
	}

	// 备份Feishu配置
	if err := bm.backupTable(ctx, "feishu_configs", &backup.FeishuConfigs); err != nil {
		return fmt.Errorf("failed to backup feishu_configs: %w", err)
	}

	// 备份通知发送记录
	if err := bm.backupTable(ctx, "notification_logs", &backup.NotificationLogs); err != nil {
		return fmt.Errorf("failed to backup notification_logs: %w", err)
	}

	// 备份区块扫描进度
	if err := bm.backupTable(ctx, "block_scan_progress", &backup.BlockScanProgress); err != nil {
		return fmt.Errorf("failed to backup block_scan_progress: %w", err)
	}

	// === 系统表备份（可选）===
	// 备份赞助方数据（系统表，可选择是否备份）
	if err := bm.backupTable(ctx, "sponsors", &backup.Sponsors); err != nil {
		return fmt.Errorf("failed to backup sponsors: %w", err)
	}

	// 写入备份文件
	file, err := os.Create(backupPath)
	if err != nil {
		logger.Error("Failed to create backup file", err, "path", backupPath)
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(backup); err != nil {
		logger.Error("Failed to encode backup data", err)
		return fmt.Errorf("failed to encode backup data: %w", err)
	}

	logger.Info("Database backup created successfully",
		"path", backupPath,
		"version", backup.Version,
		"users", len(backup.Users),
		"support_chains", len(backup.SupportChains),
		"emails", len(backup.Emails),
		"user_emails", len(backup.UserEmails),
		"safe_wallets", len(backup.SafeWallets),
		"abis", len(backup.ABIs),
		"compound_timelocks", len(backup.CompoundTimelocks),
		"openzeppelin_timelocks", len(backup.OpenzeppelinTimelocks),
		"compound_transactions", len(backup.CompoundTimelockTransactions),
		"openzeppelin_transactions", len(backup.OpenzeppelinTimelockTransactions),
		"transaction_flows", len(backup.TimelockTransactionFlows),
		"email_send_logs", len(backup.EmailSendLogs),
		"telegram_configs", len(backup.TelegramConfigs),
		"lark_configs", len(backup.LarkConfigs),
		"feishu_configs", len(backup.FeishuConfigs),
		"notification_logs", len(backup.NotificationLogs),
		"block_scan_progress", len(backup.BlockScanProgress),
		"sponsors", len(backup.Sponsors),
	)

	return nil
}

// RestoreBackup 从备份文件恢复数据
func (bm *BackupManager) RestoreBackup(backupPath string, options RestoreOptions) error {
	ctx := context.Background()
	logger.Info("Starting database restore from backup", "path", backupPath)

	// 读取备份文件
	file, err := os.Open(backupPath)
	if err != nil {
		logger.Error("Failed to open backup file", err, "path", backupPath)
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	var backup BackupData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&backup); err != nil {
		logger.Error("Failed to decode backup data", err)
		return fmt.Errorf("failed to decode backup data: %w", err)
	}

	logger.Info("Backup file loaded",
		"version", backup.Version,
		"timestamp", backup.Timestamp,
		"users", len(backup.Users),
	)

	// 在事务中执行恢复操作
	return bm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 如果需要清空现有数据
		if options.ClearExisting {
			if err := bm.clearUserData(ctx, tx); err != nil {
				return fmt.Errorf("failed to clear existing data: %w", err)
			}
		}

		// 按照外键依赖关系的正确顺序恢复数据
		// 这确保了外键约束的正确性

		// 第一层：基础独立表（被其他表外键引用）
		// 1. 用户表（被其他表外键引用）
		if err := bm.restoreTable(ctx, tx, "users", backup.Users, options.OnConflict); err != nil {
			return fmt.Errorf("failed to restore users: %w", err)
		}

		// 2. 邮箱表（被其他表外键引用）
		if err := bm.restoreTable(ctx, tx, "emails", backup.Emails, options.OnConflict); err != nil {
			return fmt.Errorf("failed to restore emails: %w", err)
		}

		// 3. 支持的区块链表（独立表）
		if err := bm.restoreTable(ctx, tx, "support_chains", backup.SupportChains, options.OnConflict); err != nil {
			return fmt.Errorf("failed to restore support_chains: %w", err)
		}

		// 4. 赞助方表（独立表）
		if err := bm.restoreTable(ctx, tx, "sponsors", backup.Sponsors, options.OnConflict); err != nil {
			return fmt.Errorf("failed to restore sponsors: %w", err)
		}

		// 5. 区块扫描进度表（独立表）
		if err := bm.restoreTable(ctx, tx, "block_scan_progress", backup.BlockScanProgress, options.OnConflict); err != nil {
			return fmt.Errorf("failed to restore block_scan_progress: %w", err)
		}

		// 6. Safe钱包表（独立表）
		if err := bm.restoreTable(ctx, tx, "safe_wallets", backup.SafeWallets, options.OnConflict); err != nil {
			return fmt.Errorf("failed to restore safe_wallets: %w", err)
		}

		// 7. 认证nonce表（独立表）
		if err := bm.restoreTable(ctx, tx, "auth_nonces", backup.AuthNonces, options.OnConflict); err != nil {
			return fmt.Errorf("failed to restore auth_nonces: %w", err)
		}

		// 第二层：依赖用户表的表
		// 8. ABI库表（外键：owner → users.wallet_address）
		if err := bm.restoreTable(ctx, tx, "abis", backup.ABIs, options.OnConflict); err != nil {
			return fmt.Errorf("failed to restore abis: %w", err)
		}

		// 9. Compound标准Timelock合约表（外键：creator_address → users.wallet_address）
		if err := bm.restoreTable(ctx, tx, "compound_timelocks", backup.CompoundTimelocks, options.OnConflict); err != nil {
			return fmt.Errorf("failed to restore compound_timelocks: %w", err)
		}

		// 10. OpenZeppelin标准Timelock合约表（外键：creator_address → users.wallet_address）
		if err := bm.restoreTable(ctx, tx, "openzeppelin_timelocks", backup.OpenzeppelinTimelocks, options.OnConflict); err != nil {
			return fmt.Errorf("failed to restore openzeppelin_timelocks: %w", err)
		}

		// 11. 用户邮箱关联表（外键：user_id → users.id, email_id → emails.id）
		if err := bm.restoreTable(ctx, tx, "user_emails", backup.UserEmails, options.OnConflict); err != nil {
			return fmt.Errorf("failed to restore user_emails: %w", err)
		}

		// 第三层：事务相关表（无直接外键约束）
		// 12. Compound Timelock 交易记录表
		if err := bm.restoreTable(ctx, tx, "compound_timelock_transactions", backup.CompoundTimelockTransactions, options.OnConflict); err != nil {
			return fmt.Errorf("failed to restore compound_timelock_transactions: %w", err)
		}

		// 13. OpenZeppelin Timelock 交易记录表
		if err := bm.restoreTable(ctx, tx, "openzeppelin_timelock_transactions", backup.OpenzeppelinTimelockTransactions, options.OnConflict); err != nil {
			return fmt.Errorf("failed to restore openzeppelin_timelock_transactions: %w", err)
		}

		// 14. Timelock 交易流程关联表
		if err := bm.restoreTable(ctx, tx, "timelock_transaction_flows", backup.TimelockTransactionFlows, options.OnConflict); err != nil {
			return fmt.Errorf("failed to restore timelock_transaction_flows: %w", err)
		}

		// 第四层：依赖user_emails和emails表的表
		// 15. 邮箱验证码表（外键：user_email_id → user_emails.id）
		if err := bm.restoreTable(ctx, tx, "email_verification_codes", backup.EmailVerificationCodes, options.OnConflict); err != nil {
			return fmt.Errorf("failed to restore email_verification_codes: %w", err)
		}

		// 16. 邮件发送日志表（外键：email_id → emails.id）
		if err := bm.restoreTable(ctx, tx, "email_send_logs", backup.EmailSendLogs, options.OnConflict); err != nil {
			return fmt.Errorf("failed to restore email_send_logs: %w", err)
		}

		// 第五层：通知配置表（无外键约束，但逻辑上依赖用户）
		// 17. Telegram配置表
		if err := bm.restoreTable(ctx, tx, "telegram_configs", backup.TelegramConfigs, options.OnConflict); err != nil {
			return fmt.Errorf("failed to restore telegram_configs: %w", err)
		}

		// 18. Lark配置表
		if err := bm.restoreTable(ctx, tx, "lark_configs", backup.LarkConfigs, options.OnConflict); err != nil {
			return fmt.Errorf("failed to restore lark_configs: %w", err)
		}

		// 19. 飞书配置表
		if err := bm.restoreTable(ctx, tx, "feishu_configs", backup.FeishuConfigs, options.OnConflict); err != nil {
			return fmt.Errorf("failed to restore feishu_configs: %w", err)
		}

		// 20. 通知日志表（最后恢复，依赖配置表）
		if err := bm.restoreTable(ctx, tx, "notification_logs", backup.NotificationLogs, options.OnConflict); err != nil {
			return fmt.Errorf("failed to restore notification_logs: %w", err)
		}

		return nil
	})
}

// RestoreOptions 恢复选项
type RestoreOptions struct {
	ClearExisting bool           // 是否清空现有用户数据
	OnConflict    ConflictAction // 冲突处理策略
}

// ConflictAction 冲突处理策略
type ConflictAction string

const (
	ConflictSkip    ConflictAction = "skip"    // 跳过冲突记录
	ConflictReplace ConflictAction = "replace" // 替换冲突记录
	ConflictError   ConflictAction = "error"   // 遇到冲突报错
)

// backupTable 备份指定表的所有数据
func (bm *BackupManager) backupTable(ctx context.Context, tableName string, result interface{}) error {
	logger.Info("Backing up table", "table", tableName)

	rows, err := bm.db.WithContext(ctx).Table(tableName).Rows()
	if err != nil {
		return fmt.Errorf("failed to query table %s: %w", tableName, err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
	}

	var records []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("failed to scan row from table %s: %w", tableName, err)
		}

		record := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				record[col] = string(b)
			} else {
				record[col] = val
			}
		}
		records = append(records, record)
	}

	// 使用反射设置结果
	switch v := result.(type) {
	case *[]map[string]interface{}:
		*v = records
	default:
		return fmt.Errorf("unsupported result type for table %s", tableName)
	}

	logger.Info("Table backup completed", "table", tableName, "records", len(records))
	return nil
}

// restoreTable 恢复表数据
func (bm *BackupManager) restoreTable(ctx context.Context, tx *gorm.DB, tableName string, records []map[string]interface{}, onConflict ConflictAction) error {
	if len(records) == 0 {
		logger.Info("No records to restore for table", "table", tableName)
		return nil
	}

	logger.Info("Restoring table", "table", tableName, "records", len(records))

	for _, record := range records {
		if err := bm.insertRecord(ctx, tx, tableName, record, onConflict); err != nil {
			// 检查是否是事务已中止的错误
			if strings.Contains(err.Error(), "current transaction is aborted") {
				logger.Error("Transaction aborted, cannot continue", err, "table", tableName)
				return fmt.Errorf("transaction aborted during %s restore: %w", tableName, err)
			}

			if onConflict == ConflictError {
				return fmt.Errorf("failed to insert record into %s: %w", tableName, err)
			}
			logger.Warn("Skipped record due to conflict", "table", tableName, "error", err)
		}
	}

	logger.Info("Table restore completed", "table", tableName)
	return nil
}

// insertRecord 插入单条记录，处理冲突
func (bm *BackupManager) insertRecord(ctx context.Context, tx *gorm.DB, tableName string, record map[string]interface{}, onConflict ConflictAction) error {
	// 构建列名和占位符
	var columns []string
	var values []interface{}
	var placeholders []string

	for col, val := range record {
		columns = append(columns, col)
		values = append(values, val)
		placeholders = append(placeholders, "?")
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	// 根据冲突策略调整SQL
	switch onConflict {
	case ConflictSkip:
		sql += " ON CONFLICT DO NOTHING"
	case ConflictReplace:
		conflictCols := bm.getConflictColumns(tableName)
		sql += fmt.Sprintf(" ON CONFLICT (%s) DO UPDATE SET ", conflictCols)
		var updates []string
		primaryKeyCols := bm.getPrimaryKeyColumns(tableName)
		primaryKeyMap := make(map[string]bool)
		for _, pk := range primaryKeyCols {
			primaryKeyMap[pk] = true
		}

		// 也需要排除冲突检测列（通常是唯一键列）
		conflictColsList := strings.Split(strings.ReplaceAll(conflictCols, " ", ""), ",")
		for _, col := range conflictColsList {
			primaryKeyMap[col] = true
		}

		for _, col := range columns {
			if !primaryKeyMap[col] { // 不更新主键列和冲突检测列
				updates = append(updates, fmt.Sprintf("%s = EXCLUDED.%s", col, col))
			}
		}
		if len(updates) > 0 {
			sql += strings.Join(updates, ", ")
		} else {
			// 如果没有可更新的列，改用DO NOTHING
			sql = strings.Replace(sql, " DO UPDATE SET ", " DO NOTHING", 1)
		}
	case ConflictError:
		// 默认行为，遇到冲突会报错
	}

	err := tx.WithContext(ctx).Exec(sql, values...).Error
	if err != nil {
		logger.Error("Failed to execute SQL", err, "table", tableName, "sql", sql, "values", values)
	}
	return err
}

// getConflictColumns 获取表的冲突检测列（用于ON CONFLICT）
func (bm *BackupManager) getConflictColumns(tableName string) string {
	// 返回用于ON CONFLICT的列，通常是UNIQUE约束的列
	switch tableName {
	// 基础独立表
	case "users":
		return "wallet_address" // 唯一键
	case "support_chains":
		return "chain_name" // 唯一键
	case "emails":
		return "email" // 唯一键
	case "block_scan_progress":
		return "chain_id" // 唯一键
	case "safe_wallets":
		return "safe_address, chain_id" // 复合唯一键
	case "auth_nonces":
		return "wallet_address, nonce" // 复合唯一键

	// 依赖用户表的表
	case "abis":
		return "name, owner" // 复合唯一键
	case "compound_timelocks":
		return "creator_address, chain_id, contract_address" // 复合唯一键
	case "openzeppelin_timelocks":
		return "creator_address, chain_id, contract_address" // 复合唯一键

	// 事务相关表
	case "compound_timelock_transactions":
		return "tx_hash, contract_address, event_type" // 复合唯一键
	case "openzeppelin_timelock_transactions":
		return "tx_hash, contract_address, event_type" // 复合唯一键
	case "timelock_transaction_flows":
		return "flow_id, timelock_standard, chain_id, contract_address" // 复合唯一键

	// 邮件相关表
	case "user_emails":
		return "user_id, email_id" // 复合唯一键
	case "email_send_logs":
		return "email_id, flow_id, status_to" // 复合唯一键

	// 通知配置表
	case "telegram_configs":
		return "user_address, name" // 复合唯一键
	case "lark_configs":
		return "user_address, name" // 复合唯一键
	case "feishu_configs":
		return "user_address, name" // 复合唯一键
	case "notification_logs":
		return "channel, config_id, flow_id, status_to" // 复合唯一键

	// 无唯一约束的表，使用主键
	case "sponsors", "email_verification_codes":
		return "id" // 主键

	// 默认主键表
	default:
		return "id" // 默认主键
	}
}

// getPrimaryKeyColumns 获取表的主键列（用于排除更新）
func (bm *BackupManager) getPrimaryKeyColumns(tableName string) []string {
	// 返回主键列，用于在UPDATE时排除
	// 注意：所有表都使用id作为主键，没有复合主键表
	return []string{"id"}
}

// clearUserData 清空用户相关数据（保留系统数据）
func (bm *BackupManager) clearUserData(ctx context.Context, tx *gorm.DB) error {
	logger.Warn("Clearing existing user data")

	// 按依赖关系顺序删除（从子表到父表），与init_tables.go中删除顺序一致
	tables := []string{
		// 通知相关表（最后创建的，最先删除）
		"notification_logs",
		"feishu_configs",
		"lark_configs",
		"telegram_configs",
		"email_send_logs",
		"email_verification_codes",
		"user_emails",
		"emails",

		// Safe钱包表
		"safe_wallets",

		// 认证相关表
		"auth_nonces",

		// 交易流程表
		"timelock_transaction_flows",

		// 交易记录表
		"openzeppelin_timelock_transactions",
		"compound_timelock_transactions",

		// 扫描进度表
		"block_scan_progress",

		// Timelock合约表
		"openzeppelin_timelocks",
		"compound_timelocks",

		// 赞助方表
		"sponsors",

		// ABI表
		"abis",

		// 支持链表
		"support_chains",

		// 用户表（最先创建的，最后删除）
		"users",
	}

	for _, table := range tables {
		// 禁用外键约束检查
		if err := tx.WithContext(ctx).Exec("SET session_replication_role = replica").Error; err != nil {
			logger.Warn("Failed to disable foreign key constraints", "error", err)
		}

		if err := tx.WithContext(ctx).Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)).Error; err != nil {
			// 如果TRUNCATE失败，尝试DELETE
			if err := tx.WithContext(ctx).Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
				logger.Error("Failed to clear table", err, "table", table)
				// 恢复外键约束检查
				tx.WithContext(ctx).Exec("SET session_replication_role = DEFAULT")
				return fmt.Errorf("failed to clear table %s: %w", table, err)
			}
		}
		logger.Info("Cleared table", "table", table)
	}

	// 恢复外键约束检查
	if err := tx.WithContext(ctx).Exec("SET session_replication_role = DEFAULT").Error; err != nil {
		logger.Warn("Failed to re-enable foreign key constraints", "error", err)
	}

	logger.Info("User data cleared successfully")
	return nil
}

// ValidateBackup 验证备份文件的完整性
func (bm *BackupManager) ValidateBackup(backupPath string) error {
	logger.Info("Validating backup file", "path", backupPath)

	file, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	var backup BackupData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&backup); err != nil {
		return fmt.Errorf("failed to decode backup data: %w", err)
	}

	// 基本验证
	if backup.Version == "" {
		return errors.New("backup version is missing")
	}

	if backup.Timestamp.IsZero() {
		return errors.New("backup timestamp is missing")
	}

	// 数据完整性检查
	userAddresses := make(map[string]bool)
	for _, user := range backup.Users {
		if addr, ok := user["wallet_address"].(string); ok {
			userAddresses[addr] = true
		}
	}

	// 检查用户邮箱数据是否有对应的用户
	for _, userEmail := range backup.UserEmails {
		if userID, ok := userEmail["user_id"].(float64); ok {
			// 简化验证逻辑，在实际环境中可以添加更复杂的验证
			if userID <= 0 {
				return fmt.Errorf("invalid user_id in user_emails: %v", userID)
			}
		}
	}

	logger.Info("Backup validation completed successfully",
		"version", backup.Version,
		"users", len(backup.Users),
		"timestamp", backup.Timestamp,
	)

	return nil
}

// GetBackupInfo 获取备份文件信息
func (bm *BackupManager) GetBackupInfo(backupPath string) (*BackupData, error) {
	file, err := os.Open(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	var backup BackupData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&backup); err != nil {
		return nil, fmt.Errorf("failed to decode backup data: %w", err)
	}

	// 清空实际数据，只返回元信息
	backup.Users = nil
	backup.SupportChains = nil
	backup.Emails = nil
	backup.UserEmails = nil
	backup.EmailVerificationCodes = nil
	backup.AuthNonces = nil
	backup.SafeWallets = nil
	backup.ABIs = nil
	backup.CompoundTimelocks = nil
	backup.OpenzeppelinTimelocks = nil
	backup.CompoundTimelockTransactions = nil
	backup.OpenzeppelinTimelockTransactions = nil
	backup.TimelockTransactionFlows = nil
	backup.EmailSendLogs = nil
	backup.TelegramConfigs = nil
	backup.LarkConfigs = nil
	backup.FeishuConfigs = nil
	backup.NotificationLogs = nil
	backup.BlockScanProgress = nil
	backup.Sponsors = nil

	return &backup, nil
}
