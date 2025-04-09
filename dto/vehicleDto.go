package dto

import "ePrometna_Server/model"

type VehicleDto struct {
	OwnerUuid string
}

// ToModel create a model from a dto
func (dto *VehicleDto) ToModel() (*model.Vehicle, error) {
	panic("unimplemented")
}

func (dto VehicleDto) FromModel(m *model.Vehicle) VehicleDto {
	panic("unimplemented")
}
