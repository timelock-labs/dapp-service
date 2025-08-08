package email

import (
	"context"
	"fmt"
	"time"
	"timelocker-backend/internal/types"

	"gorm.io/gorm"
)

// EmailRepository 邮箱仓储接口
type EmailRepository interface {
	// Email 相关
	GetOrCreateEmail(ctx context.Context, email string) (*types.Email, error)
	GetEmailByAddress(ctx context.Context, email string) (*types.Email, error)
	GetEmailByID(ctx context.Context, emailID int64) (*types.Email, error)

	// UserEmail 相关
	AddUserEmail(ctx context.Context, userID int64, emailID int64, remark *string) (*types.UserEmail, error)
	GetUserEmails(ctx context.Context, userID int64, offset, limit int) ([]types.UserEmail, int64, error)
	GetUserEmailByID(ctx context.Context, userEmailID int64, userID int64) (*types.UserEmail, error)
	UpdateUserEmailRemark(ctx context.Context, userEmailID int64, userID int64, remark *string) error
	DeleteUserEmail(ctx context.Context, userEmailID int64, userID int64) error
	VerifyUserEmail(ctx context.Context, userEmailID int64, userID int64) error
	CheckUserEmailExists(ctx context.Context, userID int64, emailID int64) (bool, error)

	// EmailVerificationCode 相关
	CreateVerificationCode(ctx context.Context, userEmailID int64, code string, expiresAt time.Time) error
	GetLatestVerificationCode(ctx context.Context, userEmailID int64) (*types.EmailVerificationCode, error)
	VerifyCode(ctx context.Context, userEmailID int64, code string) error
	CleanExpiredCodes(ctx context.Context) error

	// UserEmailSubscription 相关
	CreateSubscription(ctx context.Context, subscription *types.UserEmailSubscription) error
	GetUserSubscriptions(ctx context.Context, userID int64, offset, limit int) ([]types.SubscriptionResponse, int64, error)
	GetSubscriptionByID(ctx context.Context, subscriptionID int64, userID int64) (*types.UserEmailSubscription, error)
	UpdateSubscription(ctx context.Context, subscriptionID int64, userID int64, notifyOn []string, isActive *bool) error
	DeleteSubscription(ctx context.Context, subscriptionID int64, userID int64) error
	CheckSubscriptionExists(ctx context.Context, userEmailID int64, standard string, chainID int, contractAddress string) (bool, error)

	// 通知查询相关
	GetSubscribedEmails(ctx context.Context, standard string, chainID int, contractAddress string, statusTo string, initiatorAddress string) ([]int64, error)

	// EmailSendLog 相关
	CreateSendLog(ctx context.Context, log *types.EmailSendLog) error
	CheckSendLogExists(ctx context.Context, emailID int64, flowID string, statusTo string) (bool, error)
}

// emailRepository 邮箱仓储实现
type emailRepository struct {
	db *gorm.DB
}

// NewEmailRepository 创建邮箱仓储实例
func NewEmailRepository(db *gorm.DB) EmailRepository {
	return &emailRepository{db: db}
}

// ===== Email 相关方法 =====
// GetOrCreateEmail 获取或创建邮箱
func (r *emailRepository) GetOrCreateEmail(ctx context.Context, email string) (*types.Email, error) {
	var emailRecord types.Email

	// 先尝试查找
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&emailRecord).Error
	if err == nil {
		return &emailRecord, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to query email: %w", err)
	}

	// 不存在则创建
	emailRecord = types.Email{
		Email:         email,
		IsDeliverable: true,
	}

	if err := r.db.WithContext(ctx).Create(&emailRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create email: %w", err)
	}

	return &emailRecord, nil
}

// GetEmailByAddress 根据邮箱地址获取邮箱
func (r *emailRepository) GetEmailByAddress(ctx context.Context, email string) (*types.Email, error) {
	var emailRecord types.Email
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&emailRecord).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get email: %w", err)
	}
	return &emailRecord, nil
}

// GetEmailByID 根据邮箱ID获取邮箱
func (r *emailRepository) GetEmailByID(ctx context.Context, emailID int64) (*types.Email, error) {
	var emailRecord types.Email
	err := r.db.WithContext(ctx).Where("id = ?", emailID).First(&emailRecord).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get email by ID: %w", err)
	}
	return &emailRecord, nil
}

// ===== UserEmail 相关方法 =====
// AddUserEmail 添加用户邮箱
func (r *emailRepository) AddUserEmail(ctx context.Context, userID int64, emailID int64, remark *string) (*types.UserEmail, error) {
	userEmail := &types.UserEmail{
		UserID:     userID,
		EmailID:    emailID,
		Remark:     remark,
		IsVerified: false,
	}

	if err := r.db.WithContext(ctx).Create(userEmail).Error; err != nil {
		return nil, fmt.Errorf("failed to create user email: %w", err)
	}

	return userEmail, nil
}

// GetUserEmails 获取用户邮箱
func (r *emailRepository) GetUserEmails(ctx context.Context, userID int64, offset, limit int) ([]types.UserEmail, int64, error) {
	var userEmails []types.UserEmail
	var total int64

	// 计算总数
	if err := r.db.WithContext(ctx).Model(&types.UserEmail{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count user emails: %w", err)
	}

	// 查询数据
	query := r.db.WithContext(ctx).Preload("Email").Where("user_id = ?", userID)
	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	}

	if err := query.Find(&userEmails).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get user emails: %w", err)
	}

	return userEmails, total, nil
}

// GetUserEmailByID 根据用户邮箱ID获取用户邮箱
func (r *emailRepository) GetUserEmailByID(ctx context.Context, userEmailID int64, userID int64) (*types.UserEmail, error) {
	var userEmail types.UserEmail
	err := r.db.WithContext(ctx).Preload("Email").
		Where("id = ? AND user_id = ?", userEmailID, userID).
		First(&userEmail).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user email: %w", err)
	}
	return &userEmail, nil
}

// UpdateUserEmailRemark 更新用户邮箱备注
func (r *emailRepository) UpdateUserEmailRemark(ctx context.Context, userEmailID int64, userID int64, remark *string) error {
	result := r.db.WithContext(ctx).Model(&types.UserEmail{}).
		Where("id = ? AND user_id = ?", userEmailID, userID).
		Update("remark", remark)

	if result.Error != nil {
		return fmt.Errorf("failed to update user email remark: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// DeleteUserEmail 删除用户邮箱
func (r *emailRepository) DeleteUserEmail(ctx context.Context, userEmailID int64, userID int64) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", userEmailID, userID).
		Delete(&types.UserEmail{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete user email: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// VerifyUserEmail 验证用户邮箱
func (r *emailRepository) VerifyUserEmail(ctx context.Context, userEmailID int64, userID int64) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&types.UserEmail{}).
		Where("id = ? AND user_id = ?", userEmailID, userID).
		Updates(map[string]interface{}{
			"is_verified":      true,
			"last_verified_at": &now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to verify user email: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// CheckUserEmailExists 检查用户邮箱是否存在
func (r *emailRepository) CheckUserEmailExists(ctx context.Context, userID int64, emailID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&types.UserEmail{}).
		Where("user_id = ? AND email_id = ?", userID, emailID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check user email exists: %w", err)
	}
	return count > 0, nil
}

// ===== EmailVerificationCode 相关方法 =====
// CreateVerificationCode 创建验证码
func (r *emailRepository) CreateVerificationCode(ctx context.Context, userEmailID int64, code string, expiresAt time.Time) error {
	verificationCode := &types.EmailVerificationCode{
		UserEmailID: userEmailID,
		Code:        code,
		ExpiresAt:   expiresAt,
	}

	if err := r.db.WithContext(ctx).Create(verificationCode).Error; err != nil {
		return fmt.Errorf("failed to create verification code: %w", err)
	}

	return nil
}

// GetLatestVerificationCode 获取最新未使用的验证码
func (r *emailRepository) GetLatestVerificationCode(ctx context.Context, userEmailID int64) (*types.EmailVerificationCode, error) {
	var code types.EmailVerificationCode
	err := r.db.WithContext(ctx).
		Where("user_email_id = ? AND is_used = ?", userEmailID, false).
		Order("sent_at DESC").
		First(&code).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get latest verification code: %w", err)
	}
	return &code, nil
}

// VerifyCode 验证验证码
func (r *emailRepository) VerifyCode(ctx context.Context, userEmailID int64, code string) error {
	now := time.Now()

	// 查找有效的验证码
	var verificationCode types.EmailVerificationCode
	err := r.db.WithContext(ctx).
		Where("user_email_id = ? AND code = ? AND is_used = ? AND expires_at > ?",
			userEmailID, code, false, now).
		First(&verificationCode).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("invalid or expired verification code")
		}
		return fmt.Errorf("failed to verify code: %w", err)
	}

	// 标记验证码为已使用
	if err := r.db.WithContext(ctx).Model(&verificationCode).Update("is_used", true).Error; err != nil {
		return fmt.Errorf("failed to mark code as used: %w", err)
	}

	return nil
}

// CleanExpiredCodes 清理过期验证码
func (r *emailRepository) CleanExpiredCodes(ctx context.Context) error {
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("expires_at < ?", now).
		Delete(&types.EmailVerificationCode{}).Error
	if err != nil {
		return fmt.Errorf("failed to clean expired codes: %w", err)
	}
	return nil
}

// ===== UserEmailSubscription 相关方法 =====
// CreateSubscription 创建订阅
func (r *emailRepository) CreateSubscription(ctx context.Context, subscription *types.UserEmailSubscription) error {
	if err := r.db.WithContext(ctx).Create(subscription).Error; err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	return nil
}

// GetUserSubscriptions 获取用户订阅
func (r *emailRepository) GetUserSubscriptions(ctx context.Context, userID int64, offset, limit int) ([]types.SubscriptionResponse, int64, error) {
	var subscriptions []types.SubscriptionResponse
	var total int64

	// 构建查询
	baseQuery := r.db.WithContext(ctx).Table("user_email_subscriptions s").
		Select(`s.id, s.user_email_id, e.email, s.timelock_standard, s.chain_id, 
				sc.chain_name, s.contract_address, s.notify_on, s.is_active, 
				s.created_at, s.updated_at`).
		Joins("JOIN user_emails ue ON ue.id = s.user_email_id").
		Joins("JOIN emails e ON e.id = ue.email_id").
		Joins("LEFT JOIN support_chains sc ON sc.chain_id = s.chain_id").
		Where("ue.user_id = ?", userID)

	// 计算总数
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count user subscriptions: %w", err)
	}

	// 查询数据
	query := baseQuery
	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	}

	if err := query.Scan(&subscriptions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get user subscriptions: %w", err)
	}

	return subscriptions, total, nil
}

// GetSubscriptionByID 根据订阅ID获取订阅
func (r *emailRepository) GetSubscriptionByID(ctx context.Context, subscriptionID int64, userID int64) (*types.UserEmailSubscription, error) {
	var subscription types.UserEmailSubscription
	err := r.db.WithContext(ctx).Preload("UserEmail").
		Joins("JOIN user_emails ue ON ue.id = user_email_subscriptions.user_email_id").
		Where("user_email_subscriptions.id = ? AND ue.user_id = ?", subscriptionID, userID).
		First(&subscription).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return &subscription, nil
}

// UpdateSubscription 更新订阅
func (r *emailRepository) UpdateSubscription(ctx context.Context, subscriptionID int64, userID int64, notifyOn []string, isActive *bool) error {
	updates := map[string]interface{}{
		"notify_on": notifyOn,
	}

	if isActive != nil {
		updates["is_active"] = *isActive
	}

	result := r.db.WithContext(ctx).Table("user_email_subscriptions").
		Joins("JOIN user_emails ue ON ue.id = user_email_subscriptions.user_email_id").
		Where("user_email_subscriptions.id = ? AND ue.user_id = ?", subscriptionID, userID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update subscription: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// DeleteSubscription 删除订阅
func (r *emailRepository) DeleteSubscription(ctx context.Context, subscriptionID int64, userID int64) error {
	result := r.db.WithContext(ctx).
		Joins("JOIN user_emails ue ON ue.id = user_email_subscriptions.user_email_id").
		Where("user_email_subscriptions.id = ? AND ue.user_id = ?", subscriptionID, userID).
		Delete(&types.UserEmailSubscription{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete subscription: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// CheckSubscriptionExists 检查订阅是否存在
func (r *emailRepository) CheckSubscriptionExists(ctx context.Context, userEmailID int64, standard string, chainID int, contractAddress string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&types.UserEmailSubscription{}).
		Where("user_email_id = ? AND timelock_standard = ? AND chain_id = ? AND contract_address = ?",
			userEmailID, standard, chainID, contractAddress).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check subscription exists: %w", err)
	}
	return count > 0, nil
}

// ===== 通知查询相关方法 =====
// GetSubscribedEmails 获取订阅的邮箱
func (r *emailRepository) GetSubscribedEmails(ctx context.Context, standard string, chainID int, contractAddress string, statusTo string, initiatorAddress string) ([]int64, error) {
	var emailIDs []int64

	err := r.db.WithContext(ctx).Table("user_email_subscriptions s").
		Select("DISTINCT e.id").
		Joins("JOIN user_emails ue ON ue.id = s.user_email_id").
		Joins("JOIN emails e ON e.id = ue.email_id").
		Joins("JOIN users u ON u.id = ue.user_id").
		Where(`s.is_active = ? AND ue.is_verified = ? AND 
			   s.timelock_standard = ? AND s.chain_id = ? AND s.contract_address = ? AND 
		s.notify_on::jsonb @> ?::jsonb AND LOWER(u.wallet_address) = LOWER(?)`,
			true, true, standard, chainID, contractAddress, `"`+statusTo+`"`, initiatorAddress).
		Pluck("e.id", &emailIDs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get subscribed emails: %w", err)
	}

	return emailIDs, nil
}

// ===== EmailSendLog 相关方法 =====
// CreateSendLog 创建发送日志
func (r *emailRepository) CreateSendLog(ctx context.Context, log *types.EmailSendLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return fmt.Errorf("failed to create send log: %w", err)
	}
	return nil
}

// CheckSendLogExists 检查发送日志是否存在
func (r *emailRepository) CheckSendLogExists(ctx context.Context, emailID int64, flowID string, statusTo string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&types.EmailSendLog{}).
		Where("email_id = ? AND flow_id = ? AND status_to = ?", emailID, flowID, statusTo).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check send log exists: %w", err)
	}
	return count > 0, nil
}
