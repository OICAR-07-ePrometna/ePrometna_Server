package seed

import (
	"ePrometna_Server/app"
	"ePrometna_Server/config"
	"ePrometna_Server/model"
	"ePrometna_Server/service"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	_TEST_PASSWORD = "Pa$$w0rd"
)

var osoba *model.User
var osoba2 *model.User
var admin *model.User
var hak *model.User
var vehicle *model.Vehicle

func Insert() {
	if err := createSuperAdmin(); err != nil {
		zap.S().Panicf("Failed to create superadmin, err = %+v\n", err)
	}

	if config.AppConfig.Env == config.Test {
		if err := createUser(); err != nil {
			zap.S().Panicf("Failed to create user, err = %+v\n", err)
		}
		if err := createHakUser(); err != nil {
			zap.S().Panicf("Failed to create user(HAK), err = %+v\n", err)
		}
		if err := createDevice(); err != nil {
			zap.S().Panicf("Failed to create device, err = %+v\n", err)
		}
		if err := createVehicle(); err != nil {
			zap.S().Panicf("Failed to create vehicle, err = %+v\n", err)
		}
	}
}

func createUser() error {
	userCrud := service.NewUserCrudService()

	newUser := model.User{
		FirstName: "Test",
		LastName:  "Osoba",
		Email:     "osoba@test.hr",
		OIB:       "72352576276",
		Role:      model.RoleOsoba,
		Residence: "Zagreb",
		BirthDate: time.Now().AddDate(-20, 0, 0),
		Uuid:      uuid.New(),
	}

	user, err := userCrud.Create(&newUser, _TEST_PASSWORD)
	if err != nil {
		return err
	}

	zap.S().Infof("User created, %+v\n", user)
	osoba = user

	newUser2 := model.User{
		FirstName: "Test",
		LastName:  "Osoba",
		Email:     "osoba@test.hr",
		OIB:       "72352576276",
		Role:      model.RoleOsoba,
		Residence: "Zagreb",
		BirthDate: time.Now().AddDate(-20, 0, 0),
		Uuid:      uuid.New(),
	}

	user2, err := userCrud.Create(&newUser2, _TEST_PASSWORD)
	if err != nil {
		return err
	}

	zap.S().Infof("User2 created, %+v\n", user)
	osoba2 = user2

	return nil
}

func createHakUser() error {
	userCrud := service.NewUserCrudService()

	newUser := model.User{
		FirstName: "hak",
		LastName:  "hakovac",
		Email:     "hak@test.hr",
		OIB:       "30998630164",
		Role:      model.RoleHAK,
		Residence: "Zagreb",
		BirthDate: time.Now().AddDate(-20, 0, 0),
		Uuid:      uuid.New(),
	}

	user, err := userCrud.Create(&newUser, _TEST_PASSWORD)
	if err != nil {
		return err
	}

	zap.S().Infof("User (HAK) created, %+v\n", user)
	hak = user
	return nil
}

func createDevice() error {
	var db *gorm.DB
	app.Invoke(func(database *gorm.DB) {
		db = database
	})

	deviceInfo := model.Mobile{
		Uuid:             uuid.MustParse("495201e5-dc8c-4f3f-af41-66fcdd5e6778"),
		UserId:           osoba.ID,
		CreatorId:        osoba.ID,
		RegisteredDevice: "Google Pixel 9 (Android)",
		ActivationToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAdGVzdC5jb20iLCJ1dWlkIjoiNDk1MjAxZTUtZGM4Yy00ZjNmLWFmNDEtNjZmY2RkNWU2Nzc4Iiwicm9sZSI6Ik9TT0JBIiwiZXhwIjoxNzA5ODc2NDAwLCJpYXQiOjE3MDcyODQ0MDB9.TEST_SIGNATURE",
	}

	if err := db.Create(&deviceInfo).Error; err != nil {
		return err
	}
	return nil
}
