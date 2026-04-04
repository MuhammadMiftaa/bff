package handler

import (
	"encoding/json"

	logger "refina-web-bff/config/log"
	grpcClient "refina-web-bff/interface/grpc/client"
	"refina-web-bff/interface/grpc/interceptor"
	"refina-web-bff/internal/cache"
	"refina-web-bff/internal/types/dto"
	"refina-web-bff/internal/utils"
	"refina-web-bff/internal/utils/data"

	tpb "github.com/MuhammadMiftaa/Refina-Protobuf/transaction"
	"github.com/gofiber/fiber/v2"
)

// Budget handler error messages
const (
	errBudgetIDRequired = "budget_id is required"
)

type budgetHandler struct {
	transaction grpcClient.TransactionClient
	cache       cache.Cache
}

func NewBudgetHandler(tc grpcClient.TransactionClient, c cache.Cache) *budgetHandler {
	return &budgetHandler{
		transaction: tc,
		cache:       c,
	}
}

// GetBudgets — GET /budgets?period=YYYY-MM
func (h *budgetHandler) GetBudgets(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	period := c.Query("period")
	if period == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "period query parameter is required (format: YYYY-MM)",
		})
	}

	// Try cache first
	cacheKey := cache.BudgetList(userData.ID, period)
	if cached, err := h.cache.Get(c.UserContext(), cacheKey); err == nil && cached != nil {
		logger.Debug(data.LogCacheHit, map[string]any{"service": data.CacheService, "key": cacheKey})
		c.Set(data.ContentTypeHeader, data.ContentTypeJSON)
		return c.Send(cached)
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.transaction.GetBudgets(ctx, period)
	if err != nil {
		logger.Error(data.LogGetBudgetsFailed, map[string]any{
			"service":    data.BudgetService,
			"request_id": requestID,
			"user_id":    userData.ID,
			"period":     period,
			"error":      err.Error(),
		})
		grpcErr := utils.MapGRPCError(err)
		return c.Status(grpcErr.HTTPStatus).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: grpcErr.HTTPStatus,
			Message:    grpcErr.Message,
		})
	}

	resp := dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Budgets retrieved successfully",
		Data:       result.GetBudgets(),
	}

	// Cache the response
	if b, err := json.Marshal(resp); err == nil {
		if err := h.cache.Set(c.UserContext(), cacheKey, b, cache.TTLShort); err != nil {
			logger.Warn(data.LogCacheSetFailed, map[string]any{"service": data.CacheService, "key": cacheKey, "error": err.Error()})
		}
	}

	logger.Info(data.LogGetBudgetsSuccess, map[string]any{
		"service":    data.BudgetService,
		"request_id": requestID,
		"user_id":    userData.ID,
		"period":     period,
		"count":      len(result.GetBudgets()),
	})

	return c.JSON(resp)
}

// CreateBudget — POST /budgets
func (h *budgetHandler) CreateBudget(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	var req dto.CreateBudgetRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "invalid request body",
		})
	}

	// Validate required fields
	if req.Scope == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "scope is required (overall or category)",
		})
	}
	if req.Scope == "category" && req.CategoryID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "category_id is required for category-scoped budgets",
		})
	}
	if req.Period == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "period is required (format: YYYY-MM)",
		})
	}
	if req.MonthlyLimit <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "monthly_limit must be greater than 0",
		})
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	grpcReq := &tpb.CreateBudgetRequest{
		Scope:        req.Scope,
		CategoryId:   req.CategoryID,
		WalletId:     req.WalletID,
		MonthlyLimit: req.MonthlyLimit,
		Period:       req.Period,
	}

	result, err := h.transaction.CreateBudget(ctx, grpcReq)
	if err != nil {
		logger.Error(data.LogCreateBudgetFailed, map[string]any{
			"service":    data.BudgetService,
			"request_id": requestID,
			"user_id":    userData.ID,
			"error":      err.Error(),
		})
		grpcErr := utils.MapGRPCError(err)
		return c.Status(grpcErr.HTTPStatus).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: grpcErr.HTTPStatus,
			Message:    grpcErr.Message,
		})
	}

	// Invalidate cache
	h.invalidateBudgetCache(c, userData.ID)

	logger.Info(data.LogCreateBudgetSuccess, map[string]any{
		"service":    data.BudgetService,
		"request_id": requestID,
		"user_id":    userData.ID,
		"budget_id":  result.GetId(),
	})

	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "Budget created successfully",
		Data:       result,
	})
}

// UpdateBudget — PUT /budgets/:id
func (h *budgetHandler) UpdateBudget(c *fiber.Ctx) error {
	budgetID := c.Params("id")
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	if budgetID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    errBudgetIDRequired,
		})
	}

	var req dto.UpdateBudgetRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "invalid request body",
		})
	}

	if req.MonthlyLimit <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "monthly_limit must be greater than 0",
		})
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	grpcReq := &tpb.UpdateBudgetRequest{
		Id:           budgetID,
		MonthlyLimit: req.MonthlyLimit,
	}

	result, err := h.transaction.UpdateBudget(ctx, grpcReq)
	if err != nil {
		logger.Error(data.LogUpdateBudgetFailed, map[string]any{
			"service":    data.BudgetService,
			"request_id": requestID,
			"user_id":    userData.ID,
			"budget_id":  budgetID,
			"error":      err.Error(),
		})
		grpcErr := utils.MapGRPCError(err)
		return c.Status(grpcErr.HTTPStatus).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: grpcErr.HTTPStatus,
			Message:    grpcErr.Message,
		})
	}

	// Invalidate cache
	h.invalidateBudgetCache(c, userData.ID)

	logger.Info(data.LogUpdateBudgetSuccess, map[string]any{
		"service":    data.BudgetService,
		"request_id": requestID,
		"user_id":    userData.ID,
		"budget_id":  budgetID,
	})

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Budget updated successfully",
		Data:       result,
	})
}

// DeleteBudget — DELETE /budgets/:id
func (h *budgetHandler) DeleteBudget(c *fiber.Ctx) error {
	budgetID := c.Params("id")
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	if budgetID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    errBudgetIDRequired,
		})
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.transaction.DeleteBudget(ctx, budgetID)
	if err != nil {
		logger.Error(data.LogDeleteBudgetFailed, map[string]any{
			"service":    data.BudgetService,
			"request_id": requestID,
			"user_id":    userData.ID,
			"budget_id":  budgetID,
			"error":      err.Error(),
		})
		grpcErr := utils.MapGRPCError(err)
		return c.Status(grpcErr.HTTPStatus).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: grpcErr.HTTPStatus,
			Message:    grpcErr.Message,
		})
	}

	// Invalidate cache
	h.invalidateBudgetCache(c, userData.ID)

	logger.Info(data.LogDeleteBudgetSuccess, map[string]any{
		"service":    data.BudgetService,
		"request_id": requestID,
		"user_id":    userData.ID,
		"budget_id":  budgetID,
	})

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Budget deleted successfully",
		Data:       result,
	})
}

// ResetBudget — POST /budgets/:id/reset
func (h *budgetHandler) ResetBudget(c *fiber.Ctx) error {
	budgetID := c.Params("id")
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	if budgetID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    errBudgetIDRequired,
		})
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.transaction.ResetBudget(ctx, budgetID)
	if err != nil {
		logger.Error(data.LogResetBudgetFailed, map[string]any{
			"service":    data.BudgetService,
			"request_id": requestID,
			"user_id":    userData.ID,
			"budget_id":  budgetID,
			"error":      err.Error(),
		})
		grpcErr := utils.MapGRPCError(err)
		return c.Status(grpcErr.HTTPStatus).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: grpcErr.HTTPStatus,
			Message:    grpcErr.Message,
		})
	}

	// Invalidate cache
	h.invalidateBudgetCache(c, userData.ID)

	logger.Info(data.LogResetBudgetSuccess, map[string]any{
		"service":    data.BudgetService,
		"request_id": requestID,
		"user_id":    userData.ID,
		"budget_id":  budgetID,
	})

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Budget reset successfully",
		Data:       result,
	})
}

// invalidateBudgetCache clears all budget-related cache for a user
func (h *budgetHandler) invalidateBudgetCache(c *fiber.Ctx, userID string) {
	patterns := []string{
		cache.BudgetAllPattern(userID),
		cache.DashboardAllPattern(userID), // Budget changes may affect dashboard summaries
	}

	for _, pattern := range patterns {
		if err := h.cache.DeleteByPattern(c.UserContext(), pattern); err != nil {
			logger.Warn(data.LogCacheInvalidateFail, map[string]any{
				"service": data.CacheService,
				"pattern": pattern,
				"error":   err.Error(),
			})
		}
	}
}
