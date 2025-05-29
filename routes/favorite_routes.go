package routes

import (
	"property_lister/controllers"
	"property_lister/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupFavoriteRoutes(app *fiber.App) {
	api := app.Group("/api")

	favorites := api.Group("/favorites", middleware.AuthMiddleware())

	// Get user's favorites
	favorites.Get("/", controllers.GetFavorites)

	// Add to favorites
	favorites.Post("/:propertyId", controllers.AddToFavorites)

	// Remove from favorites
	favorites.Delete("/:propertyId", controllers.RemoveFromFavorites)
}
