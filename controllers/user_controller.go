package controllers

import (
	"os"
	"time"

	"property_lister/models"
	"property_lister/services"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

// Request structures
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Phone     string `json:"phone"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Response structures (belongs in presentation layer)
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message,omitempty"`
	Data    *AuthData `json:"data,omitempty"`
}

type AuthData struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

// toUserResponse converts User model to UserResponse DTO
func toUserResponse(user *models.User) UserResponse {
	return UserResponse{
		ID:        user.ID.Hex(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		CreatedAt: user.CreatedAt,
	}
}

func RegisterUser(c *fiber.Ctx) error {
	var req RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(AuthResponse{
			Success: false,
			Message: "Invalid request body",
		})
	}

	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return c.Status(400).JSON(AuthResponse{
			Success: false,
			Message: "Email, password, first name, and last name are required",
		})
	}

	if len(req.Password) < 6 {
		return c.Status(400).JSON(AuthResponse{
			Success: false,
			Message: "Password must be at least 6 characters long",
		})
	}

	var existingUser models.User
	err := mgm.Coll(&existingUser).FindOne(mgm.Ctx(), bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		return c.Status(409).JSON(AuthResponse{
			Success: false,
			Message: "User with this email already exists",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(AuthResponse{
			Success: false,
			Message: "Failed to hash password",
		})
	}

	user := &models.User{
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := mgm.Coll(user).Create(user); err != nil {
		return c.Status(500).JSON(AuthResponse{
			Success: false,
			Message: "Failed to create user",
		})
	}

	token, err := generateJWTToken(user.ID.Hex(), user.Email)
	if err != nil {
		return c.Status(500).JSON(AuthResponse{
			Success: false,
			Message: "Failed to generate token",
		})
	}

	return c.Status(201).JSON(AuthResponse{
		Success: true,
		Message: "User registered successfully",
		Data: &AuthData{
			User:  toUserResponse(user),
			Token: token,
		},
	})
}

func LoginUser(c *fiber.Ctx) error {
	var req LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(AuthResponse{
			Success: false,
			Message: "Invalid request body",
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(400).JSON(AuthResponse{
			Success: false,
			Message: "Email and password are required",
		})
	}

	var user models.User
	err := mgm.Coll(&user).FindOne(mgm.Ctx(), bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		return c.Status(401).JSON(AuthResponse{
			Success: false,
			Message: "Invalid email or password",
		})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return c.Status(401).JSON(AuthResponse{
			Success: false,
			Message: "Invalid email or password",
		})
	}

	token, err := generateJWTToken(user.ID.Hex(), user.Email)
	if err != nil {
		return c.Status(500).JSON(AuthResponse{
			Success: false,
			Message: "Failed to generate token",
		})
	}

	// Update user last login time
	user.UpdatedAt = time.Now()
	mgm.Coll(&user).UpdateOne(mgm.Ctx(), bson.M{"_id": user.ID}, bson.M{"$set": bson.M{"updated_at": user.UpdatedAt}})

	// Load and cache user data on login
	go services.CacheUserDataOnLogin(user.ID.Hex(), user.Email)

	return c.JSON(AuthResponse{
		Success: true,
		Message: "Login successful",
		Data: &AuthData{
			User:  toUserResponse(&user),
			Token: token,
		},
	})
}

func generateJWTToken(userID, email string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")

	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
