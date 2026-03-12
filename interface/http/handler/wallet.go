package handler

import (
	"context"
	"encoding/json"
	"fmt"

	logger "refina-web-bff/config/log"
	grpcClient "refina-web-bff/interface/grpc/client"
	"refina-web-bff/interface/grpc/interceptor"
	"refina-web-bff/internal/cache"
	"refina-web-bff/internal/types/dto"
	"refina-web-bff/internal/utils/data"

	tpb "github.com/MuhammadMiftaa/Refina-Protobuf/transaction"
	wpb "github.com/MuhammadMiftaa/Refina-Protobuf/wallet"
	"github.com/gofiber/fiber/v2"
)

type walletHandler struct {
	transaction grpcClient.TransactionClient
	wallet      grpcClient.WalletClient
	cache       cache.Cache
}

func NewWalletHandler(tc grpcClient.TransactionClient, wc grpcClient.WalletClient, c cache.Cache) *walletHandler {
	return &walletHandler{
		transaction: tc,
		wallet:      wc,
		cache:       c,
	}
}

// GetUserWallets — GET /wallets
func (h *walletHandler) GetUserWallets(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	cacheKey := cache.WalletList(userData.ID)

	if cached, err := h.cache.Get(c.UserContext(), cacheKey); err == nil && cached != nil {
		logger.Debug(data.LogCacheHit, map[string]any{"service": data.CacheService, "key": cacheKey})
		c.Set("Content-Type", "application/json")
		return c.Send(cached)
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.wallet.GetUserWallets(ctx, userData.ID)
	if err != nil {
		logger.Error(data.LogGetWalletsFailed, map[string]any{
			"service":    data.WalletService,
			"request_id": requestID,
			"user_id":    userData.ID,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    "Failed to get user wallets",
		})
	}

	walletIds := make([]string, 0, len(result.GetWallets()))
	for _, wallet := range result.GetWallets() {
		walletIds = append(walletIds, wallet.GetId())
	}

	userTransactions, err := h.transaction.GetUserTransactions(ctx, &tpb.GetUserTransactionsRequest{WalletIds: walletIds, PageSize: -1})
	if err != nil {
		logger.Error(data.LogGetTransactionsFailed, map[string]any{
			"service":    data.TransactionService,
			"request_id": requestID,
			"user_id":    userData.ID,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    "Failed to get transactions",
		})
	}

	walletTrxCount := make(map[string]int32)
	for _, tx := range userTransactions.GetTransactions() {
		walletTrxCount[tx.GetWalletId()]++
	}

	for _, wallet := range result.GetWallets() {
		wallet.TransactionCount = walletTrxCount[wallet.GetId()]
	}

	resp := dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "User wallets retrieved successfully",
		Data:       result.GetWallets(),
	}

	if b, err := json.Marshal(resp); err == nil {
		if err := h.cache.Set(c.UserContext(), cacheKey, b, cache.TTLShort); err != nil {
			logger.Warn(data.LogCacheSetFailed, map[string]any{"service": data.CacheService, "key": cacheKey, "error": err.Error()})
		}
	}

	return c.JSON(resp)
}

// GetWalletSummary — GET /wallets/summary
func (h *walletHandler) GetWalletSummary(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	cacheKey := cache.WalletSummary(userData.ID)

	if cached, err := h.cache.Get(c.UserContext(), cacheKey); err == nil && cached != nil {
		logger.Debug(data.LogCacheHit, map[string]any{"service": data.CacheService, "key": cacheKey})
		c.Set("Content-Type", "application/json")
		return c.Send(cached)
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.wallet.GetWalletSummary(ctx, userData.ID)
	if err != nil {
		logger.Error(data.LogGetWalletSummaryFailed, map[string]any{
			"service":    data.WalletService,
			"request_id": requestID,
			"user_id":    userData.ID,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    "Failed to get wallet summary",
		})
	}

	resp := dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Wallet summary retrieved successfully",
		Data:       result,
	}

	if b, err := json.Marshal(resp); err == nil {
		if err := h.cache.Set(c.UserContext(), cacheKey, b, cache.TTLShort); err != nil {
			logger.Warn(data.LogCacheSetFailed, map[string]any{"service": data.CacheService, "key": cacheKey, "error": err.Error()})
		}
	}

	return c.JSON(resp)
}

// GetWalletByID — GET /wallets/:id
func (h *walletHandler) GetWalletByID(c *fiber.Ctx) error {
	walletID := c.Params("id")
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	if walletID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "Wallet ID is required",
		})
	}

	cacheKey := cache.WalletByID(walletID)

	if cached, err := h.cache.Get(c.UserContext(), cacheKey); err == nil && cached != nil {
		logger.Debug(data.LogCacheHit, map[string]any{"service": data.CacheService, "key": cacheKey})
		c.Set("Content-Type", "application/json")
		return c.Send(cached)
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.wallet.GetWalletByID(ctx, walletID)
	if err != nil {
		logger.Error(data.LogGetWalletByIDFailed, map[string]any{
			"service":    data.WalletService,
			"request_id": requestID,
			"wallet_id":  walletID,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    "Failed to get wallet",
		})
	}

	resp := dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Wallet retrieved successfully",
		Data:       result,
	}

	if b, err := json.Marshal(resp); err == nil {
		if err := h.cache.Set(c.UserContext(), cacheKey, b, cache.TTLShort); err != nil {
			logger.Warn(data.LogCacheSetFailed, map[string]any{"service": data.CacheService, "key": cacheKey, "error": err.Error()})
		}
	}

	return c.JSON(resp)
}

// CreateWallet — POST /wallets
func (h *walletHandler) CreateWallet(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	var req dto.CreateWalletRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "Invalid request body",
		})
	}

	if req.Name == "" || req.WalletTypeID == "" || req.Number == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "Name, wallet_type_id, and number are required",
		})
	}

	grpcReq := &wpb.CreateWalletRequest{
		UserId:       userData.ID,
		WalletTypeId: req.WalletTypeID,
		Name:         req.Name,
		Balance:      req.Balance,
		Number:       req.Number,
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.wallet.CreateWallet(ctx, grpcReq)
	if err != nil {
		logger.Error(data.LogCreateWalletFailed, map[string]any{
			"service":    data.WalletService,
			"request_id": requestID,
			"user_id":    userData.ID,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to create wallet: %s", err.Error()),
		})
	}

	// Invalidate wallet & dashboard caches
	go h.invalidateWalletCaches(c.UserContext(), userData.ID)

	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "Wallet created successfully",
		Data:       result,
	})
}

// UpdateWallet — PUT /wallets/:id
func (h *walletHandler) UpdateWallet(c *fiber.Ctx) error {
	walletID := c.Params("id")
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	if walletID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "Wallet ID is required",
		})
	}

	var req dto.UpdateWalletRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "Invalid request body",
		})
	}

	grpcReq := &wpb.UpdateWalletRequest{
		Id:           walletID,
		Name:         req.Name,
		Number:       req.Number,
		WalletTypeId: req.WalletTypeID,
	}

	wallet, err := h.wallet.GetWalletByID(ctx, walletID)
	if err != nil {
		logger.Error(data.LogGetWalletByIDFailed, map[string]any{
			"service":    data.WalletService,
			"request_id": requestID,
			"wallet_id":  walletID,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    "Failed to get wallet",
		})
	}

	if wallet == nil {
		logger.Error(data.LogGetWalletByIDFailed, map[string]any{
			"service":    data.WalletService,
			"request_id": requestID,
			"wallet_id":  walletID,
		})
		return c.Status(fiber.StatusNotFound).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 404,
			Message:    "Wallet not found",
		})
	}

	if wallet.GetUserId() != userData.ID {
		return c.Status(fiber.StatusForbidden).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 403,
			Message:    "You are not allowed to update this wallet",
		})
	}

	grpcReq.Balance = wallet.GetBalance()

	result, err := h.wallet.UpdateWallet(ctx, grpcReq)
	if err != nil {
		logger.Error(data.LogUpdateWalletFailed, map[string]any{
			"service":    data.WalletService,
			"request_id": requestID,
			"wallet_id":  walletID,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to update wallet: %s", err.Error()),
		})
	}

	// Invalidate wallet & dashboard caches
	go h.invalidateWalletCaches(c.UserContext(), userData.ID)
	go h.cache.Delete(c.UserContext(), cache.WalletByID(walletID))

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Wallet updated successfully",
		Data:       result,
	})
}

// DeleteWallet — DELETE /wallets/:id
func (h *walletHandler) DeleteWallet(c *fiber.Ctx) error {
	walletID := c.Params("id")
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	if walletID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "Wallet ID is required",
		})
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	_, err := h.wallet.DeleteWallet(ctx, walletID)
	if err != nil {
		logger.Error(data.LogDeleteWalletFailed, map[string]any{
			"service":    data.WalletService,
			"request_id": requestID,
			"wallet_id":  walletID,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to delete wallet: %s", err.Error()),
		})
	}

	// Invalidate wallet & dashboard caches
	go h.invalidateWalletCaches(c.UserContext(), userData.ID)
	go h.cache.Delete(c.UserContext(), cache.WalletByID(walletID))

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Wallet deleted successfully",
	})
}

// GetWalletTypes — GET /wallet-types
func (h *walletHandler) GetWalletTypes(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	cacheKey := cache.WalletTypes()

	if cached, err := h.cache.Get(c.UserContext(), cacheKey); err == nil && cached != nil {
		logger.Debug(data.LogCacheHit, map[string]any{"service": data.CacheService, "key": cacheKey})
		c.Set("Content-Type", "application/json")
		return c.Send(cached)
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.wallet.GetWalletTypes(ctx)
	if err != nil {
		logger.Error(data.LogGetWalletTypesFailed, map[string]any{
			"service":    data.WalletService,
			"request_id": requestID,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    "Failed to get wallet types",
		})
	}

	resp := dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Wallet types retrieved successfully",
		Data:       result.GetWalletTypes(),
	}

	if b, err := json.Marshal(resp); err == nil {
		if err := h.cache.Set(c.UserContext(), cacheKey, b, cache.TTLStatic); err != nil {
			logger.Warn(data.LogCacheSetFailed, map[string]any{"service": data.CacheService, "key": cacheKey, "error": err.Error()})
		}
	}

	return c.JSON(resp)
}

// invalidateWalletCaches clears wallet-related and dashboard-related caches for a user.
func (h *walletHandler) invalidateWalletCaches(ctx context.Context, userID string) {
	patterns := []string{
		cache.WalletAllPattern(userID),
		cache.DashboardWallets(userID),
		cache.DashboardNetWorth(userID),
	}
	// Delete exact keys first
	_ = h.cache.Delete(ctx, cache.DashboardWallets(userID), cache.DashboardNetWorth(userID))
	// Then pattern-based
	for _, p := range patterns[:1] { // only wallet:* needs pattern scan
		if err := h.cache.DeleteByPattern(ctx, p); err != nil {
			logger.Warn(data.LogCacheInvalidateFail, map[string]any{"service": data.CacheService, "pattern": p, "error": err.Error()})
		}
	}
}
