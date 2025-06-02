package seed

import (
	"ePrometna_Server/config"
	"ePrometna_Server/model"
	"ePrometna_Server/service"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	_TEST_PASSWORD = "Pa$$w0rd"
)

func Insert() {
	if err := createSuperAdmin(); err != nil {
		zap.S().DPanicf("Failed to create superadmin, err = %+v\n", err)
	}

	if config.AppConfig.Env == config.Test {
		if err := createUser(); err != nil {
			zap.S().DPanicf("Failed to create user, err = %+v\n", err)
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
		BirthDate: time.Now().AddDate(-20, 0, 0),
		Uuid:      uuid.New(),
	}

	user, err := userCrud.Create(&newUser, _TEST_PASSWORD)
	if err != nil {
		return err
	}

	zap.S().Infof("User created, %+v\n", user)
	return nil
}
