package dto

import (
	"ePrometna_Server/model"

	"github.com/google/uuid"
)

type NewVehicleDto struct {
	OwnerUuid        string         `json:"ownerUuid"`
	Registration     string         `json:"registration"`
	TraveledDistance int            `json:"traveledDistance"`
	Summary          VehicleSummary `json:"summary"`
}

func (dto *NewVehicleDto) ToModel() (*model.Vehicle, error) {
	return &model.Vehicle{
		Uuid:                                   uuid.New(),
		VehicleType:                            dto.Summary.VehicleType,
		VehicleModel:                           dto.Summary.Model,
		ChassisNumber:                          dto.Summary.ChassisNumber,
		VehicleCategory:                        dto.Summary.VehicleCategory,
		Mark:                                   dto.Summary.Mark,
		HomologationType:                       dto.Summary.HomologationType,
		TradeName:                              dto.Summary.TradeName,
		BodyShape:                              dto.Summary.BodyShape,
		VehicleUse:                             dto.Summary.VehicleUse,
		DateFirstRegistration:                  dto.Summary.DateFirstRegistration,
		FirstRegistrationInCroatia:             dto.Summary.FirstRegistrationInCroatia,
		TechnicallyPermissibleMaximumLadenMass: dto.Summary.TechnicallyPermissibleMaximumLadenMass,
		PermissibleMaximumLadenMass:            dto.Summary.PermissibleMaximumLadenMass,
		UnladenMass:                            dto.Summary.UnladenMass,
		PermissiblePayload:                     dto.Summary.PermissiblePayload,
		TypeApprovalNumber:                     dto.Summary.TypeApprovalNumber,
		EngineCapacity:                         dto.Summary.EngineCapacity,
		EnginePower:                            dto.Summary.EnginePower,
		FuelOrPowerSource:                      dto.Summary.FuelOrPowerSource,
		RatedEngineSpeed:                       dto.Summary.RatedEngineSpeed,
		NumberOfSeats:                          dto.Summary.NumberOfSeats,
		ColourOfVehicle:                        dto.Summary.ColourOfVehicle,
		Length:                                 dto.Summary.Length,
		Width:                                  dto.Summary.Width,
		Height:                                 dto.Summary.Height,
		MaximumNetPower:                        dto.Summary.MaximumNetPower,
		NumberOfAxles:                          dto.Summary.NumberOfAxles,
		NumberOfDrivenAxles:                    dto.Summary.NumberOfDrivenAxles,
		Mb:                                     dto.Summary.Mb,
		StationaryNoiseLevel:                   dto.Summary.StationaryNoiseLevel,
		EngineSpeedForStationaryNoiseTest:      dto.Summary.EngineSpeedForStationaryNoiseTest,
		Co2Emissions:                           dto.Summary.Co2Emissions,
		EcCategory:                             dto.Summary.EcCategory,
		TireSize:                               dto.Summary.TireSize,
		UniqueModelCode:                        dto.Summary.UniqueModelCode,
		AdditionalTireSizes:                    dto.Summary.AdditionalTireSizes,

		Registration: &model.RegistrationInfo{
			Uuid:             uuid.New(),
			PassTechnical:    true,
			TraveledDistance: dto.TraveledDistance,
			Registration:     dto.Registration,
		},
	}, nil
}
