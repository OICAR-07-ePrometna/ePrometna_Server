package dto_test

import (
	"ePrometna_Server/dto"
	"ePrometna_Server/model"
	"ePrometna_Server/util/cerror"
	"ePrometna_Server/util/format"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserDto_ToModel(t *testing.T) {
	validUUID := uuid.New()
	validDateStr := "1990-01-15"
	validTime, _ := time.Parse(format.DateFormat, validDateStr)

	tests := []struct {
		name    string
		dto     dto.UserDto
		want    *model.User
		wantErr error
	}{
		{
			name: "Valid DTO to Model",
			dto: dto.UserDto{
				Uuid:      validUUID.String(),
				FirstName: "John",
				LastName:  "Doe",
				OIB:       "12345678901",
				Residence: "123 Main St",
				BirthDate: validDateStr,
				Email:     "john.doe@example.com",
				Role:      "osoba",
			},
			want: &model.User{
				Uuid:      validUUID,
				FirstName: "John",
				LastName:  "Doe",
				OIB:       "12345678901",
				Residence: "123 Main St",
				BirthDate: validTime,
				Email:     "john.doe@example.com",
				Role:      model.RoleOsoba,
			},
			wantErr: nil,
		},
		{
			name: "Invalid UUID",
			dto: dto.UserDto{
				Uuid:      "not-a-uuid",
				FirstName: "Jane",
				LastName:  "Doe",
				Role:      "firma",
				BirthDate: validDateStr,
			},
			want:    nil,
			wantErr: cerror.ErrBadUuid,
		},
		{
			name: "Invalid BirthDate format",
			dto: dto.UserDto{
				Uuid:      validUUID.String(),
				FirstName: "Jim",
				LastName:  "Beam",
				BirthDate: "15-01-1990", // Wrong format
				Role:      "hak",
			},
			want:    nil,
			wantErr: cerror.ErrBadDateFormat,
		},
		{
			name: "Invalid Role",
			dto: dto.UserDto{
				Uuid:      validUUID.String(),
				FirstName: "Jack",
				LastName:  "Daniels",
				BirthDate: validDateStr,
				Role:      "invalid_role",
			},
			want:    nil,
			wantErr: cerror.ErrUnknownRole,
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
				// Compare individual fields as GORM Model has unexported fields
				assert.Equal(t, tt.want.Uuid, got.Uuid)
				assert.Equal(t, tt.want.FirstName, got.FirstName)
				assert.Equal(t, tt.want.LastName, got.LastName)
				assert.Equal(t, tt.want.OIB, got.OIB)
				assert.Equal(t, tt.want.Residence, got.Residence)
				assert.Equal(t, tt.want.BirthDate, got.BirthDate)
				assert.Equal(t, tt.want.Email, got.Email)
				assert.Equal(t, tt.want.Role, got.Role)
			}
		})
	}
}

func TestUserDto_FromModel(t *testing.T) {
	userUUID := uuid.New()
	birthTime, _ := time.Parse(format.DateFormat, "1985-07-20")
	userModel := &model.User{
		Uuid:      userUUID,
		FirstName: "Alice",
		LastName:  "Smith",
		OIB:       "09876543210",
		Residence: "456 Oak Ave",
		BirthDate: birthTime,
		Email:     "alice.smith@example.com",
		Role:      model.RoleFirma,
	}

	expectedDto := dto.UserDto{
		Uuid:      userUUID.String(),
		FirstName: "Alice",
		LastName:  "Smith",
		OIB:       "09876543210",
		Residence: "456 Oak Ave",
		BirthDate: "1985-07-20",
		Email:     "alice.smith@example.com",
		Role:      "firma",
	}

	var gotDto dto.UserDto
	gotDto = gotDto.FromModel(userModel)

	assert.Equal(t, expectedDto, gotDto)
}
