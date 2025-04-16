package dto

import (
	"ePrometna_Server/model"
	"fmt"

	"github.com/google/uuid"
)

// TODO: add more properties
type VehicleDetailsDto struct {
	Uuid           string    `json:"uuid"`
	VehicleType    string    `json:"vehicleType"`
	VehicleModel   string    `json:"vehicleModel"`
	ProductionYear int       `json:"productionYear"`
	Registration   string    `json:"registration"`
	Owner          UserDto   `json:"owner"`
	Drivers        []UserDto `json:"drivers"`
	PastOwners     []UserDto `json:"pastOwners"`
	// Registration   RegistrationDto
	// PastRegistratins []RegistrationDto
}

// ToModel create a model from a dto
func (dto *VehicleDetailsDto) ToModel() (*model.Vehicle, error) {
	uuid, err := uuid.Parse(dto.Uuid)
	if err != nil {
		return nil, fmt.Errorf("invalid vehicle UUID: %w", err)
	}

	// Create a basic vehicle model
	vehicle := &model.Vehicle{
		Uuid:           uuid,
		VehicleType:    dto.VehicleType,
		VehicleModel:   dto.VehicleModel,
		ProductionYear: dto.ProductionYear,
	}

	// TODO: Converting Owner, Drivers, and PastOwners would require additional logic
	// to convert UserDto to User models

	return vehicle, nil
}

func (dto VehicleDetailsDto) FromModel(m *model.Vehicle) VehicleDetailsDto {
	result := VehicleDetailsDto{
		Uuid:           m.Uuid.String(),
		VehicleType:    m.VehicleType,
		VehicleModel:   m.VehicleModel,
		ProductionYear: m.ProductionYear,
	}

	// Add registration if available
	if m.Registration != nil {
		result.Registration = m.Registration.Registration
	}

	if m.Owner != nil {
		result.Owner = UserDto{}.FromModel(m.Owner)
	}

	// TODO: Add owner, drivers, and past owners conversion logic here
	// This would require retrieving related user information and converting to UserDto

	return result
}
