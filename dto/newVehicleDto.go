package dto

import (
	"ePrometna_Server/model"

	"github.com/google/uuid"
)

type NewVehicleDto struct {
	VehicleType      string
	VehicleModel     string
	ProductionYear   int
	ChassisNumber    string
	OwnerUuid        string
	Registration     string
	TraveledDistance int
}

// ToModel create a model from a dto
func (dto *NewVehicleDto) ToModel() (*model.Vehicle, error) {
	return &model.Vehicle{
		Uuid:           uuid.New(),
		VehicleType:    dto.VehicleType,
		VehicleModel:   dto.VehicleModel,
		ProductionYear: dto.ProductionYear,
		ChassisNumber:  dto.ChassisNumber,

		Registration: &model.RegistrationInfo{
			Uuid:             uuid.New(),
			PassTechnical:    true,
			TraveledDistance: dto.TraveledDistance,
			Registration:     dto.Registration,
		},
	}, nil
}
