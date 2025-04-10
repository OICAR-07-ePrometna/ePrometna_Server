package dto

import (
	"ePrometna_Server/model"
)

// TODO: add more properties
type VehicleDetailsDto struct {
	Uuid           string
	VehicleType    string
	VehicleModel   string
	ProductionYear int
	Registration   string
	Owner          UserDto
	Drivers        []UserDto
	PastOwners     []UserDto
	// Registration   RegistrationDto
	// PastRegistratins []RegistrationDto
}

// ToModel create a model from a dto
func (dto *VehicleDetailsDto) ToModel() (*model.Vehicle, error) {
	panic("unimplemented")
}

func (dto VehicleDetailsDto) FromModel(m *model.Vehicle) VehicleDetailsDto {
	panic("unimplemented")
}
