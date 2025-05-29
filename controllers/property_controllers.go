package controllers

import (
	"strconv"

	"property_lister/models"

	"github.com/gofiber/fiber/v2"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PropertyResponse struct {
	Success bool            `json:"success"`
	Data    interface{}     `json:"data,omitempty"`
	Message string          `json:"message,omitempty"`
	Meta    *PaginationMeta `json:"meta,omitempty"`
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// GetProperties handles GET /api/properties with filtering and pagination
func GetProperties(c *fiber.Ctx) error {
	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Build filter
	filter := bson.M{}

	// Price range filter
	if minPrice := c.Query("min_price"); minPrice != "" {
		if min, err := strconv.Atoi(minPrice); err == nil {
			filter["price"] = bson.M{"$gte": min}
		}
	}
	if maxPrice := c.Query("max_price"); maxPrice != "" {
		if max, err := strconv.Atoi(maxPrice); err == nil {
			if existing, ok := filter["price"].(bson.M); ok {
				existing["$lte"] = max
			} else {
				filter["price"] = bson.M{"$lte": max}
			}
		}
	}

	// Location filters
	if state := c.Query("state"); state != "" {
		filter["state"] = bson.M{"$regex": state, "$options": "i"}
	}
	if city := c.Query("city"); city != "" {
		filter["city"] = bson.M{"$regex": city, "$options": "i"}
	}

	// Property type filter
	if propType := c.Query("type"); propType != "" {
		filter["type"] = bson.M{"$regex": propType, "$options": "i"}
	}

	// Bedrooms filter
	if bedrooms := c.Query("bedrooms"); bedrooms != "" {
		if beds, err := strconv.Atoi(bedrooms); err == nil {
			filter["bedrooms"] = beds
		}
	}

	// Bathrooms filter
	if bathrooms := c.Query("bathrooms"); bathrooms != "" {
		if baths, err := strconv.Atoi(bathrooms); err == nil {
			filter["bathrooms"] = baths
		}
	}

	// Verified filter
	if verified := c.Query("verified"); verified != "" {
		if verified == "true" {
			filter["isVerified"] = true
		} else if verified == "false" {
			filter["isVerified"] = false
		}
	}

	// Furnished filter
	if furnished := c.Query("furnished"); furnished != "" {
		filter["furnished"] = bson.M{"$regex": furnished, "$options": "i"}
	}

	// Setup sorting
	sort := bson.D{}
	sortBy := c.Query("sort_by", "")
	sortOrder := c.Query("sort_order", "asc")

	if sortBy != "" {
		order := 1
		if sortOrder == "desc" {
			order = -1
		}

		switch sortBy {
		case "price", "rating", "areaSqFt", "bedrooms", "bathrooms":
			sort = append(sort, bson.E{Key: sortBy, Value: order})
		default:
			sort = append(sort, bson.E{Key: "price", Value: 1})
		}
	} else {
		sort = append(sort, bson.E{Key: "price", Value: 1})
	}

	// Calculate skip value
	skip := (page - 1) * limit

	// Setup find options
	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64(skip))
	findOptions.SetSort(sort)

	// Get total count
	total, err := mgm.Coll(&models.Property{}).CountDocuments(mgm.Ctx(), filter)
	if err != nil {
		return c.Status(500).JSON(PropertyResponse{
			Success: false,
			Message: "Failed to count properties",
		})
	}

	// Find properties
	var properties []models.Property
	cursor, err := mgm.Coll(&models.Property{}).Find(mgm.Ctx(), filter, findOptions)
	if err != nil {
		return c.Status(500).JSON(PropertyResponse{
			Success: false,
			Message: "Failed to fetch properties",
		})
	}
	defer cursor.Close(mgm.Ctx())

	if err = cursor.All(mgm.Ctx(), &properties); err != nil {
		return c.Status(500).JSON(PropertyResponse{
			Success: false,
			Message: "Failed to decode properties",
		})
	}

	// Calculate total pages
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return c.JSON(PropertyResponse{
		Success: true,
		Data:    properties,
		Meta: &PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// GetPropertyByID handles GET /api/properties/:id
func GetPropertyByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(PropertyResponse{
			Success: false,
			Message: "Property ID is required",
		})
	}

	var property models.Property
	err := mgm.Coll(&property).FindOne(mgm.Ctx(), bson.M{"id": id}).Decode(&property)
	if err != nil {
		return c.Status(404).JSON(PropertyResponse{
			Success: false,
			Message: "Property not found",
		})
	}

	return c.JSON(PropertyResponse{
		Success: true,
		Data:    property,
	})
}

// SearchProperties handles GET /api/properties/search with text search
func SearchProperties(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(400).JSON(PropertyResponse{
			Success: false,
			Message: "Search query is required",
		})
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Build search filter using regex for multiple fields
	searchFilter := bson.M{
		"$or": []bson.M{
			{"title": bson.M{"$regex": query, "$options": "i"}},
			{"state": bson.M{"$regex": query, "$options": "i"}},
			{"city": bson.M{"$regex": query, "$options": "i"}},
			{"type": bson.M{"$regex": query, "$options": "i"}},
			{"amenities": bson.M{"$regex": query, "$options": "i"}},
			{"tags": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	// Calculate skip
	skip := (page - 1) * limit

	// Setup find options
	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64(skip))
	findOptions.SetSort(bson.D{{Key: "rating", Value: -1}}) // Sort by rating desc

	// Get total count
	total, err := mgm.Coll(&models.Property{}).CountDocuments(mgm.Ctx(), searchFilter)
	if err != nil {
		return c.Status(500).JSON(PropertyResponse{
			Success: false,
			Message: "Failed to count search results",
		})
	}

	// Find properties
	var properties []models.Property
	cursor, err := mgm.Coll(&models.Property{}).Find(mgm.Ctx(), searchFilter, findOptions)
	if err != nil {
		return c.Status(500).JSON(PropertyResponse{
			Success: false,
			Message: "Failed to search properties",
		})
	}
	defer cursor.Close(mgm.Ctx())

	if err = cursor.All(mgm.Ctx(), &properties); err != nil {
		return c.Status(500).JSON(PropertyResponse{
			Success: false,
			Message: "Failed to decode search results",
		})
	}

	// Calculate total pages
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return c.JSON(PropertyResponse{
		Success: true,
		Data:    properties,
		Message: "Search completed successfully",
		Meta: &PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}
