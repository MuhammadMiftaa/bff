package handler

import (
	"encoding/json"
	"fmt"

	logger "refina-web-bff/config/log"
	grpcClient "refina-web-bff/interface/grpc/client"
	"refina-web-bff/interface/grpc/interceptor"
	"refina-web-bff/internal/cache"
	"refina-web-bff/internal/types/dto"
	"refina-web-bff/internal/utils"
	"refina-web-bff/internal/utils/data"

	dpb "github.com/MuhammadMiftaa/Refina-Protobuf/dashboard"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/protobuf/proto"
)

type dashboardHandler struct {
	dashboard grpcClient.DashboardClient
	cache     cache.Cache
}

func NewDashboardHandler(dc grpcClient.DashboardClient, c cache.Cache) *dashboardHandler {
	return &dashboardHandler{
		dashboard: dc,
		cache:     c,
	}
}

// GetUserFinancialSummary — POST /dashboard/financial-summary
func (h *dashboardHandler) GetUserFinancialSummary(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)

	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	var req dto.GetUserFinancialSummaryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    data.ErrInvalidRequestBody,
		})
	}

	// Build cache key
	rangeStart, rangeEnd := "", ""
	if req.Range != nil {
		rangeStart = req.Range.Start
		rangeEnd = req.Range.End
	}
	cacheKey := cache.DashboardFinancialSummary(userData.ID, req.WalletID, rangeStart, rangeEnd)

	// Try cache
	if cached, err := h.cache.Get(c.UserContext(), cacheKey); err == nil && cached != nil {
		logger.Debug(data.LogCacheHit, map[string]any{"service": data.CacheService, "key": cacheKey})
		c.Set(data.ContentTypeHeader, data.ContentTypeJSON)
		return c.Send(cached)
	}

	grpcReq := &dpb.GetUserFinancialSummaryRequest{
		UserId:   userData.ID,
		WalletId: req.WalletID,
	}
	if req.Range != nil {
		grpcReq.Range = &dpb.DateRange{
			Start: req.Range.Start,
			End:   req.Range.End,
		}
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.dashboard.GetUserFinancialSummary(ctx, grpcReq)
	if err != nil {
		logger.Error(data.LogGetFinancialSummaryFailed, map[string]any{
			"service":    data.DashboardService,
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

	resp := dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Financial summary retrieved successfully",
		Data:       result.GetSummaries(),
	}

	// Store in cache
	if b, err := json.Marshal(resp); err == nil {
		if err := h.cache.Set(c.UserContext(), cacheKey, b, cache.TTLLong); err != nil {
			logger.Warn(data.LogCacheSetFailed, map[string]any{"service": data.CacheService, "key": cacheKey, "error": err.Error()})
		}
	}

	return c.JSON(resp)
}

// GetUserBalance — POST /dashboard/balance
func (h *dashboardHandler) GetUserBalance(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)

	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	var req dto.GetUserBalanceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    data.ErrInvalidRequestBody,
		})
	}

	if req.Aggregation == "" {
		req.Aggregation = "monthly"
	}

	// Build cache key
	rangeStart, rangeEnd := "", ""
	if req.Range != nil {
		rangeStart = req.Range.Start
		rangeEnd = req.Range.End
	}
	cacheKey := cache.DashboardBalance(userData.ID, req.WalletID, req.Aggregation, rangeStart, rangeEnd)

	// Try cache
	if cached, err := h.cache.Get(c.UserContext(), cacheKey); err == nil && cached != nil {
		logger.Debug(data.LogCacheHit, map[string]any{"service": data.CacheService, "key": cacheKey})
		c.Set(data.ContentTypeHeader, data.ContentTypeJSON)
		return c.Send(cached)
	}

	grpcReq := &dpb.GetUserBalanceRequest{
		UserId:      userData.ID,
		WalletId:    req.WalletID,
		Aggregation: req.Aggregation,
	}
	if req.Range != nil {
		grpcReq.Range = &dpb.DateRange{
			Start: req.Range.Start,
			End:   req.Range.End,
		}
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.dashboard.GetUserBalance(ctx, grpcReq)
	if err != nil {
		logger.Error(data.LogGetUserBalanceFailed, map[string]any{
			"service":    data.DashboardService,
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

	resp := dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "User balance retrieved successfully",
		Data:       result.GetSnapshots(),
	}

	if b, err := json.Marshal(resp); err == nil {
		if err := h.cache.Set(c.UserContext(), cacheKey, b, cache.TTLLong); err != nil {
			logger.Warn(data.LogCacheSetFailed, map[string]any{"service": data.CacheService, "key": cacheKey, "error": err.Error()})
		}
	}

	return c.JSON(resp)
}

// GetUserTransactions — POST /dashboard/transactions
func (h *dashboardHandler) GetUserTransactions(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)

	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	var req dto.GetUserTransactionsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    data.ErrInvalidRequestBody,
		})
	}

	// Build cache key
	dateOptHash := cache.HashParams(fmt.Sprintf("%+v", req.DateOption))
	cacheKey := cache.DashboardTransactions(userData.ID, req.WalletID, dateOptHash)

	if cached, err := h.cache.Get(c.UserContext(), cacheKey); err == nil && cached != nil {
		logger.Debug(data.LogCacheHit, map[string]any{"service": data.CacheService, "key": cacheKey})
		c.Set(data.ContentTypeHeader, data.ContentTypeJSON)
		return c.Send(cached)
	}

	grpcReq := &dpb.GetUserTransactionsRequest{
		UserId:   userData.ID,
		WalletId: req.WalletID,
	}

	dateOpt := &dpb.DateOption{}
	hasDateOpt := false
	if req.DateOption.Date != nil {
		dateOpt.Date = *req.DateOption.Date
		hasDateOpt = true
	}
	if req.DateOption.Year != nil {
		dateOpt.Year = int32(*req.DateOption.Year)
		hasDateOpt = true
	}
	if req.DateOption.Month != nil {
		dateOpt.Month = int32(*req.DateOption.Month)
		hasDateOpt = true
	}
	if req.DateOption.Day != nil {
		dateOpt.Day = int32(*req.DateOption.Day)
		hasDateOpt = true
	}
	if req.DateOption.Range != nil {
		dateOpt.Range = &dpb.DateRange{
			Start: req.DateOption.Range.Start,
			End:   req.DateOption.Range.End,
		}
		hasDateOpt = true
	}
	if hasDateOpt {
		grpcReq.DateOption = dateOpt
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.dashboard.GetUserTransactions(ctx, grpcReq)
	if err != nil {
		logger.Error(data.LogGetUserTransactionsFailed, map[string]any{
			"service":    data.DashboardService,
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

	resp := dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "User transactions retrieved successfully",
		Data:       result.GetCategories(),
	}

	if b, err := json.Marshal(resp); err == nil {
		if err := h.cache.Set(c.UserContext(), cacheKey, b, cache.TTLMedium); err != nil {
			logger.Warn(data.LogCacheSetFailed, map[string]any{"service": data.CacheService, "key": cacheKey, "error": err.Error()})
		}
	}

	return c.JSON(resp)
}

// GetUserNetWorthComposition — POST /dashboard/net-worth
func (h *dashboardHandler) GetUserNetWorthComposition(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)

	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	cacheKey := cache.DashboardNetWorth(userData.ID)

	if cached, err := h.cache.Get(c.UserContext(), cacheKey); err == nil && cached != nil {
		logger.Debug(data.LogCacheHit, map[string]any{"service": data.CacheService, "key": cacheKey})
		c.Set(data.ContentTypeHeader, data.ContentTypeJSON)
		return c.Send(cached)
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.dashboard.GetUserNetWorthComposition(
		ctx,
		&dpb.GetUserNetWorthCompositionRequest{UserId: userData.ID},
	)
	if err != nil {
		logger.Error(data.LogGetNetWorthCompositionFailed, map[string]any{
			"service":    data.DashboardService,
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

	if proto.Equal(result, &dpb.NetWorthComposition{}) {
		result = nil
	}

	resp := dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Net worth composition retrieved successfully",
		Data:       result,
	}

	if b, err := json.Marshal(resp); err == nil {
		if err := h.cache.Set(c.UserContext(), cacheKey, b, cache.TTLLong); err != nil {
			logger.Warn(data.LogCacheSetFailed, map[string]any{"service": data.CacheService, "key": cacheKey, "error": err.Error()})
		}
	}

	return c.JSON(resp)
}

// GetUserWallets — GET /dashboard/wallets
func (h *dashboardHandler) GetUserWallets(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)

	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	cacheKey := cache.DashboardWallets(userData.ID)

	if cached, err := h.cache.Get(c.UserContext(), cacheKey); err == nil && cached != nil {
		logger.Debug(data.LogCacheHit, map[string]any{"service": data.CacheService, "key": cacheKey})
		c.Set(data.ContentTypeHeader, data.ContentTypeJSON)
		return c.Send(cached)
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.dashboard.GetUserWallets(ctx, userData.ID)
	if err != nil {
		logger.Error(data.LogGetUserWalletsFailed, map[string]any{
			"service":    data.DashboardService,
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

	resp := dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "User wallets retrieved successfully",
		Data:       result.GetWallets(),
	}

	if b, err := json.Marshal(resp); err == nil {
		if err := h.cache.Set(c.UserContext(), cacheKey, b, cache.TTLMedium); err != nil {
			logger.Warn(data.LogCacheSetFailed, map[string]any{"service": data.CacheService, "key": cacheKey, "error": err.Error()})
		}
	}

	return c.JSON(resp)
}