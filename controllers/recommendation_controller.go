package controllers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"property_lister/models"
)

type SendRecommendationRequest struct {
	PropertyID     string `json:"property_id" binding:"required"`
	RecipientEmail string `json:"recipient_email" binding:"required,email"`
	Message        string `json:"message"`
}

// SendRecommendation handles sending a property recommendation
func SendRecommendation(c *fiber.Ctx) error {
	var req SendRecommendationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Get current user from context (assuming you have middleware that sets this)
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	senderObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	propertyObjID, err := primitive.ObjectIDFromHex(req.PropertyID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid property id"})
	}

	// Create new recommendation
	now := time.Now()
	recommendation := &models.Recommendation{
		PropertyID:     propertyObjID,
		SenderID:       senderObjID,
		RecipientEmail: req.RecipientEmail,
		Message:        req.Message,
		Status:         "pending",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Save recommendation
	if err := mgm.Coll(recommendation).Create(recommendation); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create recommendation"})
	}

	// Update sender's recommendations sent
	sender := &models.User{}
	if err := mgm.Coll(sender).FindByID(senderObjID, sender); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to find sender"})
	}

	sender.RecommendationsSent = append(sender.RecommendationsSent, recommendation.ID)
	if err := mgm.Coll(sender).Update(sender); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update sender"})
	}

	// TODO: Send email to recipient

	return c.Status(fiber.StatusOK).JSON(recommendation)
}

// GetUserRecommendations returns all recommendations for the current user
func GetUserRecommendations(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	// Get user's recommendations
	var recommendations []models.Recommendation
	err = mgm.Coll(&models.Recommendation{}).SimpleFind(&recommendations, bson.M{
		"$or": []bson.M{
			{"sender_id": userObjID},
			{"recipient_email": c.Locals("user_email").(string)}, // assuming email is set in context
		},
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch recommendations"})
	}

	return c.Status(fiber.StatusOK).JSON(recommendations)
}
