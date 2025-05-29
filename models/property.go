package models

import (
	"time"

	"github.com/kamva/mgm/v3"
)

type Property struct {
	mgm.DefaultModel `bson:",inline"`

	ID            string    `csv:"id" bson:"id"`
	Title         string    `csv:"title" bson:"title"`
	Type          string    `csv:"type" bson:"type"`
	Price         int       `csv:"price" bson:"price"`
	State         string    `csv:"state" bson:"state"`
	City          string    `csv:"city" bson:"city"`
	AreaSqFt      int       `csv:"areaSqFt" bson:"areaSqFt"`
	Bedrooms      int       `csv:"bedrooms" bson:"bedrooms"`
	Bathrooms     int       `csv:"bathrooms" bson:"bathrooms"`
	Amenities     []string  `csv:"amenities" bson:"amenities"`
	Furnished     string    `csv:"furnished" bson:"furnished"`
	AvailableFrom string    `csv:"availableFrom" bson:"availableFrom"`
	ListedBy      string    `csv:"listedBy" bson:"listedBy"`
	Tags          []string  `csv:"tags" bson:"tags"`
	ColorTheme    string    `csv:"colorTheme" bson:"colorTheme"`
	Rating        float64   `csv:"rating" bson:"rating"`
	IsVerified    bool      `csv:"isVerified" bson:"isVerified"`
	ListingType   string    `csv:"listingType" bson:"listingType"`
	CreatedBy     string    `json:"created_by" bson:"created_by"`
	CreatedAt     time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" bson:"updated_at"`
}
