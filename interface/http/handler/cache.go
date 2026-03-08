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
	service := c.Query("service") // "" means all

	ctx := c.UserContext()
	userID := userData.ID

	// Collect all patterns that need to be wiped for the requested service(s).
	type invalidation struct {
		kind    string // "pattern" | "key"
		subject string
	}

	build := func(svc string) []invalidation {
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
				{kind: "key", subject: cache.TransactionCategories("")}, // all-type bucket
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

	services := []string{"dashboard", "wallet", "transaction", "investment"}
	if service != "" {
		// Validate
		valid := false
		for _, s := range services {
			if s == service {
				valid = true
				break
			}
		}
		if !valid {
			return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
				Status:     false,
				StatusCode: 400,
				Message:    "Invalid service. Must be one of: dashboard, wallet, transaction, investment",
			})
		}
		services = []string{service}
	}

	var invalidations []invalidation
	for _, svc := range services {
		invalidations = append(invalidations, build(svc)...)
	}

	var failed []string
	for _, inv := range invalidations {
		var err error
		if inv.kind == "pattern" {
			err = h.cache.DeleteByPattern(ctx, inv.subject)
		} else {
			err = h.cache.Delete(ctx, inv.subject)
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
		"user_id":    userID,
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
