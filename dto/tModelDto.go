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

func (dto *TmodelDto) Map() *model.Tmodel {
	return &model.Tmodel{
		Name: dto.Name,
		Age:  dto.Age,
		Uuid: uuid.New(),
	}
}
