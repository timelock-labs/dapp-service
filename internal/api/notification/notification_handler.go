package notification

import (
	"net/http"
	"strings"
	"timelocker-backend/internal/middleware"
	"timelocker-backend/internal/service/auth"
	"timelocker-backend/internal/service/notification"
	"timelocker-backend/internal/types"
	"timelocker-backend/pkg/logger"

	"github.com/gin-gonic/gin"
)

// NotificationHandler 通知API处理器
type NotificationHandler struct {
	notificationService notification.NotificationService
	authService         auth.Service
}

// NewNotificationHandler 创建通知处理器实例
func NewNotificationHandler(notificationService notification.NotificationService, authService auth.Service) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		authService:         authService,
	}
}

// RegisterRoutes 注册通知相关路由
func (h *NotificationHandler) RegisterRoutes(router *gin.RouterGroup) {
	// 通知API组 - 需要认证
	notificationGroup := router.Group("/notifications", middleware.AuthMiddleware(h.authService))
	{
		// 获取所有通知配置
		// POST /api/v1/notifications/configs
		// http://localhost:8080/api/v1/notifications/configs
		notificationGroup.POST("/configs", h.GetAllNotificationConfigs)

		// 创建通知配置
		// POST /api/v1/notifications/create
		// http://localhost:8080/api/v1/notifications/create
		notificationGroup.POST("/create", h.CreateNotificationConfig)

		// 更新通知配置
		// POST /api/v1/notifications/update
		// http://localhost:8080/api/v1/notifications/update
		notificationGroup.POST("/update", h.UpdateNotificationConfig)

		// 删除通知配置
		// POST /api/v1/notifications/delete
		// http://localhost:8080/api/v1/notifications/delete
		notificationGroup.POST("/delete", h.DeleteNotificationConfig)
	}
}

// ===== 通用API =====

// GetAllNotificationConfigs 获取所有通知配置
// @Summary 获取所有通知配置
// @Description 获取当前用户的所有通知渠道配置
// @Tags Notification
// @Accept json
// @Produce json
// @Success 200 {object} types.APIResponse{data=types.NotificationConfigListResponse} "获取成功"
// @Failure 401 {object} types.APIResponse{error=types.APIError} "未认证"
// @Failure 500 {object} types.APIResponse{error=types.APIError} "服务器内部错误"
// @Router /api/v1/notifications/configs [post]
func (h *NotificationHandler) GetAllNotificationConfigs(c *gin.Context) {
	// 从上下文获取用户信息
	_, userAddress, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, types.APIResponse{
			Success: false,
			Error: &types.APIError{
				Code:    "UNAUTHORIZED",
				Message: "User not authenticated",
			},
		})
		logger.Error("GetAllNotificationConfigs error", nil, "message", "user not authenticated")
		return
	}

	// 调用service层
	response, err := h.notificationService.GetAllNotificationConfigs(c.Request.Context(), userAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.APIResponse{
			Success: false,
			Error: &types.APIError{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})
		logger.Error("GetAllNotificationConfigs error", err, "user_address", userAddress)
		return
	}

	logger.Info("GetAllNotificationConfigs success", "user_address", userAddress)
	c.JSON(http.StatusOK, types.APIResponse{
		Success: true,
		Data:    response,
	})
}

// CreateNotificationConfig 创建通知配置
// @Summary 创建通知配置
// @Description 为当前用户创建新的通知配置, 名字的空格会被自动去除, 防止攻击者通过空格来绕过名称验证
// @Tags Notification
// @Accept json
// @Produce json
// @Param request body types.CreateNotificationRequest true "创建请求"
// @Success 200 {object} types.APIResponse{data=types.NotificationConfig} "创建成功"
// @Failure 400 {object} types.APIResponse{error=types.APIError} "请求参数错误"
// @Failure 401 {object} types.APIResponse{error=types.APIError} "未认证"
// @Failure 409 {object} types.APIResponse{error=types.APIError} "配置名称已存在"
// @Failure 500 {object} types.APIResponse{error=types.APIError} "服务器内部错误"
// @Router /api/v1/notifications/create [post]
func (h *NotificationHandler) CreateNotificationConfig(c *gin.Context) {
	// 从上下文获取用户信息
	_, userAddress, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, types.APIResponse{
			Success: false,
			Error: &types.APIError{
				Code:    "UNAUTHORIZED",
				Message: "User not authenticated",
			},
		})
		logger.Error("CreateNotificationConfig error", nil, "message", "user not authenticated")
		return
	}

	var req types.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Success: false,
			Error: &types.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request parameters",
				Details: err.Error(),
			},
		})
		logger.Error("CreateNotificationConfig error", err, "message", "invalid request parameters", "user_address", userAddress)
		return
	}

	// 标准化名称
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Success: false,
			Error: &types.APIError{
				Code:    "INVALID_NAME",
				Message: "Name cannot be empty",
			},
		})
		return
	}

	// 调用service层
	err := h.notificationService.CreateNotificationConfig(c.Request.Context(), userAddress, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.APIResponse{
			Success: false,
			Error: &types.APIError{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})
		logger.Error("CreateNotificationConfig error", err, "user_address", userAddress, "name", req.Name)
		return
	}

	logger.Info("CreateNotificationConfig success", "user_address", userAddress, "name", req.Name)
	c.JSON(http.StatusOK, types.APIResponse{
		Success: true,
		Data:    gin.H{"message": "Notification config created successfully"},
	})
}

// UpdateNotificationConfig 更新通知配置
// @Summary 更新通知配置
// @Description 更新当前用户的通知配置, 如果不需要更新某个字段, 可以不传该字段, 但至少传一个字段
// @Tags Notification
// @Accept json
// @Produce json
// @Param request body types.UpdateNotificationRequest true "更新请求"
// @Success 200 {object} types.APIResponse "更新成功"
// @Failure 400 {object} types.APIResponse{error=types.APIError} "请求参数错误"
// @Failure 401 {object} types.APIResponse{error=types.APIError} "未认证"
// @Failure 404 {object} types.APIResponse{error=types.APIError} "配置不存在"
// @Failure 500 {object} types.APIResponse{error=types.APIError} "服务器内部错误"
// @Router /api/v1/notifications/update [post]
func (h *NotificationHandler) UpdateNotificationConfig(c *gin.Context) {

	// 从上下文获取用户信息
	_, userAddress, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, types.APIResponse{
			Success: false,
			Error: &types.APIError{

				Code:    "UNAUTHORIZED",
				Message: "User not authenticated",
			},
		})
		logger.Error("UpdateNotificationConfig error", nil, "message", "user not authenticated")
		return
	}

	var req types.UpdateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Success: false,
			Error: &types.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request parameters",
				Details: err.Error(),
			},
		})
		logger.Error("UpdateNotificationConfig error", err, "message", "invalid request parameters", "user_address", userAddress)
		return
	}

	// 验证名称
	if req.Name == nil || strings.TrimSpace(*req.Name) == "" {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Success: false,
			Error: &types.APIError{
				Code:    "INVALID_NAME",
				Message: "Name is required",
			},
		})
		return
	}

	// 调用service层
	err := h.notificationService.UpdateNotificationConfig(c.Request.Context(), userAddress, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.APIResponse{
			Success: false,
			Error: &types.APIError{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})
		logger.Error("UpdateNotificationConfig error", err, "user_address", userAddress, "name", *req.Name)
		return
	}

	logger.Info("UpdateNotificationConfig success", "user_address", userAddress, "name", *req.Name)
	c.JSON(http.StatusOK, types.APIResponse{
		Success: true,
		Data:    gin.H{"message": "Notification config updated successfully"},
	})
}

// DeleteNotificationConfig 删除通知配置
// @Summary 删除通知配置
// @Description 删除当前用户的通知配置, 名字的空格会被自动去除, 防止攻击者通过空格来绕过名称验证
// @Tags Notification
// @Accept json
// @Produce json
// @Param request body types.DeleteNotificationRequest true "删除请求"
// @Success 200 {object} types.APIResponse "删除成功"
// @Failure 400 {object} types.APIResponse{error=types.APIError} "请求参数错误"
// @Failure 401 {object} types.APIResponse{error=types.APIError} "未认证"
// @Failure 404 {object} types.APIResponse{error=types.APIError} "配置不存在"
// @Failure 500 {object} types.APIResponse{error=types.APIError} "服务器内部错误"
// @Router /api/v1/notifications/delete [post]
func (h *NotificationHandler) DeleteNotificationConfig(c *gin.Context) {
	// 从上下文获取用户信息
	_, userAddress, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, types.APIResponse{
			Success: false,
			Error: &types.APIError{
				Code:    "UNAUTHORIZED",
				Message: "User not authenticated",
			},
		})
		logger.Error("DeleteNotificationConfig error", nil, "message", "user not authenticated")
		return
	}

	var req types.DeleteNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Success: false,
			Error: &types.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request parameters",
				Details: err.Error(),
			},
		})
		logger.Error("DeleteNotificationConfig error", err, "message", "invalid request parameters", "user_address", userAddress)
		return
	}

	// 标准化名称
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Success: false,
			Error: &types.APIError{
				Code:    "INVALID_NAME",
				Message: "Name cannot be empty",
			},
		})
		return
	}

	// 调用service层
	err := h.notificationService.DeleteNotificationConfig(c.Request.Context(), userAddress, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.APIResponse{
			Success: false,
			Error: &types.APIError{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})
		logger.Error("DeleteNotificationConfig error", err, "user_address", userAddress, "name", req.Name)
		return
	}

	logger.Info("DeleteNotificationConfig success", "user_address", userAddress, "name", req.Name)
	c.JSON(http.StatusOK, types.APIResponse{
		Success: true,
		Data:    gin.H{"message": "Notification config deleted successfully"},
	})
}
