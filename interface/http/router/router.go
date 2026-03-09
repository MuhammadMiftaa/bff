package router

import (
	grpcClient "refina-web-bff/interface/grpc/client"
	"refina-web-bff/interface/http/middleware"
	"refina-web-bff/interface/http/routes"
	"refina-web-bff/internal/cache"
	"refina-web-bff/internal/types/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/redis/go-redis/v9"
)

func SetupHTTPServer(dc grpcClient.DashboardClient, wc grpcClient.WalletClient, tc grpcClient.TransactionClient, ic grpcClient.InvestmentClient, c cache.Cache, redisClient *redis.Client) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Refina BFF",
		ServerHeader: "Refina",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(dto.APIResponse{
				Status:     false,
				StatusCode: code,
				Message:    err.Error(),
			})
		},
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(middleware.RequestIDMiddleware())
	app.Use(middleware.CORSMiddleware())
	app.Use(middleware.LoggerMiddleware())
	app.Use(middleware.RateLimiterMiddleware(redisClient))

	// Health check
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(dto.APIResponse{
			Status:     true,
			StatusCode: 200,
			Message:    "BFF is healthy",
		})
	})

	// Register route groups
	routes.DashboardRoutes(app, dc, c)
	routes.WalletRoutes(app, tc, wc, c)
	routes.TransactionRoutes(app, tc, wc, c)
	routes.InvestmentRoutes(app, ic, wc, c)
	routes.CacheRoutes(app, c)

	return app
}
