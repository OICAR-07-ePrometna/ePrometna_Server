package dto

import "ePrometna_Server/model"

type RegistrationDto struct {
	PassTechnical    bool   `json:"passTechnical"`
	TraveledDistance int    `json:"traveledDistance"`
	Registration     string `json:"registration"`
	Note             string `json:"note"`
}

func (dto *RegistrationDto) ToModel() (model.RegistrationInfo, error) {
	m := model.RegistrationInfo{
		PassTechnical:    dto.PassTechnical,
		TraveledDistance: dto.TraveledDistance,
		Registration:     dto.Registration,
		Note:             &dto.Note,
	}

	return m, nil
}
