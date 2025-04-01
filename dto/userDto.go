package dto

import "ePrometna_Server/model"

type UserDto struct{}

// ToModel create a model from a dto
func (dto *UserDto) ToModel() *model.User {
	return &model.User{}
}

// FromModel returns a dto from model struct
func (dto *UserDto) FromModel(m *model.User) *UserDto {
	dto = &UserDto{}
	return dto
}
