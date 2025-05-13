package dto_test

import (
	"ePrometna_Server/dto"
	"ePrometna_Server/model"
	"ePrometna_Server/util/cerror"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestVehicleDto_ToModel(t *testing.T) {
	validUUID := uuid.New()

	tests := []struct {
		name    string
		dto     dto.VehicleDto
		want    *model.Vehicle
		wantErr error
	}{
		{
			name: "Valid DTO to Model",
			dto: dto.VehicleDto{
				Uuid:         validUUID.String(),
				VehicleType:  "Motorcycle",
				Model:        "Ninja",
				Registration: "KA987ZX", // Not directly used by ToModel for model.Vehicle.Registration
				AllowedTo:    "2025-12-31",
			},
			want: &model.Vehicle{
				Uuid:         validUUID,
				VehicleType:  "Motorcycle",
				VehicleModel: "Ninja",
				// model.Vehicle.Registration is *model.RegistrationInfo, set by service
				// model.Vehicle.AllowedTo is not a direct field, this DTO field is for display
			},
			wantErr: nil,
		},
		{
			name: "Invalid UUID",
			dto: dto.VehicleDto{
				Uuid: "invalid-uuid-string",
			},
			want:    nil,
			wantErr: cerror.ErrBadUuid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.dto.ToModel()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.want.Uuid, got.Uuid)
				assert.Equal(t, tt.want.VehicleType, got.VehicleType)
				assert.Equal(t, tt.want.VehicleModel, got.VehicleModel)
			}
		})
	}
}

func TestVehicleDto_FromModel(t *testing.T) {
	vehicleUUID := uuid.New()
	regUUID := uuid.New()
	techDate := time.Now()

	tests := []struct {
		name  string
		model *model.Vehicle
		want  dto.VehicleDto
	}{
		{
			name: "Model with Registration",
			model: &model.Vehicle{
				Uuid:         vehicleUUID,
				VehicleType:  "Truck",
				VehicleModel: "Actros",
				Registration: &model.RegistrationInfo{ // Current registration
					Uuid:             regUUID,
					VehicleId:        1, // Assuming a vehicle ID
					PassTechnical:    true,
					TraveledDistance: 250000,
					TechnicalDate:    techDate,
					Registration:     "DA123TR",
				},
				// AllowedTo is not directly in model.Vehicle, usually derived
			},
			want: dto.VehicleDto{
				Uuid:         vehicleUUID.String(),
				VehicleType:  "Truck",
				Model:        "Actros",
				Registration: "DA123TR",
				AllowedTo:    "", // VehicleDto.FromModel doesn't set AllowedTo currently
			},
		},
		{
			name: "Model without Registration (nil)",
			model: &model.Vehicle{
				Uuid:         vehicleUUID,
				VehicleType:  "Van",
				VehicleModel: "Sprinter",
				Registration: nil,
			},
			want: dto.VehicleDto{
				Uuid:         vehicleUUID.String(),
				VehicleType:  "Van",
				Model:        "Sprinter",
				Registration: "", // Expected empty if model.Registration is nil
				AllowedTo:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d dto.VehicleDto
			got := d.FromModel(tt.model)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVehiclesDto_FromModel(t *testing.T) {
	v1UUID := uuid.New()
	v2UUID := uuid.New()

	models := []model.Vehicle{
		{
			Uuid:         v1UUID,
			VehicleType:  "Bus",
			VehicleModel: "Lion's City",
			Registration: &model.RegistrationInfo{Registration: "ZG555BUS"},
		},
		{
			Uuid:         v2UUID,
			VehicleType:  "Scooter",
			VehicleModel: "Vespa",
			Registration: nil, // No current registration
		},
	}

	expectedDtos := dto.VehiclesDto{
		{
			Uuid:         v1UUID.String(),
			VehicleType:  "Bus",
			Model:        "Lion's City",
			Registration: "ZG555BUS",
			AllowedTo:    "",
		},
		{
			Uuid:         v2UUID.String(),
			VehicleType:  "Scooter",
			Model:        "Vespa",
			Registration: "",
			AllowedTo:    "",
		},
	}

	var d dto.VehiclesDto
	gotDtos := d.FromModel(models)
	assert.Equal(t, expectedDtos, gotDtos)

	// Test with empty slice
	emptyModels := []model.Vehicle{}
	expectedEmptyDtos := dto.VehiclesDto{}
	gotEmptyDtos := d.FromModel(emptyModels)
	assert.Equal(t, expectedEmptyDtos, gotEmptyDtos, "FromModel with empty slice should return empty dto slice")
	assert.NotNil(t, gotEmptyDtos, "FromModel with empty slice should return non-nil empty dto slice")

	// Test with nil slice
	var nilModels []model.Vehicle = nil
	// The FromModel for VehiclesDto initializes the slice, so it should return an empty, non-nil slice.
	expectedNilDtos := dto.VehiclesDto{}
	gotNilDtos := d.FromModel(nilModels)
	assert.Equal(t, expectedNilDtos, gotNilDtos, "FromModel with nil slice should return empty dto slice")
	assert.NotNil(t, gotNilDtos, "FromModel with nil slice should return non-nil empty dto slice")
}
