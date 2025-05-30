package services

import (
	"encoding/json"
	"fmt"
	"time"

	"property_lister/config"
	"property_lister/models"

	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CacheExpiry = 10 * time.Minute

func GetCacheKey(prefix, userID, suffix string) string {
	if suffix != "" {
		return fmt.Sprintf("%s:%s:%s", prefix, userID, suffix)
	}
	return fmt.Sprintf("%s:%s", prefix, userID)
}

// SetCache stores data in Redis with expiry
func SetCache(key string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return config.RedisClient.Set(config.Ctx, key, jsonData, CacheExpiry).Err()
}

// GetCache retrieves and unmarshals data from Redis
func GetCache(key string, result interface{}) error {
	val, err := config.RedisClient.Get(config.Ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), result)
}

// DeleteCache removes a key from Redis
func DeleteCache(key string) error {
	return config.RedisClient.Del(config.Ctx, key).Err()
}

// DeleteCachePattern removes all keys matching a pattern
func DeleteCachePattern(pattern string) error {
	keys, err := config.RedisClient.Keys(config.Ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return config.RedisClient.Del(config.Ctx, keys...).Err()
	}
	return nil
}

// InvalidateUserCache removes all cached data for a user
func InvalidateUserCache(userID string) error {
	pattern := fmt.Sprintf("*:%s:*", userID)
	return DeleteCachePattern(pattern)
}

// User-specific cache operations

// CacheUserFavorites caches user's favorite properties
func CacheUserFavorites(userID string) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return
	}

	var user models.User
	err = mgm.Coll(&user).FindOne(mgm.Ctx(), bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		return
	}

	// Cache the favorites list
	favKey := GetCacheKey("user_favorites", userID, "")
	SetCache(favKey, user.Favorites)

	// If user has favorites, also cache the full property details
	if len(user.Favorites) > 0 {
		var properties []models.Property
		cursor, err := mgm.Coll(&models.Property{}).Find(mgm.Ctx(), bson.M{
			"id": bson.M{"$in": user.Favorites},
		})
		if err == nil {
			cursor.All(mgm.Ctx(), &properties)
			cursor.Close(mgm.Ctx())

			// Cache favorite properties details
			favPropsKey := GetCacheKey("user_favorite_properties", userID, "")
			SetCache(favPropsKey, properties)
		}
	}
}

// CacheUserRecommendations caches user's sent and received recommendations
func CacheUserRecommendations(userID, email string) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return
	}

	// Cache sent recommendations
	var sentRecommendations []models.Recommendation
	err = mgm.Coll(&models.Recommendation{}).SimpleFind(&sentRecommendations, bson.M{
		"sender_id": objID,
	})
	if err == nil {
		sentKey := GetCacheKey("user_recommendations", userID, "sent")
		SetCache(sentKey, sentRecommendations)
	}

	// Cache received recommendations
	var receivedRecommendations []models.Recommendation
	err = mgm.Coll(&models.Recommendation{}).SimpleFind(&receivedRecommendations, bson.M{
		"recipient_email": email,
	})
	if err == nil {
		receivedKey := GetCacheKey("user_recommendations", userID, "received")
		SetCache(receivedKey, receivedRecommendations)
	}
}

// CacheUserListings caches user's property listings
func CacheUserListings(userID string) {
	var listings []models.Property
	cursor, err := mgm.Coll(&models.Property{}).Find(mgm.Ctx(), bson.M{
		"created_by": userID,
	}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return
	}
	defer cursor.Close(mgm.Ctx())

	if err = cursor.All(mgm.Ctx(), &listings); err != nil {
		return
	}

	listingsKey := GetCacheKey("user_listings", userID, "")
	SetCache(listingsKey, listings)
}

// CacheUserDataOnLogin loads and caches all user-related data
func CacheUserDataOnLogin(userID, email string) {
	// Cache user favorites
	CacheUserFavorites(userID)

	// Cache user recommendations (sent and received)
	CacheUserRecommendations(userID, email)

	// Cache user listings
	CacheUserListings(userID)
}

// UpdateFavoritesCache updates the favorites cache after database changes
func UpdateFavoritesCache(userID string) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return
	}

	// Get updated user data
	var user models.User
	err = mgm.Coll(&user).FindOne(mgm.Ctx(), bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		// If we can't fetch updated data, just invalidate the cache
		favKey := GetCacheKey("user_favorites", userID, "")
		favPropsKey := GetCacheKey("user_favorite_properties", userID, "")
		DeleteCache(favKey)
		DeleteCache(favPropsKey)
		return
	}

	// Update favorites list cache
	favKey := GetCacheKey("user_favorites", userID, "")
	SetCache(favKey, user.Favorites)

	// Update favorite properties cache
	if len(user.Favorites) > 0 {
		var properties []models.Property
		cursor, err := mgm.Coll(&models.Property{}).Find(mgm.Ctx(), bson.M{
			"id": bson.M{"$in": user.Favorites},
		})
		if err == nil {
			cursor.All(mgm.Ctx(), &properties)
			cursor.Close(mgm.Ctx())

			favPropsKey := GetCacheKey("user_favorite_properties", userID, "")
			SetCache(favPropsKey, properties)
		}
	} else {
		// Cache empty array if no favorites
		favPropsKey := GetCacheKey("user_favorite_properties", userID, "")
		SetCache(favPropsKey, []models.Property{})
	}
}

// UpdateSentRecommendationsCache updates the cache for user's sent recommendations
func UpdateSentRecommendationsCache(userID string) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return
	}

	var sentRecommendations []models.Recommendation
	err = mgm.Coll(&models.Recommendation{}).SimpleFind(&sentRecommendations, bson.M{
		"sender_id": objID,
	})
	if err == nil {
		sentKey := GetCacheKey("user_recommendations", userID, "sent")
		SetCache(sentKey, sentRecommendations)
	}
}

// UpdateReceivedRecommendationsCacheByEmail updates cache for received recommendations by email
func UpdateReceivedRecommendationsCacheByEmail(email string) {
	// Find user by email to get their ID
	var user models.User
	err := mgm.Coll(&user).FindOne(mgm.Ctx(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		return // User doesn't exist in our system yet
	}

	var receivedRecommendations []models.Recommendation
	err = mgm.Coll(&models.Recommendation{}).SimpleFind(&receivedRecommendations, bson.M{
		"recipient_email": email,
	})
	if err == nil {
		receivedKey := GetCacheKey("user_recommendations", user.ID.Hex(), "received")
		SetCache(receivedKey, receivedRecommendations)
	}
}

// UpdateListingsCache updates the cache for user's listings
func UpdateListingsCache(userID string) {
	var listings []models.Property
	cursor, err := mgm.Coll(&models.Property{}).Find(mgm.Ctx(), bson.M{
		"created_by": userID,
	}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return
	}
	defer cursor.Close(mgm.Ctx())

	if err = cursor.All(mgm.Ctx(), &listings); err != nil {
		return
	}

	listingsKey := GetCacheKey("user_listings", userID, "")
	SetCache(listingsKey, listings)
}
