package handler

import (
	logger "refina-web-bff/config/log"
	"refina-web-bff/internal/cache"
	"refina-web-bff/internal/types/dto"
	"refina-web-bff/internal/utils/data"

	"github.com/gofiber/fiber/v2"
)

type cacheHandler struct {
	cache cache.Cache
}

func NewCacheHandler(c cache.Cache) *cacheHandler {
	return &cacheHandler{cache: c}
}

// invalidation represents a single cache key or pattern to be wiped.
type invalidation struct {
	kind    string // "pattern" | "key"
	subject string
}

// invalidationsForService returns the list of cache keys/patterns to wipe
// for a given service name. Extracted to reduce cognitive complexity (go:S3776).
func invalidationsForService(svc, userID string) []invalidation {
	switch svc {
	case "dashboard":
		return []invalidation{
			{kind: "pattern", subject: cache.DashboardAllPattern(userID)},
		}
	case "wallet":
		return []invalidation{
			{kind: "pattern", subject: cache.WalletAllPattern(userID)},
			{kind: "key", subject: cache.WalletTypes()},
		}
	case "transaction":
		return []invalidation{
			{kind: "pattern", subject: cache.TransactionListPattern(userID)},
			{kind: "key", subject: cache.TransactionCategories("")},
			{kind: "key", subject: cache.TransactionCategories("income")},
			{kind: "key", subject: cache.TransactionCategories("expense")},
			{kind: "key", subject: cache.TransactionCategories("fund_transfer")},
		}
	case "investment":
		return []invalidation{
			{kind: "pattern", subject: cache.InvestmentAllPattern(userID)},
			{kind: "key", subject: cache.InvestmentAssetCodes()},
		}
	default:
		return nil
	}
}

// collectInvalidations gathers all invalidation entries for the requested services.
// Extracted to reduce cognitive complexity (go:S3776).
func collectInvalidations(services []string, userID string) []invalidation {
	var result []invalidation
	for _, svc := range services {
		result = append(result, invalidationsForService(svc, userID)...)
	}
	return result
}

// isValidService reports whether svc is one of the supported service names.
func isValidService(svc string, services []string) bool {
	for _, s := range services {
		if s == svc {
			return true
		}
	}
	return false
}

// RefreshCache — DELETE /cache/refresh?service=<dashboard|wallet|transaction|investment>
//
// Invalidates all cache entries belonging to the authenticated user for the
// requested service(s). If ?service is omitted, all four service caches are
// cleared at once.
//
// Query param:
//
//	service  (optional) — one of: dashboard | wallet | transaction | investment
//	                       omit to refresh everything.
func (h *cacheHandler) RefreshCache(c *fiber.Ctx) error {
	userData := c.Locals("user_data").(dto.UserData)
	requestID, _ := c.Locals(data.REQUEST_ID_LOCAL_KEY).(string)
	service := c.Query("service")

	allServices := []string{"dashboard", "wallet", "transaction", "investment"}

	if service != "" && !isValidService(service, allServices) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 400,
			Message:    "Invalid service. Must be one of: dashboard, wallet, transaction, investment",
		})
	}

	services := allServices
	if service != "" {
		services = []string{service}
	}

	invalidations := collectInvalidations(services, userData.ID)
	failed := h.applyInvalidations(c, invalidations, requestID, userData.ID)

	if len(failed) > 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Status:     false,
			StatusCode: 500,
			Message:    "Cache refresh partially failed",
			Data: map[string]any{
				"failed_patterns": failed,
			},
		})
	}

	scope := service
	if scope == "" {
		scope = "all"
	}

	logger.Info(data.LogCacheInvalidated, map[string]any{
		"service":    data.CacheService,
		"request_id": requestID,
		"user_id":    userData.ID,
		"scope":      scope,
	})

	return c.JSON(dto.APIResponse{
		Status:     true,
		StatusCode: 200,
		Message:    "Cache refreshed successfully",
		Data: map[string]any{
			"scope": scope,
		},
	})
}

// applyInvalidations executes all cache deletions and returns the list of subjects that failed.
// Extracted to reduce cognitive complexity (go:S3776).
func (h *cacheHandler) applyInvalidations(c *fiber.Ctx, invalidations []invalidation, requestID, userID string) []string {
	var failed []string
	for _, inv := range invalidations {
		var err error
		if inv.kind == "pattern" {
			err = h.cache.DeleteByPattern(c.UserContext(), inv.subject)
		} else {
			err = h.cache.Delete(c.UserContext(), inv.subject)
		}
		if err != nil {
			failed = append(failed, inv.subject)
			logger.Warn(data.LogCacheInvalidateFail, map[string]any{
				"service":    data.CacheService,
				"request_id": requestID,
				"user_id":    userID,
				"subject":    inv.subject,
				"error":      err.Error(),
			})
		}
	}
	return failed
}