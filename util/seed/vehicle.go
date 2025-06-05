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
	return nil
}

func createSecondVehicle() error {
	vservice := service.NewVehicleService()

	vehicleInfo := model.Vehicle{
		Uuid:                                   uuid.New(),
		VehicleType:                            "Car",
		VehicleModel:                           "A4",
		Mark:                                   "Audi",
		VehicleCategory:                        "M1",
		ChassisNumber:                          "WAUZZZ8K9BA123456",
		BodyShape:                              "Sedan",
		VehicleUse:                             "Personal",
		DateFirstRegistration:                  "2021-03-20",
		EngineCapacity:                         "1968",
		EnginePower:                            "150",
		FuelOrPowerSource:                      "Diesel",
		NumberOfSeats:                          "5",
		ColourOfVehicle:                        "Black",
		HomologationType:                       "e1*2007/46*0063*00",
		TradeName:                              "Audi A4 2.0 TDI",
		FirstRegistrationInCroatia:             "2021-03-20",
		TechnicallyPermissibleMaximumLadenMass: "2100",
		PermissibleMaximumLadenMass:            "2100",
		UnladenMass:                            "1650",
		PermissiblePayload:                     "450",
		TypeApprovalNumber:                     "e1*2007/46*0063*00",
		RatedEngineSpeed:                       "4200",
		Length:                                 "4762",
		Width:                                  "1847",
		Height:                                 "1436",
		MaximumNetPower:                        "110",
		NumberOfAxles:                          "2",
		NumberOfDrivenAxles:                    "2",
		Mb:                                     "Audi",
		StationaryNoiseLevel:                   "71",
		EngineSpeedForStationaryNoiseTest:      "2750",
		Co2Emissions:                           "114",
		EcCategory:                             "M1",
		TireSize:                               "225/50R17",
		UniqueModelCode:                        "8K",
		AdditionalTireSizes:                    "225/45R18",
		Registration: &model.RegistrationInfo{
			Uuid:             uuid.New(),
			PassTechnical:    true,
			TraveledDistance: 15000,
			TechnicalDate:    time.Now().AddDate(0, 8, 0),
			Registration:     "ZG5678BB",
		},
	}

	newVehicle, err := vservice.Create(&vehicleInfo, osoba.Uuid)
	if err != nil {
		return err
	}

	zap.S().Infof("Second vehicle created, %+v\n", newVehicle)
	return nil
}
