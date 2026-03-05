package handler

import (
	"fmt"

	logger "refina-web-bff/config/log"
	grpcClient "refina-web-bff/interface/grpc/client"
	"refina-web-bff/interface/grpc/interceptor"
	"refina-web-bff/internal/types/dto"
	"refina-web-bff/internal/utils/data"

	tpb "github.com/MuhammadMiftaa/Refina-Protobuf/transaction"
	wpb "github.com/MuhammadMiftaa/Refina-Protobuf/wallet"
	"github.com/gofiber/fiber/v2"
)

type walletHandler struct {
	transaction grpcClient.TransactionClient
	wallet      grpcClient.WalletClient
}

func NewWalletHandler(tc grpcClient.TransactionClient, wc grpcClient.WalletClient) *walletHandler {
	return &walletHandler{
		transaction: tc,
		wallet:      wc,
	}
}

// GetUserWallets — GET /wallets
func (h *walletHandler) GetUserWallets(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

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

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "User wallets retrieved successfully",
		Data:       result.GetWallets(),
	})
}

// GetWalletSummary — GET /wallets/summary
func (h *walletHandler) GetWalletSummary(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

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

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Wallet summary retrieved successfully",
		Data:       result,
	})
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

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Wallet retrieved successfully",
		Data:       result,
	})
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

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

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

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Wallet types retrieved successfully",
		Data:       result.GetWalletTypes(),
	})
}
