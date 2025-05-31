package routes

import (
	"property_lister/controllers"
	"property_lister/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRecommendationRoutes(app *fiber.App) {
	api := app.Group("/api")
	recommendations := api.Group("/recommendations", middleware.AuthMiddleware())

	recommendations.Post("/send", controllers.SendRecommendation)
	recommendations.Get("/", controllers.GetUserRecommendations)
	recommendations.Get("/sent", controllers.GetSentRecommendations)
	recommendations.Get("/received", controllers.GetReceivedRecommendations)
}
