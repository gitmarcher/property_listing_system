package controllers

import (
	"fmt"
	"strconv"
	"time"

	"property_lister/models"
	"property_lister/services"
	"property_lister/types"

	"github.com/gofiber/fiber/v2"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ListingResponse struct {
	Success bool                  `json:"success"`
	Data    interface{}           `json:"data,omitempty"`
	Message string                `json:"message,omitempty"`
	Meta    *types.PaginationMeta `json:"meta,omitempty"`
}

type CreateListingRequest struct {
	Title         string   `json:"title" validate:"required"`
	Type          string   `json:"type" validate:"required"`
	Price         int      `json:"price" validate:"required"`
	State         string   `json:"state" validate:"required"`
	City          string   `json:"city" validate:"required"`
	AreaSqFt      int      `json:"areaSqFt" validate:"required"`
	Bedrooms      int      `json:"bedrooms" validate:"required"`
	Bathrooms     int      `json:"bathrooms" validate:"required"`
	Amenities     []string `json:"amenities"`
	Furnished     string   `json:"furnished" validate:"required"`
	AvailableFrom string   `json:"availableFrom" validate:"required"`
	Tags          []string `json:"tags"`
	ListingType   string   `json:"listingType" validate:"required,oneof=rent sale"`
}

type UpdateListingRequest struct {
	Title         string   `json:"title"`
	Type          string   `json:"type"`
	Price         int      `json:"price"`
	State         string   `json:"state"`
	City          string   `json:"city"`
	AreaSqFt      int      `json:"areaSqFt"`
	Bedrooms      int      `json:"bedrooms"`
	Bathrooms     int      `json:"bathrooms"`
	Amenities     []string `json:"amenities"`
	Furnished     string   `json:"furnished"`
	AvailableFrom string   `json:"availableFrom"`
	Tags          []string `json:"tags"`
	ListingType   string   `json:"listingType" validate:"omitempty,oneof=rent sale"`
}

// GetListings handles GET /api/listings
func GetListings(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID := c.Locals("user_id").(string)

	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Try to get from cache first (only for first page with default limit)
	if page == 1 && limit == 10 {
		listingsKey := services.GetCacheKey("user_listings", userID, "")
		var cachedListings []models.Property
		if err := services.GetCache(listingsKey, &cachedListings); err == nil {
			// Calculate pagination metadata for cached data
			total := int64(len(cachedListings))
			totalPages := int((total + int64(limit) - 1) / int64(limit))

			// Return first page from cache
			end := limit
			if end > len(cachedListings) {
				end = len(cachedListings)
			}

			return c.JSON(ListingResponse{
				Success: true,
				Data:    cachedListings[:end],
				Meta: &types.PaginationMeta{
					Page:       page,
					Limit:      limit,
					Total:      total,
					TotalPages: totalPages,
				},
			})
		}
	}

	// Cache miss or non-default pagination - fetch from database
	// Build filter
	filter := bson.M{"created_by": userID}

	// Calculate skip value
	skip := (page - 1) * limit

	// Setup find options
	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64(skip))
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}}) // Sort by newest first

	// Get total count
	total, err := mgm.Coll(&models.Property{}).CountDocuments(mgm.Ctx(), filter)
	if err != nil {
		return c.Status(500).JSON(ListingResponse{
			Success: false,
			Message: "Failed to count listings",
		})
	}

	// Find properties
	var properties []models.Property
	cursor, err := mgm.Coll(&models.Property{}).Find(mgm.Ctx(), filter, findOptions)
	if err != nil {
		return c.Status(500).JSON(ListingResponse{
			Success: false,
			Message: "Failed to fetch listings",
		})
	}
	defer cursor.Close(mgm.Ctx())

	if err = cursor.All(mgm.Ctx(), &properties); err != nil {
		return c.Status(500).JSON(ListingResponse{
			Success: false,
			Message: "Failed to decode listings",
		})
	}

	// Cache the results if it's the full dataset (no pagination)
	if page == 1 && limit == 10 {
		// Fetch all listings for caching (not just the page)
		var allListings []models.Property
		cursor, err := mgm.Coll(&models.Property{}).Find(mgm.Ctx(), filter,
			options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
		if err == nil {
			cursor.All(mgm.Ctx(), &allListings)
			cursor.Close(mgm.Ctx())

			listingsKey := services.GetCacheKey("user_listings", userID, "")
			services.SetCache(listingsKey, allListings)
		}
	}

	// Calculate total pages
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return c.JSON(ListingResponse{
		Success: true,
		Data:    properties,
		Meta: &types.PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// CreateListing handles PUT /api/listings
func CreateListing(c *fiber.Ctx) error {
	var req CreateListingRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(ListingResponse{
			Success: false,
			Message: "Invalid request body",
		})
	}

	// Get user ID from context (set by auth middleware)
	userID := c.Locals("user_id").(string)

	// Find the highest property ID
	var lastProperty models.Property
	err := mgm.Coll(&models.Property{}).
		FindOne(
			mgm.Ctx(),
			bson.M{},
			options.FindOne().SetSort(bson.D{{Key: "id", Value: -1}}),
		).Decode(&lastProperty)

	// Generate new ID
	var newIDNum int
	if err != nil {
		// No properties exist yet, start from 1000
		newIDNum = 1000
	} else {
		// Extract number from last ID (remove "PROP" prefix)
		lastIDNum, err := strconv.Atoi(lastProperty.ID[4:])
		if err != nil {
			return c.Status(500).JSON(ListingResponse{
				Success: false,
				Message: "Failed to generate property ID",
			})
		}
		newIDNum = lastIDNum + 1
	}

	// Format new ID with PROP prefix and padded number
	newID := fmt.Sprintf("PROP%d", newIDNum)

	now := time.Now()
	property := &models.Property{
		ID:            newID,
		Title:         req.Title,
		Type:          req.Type,
		Price:         req.Price,
		State:         req.State,
		City:          req.City,
		AreaSqFt:      req.AreaSqFt,
		Bedrooms:      req.Bedrooms,
		Bathrooms:     req.Bathrooms,
		Amenities:     req.Amenities,
		Furnished:     req.Furnished,
		AvailableFrom: req.AvailableFrom,
		Tags:          req.Tags,
		ListingType:   req.ListingType,
		CreatedBy:     userID,
		CreatedAt:     now,
		UpdatedAt:     now,
		IsVerified:    false, // New listings start as unverified
		Rating:        0,     // New listings start with 0 rating
	}

	if err := mgm.Coll(property).Create(property); err != nil {
		return c.Status(500).JSON(ListingResponse{
			Success: false,
			Message: "Failed to create listing",
		})
	}

	// Update cache after successful creation
	go services.UpdateListingsCache(userID)

	return c.Status(201).JSON(ListingResponse{
		Success: true,
		Message: "Listing created successfully",
		Data:    property,
	})
}

// UpdateListing handles PATCH /api/listings/:id
func UpdateListing(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(ListingResponse{
			Success: false,
			Message: "Listing ID is required",
		})
	}

	// Get user ID from context (set by auth middleware)
	userID := c.Locals("user_id").(string)

	// Find the existing property
	var property models.Property
	err := mgm.Coll(&property).FindOne(mgm.Ctx(), bson.M{"id": id}).Decode(&property)
	if err != nil {
		return c.Status(404).JSON(ListingResponse{
			Success: false,
			Message: "Listing not found",
		})
	}

	// Check ownership
	if property.CreatedBy != userID && property.CreatedBy != "SYSTEM" {
		return c.Status(403).JSON(ListingResponse{
			Success: false,
			Message: "You don't have permission to update this listing",
		})
	}

	var req UpdateListingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(ListingResponse{
			Success: false,
			Message: "Invalid request body",
		})
	}

	// Update only provided fields
	update := bson.M{"updated_at": time.Now()}
	if req.Title != "" {
		update["title"] = req.Title
	}
	if req.Type != "" {
		update["type"] = req.Type
	}
	if req.Price != 0 {
		update["price"] = req.Price
	}
	if req.State != "" {
		update["state"] = req.State
	}
	if req.City != "" {
		update["city"] = req.City
	}
	if req.AreaSqFt != 0 {
		update["areaSqFt"] = req.AreaSqFt
	}
	if req.Bedrooms != 0 {
		update["bedrooms"] = req.Bedrooms
	}
	if req.Bathrooms != 0 {
		update["bathrooms"] = req.Bathrooms
	}
	if req.Amenities != nil {
		update["amenities"] = req.Amenities
	}
	if req.Furnished != "" {
		update["furnished"] = req.Furnished
	}
	if req.AvailableFrom != "" {
		update["availableFrom"] = req.AvailableFrom
	}
	if req.Tags != nil {
		update["tags"] = req.Tags
	}
	if req.ListingType != "" {
		update["listingType"] = req.ListingType
	}

	_, err = mgm.Coll(&property).UpdateOne(
		mgm.Ctx(),
		bson.M{"id": id},
		bson.M{"$set": update},
	)

	if err != nil {
		return c.Status(500).JSON(ListingResponse{
			Success: false,
			Message: "Failed to update listing",
		})
	}

	// Fetch updated property
	err = mgm.Coll(&property).FindOne(mgm.Ctx(), bson.M{"id": id}).Decode(&property)
	if err != nil {
		return c.Status(500).JSON(ListingResponse{
			Success: false,
			Message: "Failed to fetch updated listing",
		})
	}

	// Update cache after successful update
	go services.UpdateListingsCache(userID)

	return c.JSON(ListingResponse{
		Success: true,
		Message: "Listing updated successfully",
		Data:    property,
	})
}

// DeleteListing handles DELETE /api/listings/:id
func DeleteListing(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(ListingResponse{
			Success: false,
			Message: "Listing ID is required",
		})
	}

	// Get user ID from context (set by auth middleware)
	userID := c.Locals("user_id").(string)

	// Find the property first to check ownership
	var property models.Property
	err := mgm.Coll(&property).FindOne(mgm.Ctx(), bson.M{"id": id}).Decode(&property)
	if err != nil {
		return c.Status(404).JSON(ListingResponse{
			Success: false,
			Message: "Listing not found",
		})
	}

	// Check ownership
	if property.CreatedBy != userID && property.CreatedBy != "SYSTEM" {
		return c.Status(403).JSON(ListingResponse{
			Success: false,
			Message: "You don't have permission to delete this listing",
		})
	}

	_, err = mgm.Coll(&property).DeleteOne(mgm.Ctx(), bson.M{"id": id})
	if err != nil {
		return c.Status(500).JSON(ListingResponse{
			Success: false,
			Message: "Failed to delete listing",
		})
	}

	// Update cache after successful deletion
	go services.UpdateListingsCache(userID)

	return c.JSON(ListingResponse{
		Success: true,
		Message: "Listing deleted successfully",
	})
}
