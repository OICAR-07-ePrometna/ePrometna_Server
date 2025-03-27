package dto

import (
	"ePrometna_Server/model"

	"github.com/google/uuid"
)

type TmodelDto struct {
	Name string
	Age  int
	Uuid string
}

// ToModel create a model from a dto
func (dto *TmodelDto) ToModel() *model.Tmodel {
	return &model.Tmodel{
		Name: dto.Name,
		Age:  dto.Age,
		Uuid: uuid.New(),
	}
}

// FromModel returns a dto from model struct
func (dto *TmodelDto) FromModel(m *model.Tmodel) *TmodelDto {
	dto = &TmodelDto{
		Name: m.Name,
		Age:  m.Age,
		Uuid: m.Uuid.String(),
	}
	return dto
}
