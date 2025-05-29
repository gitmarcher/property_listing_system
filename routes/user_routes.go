package routes

import (
	"property_lister/controllers"

	"github.com/gofiber/fiber/v2"
)

func SetupUserRoutes(app *fiber.App) {
	api := app.Group("/api")

	auth := api.Group("/auth")

	auth.Post("/register", controllers.RegisterUser)
	auth.Post("/login", controllers.LoginUser)
}
