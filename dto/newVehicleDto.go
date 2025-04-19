package dto

import (
	"ePrometna_Server/model"

	"github.com/google/uuid"
)

type NewVehicleDto struct {
	OwnerUuid                              string `json:"ownerUuid"`
	Registration                           string `json:"registration"`
	TraveledDistance                       int    `json:"traveledDistance"`
	VehicleCategory                        string `json:"vehicleCategory"`                        // Kategorija vozila // J
	Mark                                   string `json:"mark"`                                   // Marka // D1
	Model                                  string `json:"model"`                                  // Model // (14)
	HomologationType                       string `json:"homologationType"`                       // Homologacijski tip // D2
	TradeName                              string `json:"tradeName"`                              // Trgovački naziv // D3
	ChassisNumber                          string `json:"chassisNumber"`                          // Broj šasije // E
	BodyShape                              string `json:"bodyShape"`                              // Oblik karoserije // (2)
	VehicleUse                             string `json:"vehicleUse"`                             // Namjena vozila // (3)
	DateFirstRegistration                  string `json:"dateFirstRegistration"`                  // Datum prve registracije // B
	FirstRegistrationInCroatia             string `json:"firstRegistrationInCroatia"`             // Prva registracija u Hrvatskoj // (4)
	TechnicallyPermissibleMaximumLadenMass string `json:"technicallyPermissibleMaximumLadenMass"` // Tehnički dopuštena najveća masa // F1
	PermissibleMaximumLadenMass            string `json:"permissibleMaximumLadenMass"`            // Dopuštena najveća masa // F2
	UnladenMass                            string `json:"unladenMass"`                            // Masa praznog vozila // G
	PermissiblePayload                     string `json:"permissiblePayload"`                     // Dopuštena nosivost // (5)
	TypeApprovalNumber                     string `json:"typeApprovalNumber"`                     // Broj homologacije // K
	EngineCapacity                         string `json:"engineCapacity"`                         // Obujam motora // P1
	EnginePower                            string `json:"enginePower"`                            // Snaga motora // P2
	FuelOrPowerSource                      string `json:"fuelOrPowerSource"`                      // Gorivo ili izvor energije // P3
	RatedEngineSpeed                       string `json:"ratedEngineSpeed"`                       // Nazivni broj okretaja motora // P4
	NumberOfSeats                          string `json:"numberOfSeats"`                          // Broj sjedala // S1
	ColourOfVehicle                        string `json:"colourOfVehicle"`                        // Boja vozila // R
	Length                                 string `json:"length"`                                 // Dužina // (6)
	Width                                  string `json:"width"`                                  // Širina // (7)
	Height                                 string `json:"height"`                                 // Visina // (8)
	MaximumNetPower                        string `json:"maximumNetPower"`                        // Najveća neto snaga // T
	NumberOfAxles                          string `json:"numberOfAxles"`                          // Broj osovina // L
	NumberOfDrivenAxles                    string `json:"numberOfDrivenAxles"`                    // Broj pogonskih osovina // (9)
	Mb                                     string `json:"mb"`                                     // MB (pretpostavka: proizvođač) // (13)
	StationaryNoiseLevel                   string `json:"stationaryNoiseLevel"`                   // Razina buke u stacionarnom stanju // U1
	EngineSpeedForStationaryNoiseTest      string `json:"engineSpeedForStationaryNoiseTest"`      // Broj okretaja motora pri ispitivanju buke u stacionarnom stanju // U2
	Co2Emissions                           string `json:"co2Emissions"`                           // Emisija CO2 // V7
	EcCategory                             string `json:"ecCategory"`                             // EC kategorija // V9
	TireSize                               string `json:"tireSize"`                               // Dimenzije guma // (11)
	UniqueModelCode                        string `json:"uniqueModelCode"`                        // Jedinstvena oznaka modela // (12)
	AdditionalTireSizes                    string `json:"additionalTireSizes"`                    // Dodatne dimenzije guma // (15)
	VehicleType                            string `json:"vehicleType"`                            // Tip vozila (16) // (16)
}

func (dto *NewVehicleDto) ToModel() (*model.Vehicle, error) {
	return &model.Vehicle{
		Uuid:                                   uuid.New(),
		VehicleType:                            dto.VehicleType,
		VehicleModel:                           dto.Model,
		ChassisNumber:                          dto.ChassisNumber,
		VehicleCategory:                        dto.VehicleCategory,
		Mark:                                   dto.Mark,
		HomologationType:                       dto.HomologationType,
		TradeName:                              dto.TradeName,
		BodyShape:                              dto.BodyShape,
		VehicleUse:                             dto.VehicleUse,
		DateFirstRegistration:                  dto.DateFirstRegistration,
		FirstRegistrationInCroatia:             dto.FirstRegistrationInCroatia,
		TechnicallyPermissibleMaximumLadenMass: dto.TechnicallyPermissibleMaximumLadenMass,
		PermissibleMaximumLadenMass:            dto.PermissibleMaximumLadenMass,
		UnladenMass:                            dto.UnladenMass,
		PermissiblePayload:                     dto.PermissiblePayload,
		TypeApprovalNumber:                     dto.TypeApprovalNumber,
		EngineCapacity:                         dto.EngineCapacity,
		EnginePower:                            dto.EnginePower,
		FuelOrPowerSource:                      dto.FuelOrPowerSource,
		RatedEngineSpeed:                       dto.RatedEngineSpeed,
		NumberOfSeats:                          dto.NumberOfSeats,
		ColourOfVehicle:                        dto.ColourOfVehicle,
		Length:                                 dto.Length,
		Width:                                  dto.Width,
		Height:                                 dto.Height,
		MaximumNetPower:                        dto.MaximumNetPower,
		NumberOfAxles:                          dto.NumberOfAxles,
		NumberOfDrivenAxles:                    dto.NumberOfDrivenAxles,
		Mb:                                     dto.Mb,
		StationaryNoiseLevel:                   dto.StationaryNoiseLevel,
		EngineSpeedForStationaryNoiseTest:      dto.EngineSpeedForStationaryNoiseTest,
		Co2Emissions:                           dto.Co2Emissions,
		EcCategory:                             dto.EcCategory,
		TireSize:                               dto.TireSize,
		UniqueModelCode:                        dto.UniqueModelCode,
		AdditionalTireSizes:                    dto.AdditionalTireSizes,

		Registration: &model.RegistrationInfo{
			Uuid:             uuid.New(),
			PassTechnical:    true,
			TraveledDistance: dto.TraveledDistance,
			Registration:     dto.Registration,
		},
	}, nil
}
