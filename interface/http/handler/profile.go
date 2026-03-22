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

	"github.com/gofiber/fiber/v2"
)

type profileHandler struct {
	profile grpcClient.ProfileClient
	cache   cache.Cache
}

func NewProfileHandler(pc grpcClient.ProfileClient, c cache.Cache) *profileHandler {
	return &profileHandler{
		profile: pc,
		cache:   c,
	}
}

// GetProfile — GET /profile
func (h *profileHandler) GetProfile(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	// Build cache key
	cacheKey := cache.ProfileData(userData.ID)

	// Try cache
	if cached, err := h.cache.Get(c.UserContext(), cacheKey); err == nil && cached != nil {
		logger.Debug(data.LogCacheHit, map[string]any{"service": data.CacheService, "key": cacheKey})
		c.Set("Content-Type", "application/json")
		return c.Send(cached)
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.profile.GetProfile(ctx, userData.ID)
	if err != nil {
		logger.Error(data.LogGetProfileFailed, map[string]any{
			"service":    data.ProfileService,
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
		Message:    "Profile retrieved successfully",
		Data: dto.ProfileResponse{
			ID:        result.GetId(),
			UserID:    result.GetUserId(),
			Fullname:  result.GetFullname(),
			PhotoURL:  result.GetPhotoUrl(),
			CreatedAt: result.GetCreatedAt(),
			UpdatedAt: result.GetUpdatedAt(),
		},
	}

	// Store in cache
	if b, err := json.Marshal(resp); err == nil {
		if err := h.cache.Set(c.UserContext(), cacheKey, b, cache.TTLShort); err != nil {
			logger.Warn(data.LogCacheSetFailed, map[string]any{
				"service": data.CacheService,
				"key":     cacheKey,
				"error":   err.Error(),
			})
		}
	}

	logger.Info(data.LogGetProfileSuccess, map[string]any{
		"service":    data.ProfileService,
		"request_id": requestID,
		"user_id":    userData.ID,
	})

	return c.JSON(resp)
}

// UpdateProfile — PUT /profile
func (h *profileHandler) UpdateProfile(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	var req dto.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "Invalid request body",
		})
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.profile.UpdateProfile(ctx, userData.ID, req.Fullname)
	if err != nil {
		logger.Error(data.LogUpdateProfileFailed, map[string]any{
			"service":    data.ProfileService,
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
	cacheKey := cache.ProfileData(userData.ID)
	_ = h.cache.Delete(c.UserContext(), cacheKey)

	logger.Info(data.LogUpdateProfileSuccess, map[string]any{
		"service":    data.ProfileService,
		"request_id": requestID,
		"user_id":    userData.ID,
	})

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Profile updated successfully",
		Data: dto.ProfileResponse{
			ID:        result.GetId(),
			UserID:    result.GetUserId(),
			Fullname:  result.GetFullname(),
			PhotoURL:  result.GetPhotoUrl(),
			CreatedAt: result.GetCreatedAt(),
			UpdatedAt: result.GetUpdatedAt(),
		},
	})
}

// UploadProfilePhoto — POST /profile/photo
func (h *profileHandler) UploadProfilePhoto(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	var req dto.UploadPhotoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "Invalid request body",
		})
	}

	if req.Base64Image == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "base64_image is required",
		})
	}

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.profile.UploadProfilePhoto(ctx, userData.ID, req.Base64Image)
	if err != nil {
		logger.Error(data.LogUploadPhotoFailed, map[string]any{
			"service":    data.ProfileService,
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

	if !result.GetSuccess() {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    result.GetMessage(),
		})
	}

	// Invalidate cache
	cacheKey := cache.ProfileData(userData.ID)
	_ = h.cache.Delete(c.UserContext(), cacheKey)

	logger.Info(data.LogUploadPhotoSuccess, map[string]any{
		"service":    data.ProfileService,
		"request_id": requestID,
		"user_id":    userData.ID,
	})

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    result.GetMessage(),
		Data: dto.UploadPhotoResponse{
			Success:  result.GetSuccess(),
			PhotoURL: result.GetPhotoUrl(),
			Message:  result.GetMessage(),
		},
	})
}

// DeleteProfilePhoto — DELETE /profile/photo
func (h *profileHandler) DeleteProfilePhoto(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)

	ctx := interceptor.ContextWithUserData(c.UserContext(), userData)

	result, err := h.profile.DeleteProfilePhoto(ctx, userData.ID)
	if err != nil {
		logger.Error(data.LogDeletePhotoFailed, map[string]any{
			"service":    data.ProfileService,
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

	if !result.GetSuccess() {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    result.GetMessage(),
		})
	}

	// Invalidate cache
	cacheKey := cache.ProfileData(userData.ID)
	_ = h.cache.Delete(c.UserContext(), cacheKey)

	logger.Info(data.LogDeletePhotoSuccess, map[string]any{
		"service":    data.ProfileService,
		"request_id": requestID,
		"user_id":    userData.ID,
	})

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    result.GetMessage(),
	})
}
