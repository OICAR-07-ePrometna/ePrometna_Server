package dto

import (
	"ePrometna_Server/model"
	"time"

	"github.com/google/uuid"
)

type RegistrationDto struct {
	PassTechnical    bool   `json:"passTechnical"`
	TraveledDistance int    `json:"traveledDistance"`
	Registration     string `json:"registration"`
	Note             string `json:"note"`
	TechnicalDate    string `json:"technicalDate"`
}

func (dto *RegistrationDto) ToModel() (model.RegistrationInfo, error) {
	m := model.RegistrationInfo{
		Uuid:             uuid.New(),
		PassTechnical:    dto.PassTechnical,
		TraveledDistance: dto.TraveledDistance,
		Registration:     dto.Registration,
		Note:             &dto.Note,
		TechnicalDate:    time.Now(), // Always use current time for new registrations
	}

	return m, nil
}
