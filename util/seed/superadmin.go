package seed

import (
	"ePrometna_Server/dto"
	"ePrometna_Server/service"
	"errors"
	"os"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	_NAME = "SUPERADMIN_PASSWORD"
	_OIB  = "11111111111"
)

func CreateSuperAdmin() {
	userCrud := service.NewUserCrudService()

	// Check if SuperAdmin exists
	{
		user, err := userCrud.GetUserByOIB(_OIB)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				zap.S().Errorf("SuperAdmin not found, err %+v", err)
			} else {
				zap.S().DPanicf("Error reading users, err %+v", err)
				return
			}
		} else if user.OIB == _OIB {
			zap.S().Infoln("SuperAdmin found")
			zap.S().Infoln("Skipping superadmin creation")
			return
		}
	}

	password := os.Getenv(_NAME)
	if password == "" {
		zap.S().DPanicf("Env variable %s is empty\n", _NAME)
		return
	}

	dto := dto.NewUserDto{
		FirstName: "Super",
		LastName:  "Admin",
		Email:     "superadmin@test.hr",
		Password:  password,
		BirthDate: "2000-01-01",
		Role:      "superadmin",
		OIB:       _OIB,
	}
	newUser, err := dto.ToModel()
	if err != nil {
		zap.S().DPanicf("Failed to map superadmin, err = %+v\n", err)
		return
	}

	user, err := userCrud.Create(newUser, dto.Password)
	if err != nil {
		zap.S().DPanicf("Failed to crate superadmin, err = %+v\n", err)
		return
	}
	zap.S().Infof("superadmin created, %+v\n", user)
}
