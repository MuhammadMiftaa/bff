package handler

import (
	"context"

	logger "refina-web-bff/config/log"
	grpcClient "refina-web-bff/interface/grpc/client"
	"refina-web-bff/internal/types/dto"
	"refina-web-bff/internal/utils/data"

	dpb "github.com/MuhammadMiftaa/Refina-Protobuf/dashboard"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/protobuf/proto"
)

type dashboardHandler struct {
	dashboard grpcClient.DashboardClient
}

func NewDashboardHandler(dc grpcClient.DashboardClient) *dashboardHandler {
	return &dashboardHandler{
		dashboard: dc,
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
			Message:    "Invalid request body",
		})
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

	result, err := h.dashboard.GetUserFinancialSummary(context.Background(), grpcReq)
	if err != nil {
		logger.Error(data.LogGetFinancialSummaryFailed, map[string]any{
			"service":    data.DashboardService,
			"request_id": requestID,
			"user_id":    userData.ID,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    "Failed to get financial summary",
		})
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Financial summary retrieved successfully",
		Data:       result.GetSummaries(),
	})
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
			Message:    "Invalid request body",
		})
	}

	if req.Aggregation == "" {
		req.Aggregation = "monthly"
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

	result, err := h.dashboard.GetUserBalance(context.Background(), grpcReq)
	if err != nil {
		logger.Error(data.LogGetUserBalanceFailed, map[string]any{
			"service":    data.DashboardService,
			"request_id": requestID,
			"user_id":    userData.ID,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    "Failed to get user balance",
		})
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "User balance retrieved successfully",
		Data:       result.GetSnapshots(),
	})
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
			Message:    "Invalid request body",
		})
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

	result, err := h.dashboard.GetUserTransactions(context.Background(), grpcReq)
	if err != nil {
		logger.Error(data.LogGetUserTransactionsFailed, map[string]any{
			"service":    data.DashboardService,
			"request_id": requestID,
			"user_id":    userData.ID,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    "Failed to get user transactions",
		})
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "User transactions retrieved successfully",
		Data:       result.GetCategories(),
	})
}

// GetUserNetWorthComposition — POST /dashboard/net-worth
func (h *dashboardHandler) GetUserNetWorthComposition(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)

	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	result, err := h.dashboard.GetUserNetWorthComposition(
		context.Background(),
		&dpb.GetUserNetWorthCompositionRequest{UserId: userData.ID},
	)
	if err != nil {
		logger.Error(data.LogGetNetWorthCompositionFailed, map[string]any{
			"service":    data.DashboardService,
			"request_id": requestID,
			"user_id":    userData.ID,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    "Failed to get net worth composition",
		})
	}
	
	if proto.Equal(result, &dpb.NetWorthComposition{}) {
		result = nil
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Net worth composition retrieved successfully",
		Data:       result,
	})
}

// GetUserWallets — GET /dashboard/wallets
func (h *dashboardHandler) GetUserWallets(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)

	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	result, err := h.dashboard.GetUserWallets(context.Background(), userData.ID)
	if err != nil {
		logger.Error(data.LogGetUserWalletsFailed, map[string]any{
			"service":    data.DashboardService,
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

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "User wallets retrieved successfully",
		Data:       result.GetWallets(),
	})
}
