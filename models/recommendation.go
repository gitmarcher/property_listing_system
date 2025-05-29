package models

import (
	"time"

	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Recommendation struct {
	mgm.DefaultModel `bson:",inline"`

	PropertyID     primitive.ObjectID `json:"property_id" bson:"property_id" validate:"required"`
	SenderID       primitive.ObjectID `json:"sender_id" bson:"sender_id" validate:"required"`
	RecipientEmail string             `json:"recipient_email" bson:"recipient_email" validate:"required,email"`
	Message        string             `json:"message" bson:"message"`
	Status         string             `json:"status" bson:"status"` // pending, accepted, rejected
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
}
