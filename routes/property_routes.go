package routes

import (
	"property_lister/controllers"

	"github.com/gofiber/fiber/v2"
)

func SetupPropertyRoutes(app *fiber.App) {
	api := app.Group("/api")

	properties := api.Group("/properties")

	properties.Get("/", controllers.GetProperties)
	properties.Get("/:id", controllers.GetPropertyByID)
	properties.Get("/search", controllers.SearchProperties)
}
