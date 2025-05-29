package controllers

import (
	"property_lister/models"

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

	// If user has no favorites, return empty array
	if len(user.Favorites) == 0 {
		return c.JSON(FavoriteResponse{
			Success: true,
			Data:    []models.Property{},
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
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(400).JSON(FavoriteResponse{
			Success: false,
			Message: "Invalid user ID",
		})
	}

	// Verify property exists
	var property models.Property
	err = mgm.Coll(&property).FindOne(mgm.Ctx(), bson.M{"id": propertyID}).Decode(&property)
	if err != nil {
		return c.Status(404).JSON(FavoriteResponse{
			Success: false,
			Message: "Property not found",
		})
	}

	// Add to favorites if not already in list
	result := mgm.Coll(&models.User{}).FindOneAndUpdate(
		mgm.Ctx(),
		bson.M{
			"_id": objID,
			"favorites": bson.M{
				"$ne": propertyID,
			},
		},
		bson.M{
			"$push": bson.M{"favorites": propertyID},
		},
	)

	if result.Err() != nil {
		// Check if it's because property is already in favorites
		var user models.User
		err = mgm.Coll(&user).FindOne(mgm.Ctx(), bson.M{
			"_id":       objID,
			"favorites": propertyID,
		}).Decode(&user)

		if err == nil {
			return c.Status(400).JSON(FavoriteResponse{
				Success: false,
				Message: "Property is already in favorites",
			})
		}

		return c.Status(500).JSON(FavoriteResponse{
			Success: false,
			Message: "Failed to add to favorites",
		})
	}

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

	return c.JSON(FavoriteResponse{
		Success: true,
		Message: "Removed from favorites successfully",
	})
}
