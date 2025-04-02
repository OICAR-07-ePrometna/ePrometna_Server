package dto

import (
	"ePrometna_Server/model"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DriverLicenseDto struct {
	Uuid          string
	LicenseNumber string
	IssueDate     string
	ExpiringDate  string
	Category      string
}

// ToModel create a model from a dto
func (dto *DriverLicenseDto) ToModel() *model.DriverLicense {
	issueDate, err := time.Parse(DateFormat, dto.IssueDate)
	if err != nil {
		zap.S().DPanicf("Bad date time format need %s has %s", DateFormat, dto.IssueDate)
		// TODO: see what to have to happen in prod
		return nil
	}
	exp, err := time.Parse(DateFormat, dto.ExpiringDate)
	if err != nil {
		zap.S().DPanicf("Bad date time format need %s has %s", DateFormat, dto.ExpiringDate)
		// TODO: see what to have to happen in prod
		return nil
	}
	return &model.DriverLicense{
		Uuid:          uuid.New(),
		LicenseNumber: dto.LicenseNumber,
		Category:      dto.Category,
		IssueDate:     issueDate,
		ExpiringDate:  exp,
	}
}

// FromModel returns a dto from model struct
func (dto *DriverLicenseDto) FromModel(m *model.DriverLicense) *DriverLicenseDto {
	dto = &DriverLicenseDto{
		Uuid:          m.Uuid.String(),
		LicenseNumber: m.LicenseNumber,
		Category:      m.Category,
		IssueDate:     m.IssueDate.Format(DateFormat),
		ExpiringDate:  m.ExpiringDate.Format(DateFormat),
	}
	return dto
}
