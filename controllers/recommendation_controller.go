package controllers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"property_lister/models"
	"property_lister/services"
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

	// Verify property exists using the string ID format (PROP prefixed)
	var property models.Property
	err = mgm.Coll(&property).FindOne(mgm.Ctx(), bson.M{"id": req.PropertyID}).Decode(&property)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "property not found"})
	}

	// Verify recipient email exists in the system
	var recipient models.User
	err = mgm.Coll(&recipient).FindOne(mgm.Ctx(), bson.M{"email": req.RecipientEmail}).Decode(&recipient)
	if err != nil {
		log.Printf("Recipient email %s not found in system: %v", req.RecipientEmail, err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "recipient email not found in system",
			"message": "The recipient must be a registered user to receive recommendations",
		})
	}
	log.Printf("Recipient email %s verified, user ID: %s", req.RecipientEmail, recipient.ID.Hex())

	// Verify sender is not trying to recommend to themselves
	if req.RecipientEmail == c.Locals("email").(string) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot send recommendation to yourself",
		})
	}

	// Check for duplicate recommendations (same sender, property, and recipient)
	var existingRecommendation models.Recommendation
	err = mgm.Coll(&existingRecommendation).FindOne(mgm.Ctx(), bson.M{
		"sender_id":       senderObjID,
		"property_id":     req.PropertyID,
		"recipient_email": req.RecipientEmail,
	}).Decode(&existingRecommendation)
	if err == nil {
		log.Printf("Duplicate recommendation found for property %s to %s", req.PropertyID, req.RecipientEmail)
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error":   "recommendation already exists",
			"message": "You have already recommended this property to this user",
		})
	}
	log.Printf("No duplicate recommendation found, proceeding with creation")

	// Create new recommendation
	now := time.Now()
	recommendation := &models.Recommendation{
		PropertyID:     req.PropertyID, // Use string ID directly
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

	// Update cache for sender's sent recommendations
	go services.UpdateSentRecommendationsCache(userID)

	// Update cache for recipient's received recommendations (if they exist)
	go services.UpdateReceivedRecommendationsCacheByEmail(req.RecipientEmail)

	// TODO: Send email to recipient

	return c.Status(fiber.StatusOK).JSON(recommendation)
}

// GetUserRecommendations returns all recommendations for the current user
func GetUserRecommendations(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	userEmail := c.Locals("email").(string) // Note: using "email" not "user_email"
	log.Printf("Getting recommendations for user %s (email: %s)", userID, userEmail)

	// Try to get sent recommendations from cache
	sentKey := services.GetCacheKey("user_recommendations", userID, "sent")
	var cachedSentRecommendations []models.Recommendation
	sentFromCache := services.GetCache(sentKey, &cachedSentRecommendations) == nil

	// Try to get received recommendations from cache
	receivedKey := services.GetCacheKey("user_recommendations", userID, "received")
	var cachedReceivedRecommendations []models.Recommendation
	receivedFromCache := services.GetCache(receivedKey, &cachedReceivedRecommendations) == nil

	if sentFromCache && receivedFromCache {
		// Both found in cache, combine and return
		log.Printf("Returning cached recommendations for user %s", userID)
		allRecommendations := append(cachedSentRecommendations, cachedReceivedRecommendations...)
		return c.Status(fiber.StatusOK).JSON(allRecommendations)
	}

	// Cache miss - fetch from database
	log.Printf("Cache miss, fetching recommendations from database for user %s", userID)
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	// Get user's recommendations
	var recommendations []models.Recommendation
	err = mgm.Coll(&models.Recommendation{}).SimpleFind(&recommendations, bson.M{
		"$or": []bson.M{
			{"sender_id": userObjID},
			{"recipient_email": userEmail},
		},
	})

	if err != nil {
		log.Printf("Failed to fetch recommendations for user %s: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch recommendations"})
	}

	// Separate sent and received recommendations
	var sentRecommendations []models.Recommendation
	var receivedRecommendations []models.Recommendation

	for _, rec := range recommendations {
		if rec.SenderID == userObjID {
			sentRecommendations = append(sentRecommendations, rec)
		} else {
			receivedRecommendations = append(receivedRecommendations, rec)
		}
	}

	// Cache the results
	if !sentFromCache {
		services.SetCache(sentKey, sentRecommendations)
	}
	if !receivedFromCache {
		services.SetCache(receivedKey, receivedRecommendations)
	}

	log.Printf("Successfully fetched %d recommendations for user %s", len(recommendations), userID)
	return c.Status(fiber.StatusOK).JSON(recommendations)
}

// GetSentRecommendations returns recommendations sent by the current user
func GetSentRecommendations(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	log.Printf("Getting sent recommendations for user %s", userID)

	// Try to get sent recommendations from cache
	sentKey := services.GetCacheKey("user_recommendations", userID, "sent")
	var cachedSentRecommendations []models.Recommendation
	sentFromCache := services.GetCache(sentKey, &cachedSentRecommendations) == nil

	if sentFromCache {
		log.Printf("Returning cached sent recommendations for user %s", userID)
		return c.Status(fiber.StatusOK).JSON(cachedSentRecommendations)
	}

	// Cache miss - fetch from database
	log.Printf("Cache miss, fetching sent recommendations from database for user %s", userID)
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	// Get sent recommendations
	var sentRecommendations []models.Recommendation
	err = mgm.Coll(&models.Recommendation{}).SimpleFind(&sentRecommendations, bson.M{
		"sender_id": userObjID,
	})

	if err != nil {
		log.Printf("Failed to fetch sent recommendations for user %s: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch sent recommendations"})
	}

	// Cache the results
	services.SetCache(sentKey, sentRecommendations)

	log.Printf("Successfully fetched %d sent recommendations for user %s", len(sentRecommendations), userID)
	return c.Status(fiber.StatusOK).JSON(sentRecommendations)
}

// GetReceivedRecommendations returns recommendations received by the current user
func GetReceivedRecommendations(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	userEmail := c.Locals("email").(string)
	log.Printf("Getting received recommendations for user %s (email: %s)", userID, userEmail)

	// Try to get received recommendations from cache
	receivedKey := services.GetCacheKey("user_recommendations", userID, "received")
	var cachedReceivedRecommendations []models.Recommendation
	receivedFromCache := services.GetCache(receivedKey, &cachedReceivedRecommendations) == nil

	if receivedFromCache {
		log.Printf("Returning cached received recommendations for user %s", userID)
		return c.Status(fiber.StatusOK).JSON(cachedReceivedRecommendations)
	}

	// Cache miss - fetch from database
	log.Printf("Cache miss, fetching received recommendations from database for user %s", userID)

	// Get received recommendations
	var receivedRecommendations []models.Recommendation
	err := mgm.Coll(&models.Recommendation{}).SimpleFind(&receivedRecommendations, bson.M{
		"recipient_email": userEmail,
	})

	if err != nil {
		log.Printf("Failed to fetch received recommendations for user %s: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch received recommendations"})
	}

	// Cache the results
	services.SetCache(receivedKey, receivedRecommendations)

	log.Printf("Successfully fetched %d received recommendations for user %s", len(receivedRecommendations), userID)
	return c.Status(fiber.StatusOK).JSON(receivedRecommendations)
}
