package routes

import (
	grpcClient "refina-web-bff/interface/grpc/client"
	"refina-web-bff/interface/http/handler"
	"refina-web-bff/interface/http/middleware"
	"refina-web-bff/internal/cache"

	"github.com/gofiber/fiber/v2"
)

func ProfileRoutes(app *fiber.App, pc grpcClient.ProfileClient, c cache.Cache) {
	h := handler.NewProfileHandler(pc, c)

	profile := app.Group("/profile")
	profile.Use(middleware.AuthMiddleware())

	profile.Get("/", h.GetProfile)
	profile.Put("/", h.UpdateProfile)
	profile.Post("/photo", h.UploadProfilePhoto)
	profile.Delete("/photo", h.DeleteProfilePhoto)
}
