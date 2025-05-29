package data_ingestion

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"property_lister/models"

	"github.com/kamva/mgm/v3"
)

func IngestCSV(csvPath string) error {
	absPath, err := filepath.Abs(csvPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read csv: %w", err)
	}

	for i, row := range records {
		if i == 0 {
			continue
		}

		price, _ := strconv.Atoi(row[3])
		area, _ := strconv.Atoi(row[6])
		bed, _ := strconv.Atoi(row[7])
		bath, _ := strconv.Atoi(row[8])
		rating, _ := strconv.ParseFloat(row[15], 64)
		isVerified := row[16] == "True"

		now := time.Now()

		property := &models.Property{
			ID:            row[0],
			Title:         row[1],
			Type:          row[2],
			Price:         price,
			State:         row[4],
			City:          row[5],
			AreaSqFt:      area,
			Bedrooms:      bed,
			Bathrooms:     bath,
			Amenities:     strings.Split(row[9], "|"),
			Furnished:     row[10],
			AvailableFrom: row[11],
			ListedBy:      row[12],
			Tags:          strings.Split(row[13], "|"),
			ColorTheme:    row[14],
			Rating:        rating,
			IsVerified:    isVerified,
			ListingType:   row[17],
			CreatedBy:     "SYSTEM",
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if err := mgm.Coll(property).Create(property); err != nil {
			fmt.Printf("Failed to insert row %d: %v\n", i, err)
		} else {
			fmt.Printf("Inserted property ID %s\n", property.ID)
		}
	}

	return nil
}
