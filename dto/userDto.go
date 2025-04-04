package dto

import (
	"ePrometna_Server/model"
	"ePrometna_Server/util/cerror"
	"ePrometna_Server/util/format"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserDto struct {
	Uuid      string
	FirstName string
	LastName  string
	OIB       string
	Residence string
	BirthDate string
	Email     string
	Role      string
}

// ToModel create a model from a dto
func (dto *UserDto) ToModel() (*model.User, error) {
	uuid, err := uuid.Parse(dto.Uuid)
	if err != nil {
		zap.S().Error("Failed to parse uuid err = %+v", err)
		return nil, cerror.ErrBadUuid
	}

	bod, err := time.Parse(format.DateFormat, dto.BirthDate)
	if err != nil {
		zap.S().Error("Failed to parse BirthDate err = %+v", err)
		return nil, cerror.ErrBadDateFormat
	}

	role := model.RoleSuperAdmin
	return &model.User{
		Uuid:      uuid,
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
func (dto UserDto) FromModel(m *model.User) UserDto {
	dto = UserDto{
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
