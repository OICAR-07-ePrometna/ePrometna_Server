package dto

import (
	"ePrometna_Server/model"
	"ePrometna_Server/util/format"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DriverLicenseDto struct {
	Uuid          string `json:"uuid"`
	LicenseNumber string `json:"licenseNumber"`
	IssueDate     string `json:"issueDate"`
	ExpiringDate  string `json:"expiringDate"`
	Category      string `json:"category"`
}

// ToModel create a model from a dto
func (dto *DriverLicenseDto) ToModel() (*model.DriverLicense, error) {
	issueDate, err := time.Parse(format.DateFormat, dto.IssueDate)
	if err != nil {
		zap.S().Errorf("Bad date time format need %s has %s", format.DateFormat, dto.IssueDate)
		return nil, err
	}
	exp, err := time.Parse(format.DateFormat, dto.ExpiringDate)
	if err != nil {
		zap.S().Errorf("Bad date time format need %s has %s", format.DateFormat, dto.ExpiringDate)
		return nil, err
	}
	return &model.DriverLicense{
		Uuid:          uuid.New(),
		LicenseNumber: dto.LicenseNumber,
		Category:      dto.Category,
		IssueDate:     issueDate,
		ExpiringDate:  exp,
	}, nil
}

// FromModel returns a dto from model struct
func (dto *DriverLicenseDto) FromModel(m *model.DriverLicense) *DriverLicenseDto {
	dto = &DriverLicenseDto{
		Uuid:          m.Uuid.String(),
		LicenseNumber: m.LicenseNumber,
		Category:      m.Category,
		IssueDate:     m.IssueDate.Format(format.DateFormat),
		ExpiringDate:  m.ExpiringDate.Format(format.DateFormat),
	}
	return dto
}
