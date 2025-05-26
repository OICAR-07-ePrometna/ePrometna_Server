package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Vehicle struct {
	gorm.Model
	Uuid             uuid.UUID          `gorm:"type:uuid;unique;not null"`
	UserId           *uint              `gorm:"type:uint;null"`
	Owner            *User              `gorm:"foreignKey:UserId;OnDelete:SET NULL"`
	Drivers          []VehicleDrivers   `gorm:"foreignKey:VehicleId;null"`
	PastOwners       []OwnerHistory     `gorm:"foreignKey:VehicleId;null"`
	TemporaryData    *TempData          `gorm:"foreignKey:VehicleId;null"`
	Registration     *RegistrationInfo  `gorm:"null"`
	PastRegistration []RegistrationInfo `gorm:"foreignKey:VehicleId;null"`
	RegistrationID   *uint

	VehicleCategory                        string // Kategorija vozila // J
	Mark                                   string // Marka // D1
	VehicleModel                           string // Model // (14) NOTE: must NOT be model becouse of gorm.Model
	HomologationType                       string // Homologacijski tip // D2
	TradeName                              string // Trgovački naziv // D3
	ChassisNumber                          string // Broj šasije // E
	BodyShape                              string // Oblik karoserije // (2)
	VehicleUse                             string // Namjena vozila // (3)
	DateFirstRegistration                  string // Datum prve registracije // B
	FirstRegistrationInCroatia             string // Prva registracija u Hrvatskoj // (4)
	TechnicallyPermissibleMaximumLadenMass string // Tehnički dopuštena najveća masa // F1
	PermissibleMaximumLadenMass            string // Dopuštena najveća masa // F2
	UnladenMass                            string // Masa praznog vozila // G
	PermissiblePayload                     string // Dopuštena nosivost // (5)
	TypeApprovalNumber                     string // Broj homologacije // K
	EngineCapacity                         string // Obujam motora // P1
	EnginePower                            string // Snaga motora // P2
	FuelOrPowerSource                      string // Gorivo ili izvor energije // P3
	RatedEngineSpeed                       string // Nazivni broj okretaja motora // P4
	NumberOfSeats                          string // Broj sjedala // S1
	ColourOfVehicle                        string // Boja vozila // R
	Length                                 string // Dužina // (6)
	Width                                  string // Širina // (7)
	Height                                 string // Visina // (8)
	MaximumNetPower                        string // Najveća neto snaga // T
	NumberOfAxles                          string // Broj osovina // L
	NumberOfDrivenAxles                    string // Broj pogonskih osovina // (9)
	Mb                                     string // MB (pretpostavka: proizvođač) // (13)
	StationaryNoiseLevel                   string // Razina buke u stacionarnom stanju // U1
	EngineSpeedForStationaryNoiseTest      string // Broj okretaja motora pri ispitivanju buke u stacionarnom stanju // U2
	Co2Emissions                           string // Emisija CO2 // V7
	EcCategory                             string // EC kategorija // V9
	TireSize                               string // Dimenzije guma // (11)
	UniqueModelCode                        string // Jedinstvena oznaka modela // (12)
	AdditionalTireSizes                    string // Dodatne dimenzije guma // (15)
	VehicleType                            string // Tip vozila (16) // (16)
}

func (v *Vehicle) Update(newVehicle Vehicle) *Vehicle {
	v.HomologationType = newVehicle.HomologationType
	v.BodyShape = newVehicle.BodyShape
	v.VehicleUse = newVehicle.VehicleUse
	v.TechnicallyPermissibleMaximumLadenMass = newVehicle.TechnicallyPermissibleMaximumLadenMass
	v.PermissibleMaximumLadenMass = newVehicle.PermissibleMaximumLadenMass
	v.UnladenMass = newVehicle.UnladenMass
	v.PermissiblePayload = newVehicle.PermissiblePayload
	v.TypeApprovalNumber = newVehicle.TypeApprovalNumber
	v.EngineCapacity = newVehicle.EngineCapacity
	v.EnginePower = newVehicle.EnginePower
	v.FuelOrPowerSource = newVehicle.FuelOrPowerSource
	v.RatedEngineSpeed = newVehicle.RatedEngineSpeed
	v.NumberOfSeats = newVehicle.NumberOfSeats
	v.ColourOfVehicle = newVehicle.ColourOfVehicle
	v.Length = newVehicle.Length
	v.Width = newVehicle.Width
	v.Height = newVehicle.Height
	v.MaximumNetPower = newVehicle.MaximumNetPower
	v.NumberOfAxles = newVehicle.NumberOfAxles
	v.NumberOfDrivenAxles = newVehicle.NumberOfDrivenAxles
	v.Mb = newVehicle.Mb
	v.StationaryNoiseLevel = newVehicle.StationaryNoiseLevel
	v.EngineSpeedForStationaryNoiseTest = newVehicle.EngineSpeedForStationaryNoiseTest
	v.Co2Emissions = newVehicle.Co2Emissions
	v.EcCategory = newVehicle.EcCategory
	v.AdditionalTireSizes = newVehicle.AdditionalTireSizes
	return v
}
