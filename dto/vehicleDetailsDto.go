package dto

import (
	"ePrometna_Server/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TODO: add more properties
type VehicleDetailsDto struct {
	Uuid             string            `json:"uuid"`
	Registration     string            `json:"registration"`
	Owner            UserDto           `json:"owner"`
	Drivers          []UserDto         `json:"drivers"`
	PastOwners       []UserDto         `json:"pastOwners"`
	PastRegistration []RegistrationDto `json:"pastRegistration"`
	Summary          VehicleSummary    `json:"summary"`
}
type VehicleSummary struct {
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

// ToModel create a model from a dto
func (dto *VehicleDetailsDto) ToModel() (*model.Vehicle, error) {
	uuid, err := uuid.Parse(dto.Uuid)
	if err != nil {
		//	return nil, fmt.Errorf("invalid vehicle UUID: %w", err)
	}

	// Create a basic vehicle model
	vehicle := &model.Vehicle{
		Uuid:                                   uuid,
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
	}

	return vehicle, nil
}

func (dto VehicleDetailsDto) FromModel(m *model.Vehicle) VehicleDetailsDto {
	result := VehicleDetailsDto{
		Uuid: m.Uuid.String(),
		Summary: VehicleSummary{
			VehicleType:                            m.VehicleType,
			Model:                                  m.VehicleModel,
			ChassisNumber:                          m.ChassisNumber,
			VehicleCategory:                        m.VehicleCategory,
			Mark:                                   m.Mark,
			HomologationType:                       m.HomologationType,
			TradeName:                              m.TradeName,
			BodyShape:                              m.BodyShape,
			VehicleUse:                             m.VehicleUse,
			DateFirstRegistration:                  m.DateFirstRegistration,
			FirstRegistrationInCroatia:             m.FirstRegistrationInCroatia,
			TechnicallyPermissibleMaximumLadenMass: m.TechnicallyPermissibleMaximumLadenMass,
			PermissibleMaximumLadenMass:            m.PermissibleMaximumLadenMass,
			UnladenMass:                            m.UnladenMass,
			PermissiblePayload:                     m.PermissiblePayload,
			TypeApprovalNumber:                     m.TypeApprovalNumber,
			EngineCapacity:                         m.EngineCapacity,
			EnginePower:                            m.EnginePower,
			FuelOrPowerSource:                      m.FuelOrPowerSource,
			RatedEngineSpeed:                       m.RatedEngineSpeed,
			NumberOfSeats:                          m.NumberOfSeats,
			ColourOfVehicle:                        m.ColourOfVehicle,
			Length:                                 m.Length,
			Width:                                  m.Width,
			Height:                                 m.Height,
			MaximumNetPower:                        m.MaximumNetPower,
			NumberOfAxles:                          m.NumberOfAxles,
			NumberOfDrivenAxles:                    m.NumberOfDrivenAxles,
			Mb:                                     m.Mb,
			StationaryNoiseLevel:                   m.StationaryNoiseLevel,
			EngineSpeedForStationaryNoiseTest:      m.EngineSpeedForStationaryNoiseTest,
			Co2Emissions:                           m.Co2Emissions,
			EcCategory:                             m.EcCategory,
			TireSize:                               m.TireSize,
			UniqueModelCode:                        m.UniqueModelCode,
			AdditionalTireSizes:                    m.AdditionalTireSizes,
		},
	}
	// Add registration if available
	if m.RegistrationID != nil {
		if m.Registration != nil {
			result.Registration = m.Registration.Registration
		}
	}

	// Add past registrations
	if len(m.PastRegistration) > 0 {
		result.PastRegistration = make([]RegistrationDto, 0, len(m.PastRegistration))
		for _, reg := range m.PastRegistration {
			note := ""
			if reg.Note != nil {
				note = *reg.Note
			}
			result.PastRegistration = append(result.PastRegistration, RegistrationDto{
				PassTechnical:    reg.PassTechnical,
				TraveledDistance: reg.TraveledDistance,
				Registration:     reg.Registration,
				Note:             note,
			})
		}
	} else {
		result.PastRegistration = []RegistrationDto{}
	}

	if m.Owner != nil {
		result.Owner = UserDto{}.FromModel(m.Owner)
	}

	// Convert drivers
	if len(m.Drivers) != 0 {
		for _, driverEntry := range m.Drivers {
			// Check if the User field within VehicleDrivers was preloaded
			if driverEntry.User.ID != 0 { // Check if User is a valid, non-zero-ID user
				var driverUserDto UserDto
				result.Drivers = append(result.Drivers, driverUserDto.FromModel(&driverEntry.User))
			} else {
				zap.S().Warnf("Vehicle UUID %s: Driver entry (VehicleDrivers ID: %d) has no preloaded User details (or User ID is 0). Skipping.", m.Uuid, driverEntry.ID)
			}
		}
	} else {
		result.Drivers = []UserDto{}
	}

	// Convert past owners
	if len(m.PastOwners) != 0 {
		for _, pastOwnerEntry := range m.PastOwners {
			// Check if the User field within OwnerHistory was preloaded
			if pastOwnerEntry.User.ID != 0 { // Check if User is a valid, non-zero-ID user
				var pastOwnerUserDto UserDto
				result.PastOwners = append(result.PastOwners, pastOwnerUserDto.FromModel(&pastOwnerEntry.User))
			} else {
				zap.S().Warnf("Vehicle UUID %s: Past owner entry (OwnerHistory ID: %d) has no preloaded User details (or User ID is 0). Skipping.", m.Uuid, pastOwnerEntry.ID)
			}
		}
	} else {
		result.PastOwners = []UserDto{}
	}
	return result
}
