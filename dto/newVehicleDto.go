package dto

import (
	"ePrometna_Server/model"

	"github.com/google/uuid"
)

type NewVehicleDto struct {
	VehicleType      string `json:"vehicleType"`
	VehicleModel     string `json:"vehicleModel"`
	ProductionYear   int    `json:"productionYear"`
	ChassisNumber    string `json:"chassisNumber"`
	OwnerUuid        string `json:"ownerUuid"`
	Registration     string `json:"registration"`
	TraveledDistance int    `json:"traveledDistance"`
}

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
