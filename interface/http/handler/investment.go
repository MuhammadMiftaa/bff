package handler

import (
	"strconv"

	logger "refina-web-bff/config/log"
	grpcClient "refina-web-bff/interface/grpc/client"
	"refina-web-bff/interface/grpc/interceptor"
	"refina-web-bff/internal/types/dto"
	"refina-web-bff/internal/utils/data"

	ipb "github.com/MuhammadMiftaa/Refina-Protobuf/investment"
	"github.com/gofiber/fiber/v2"
)

type investmentHandler struct {
	investment grpcClient.InvestmentClient
}

func NewInvestmentHandler(ic grpcClient.InvestmentClient) *investmentHandler {
	return &investmentHandler{investment: ic}
}

// GetUserInvestments — GET /investments
func (h *investmentHandler) GetUserInvestments(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))

	req := &ipb.GetUserInvestmentListRequest{
		UserId:    userData.ID,
		Page:      int32(page),
		PageSize:  int32(pageSize),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
		Search:    c.Query("search"),
		Code:      c.Query("code"),
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.investment.GetUserInvestmentList(ctx, req)
	if err != nil {
		logger.Error(data.LogGetInvestmentsFailed, map[string]any{
			"service":    data.InvestmentService,
			"request_id": requestID,
			"user_id":    userData.ID,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    "Failed to get investments",
		})
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Investments retrieved successfully",
		Data:       result,
	})
}

// GetInvestmentDetail — GET /investments/:id
func (h *investmentHandler) GetInvestmentDetail(c *fiber.Ctx) error {
	investmentID := c.Params("id")
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	if investmentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "Investment ID is required",
		})
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.investment.GetInvestmentDetail(ctx, &ipb.GetInvestmentDetailRequest{
		UserId:       userData.ID,
		InvestmentId: investmentID,
	})
	if err != nil {
		logger.Error(data.LogGetInvestmentDetailFailed, map[string]any{
			"service":       data.InvestmentService,
			"request_id":    requestID,
			"investment_id": investmentID,
			"error":         err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    "Failed to get investment detail",
		})
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Investment detail retrieved successfully",
		Data:       result,
	})
}

// CreateInvestment — POST /investments
func (h *investmentHandler) CreateInvestment(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	var body dto.CreateInvestmentRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "Invalid request body",
		})
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.investment.CreateInvestment(ctx, &ipb.CreateInvestmentRequest{
		UserId:           userData.ID,
		Code:             body.Code,
		Quantity:         body.Quantity,
		Amount:           body.Amount,
		InitialValuation: body.InitialValuation,
		Date:             body.Date,
		Description:      body.Description,
	})
	if err != nil {
		logger.Error(data.LogCreateInvestmentFailed, map[string]any{
			"service":    data.InvestmentService,
			"request_id": requestID,
			"user_id":    userData.ID,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    "Failed to create investment",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 201,
		Message:    "Investment created successfully",
		Data:       result,
	})
}

// SellInvestment — POST /investments/sell
func (h *investmentHandler) SellInvestment(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	var body dto.SellInvestmentRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "Invalid request body",
		})
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.investment.SellInvestment(ctx, &ipb.SellInvestmentRequest{
		UserId:      userData.ID,
		AssetCode:   body.AssetCode,
		Quantity:    body.Quantity,
		Amount:      body.Amount,
		Date:        body.Date,
		Description: body.Description,
	})
	if err != nil {
		logger.Error(data.LogSellInvestmentFailed, map[string]any{
			"service":    data.InvestmentService,
			"request_id": requestID,
			"user_id":    userData.ID,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    "Failed to sell investment",
		})
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Investment sold successfully",
		Data:       result,
	})
}

// GetInvestmentSummary — GET /investments/summary
func (h *investmentHandler) GetInvestmentSummary(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.investment.GetInvestmentSummary(ctx, userData.ID)
	if err != nil {
		logger.Error(data.LogGetInvestmentSummaryFailed, map[string]any{
			"service":    data.InvestmentService,
			"request_id": requestID,
			"user_id":    userData.ID,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    "Failed to get investment summary",
		})
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Investment summary retrieved successfully",
		Data:       result,
	})
}

// GetAssetCodes — GET /investments/asset-codes
func (h *investmentHandler) GetAssetCodes(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.investment.GetAssetCodes(ctx)
	if err != nil {
		logger.Error(data.LogGetAssetCodesFailed, map[string]any{
			"service":    data.InvestmentService,
			"request_id": requestID,
			"user_id":    userData.ID,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    "Failed to get asset codes",
		})
	}

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Asset codes retrieved successfully",
		Data:       result,
	})
}
