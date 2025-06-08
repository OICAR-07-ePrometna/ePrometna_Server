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
var osoba3 *model.User
var admin *model.User
var hak *model.User
var mup *model.User
var officer *model.User
var officer2 *model.User
var vehicle *model.Vehicle
var vehicle2 *model.Vehicle

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
		if err := createMupAdminUser(); err != nil {
			zap.S().Panicf("Failed to create user(MUP admin), err = %+v\n", err)
		}
		if err := createMupOfficerUser(); err != nil {
			zap.S().Panicf("Failed to create user(MUP officer), err = %+v\n", err)
		}
		if err := createDevice(); err != nil {
			zap.S().Panicf("Failed to create device, err = %+v\n", err)
		}
		if err := createVehicles(); err != nil {
			zap.S().Panicf("Failed to create vehicle, err = %+v\n", err)
		}
		if err := CreateTempData(); err != nil {
			zap.S().Panicf("Failed to temp data, err = %+v\n", err)
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
		FirstName: "Test2",
		LastName:  "Osoba2",
		Email:     "osoba2@test.hr",
		OIB:       "89190011773",
		Role:      model.RoleOsoba,
		Residence: "Zagreb",
		BirthDate: time.Now().AddDate(-20, 0, 0),
		Uuid:      uuid.New(),
	}

	user2, err := userCrud.Create(&newUser2, _TEST_PASSWORD)
	if err != nil {
		return err
	}

	zap.S().Infof("User2 created, %+v\n", user2)
	osoba2 = user2

	newUser3 := model.User{
		FirstName: "Test3",
		LastName:  "Osoba3",
		Email:     "osoba3@test.hr",
		OIB:       "02535077085",
		Role:      model.RoleOsoba,
		Residence: "Zagreb",
		BirthDate: time.Now().AddDate(-20, 0, 0),
		Uuid:      uuid.New(),
	}

	user3, err := userCrud.Create(&newUser3, _TEST_PASSWORD)
	if err != nil {
		return err
	}

	zap.S().Infof("User2 created, %+v\n", user3)
	osoba3 = user3

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

func createMupAdminUser() error {
	userCrud := service.NewUserCrudService()

	newUser := model.User{
		FirstName: "mup",
		LastName:  "mupovac",
		Email:     "mup@test.hr",
		OIB:       "18558015701",
		Role:      model.RoleMupADMIN,
		Residence: "Zagreb",
		BirthDate: time.Now().AddDate(-20, 0, 0),
		Uuid:      uuid.New(),
	}

	user, err := userCrud.Create(&newUser, _TEST_PASSWORD)
	if err != nil {
		return err
	}

	zap.S().Infof("User (MUP admin) created, %+v\n", user)
	mup = user
	return nil
}

func createMupOfficerUser() error {
	userCrud := service.NewUserCrudService()
	token := "ABCDEFGH"
	newUser := model.User{
		FirstName: "officer",
		LastName:  "mup Officer",
		Email:     "mupOfficer@test.hr",
		OIB:       "22978358568",
		Role:      model.RolePolicija,
		Residence: "Zagreb",
		BirthDate: time.Now().AddDate(-20, 0, 0),
		Uuid:      uuid.New(),
	}

	user, err := userCrud.Create(&newUser, _TEST_PASSWORD)
	if err != nil {
		return err
	}

	zap.S().Infof("User (MUP officer) created, %+v\n", user)
	officer2 = user

	newUser2 := model.User{
		FirstName:   "officer",
		LastName:    "mup Officer2",
		Email:       "mupOfficer2@test.hr",
		OIB:         "12308831323",
		Role:        model.RolePolicija,
		Residence:   "Zagreb",
		BirthDate:   time.Now().AddDate(-20, 0, 0),
		Uuid:        uuid.New(),
		PoliceToken: &token,
	}

	user2, err := userCrud.Create(&newUser2, _TEST_PASSWORD)
	if err != nil {
		return err
	}

	zap.S().Infof("User (MUP officer2) created, %+v\n", user)
	officer = user2
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

func CreateTempData() error {
	var db *gorm.DB
	app.Invoke(func(database *gorm.DB) {
		db = database
	})

	tempData := model.TempData{
		Uuid:      uuid.MustParse("8f22daa2-6331-4dab-9ad2-636e7df18a14"),
		DriverId:  osoba3.ID,
		VehicleId: vehicle2.ID,
		Expiring:  time.Now().Add(10 * 365 * 24 * time.Hour), // expires in 10 years
	}

	if err := db.Create(&tempData).Error; err != nil {
		return err
	}
	zap.S().Infof("Create new temp data %+v", tempData)
	return nil
}
