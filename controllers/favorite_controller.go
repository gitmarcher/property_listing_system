package controllers

import (
	"log"
	"property_lister/models"
	"property_lister/services"

	"github.com/gofiber/fiber/v2"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FavoriteResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// GetFavorites returns all favorite properties of the authenticated user
func GetFavorites(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID := c.Locals("user_id").(string)

	// Try to get from cache first
	favPropsKey := services.GetCacheKey("user_favorite_properties", userID, "")
	var cachedProperties []models.Property

	err := services.GetCache(favPropsKey, &cachedProperties)
	if err == nil && len(cachedProperties) >= 0 {
		// Cache hit - return cached data
		return c.JSON(FavoriteResponse{
			Success: true,
			Data:    cachedProperties,
		})
	}

	// Cache miss - fetch from database
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(400).JSON(FavoriteResponse{
			Success: false,
			Message: "Invalid user ID",
		})
	}

	// Find user
	var user models.User
	err = mgm.Coll(&user).FindOne(mgm.Ctx(), bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		return c.Status(404).JSON(FavoriteResponse{
			Success: false,
			Message: "User not found",
		})
	}

	// If user has no favorites, return empty array and cache it
	if len(user.Favorites) == 0 {
		emptyProperties := []models.Property{}
		services.SetCache(favPropsKey, emptyProperties)
		return c.JSON(FavoriteResponse{
			Success: true,
			Data:    emptyProperties,
		})
	}

	// Find all favorite properties
	var properties []models.Property
	cursor, err := mgm.Coll(&models.Property{}).Find(mgm.Ctx(), bson.M{
		"id": bson.M{"$in": user.Favorites},
	})
	if err != nil {
		return c.Status(500).JSON(FavoriteResponse{
			Success: false,
			Message: "Failed to fetch favorite properties",
		})
	}
	defer cursor.Close(mgm.Ctx())

	if err = cursor.All(mgm.Ctx(), &properties); err != nil {
		return c.Status(500).JSON(FavoriteResponse{
			Success: false,
			Message: "Failed to decode properties",
		})
	}

	// Cache the result for future requests
	services.SetCache(favPropsKey, properties)

	// Also cache the favorites list
	favKey := services.GetCacheKey("user_favorites", userID, "")
	services.SetCache(favKey, user.Favorites)

	return c.JSON(FavoriteResponse{
		Success: true,
		Data:    properties,
	})
}

// AddToFavorites adds a property to user's favorites
func AddToFavorites(c *fiber.Ctx) error {
	propertyID := c.Params("propertyId")
	if propertyID == "" {
		return c.Status(400).JSON(FavoriteResponse{
			Success: false,
			Message: "Property ID is required",
		})
	}

	// Get user ID from context
	userID := c.Locals("user_id").(string)
	log.Printf("Adding property %s to favorites for user %s", propertyID, userID)

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Printf("Invalid user ID format: %s, error: %v", userID, err)
		return c.Status(400).JSON(FavoriteResponse{
			Success: false,
			Message: "Invalid user ID",
		})
	}

	// Verify property exists
	var property models.Property
	err = mgm.Coll(&property).FindOne(mgm.Ctx(), bson.M{"id": propertyID}).Decode(&property)
	if err != nil {
		log.Printf("Property %s not found: %v", propertyID, err)
		return c.Status(404).JSON(FavoriteResponse{
			Success: false,
			Message: "Property not found",
		})
	}
	log.Printf("Property %s found successfully", propertyID)

	// Add to favorites if not already in list
	// First, ensure the favorites field exists as an array (handle null case)
	_, err = mgm.Coll(&models.User{}).UpdateOne(
		mgm.Ctx(),
		bson.M{
			"_id":       objID,
			"favorites": bson.M{"$exists": false},
		},
		bson.M{
			"$set": bson.M{"favorites": []string{}},
		},
	)
	if err != nil {
		log.Printf("Failed to initialize favorites array: %v", err)
	}

	// Handle the case where favorites is null
	_, err = mgm.Coll(&models.User{}).UpdateOne(
		mgm.Ctx(),
		bson.M{
			"_id":       objID,
			"favorites": nil,
		},
		bson.M{
			"$set": bson.M{"favorites": []string{}},
		},
	)
	if err != nil {
		log.Printf("Failed to convert null favorites to array: %v", err)
	}

	// Now add to favorites using $addToSet to avoid duplicates
	result, err := mgm.Coll(&models.User{}).UpdateOne(
		mgm.Ctx(),
		bson.M{"_id": objID},
		bson.M{
			"$addToSet": bson.M{"favorites": propertyID},
		},
	)

	if err != nil {
		log.Printf("Failed to add to favorites: %v", err)
		return c.Status(500).JSON(FavoriteResponse{
			Success: false,
			Message: "Failed to add to favorites",
		})
	}

	// Check if any document was modified (if 0, property was already in favorites)
	if result.ModifiedCount == 0 {
		log.Printf("Property %s is already in favorites for user %s", propertyID, userID)
		return c.Status(400).JSON(FavoriteResponse{
			Success: false,
			Message: "Property is already in favorites",
		})
	}

	log.Printf("Successfully added property %s to favorites for user %s", propertyID, userID)
	// Update cache after successful database update
	go services.UpdateFavoritesCache(userID)

	return c.JSON(FavoriteResponse{
		Success: true,
		Message: "Added to favorites successfully",
	})
}

// RemoveFromFavorites removes a property from user's favorites
func RemoveFromFavorites(c *fiber.Ctx) error {
	propertyID := c.Params("propertyId")
	if propertyID == "" {
		return c.Status(400).JSON(FavoriteResponse{
			Success: false,
			Message: "Property ID is required",
		})
	}

	// Get user ID from context
	userID := c.Locals("user_id").(string)
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(400).JSON(FavoriteResponse{
			Success: false,
			Message: "Invalid user ID",
		})
	}

	// Remove from favorites
	result := mgm.Coll(&models.User{}).FindOneAndUpdate(
		mgm.Ctx(),
		bson.M{"_id": objID},
		bson.M{
			"$pull": bson.M{"favorites": propertyID},
		},
	)

	if result.Err() != nil {
		return c.Status(500).JSON(FavoriteResponse{
			Success: false,
			Message: "Failed to remove from favorites",
		})
	}

	// Update cache after successful database update
	go services.UpdateFavoritesCache(userID)

	return c.JSON(FavoriteResponse{
		Success: true,
		Message: "Removed from favorites successfully",
	})
}
