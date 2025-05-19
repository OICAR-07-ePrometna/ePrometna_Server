package dto

import (
	"time"
)

type ShareVehicleDto struct {
	VehicleUuid string `json:"vehicleUuid" binding:"required"`
	UserUuid    string `json:"driverUuid" binding:"required"`
	Until       string `json:"until"` // optional expiration date in YYYY-MM-DD format
}

func (dto *ShareVehicleDto) ParseUntil() (*time.Time, error) {
	if dto.Until == "" {
		return nil, nil // No expiration date
	}
	parsedDate, err := time.Parse("2006-01-02", dto.Until)
	if err != nil {
		return nil, err
	}

	return &parsedDate, nil
}
