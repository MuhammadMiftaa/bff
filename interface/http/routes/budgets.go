package routes

import (
	grpcClient "refina-web-bff/interface/grpc/client"
	"refina-web-bff/interface/http/handler"
	"refina-web-bff/interface/http/middleware"
	"refina-web-bff/internal/cache"

	"github.com/gofiber/fiber/v2"
)

func BudgetRoutes(app *fiber.App, tc grpcClient.TransactionClient, c cache.Cache) {
	h := handler.NewBudgetHandler(tc, c)

	// Budget CRUD routes
	budgets := app.Group("/budgets")
	budgets.Use(middleware.AuthMiddleware())

	budgets.Get("/", h.GetBudgets)            // GET /budgets?period=2024-01
	budgets.Post("/", h.CreateBudget)         // POST /budgets
	budgets.Put("/:id", h.UpdateBudget)       // PUT /budgets/:id
	budgets.Delete("/:id", h.DeleteBudget)    // DELETE /budgets/:id
	budgets.Post("/:id/reset", h.ResetBudget) // POST /budgets/:id/reset
}
