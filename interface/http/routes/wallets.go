package routes

import (
	grpcClient "refina-web-bff/interface/grpc/client"
	"refina-web-bff/interface/http/handler"
	"refina-web-bff/interface/http/middleware"

	"github.com/gofiber/fiber/v2"
)

func WalletRoutes(app *fiber.App, tc grpcClient.TransactionClient, wc grpcClient.WalletClient) {
	h := handler.NewWalletHandler(tc, wc)

	// Wallet CRUD routes
	wallets := app.Group("/wallets")
	wallets.Use(middleware.AuthMiddleware())

	wallets.Get("/summary", h.GetWalletSummary)
	wallets.Get("/:id", h.GetWalletByID)
	wallets.Get("/", h.GetUserWallets)
	wallets.Post("/", h.CreateWallet)
	wallets.Put("/:id", h.UpdateWallet)
	wallets.Delete("/:id", h.DeleteWallet)

	// Wallet types route
	walletTypes := app.Group("/wallet-types")
	walletTypes.Use(middleware.AuthMiddleware())

	walletTypes.Get("/", h.GetWalletTypes)
}
