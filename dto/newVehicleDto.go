package dto

import "ePrometna_Server/model"

type NewVehicleDto struct {
	OwnerUuid string
}

// ToModel create a model from a dto
func (dto *NewVehicleDto) ToModel() (*model.Vehicle, error) {
	panic("Unimplemented")
}
