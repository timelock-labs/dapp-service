# TimeLocker 备份文件目录

此目录用于存储 TimeLocker 数据库备份文件。

## 文件命名规则

### 手动备份
- 默认格式：`timelocker_backup_YYYYMMDD_HHMMSS.json`
- 自定义格式：用户指定的文件名

### 自动备份
- 格式：`timelocker_auto_YYYYMMDD_HHMMSS.json`
- 可通过 `BACKUP_PREFIX` 环境变量自定义前缀

## 文件示例

```
backups/
├── README.md                           # 本文件
├── backup.log                          # 备份操作日志
├── timelocker_backup_20241220_143000.json    # 手动备份
├── timelocker_auto_20241220_020000.json      # 自动备份
├── timelocker_auto_20241221_020000.json      # 自动备份
└── my_custom_backup.json               # 自定义名称备份
```

## 文件管理

### 自动清理
自动备份脚本会根据配置自动清理旧文件：
- 默认保留 30 天内的备份
- 最多保留 50 个备份文件
- 可通过环境变量调整

### 手动清理
```bash
# 删除 30 天前的备份文件
find . -name "timelocker_*.json" -mtime +30 -delete

# 只保留最新的 10 个备份
ls -t timelocker_*.json | tail -n +11 | xargs rm -f
```

## 备份文件安全

### 权限设置
```bash
# 设置备份目录权限
chmod 750 backups/
chmod 640 backups/*.json
```

### 加密备份
```bash
# 加密备份文件
gpg --symmetric --cipher-algo AES256 backup.json

# 解密备份文件
gpg --decrypt backup.json.gpg > backup.json
```

## 存储建议

1. **本地存储**：确保有足够的磁盘空间
2. **远程备份**：定期上传到云存储或远程服务器
3. **多地备份**：重要数据建议多地存储
4. **定期测试**：定期测试备份文件的可用性

## 注意事项

⚠️ **重要提醒**：
- 备份文件包含敏感数据，请妥善保管
- 不要将备份文件提交到版本控制系统
- 定期验证备份文件的完整性
- 在生产环境恢复前，请先在测试环境验证
