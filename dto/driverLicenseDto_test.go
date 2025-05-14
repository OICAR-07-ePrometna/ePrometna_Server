package dto_test

import (
	"ePrometna_Server/dto"
	"ePrometna_Server/model"
	"ePrometna_Server/util/format"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDriverLicenseDto_ToModel(t *testing.T) {
	validIssueDateStr := "2020-01-01"
	validExpDateStr := "2030-01-01"
	issueTime, _ := time.Parse(format.DateFormat, validIssueDateStr)
	expTime, _ := time.Parse(format.DateFormat, validExpDateStr)

	tests := []struct {
		name    string
		dto     dto.DriverLicenseDto
		want    *model.DriverLicense
		wantErr bool // Simplified error check for DPanicf cases
	}{
		{
			name: "Valid DTO to Model",
			dto: dto.DriverLicenseDto{
				// Uuid is generated in ToModel, so not provided here
				LicenseNumber: "DL12345",
				IssueDate:     validIssueDateStr,
				ExpiringDate:  validExpDateStr,
				Category:      "B",
			},
			want: &model.DriverLicense{
				// Uuid will be compared for non-nil
				LicenseNumber: "DL12345",
				IssueDate:     issueTime,
				ExpiringDate:  expTime,
				Category:      "B",
			},
			wantErr: false,
		},
		{
			name: "Invalid IssueDate format",
			dto: dto.DriverLicenseDto{
				LicenseNumber: "DL67890",
				IssueDate:     "01-01-2020", // Wrong format
				ExpiringDate:  validExpDateStr,
				Category:      "C",
			},
			want:    nil,
			wantErr: true, // DPanicf will cause panic in test, or return nil in prod
		},
		{
			name: "Invalid ExpiringDate format",
			dto: dto.DriverLicenseDto{
				LicenseNumber: "DL11223",
				IssueDate:     validIssueDateStr,
				ExpiringDate:  "01/01/2030", // Wrong format
				Category:      "A",
			},
			want:    nil,
			wantErr: true, // DPanicf
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The DPanicf in the original ToModel makes direct error comparison tricky.
			// In a test environment, DPanicf will panic. In production, it logs and continues.
			// We'll assume for testing it might return nil if parsing fails.
			got, err := tt.dto.ToModel()

			if tt.wantErr {
				assert.NotNil(t, err, "Expected nil for DPanicf cases or parse errors")
			} else {
				assert.NotNil(t, got)
				assert.NotEqual(t, uuid.Nil, got.Uuid, "UUID should be generated and not nil")
				assert.Equal(t, tt.want.LicenseNumber, got.LicenseNumber)
				assert.Equal(t, tt.want.IssueDate, got.IssueDate)
				assert.Equal(t, tt.want.ExpiringDate, got.ExpiringDate)
				assert.Equal(t, tt.want.Category, got.Category)
			}
		})
	}
}

func TestDriverLicenseDto_FromModel(t *testing.T) {
	licenseUUID := uuid.New()
	issueTime, _ := time.Parse(format.DateFormat, "2018-05-10")
	expTime, _ := time.Parse(format.DateFormat, "2028-05-09")

	licenseModel := &model.DriverLicense{
		Uuid:          licenseUUID,
		UserId:        1, // Assuming a UserID
		LicenseNumber: "LN789XYZ",
		IssueDate:     issueTime,
		ExpiringDate:  expTime,
		Category:      "B, C1",
	}

	expectedDto := &dto.DriverLicenseDto{
		Uuid:          licenseUUID.String(),
		LicenseNumber: "LN789XYZ",
		IssueDate:     "2018-05-10",
		ExpiringDate:  "2028-05-09",
		Category:      "B, C1",
	}

	var d dto.DriverLicenseDto
	gotDto := d.FromModel(licenseModel)

	assert.Equal(t, expectedDto, gotDto)
}
