package dto

import (
	"ePrometna_Server/model"
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
	panic("unimplemented")
}

func (dto VehicleDto) FromModel(m *model.Vehicle) VehicleDto {
	panic("unimplemented")
}
