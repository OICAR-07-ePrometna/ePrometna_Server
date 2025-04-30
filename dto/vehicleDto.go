package dto

import (
	"ePrometna_Server/model"
	"ePrometna_Server/util/cerror"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type VehicleDto struct {
	Uuid         string `json:"uuid"`
	VehicleType  string `json:"vehicleType"`
	Model        string `json:"model"`
	Registration string `json:"registration"`

	// NOTE: can be date or empty if empty then it is allowed forever
	AllowedTo string `json:"allowedTo"`
}

func (dto *VehicleDto) ToModel() (*model.Vehicle, error) {
	uuid, err := uuid.Parse(dto.Uuid)
	if err != nil {
		zap.S().Errorf("Failed to parse uuid = %s, err = %+v", dto.Uuid, err)
		return nil, cerror.ErrBadUuid
	}

	return &model.Vehicle{
		Uuid:         uuid,
		VehicleType:  dto.VehicleType,
		VehicleModel: dto.Model,
		Registration: nil,
	}, nil
}

func (dto VehicleDto) FromModel(m *model.Vehicle) VehicleDto {
	reg := ""
	if m.Registration == nil {
		zap.S().Errorf("Registration is nil on car with uuid = %s", m.Uuid)
	} else {
		reg = m.Registration.Registration
	}

	dto = VehicleDto{
		Uuid:         m.Uuid.String(),
		VehicleType:  m.VehicleType,
		Model:        m.VehicleModel,
		Registration: reg,
	}
	return dto
}

type VehiclesDto []VehicleDto

func (dto VehiclesDto) FromModel(m []model.Vehicle) VehiclesDto {
	dto = make([]VehicleDto, 0, len(m))
	for _, v := range m {
		dto = append(dto, VehicleDto{}.FromModel(&v))
	}

	return dto
}
