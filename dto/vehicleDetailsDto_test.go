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

func TestVehicleDetailsDto_ToModel(t *testing.T) {
	validUUID := uuid.New()

	tests := []struct {
		name    string
		dto     dto.VehicleDetailsDto
		want    *model.Vehicle // Comparing main fields, not associations like Owner, Drivers
		wantErr bool
	}{
		{
			name: "Valid DTO to Model - Basic fields",
			dto: dto.VehicleDetailsDto{
				Uuid:         validUUID.String(),
				Registration: "ZG001TEST",
				Summary: dto.VehicleSummary{
					VehicleType:           "SUV",
					Model:                 "Explorer",
					ChassisNumber:         "CHASSISDETAIL01",
					Mark:                  "Ford",
					DateFirstRegistration: "2021-03-15",
				},
				// Owner, Drivers, PastOwners are not converted by ToModel in this version
			},
			want: &model.Vehicle{
				Uuid:                  validUUID,
				VehicleType:           "SUV",
				VehicleModel:          "Explorer",
				ChassisNumber:         "CHASSISDETAIL01",
				Mark:                  "Ford",
				DateFirstRegistration: "2021-03-15",
				// Registration field in model.Vehicle is *RegistrationInfo, not string
			},
			wantErr: false,
		},
		{
			name: "Invalid UUID",
			dto: dto.VehicleDetailsDto{
				Uuid: "not-a-real-uuid",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.dto.ToModel()
			if tt.wantErr {
				assert.Error(t, err) // Check if an error is returned
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.want.Uuid, got.Uuid)
				assert.Equal(t, tt.want.VehicleType, got.VehicleType)
				assert.Equal(t, tt.want.VehicleModel, got.VehicleModel)
				assert.Equal(t, tt.want.ChassisNumber, got.ChassisNumber)
				assert.Equal(t, tt.want.Mark, got.Mark)
				assert.Equal(t, tt.want.DateFirstRegistration, got.DateFirstRegistration)
				// Note: The DTO's Registration string is not directly mapped to model.Vehicle.Registration (*RegistrationInfo)
				// The model's Registration field would be populated by the service layer.
			}
		})
	}
}

func TestVehicleDetailsDto_FromModel(t *testing.T) {
	vehicleUUID := uuid.New()
	ownerUUID := uuid.New()
	ownerBirthDate, _ := time.Parse(format.DateFormat, "1980-01-01")

	vehicleModel := &model.Vehicle{
		Uuid:                  vehicleUUID,
		VehicleType:           "Sedan",
		VehicleModel:          "Accord",
		ChassisNumber:         "CHASSISFROMMODEL",
		Mark:                  "Honda",
		DateFirstRegistration: "2019-07-20",
		Registration: &model.RegistrationInfo{
			Registration: "ZG777FROM",
		},
		Owner: &model.User{ // Simplified Owner for testing FromModel
			Uuid:      ownerUUID,
			FirstName: "OwnerF",
			LastName:  "OwnerL",
			OIB:       "12312312312",
			Residence: "Owner Residence",
			BirthDate: ownerBirthDate,
			Email:     "owner@example.com",
			Role:      model.RoleOsoba,
		},
		// Drivers and PastOwners would be slices of model.User
		// For simplicity, leaving them empty for this test, but a full test would populate them.
		Drivers:    []model.VehicleDrivers{},
		PastOwners: []model.OwnerHistory{},
		// Populate other summary fields from model.Vehicle
		BodyShape:   "Saloon",
		EnginePower: "150kW",
	}

	expectedDto := dto.VehicleDetailsDto{
		Uuid:         vehicleUUID.String(),
		Registration: "ZG777FROM",
		Summary: dto.VehicleSummary{
			VehicleType:           "Sedan",
			Model:                 "Accord",
			ChassisNumber:         "CHASSISFROMMODEL",
			Mark:                  "Honda",
			DateFirstRegistration: "2019-07-20",
			BodyShape:             "Saloon",
			EnginePower:           "150kW",
		},
		Owner: dto.UserDto{
			Uuid:      ownerUUID.String(),
			FirstName: "OwnerF",
			LastName:  "OwnerL",
			OIB:       "12312312312",
			Residence: "Owner Residence",
			BirthDate: "1980-01-01",
			Email:     "owner@example.com",
			Role:      "osoba",
		},
		Drivers:    []dto.UserDto{},
		PastOwners: []dto.UserDto{},
	}

	var d dto.VehicleDetailsDto
	gotDto := d.FromModel(vehicleModel)

	assert.Equal(t, expectedDto.Uuid, gotDto.Uuid)
	assert.Equal(t, expectedDto.Registration, gotDto.Registration)
	assert.Equal(t, expectedDto.Summary, gotDto.Summary)
	assert.Equal(t, expectedDto.Owner, gotDto.Owner)
	assert.Equal(t, expectedDto.Drivers, gotDto.Drivers)
	assert.Equal(t, expectedDto.PastOwners, gotDto.PastOwners)
}
