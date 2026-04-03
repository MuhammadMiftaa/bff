package routes

import (
	grpcClient "refina-web-bff/interface/grpc/client"
	"refina-web-bff/interface/http/handler"
	"refina-web-bff/interface/http/middleware"
	"refina-web-bff/internal/cache"

	"github.com/gofiber/fiber/v2"
)

func DashboardRoutes(app *fiber.App, dc grpcClient.DashboardClient, c cache.Cache) {
	h := handler.NewDashboardHandler(dc, c)

	dashboard := app.Group("/dashboard")
	dashboard.Use(middleware.AuthMiddleware())

	dashboard.Get("/wallets", h.GetUserWallets)
	dashboard.Post("/financial-summary", h.GetUserFinancialSummary)
	dashboard.Post("/balance", h.GetUserBalance)
	dashboard.Post("/transactions", h.GetUserTransactions)
	dashboard.Post("/net-worth", h.GetUserNetWorthComposition)
	dashboard.Post("/category-transactions", h.GetCategoryTransactions)
}
