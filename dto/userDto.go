package dto

import "ePrometna_Server/model"

type UserDto struct {
	Uuid      string
	FirstName string
	LastName  string
	OIB       string
	Residence string
	BirthDate string
	Email     string
	Password  string
	Role      string
	License   DriverLicenseDto
}

// ToModel create a model from a dto
func (dto *UserDto) ToModel() *model.User {
	return &model.User{}
}

// FromModel returns a dto from model struct
func (dto *UserDto) FromModel(m *model.User) *UserDto {
	dto = &UserDto{}
	return dto
}
