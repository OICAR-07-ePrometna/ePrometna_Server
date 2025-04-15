package dto

import (
	"ePrometna_Server/model"
	"ePrometna_Server/util/cerror"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type VehicleDto struct {
	Uuid           string
	VehicleType    string
	VehicleModel   string
	ProductionYear int
	Registration   string
}

// ToModel create a model from a dto
func (dto *VehicleDto) ToModel() (*model.Vehicle, error) {
	uuid, err := uuid.Parse(dto.Uuid)
	if err != nil {
		zap.S().Errorf("Failed to parse uuid = %s, err = %+v", dto.Uuid, err)
		return nil, cerror.ErrBadUuid
	}

	return &model.Vehicle{
		Uuid:           uuid,
		VehicleType:    dto.VehicleType,
		VehicleModel:   dto.VehicleModel,
		ProductionYear: dto.ProductionYear,
		Registration:   nil,
	}, nil
}

func (dto VehicleDto) FromModel(m *model.Vehicle) VehicleDto {
	reg := ""
	if m.Registration != nil {
		reg = m.Registration.Registration
		zap.S().Errorf("Registration is nil on car with uuid = %s", m.Uuid)
	}

	dto = VehicleDto{
		Uuid:           m.Uuid.String(),
		VehicleType:    m.VehicleType,
		VehicleModel:   m.VehicleModel,
		ProductionYear: m.ProductionYear,
		Registration:   reg,
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
