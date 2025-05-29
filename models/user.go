package models

import (
	"time"

	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	mgm.DefaultModel `bson:",inline"`

	Email                   string               `json:"email" bson:"email" validate:"required,email"`
	Password                string               `json:"-" bson:"password" validate:"required,min=6"`
	FirstName               string               `json:"first_name" bson:"first_name" validate:"required"`
	LastName                string               `json:"last_name" bson:"last_name" validate:"required"`
	Phone                   string               `json:"phone" bson:"phone"`
	CreatedAt               time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt               time.Time            `json:"updated_at" bson:"updated_at"`
	Favorites               []string             `json:"favorites" bson:"favorites"`
	RecommendationsSent     []primitive.ObjectID `json:"recommendations_sent" bson:"recommendations_sent"`
	RecommendationsReceived []primitive.ObjectID `json:"recommendations_received" bson:"recommendations_received"`
}
