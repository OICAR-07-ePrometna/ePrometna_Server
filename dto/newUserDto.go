package dto

import (
	"ePrometna_Server/model"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NOTE: date format
const DateFormat = "2006-01-02"

type NewUserDto struct {
	Uuid      string
	FirstName string
	LastName  string
	OIB       string
	Residence string
	BirthDate string
	Email     string
	Password  string
	Role      string
	License   DriverLicenseDto
}

// ToModel create a model from a dto
func (dto *NewUserDto) ToModel() *model.User {
	bod, err := time.Parse(DateFormat, dto.BirthDate)
	if err != nil {
		zap.S().DPanicf("Bad date time format need %s has %s", DateFormat, dto.BirthDate)
		// TODO: see what to have to happen in prod
		return nil
	}
	role := model.RoleSuperAdmin

	return &model.User{
		Uuid:      uuid.New(),
		FirstName: dto.FirstName,
		LastName:  dto.LastName,
		OIB:       dto.OIB,
		Residence: dto.Residence,
		BirthDate: bod,
		Email:     dto.Email,
		Role:      role,
		// TODO: see what to do with license
	}
}

// FromModel returns a dto from model struct
func (dto *NewUserDto) FromModel(m *model.User) *NewUserDto {
	license := DriverLicenseDto{}
	dto = &NewUserDto{
		Uuid:      m.Uuid.String(),
		FirstName: m.FirstName,
		LastName:  m.LastName,
		OIB:       m.OIB,
		Residence: m.Residence,
		BirthDate: m.BirthDate.Format(DateFormat),
		Email:     m.Email,
		Role:      string(m.Role),
		License:   *license.FromModel(&m.License),
	}
	return dto
}
