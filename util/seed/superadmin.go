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
	_PASSWORD_ENV = "SUPERADMIN_PASSWORD"
	_OIB          = "11111111111"
)

// CreateSuperAdmin creates a SuperAdmin user if one doesn't already exist.
// It reads the password from the SUPERADMIN_PASSWORD environment variable.
// The function will panic if required environment variables are missing or
// if user creation fails, as this is critical for application bootstrap.
func CreateSuperAdmin() {
	userCrud := service.NewUserCrudService()

	// Check if SuperAdmin exists
	{
		_, err := userCrud.GetUserByOIB(_OIB)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				zap.S().Infof("SuperAdmin not found, err %+v", err)
			} else {
				zap.S().DPanicf("Error reading users, err %+v", err)
				return
			}
		} else {
			zap.S().Infoln("SuperAdmin found")
			zap.S().Infoln("Skipping superadmin creation")
			return
		}
	}

	zap.S().Infoln("Crating superadmin creation")
	password := os.Getenv(_PASSWORD_ENV)
	if password == "" {
		zap.S().DPanicf("Env variable %s is empty\n", _PASSWORD_ENV)
		return
	}
	if len(password) < 8 {
		zap.S().DPanicln("SuperAdmin password must be at least 8 characters long")
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
		zap.S().DPanicf("Failed to create superadmin, err = %+v\n", err)
		return
	}
	zap.S().Infof("superadmin created, %+v\n", user)
}
