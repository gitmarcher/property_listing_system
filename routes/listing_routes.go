package routes

import (
	"property_lister/controllers"
	"property_lister/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupListingRoutes(app *fiber.App) {
	api := app.Group("/api")

	listings := api.Group("/listings", middleware.AuthMiddleware())

	listings.Get("/", controllers.GetListings)
	listings.Put("/", controllers.CreateListing)
	listings.Patch("/:id", controllers.UpdateListing)
	listings.Delete("/:id", controllers.DeleteListing)
}
