package routes

import (
	"refina-web-bff/interface/http/handler"
	"refina-web-bff/interface/http/middleware"
	"refina-web-bff/internal/cache"

	"github.com/gofiber/fiber/v2"
)

// CacheRoutes registers the manual cache-refresh endpoint.
//
//	DELETE /cache/refresh?service=<dashboard|wallet|transaction|investment>
//
// Omitting ?service refreshes the entire user-scoped cache across all domains.
// Requires a valid JWT (AuthMiddleware).
func CacheRoutes(app *fiber.App, c cache.Cache) {
	h := handler.NewCacheHandler(c)

	api := app.Group("/cache", middleware.AuthMiddleware())
	api.Delete("/refresh", h.RefreshCache)
}
