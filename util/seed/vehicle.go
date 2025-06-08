package seed

import (
	"ePrometna_Server/model"
	"ePrometna_Server/service"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func createVehicle() error {
	vservice := service.NewVehicleService()

	vehicleInfo := model.Vehicle{
		Uuid:                  uuid.New(),
		VehicleType:           "Car",
		VehicleModel:          "Golf",
		Mark:                  "Volkswagen",
		VehicleCategory:       "M1",
		ChassisNumber:         "WVWZZZ1KZAW123456",
		BodyShape:             "Hatchback",
		VehicleUse:            "Personal",
		DateFirstRegistration: "2020-01-15",
		EngineCapacity:        "1598",
		EnginePower:           "110",
		FuelOrPowerSource:     "Petrol",
		NumberOfSeats:         "5",
		ColourOfVehicle:       "Silver",
		Registration: &model.RegistrationInfo{
			Uuid:             uuid.New(),
			PassTechnical:    true,
			TraveledDistance: 25000,
			TechnicalDate:    time.Now().AddDate(0, 6, 0),
			Registration:     "ZG1234AA",
		},
	}

	newVehicle, err := vservice.Create(&vehicleInfo, osoba.Uuid)
	if err != nil {
		return err
	}

	zap.S().Infof("Vehicle created, %+v\n", newVehicle)
	vehicle = newVehicle

	vehicleInfo2 := model.Vehicle{
		Uuid:                  uuid.New(),
		VehicleType:           "Car",
		VehicleModel:          "Golf",
		Mark:                  "Volkswagen",
		VehicleCategory:       "M1",
		ChassisNumber:         "WVWYYY1KZAW123456",
		BodyShape:             "Hatchback",
		VehicleUse:            "Personal",
		DateFirstRegistration: "2020-01-15",
		EngineCapacity:        "1598",
		EnginePower:           "110",
		FuelOrPowerSource:     "Petrol",
		NumberOfSeats:         "5",
		ColourOfVehicle:       "Silver",
		Registration: &model.RegistrationInfo{
			Uuid:             uuid.New(),
			PassTechnical:    true,
			TraveledDistance: 25000,
			TechnicalDate:    time.Now().AddDate(0, 6, 0),
			Registration:     "ZG1234BB",
		},
	}

	newVehicle2, err := vservice.Create(&vehicleInfo2, osoba3.Uuid)
	if err != nil {
		return err
	}

	zap.S().Infof("Vehicle created, %+v\n", newVehicle2)
	vehicle2 = newVehicle2
	return nil
}
