package dto

import (
	"ePrometna_Server/model"
	"ePrometna_Server/util/cerror"
	"ePrometna_Server/util/format"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type NewUserDto struct {
	Uuid      string `json:"uuid"`
	FirstName string `json:"firstName" binding:"required,min=2,max=100"`
	LastName  string `json:"lastName" binding:"required,min=2,max=100"`
	OIB       string `json:"oib" binding:"required,len=11"`
	Residence string `json:"residence" binding:"required,max=255"`
	BirthDate string `json:"birthDate" binding:"required,datetime=2006-01-02"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	Role      string `json:"role" binding:"required,oneof=hak mupadmin osoba firma policija superadmin"`
}

// ToModel create a model from a dto
func (dto *NewUserDto) ToModel() (*model.User, error) {
	bod, err := time.Parse(format.DateFormat, dto.BirthDate)
	if err != nil {
		zap.S().Errorf("Bad date time format need %s has %s", format.DateFormat, dto.BirthDate)
		return nil, cerror.ErrBadDateFormat
	}
	role, err := model.StoUserRole(dto.Role)
	if err != nil {
		zap.S().Error("Failed to parse role = %+v, err = %+v", dto.Role, err)
		return nil, cerror.ErrUnknownRole
	}
	if dto.Uuid != "" {
		_, err := uuid.Parse(dto.Uuid)
		if err != nil {
			zap.S().Errorf("Failed to parse uuid = %s, err = %+v", dto.Uuid, err)
			return nil, cerror.ErrBadUuid
		}
	}

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
		Role:      fmt.Sprint(m.Role),
	}
	return dto
}
