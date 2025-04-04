package dto

import (
	"ePrometna_Server/model"
	"ePrometna_Server/util/cerror"
	"ePrometna_Server/util/format"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

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
}

// ToModel create a model from a dto
func (dto *NewUserDto) ToModel() (*model.User, error) {
	bod, err := time.Parse(format.DateFormat, dto.BirthDate)
	if err != nil {
		zap.S().Errorf("Bad date time format need %s has %s", format.DateFormat, dto.BirthDate)
		return nil, cerror.ErrBadDateFormat
	}
	// TODO: map role
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
	}, nil
}

// FromModel returns a dto from model struct
func (dto *NewUserDto) FromModel(m *model.User) *NewUserDto {
	dto = &NewUserDto{
		Uuid:      m.Uuid.String(),
		FirstName: m.FirstName,
		LastName:  m.LastName,
		OIB:       m.OIB,
		Residence: m.Residence,
		BirthDate: m.BirthDate.Format(format.DateFormat),
		Email:     m.Email,
		Role:      string(m.Role),
	}
	return dto
}
