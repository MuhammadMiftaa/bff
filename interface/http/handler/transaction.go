package handler

import (
	"fmt"
	"strconv"

	logger "refina-web-bff/config/log"
	grpcClient "refina-web-bff/interface/grpc/client"
	"refina-web-bff/interface/grpc/interceptor"
	"refina-web-bff/internal/types/dto"
	"refina-web-bff/internal/utils/data"

	tpb "github.com/MuhammadMiftaa/Refina-Protobuf/transaction"
	"github.com/gofiber/fiber/v2"
)

type transactionHandler struct {
	transaction grpcClient.TransactionClient
}

func NewTransactionHandler(tc grpcClient.TransactionClient) *transactionHandler {
	return &transactionHandler{
		transaction: tc,
	}
}

// ── Transaction Handlers ──

// GetUserTransactions — GET /transactions
func (h *transactionHandler) GetUserTransactions(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	// Build gRPC request from query params
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))

	req := &tpb.GetUserTransactionsRequest{
		Page:         int32(page),
		PageSize:     int32(pageSize),
		SortBy:       c.Query("sort_by"),
		SortOrder:    c.Query("sort_order"),
		Search:       c.Query("search"),
		WalletId:     c.Query("wallet_id"),
		CategoryId:   c.Query("category_id"),
		CategoryType: c.Query("category_type"),
		DateFrom:     c.Query("date_from"),
		DateTo:       c.Query("date_to"),
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.transaction.GetUserTransactions(ctx, req)
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

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Transactions retrieved successfully",
		Data:       result,
	})
}

// GetTransactionByID — GET /transactions/:id
func (h *transactionHandler) GetTransactionByID(c *fiber.Ctx) error {
	transactionID := c.Params("id")
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	if transactionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "Transaction ID is required",
		})
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.transaction.GetTransactionByID(ctx, transactionID)
	if err != nil {
		logger.Error(data.LogGetTransactionByIDFailed, map[string]any{
			"service":        data.TransactionService,
			"request_id":     requestID,
			"transaction_id": transactionID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    "Failed to get transaction",
		})
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Transaction retrieved successfully",
		Data:       result,
	})
}

// CreateTransaction — POST /transactions
func (h *transactionHandler) CreateTransaction(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	var req dto.CreateTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "Invalid request body",
		})
	}

	if req.WalletID == "" || req.CategoryID == "" || req.TransactionDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "wallet_id, category_id, and transaction_date are required",
		})
	}

	grpcReq := &tpb.CreateTransactionRequest{
		UserId:          userData.ID,
		WalletId:        req.WalletID,
		CategoryId:      req.CategoryID,
		Amount:          req.Amount,
		TransactionDate: req.TransactionDate,
		Description:     req.Description,
		Attachments:     req.Attachments,
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.transaction.CreateTransaction(ctx, grpcReq)
	if err != nil {
		logger.Error(data.LogCreateTransactionFailed, map[string]any{
			"service":    data.TransactionService,
			"request_id": requestID,
			"user_id":    userData.ID,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to create transaction: %s", err.Error()),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "Transaction created successfully",
		Data:       result,
	})
}

// CreateFundTransfer — POST /transactions/transfer
func (h *transactionHandler) CreateFundTransfer(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	var req dto.CreateFundTransferRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "Invalid request body",
		})
	}

	if req.FromWalletID == "" || req.ToWalletID == "" || req.CashOutCategoryID == "" || req.CashInCategoryID == "" || req.TransactionDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "from_wallet_id, to_wallet_id, cash_out_category_id, cash_in_category_id, and transaction_date are required",
		})
	}

	grpcReq := &tpb.CreateFundTransferRequest{
		UserId:            userData.ID,
		FromWalletId:      req.FromWalletID,
		ToWalletId:        req.ToWalletID,
		Amount:            req.Amount,
		AdminFee:          req.AdminFee,
		CashOutCategoryId: req.CashOutCategoryID,
		CashInCategoryId:  req.CashInCategoryID,
		TransactionDate:   req.TransactionDate,
		Description:       req.Description,
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.transaction.CreateFundTransfer(ctx, grpcReq)
	if err != nil {
		logger.Error(data.LogCreateFundTransferFailed, map[string]any{
			"service":    data.TransactionService,
			"request_id": requestID,
			"user_id":    userData.ID,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to create fund transfer: %s", err.Error()),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "Fund transfer created successfully",
		Data:       result,
	})
}

// UpdateTransaction — PUT /transactions/:id
func (h *transactionHandler) UpdateTransaction(c *fiber.Ctx) error {
	transactionID := c.Params("id")
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	if transactionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "Transaction ID is required",
		})
	}

	var req dto.UpdateTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "Invalid request body",
		})
	}

	// Convert attachment actions to proto
	var attachmentActions []*tpb.UpdateAttachmentAction
	for _, action := range req.AttachmentActions {
		attachmentActions = append(attachmentActions, &tpb.UpdateAttachmentAction{
			Status: action.Status,
			Files:  action.Files,
		})
	}

	grpcReq := &tpb.UpdateTransactionRequest{
		Id:                transactionID,
		WalletId:          req.WalletID,
		CategoryId:        req.CategoryID,
		Amount:            req.Amount,
		TransactionDate:   req.TransactionDate,
		Description:       req.Description,
		AttachmentActions: attachmentActions,
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.transaction.UpdateTransaction(ctx, grpcReq)
	if err != nil {
		logger.Error(data.LogUpdateTransactionFailed, map[string]any{
			"service":        data.TransactionService,
			"request_id":     requestID,
			"transaction_id": transactionID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to update transaction: %s", err.Error()),
		})
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Transaction updated successfully",
		Data:       result,
	})
}

// DeleteTransaction — DELETE /transactions/:id
func (h *transactionHandler) DeleteTransaction(c *fiber.Ctx) error {
	transactionID := c.Params("id")
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	if transactionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "Transaction ID is required",
		})
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	_, err := h.transaction.DeleteTransaction(ctx, transactionID)
	if err != nil {
		logger.Error(data.LogDeleteTransactionFailed, map[string]any{
			"service":        data.TransactionService,
			"request_id":     requestID,
			"transaction_id": transactionID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to delete transaction: %s", err.Error()),
		})
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Transaction deleted successfully",
	})
}

// ── Category Handlers ──

// GetCategories — GET /categories
func (h *transactionHandler) GetCategories(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	categoryType := c.Query("type") // optional filter: income, expense, fund_transfer

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.transaction.GetCategories(ctx, categoryType)
	if err != nil {
		logger.Error(data.LogGetCategoriesFailed, map[string]any{
			"service":    data.TransactionService,
			"request_id": requestID,
			"type":       categoryType,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    "Failed to get categories",
		})
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Categories retrieved successfully",
		Data:       result.GetCategories(),
	})
}

// ── Attachment Handlers ──

// GetAttachmentsByTransaction — GET /attachments/transaction/:transactionId
func (h *transactionHandler) GetAttachmentsByTransaction(c *fiber.Ctx) error {
	transactionID := c.Params("transactionId")
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	if transactionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "Transaction ID is required",
		})
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.transaction.GetAttachmentsByTransactionID(ctx, transactionID)
	if err != nil {
		logger.Error(data.LogGetAttachmentsFailed, map[string]any{
			"service":        data.TransactionService,
			"request_id":     requestID,
			"transaction_id": transactionID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    "Failed to get attachments",
		})
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Attachments retrieved successfully",
		Data:       result.GetAttachments(),
	})
}

// CreateAttachment — POST /attachments
func (h *transactionHandler) CreateAttachment(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	var req dto.CreateAttachmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "Invalid request body",
		})
	}

	if req.TransactionID == "" || req.Image == "" || req.Format == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "transaction_id, image, and format are required",
		})
	}

	grpcReq := &tpb.CreateAttachmentRequest{
		TransactionId: req.TransactionID,
		Image:         req.Image,
		Format:        req.Format,
		Size:          req.Size,
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.transaction.CreateAttachment(ctx, grpcReq)
	if err != nil {
		logger.Error(data.LogCreateAttachmentFailed, map[string]any{
			"service":        data.TransactionService,
			"request_id":     requestID,
			"transaction_id": req.TransactionID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to create attachment: %s", err.Error()),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "Attachment created successfully",
		Data:       result,
	})
}

// DeleteAttachment — DELETE /attachments/:id
func (h *transactionHandler) DeleteAttachment(c *fiber.Ctx) error {
	attachmentID := c.Params("id")
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	if attachmentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "Attachment ID is required",
		})
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	_, err := h.transaction.DeleteAttachment(ctx, attachmentID)
	if err != nil {
		logger.Error(data.LogDeleteAttachmentFailed, map[string]any{
			"service":       data.TransactionService,
			"request_id":    requestID,
			"attachment_id": attachmentID,
			"error":         err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to delete attachment: %s", err.Error()),
		})
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Attachment deleted successfully",
	})
}
