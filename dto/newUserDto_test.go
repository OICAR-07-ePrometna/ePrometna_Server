package dto_test

import (
	"ePrometna_Server/dto"
	"ePrometna_Server/model"
	"ePrometna_Server/util/cerror"
	"ePrometna_Server/util/format"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewUserDto_ToModel(t *testing.T) {
	validDateStr := "1995-03-20"
	validTime, _ := time.Parse(format.DateFormat, validDateStr)
	userUUID := uuid.New()

	tests := []struct {
		name    string
		dto     dto.NewUserDto
		want    *model.User // Uuid will be generated, so we compare other fields
		wantErr error
	}{
		{
			name: "Valid DTO to Model - No UUID in DTO",
			dto: dto.NewUserDto{
				// Uuid: "", // Intentionally empty, ToModel should generate it
				FirstName: "Test",
				LastName:  "User",
				OIB:       "11223344556",
				Residence: "Test Residence",
				BirthDate: validDateStr,
				Email:     "test.user@example.com",
				Password:  "password123",
				Role:      "osoba",
			},
			want: &model.User{
				// Uuid will be checked for non-nil
				FirstName: "Test",
				LastName:  "User",
				OIB:       "11223344556",
				Residence: "Test Residence",
				BirthDate: validTime,
				Email:     "test.user@example.com",
				Role:      model.RoleOsoba,
				// PasswordHash is set by service, not DTO
			},
			wantErr: nil,
		},
		{
			name: "Valid DTO to Model - With UUID in DTO",
			dto: dto.NewUserDto{
				Uuid:      userUUID.String(),
				FirstName: "Another",
				LastName:  "User",
				OIB:       "66778899001",
				Residence: "Another Residence",
				BirthDate: validDateStr,
				Email:     "another.user@example.com",
				Password:  "securepass",
				Role:      "firma",
			},
			want: &model.User{
				// Uuid will be checked for non-nil (ToModel generates a new one regardless of DTO input)
				FirstName: "Another",
				LastName:  "User",
				OIB:       "66778899001",
				Residence: "Another Residence",
				BirthDate: validTime,
				Email:     "another.user@example.com",
				Role:      model.RoleFirma,
			},
			wantErr: nil,
		},
		{
			name: "Invalid BirthDate format",
			dto: dto.NewUserDto{
				FirstName: "Bad",
				LastName:  "Date",
				BirthDate: "20-03-1995", // Wrong format
				Role:      "hak",
				OIB:       "12345678901",
				Email:     "bad.date@example.com",
				Password:  "password",
				Residence: "Some place",
			},
			want:    nil,
			wantErr: cerror.ErrBadDateFormat,
		},
		{
			name: "Invalid Role",
			dto: dto.NewUserDto{
				FirstName: "Bad",
				LastName:  "Role",
				BirthDate: validDateStr,
				Role:      "nonexistent_role",
				OIB:       "12345678901",
				Email:     "bad.role@example.com",
				Password:  "password",
				Residence: "Some place",
			},
			want:    nil,
			wantErr: cerror.ErrUnknownRole,
		},
		{
			name: "Invalid UUID in DTO",
			dto: dto.NewUserDto{
				Uuid:      "this-is-not-a-uuid",
				FirstName: "Bad",
				LastName:  "UUID",
				BirthDate: validDateStr,
				Role:      "osoba",
				OIB:       "12345678901",
				Email:     "bad.uuid@example.com",
				Password:  "password",
				Residence: "Some place",
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
				assert.NotEqual(t, uuid.Nil, got.Uuid, "Generated UUID should not be nil")
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

func TestNewUserDto_FromModel(t *testing.T) {
	userUUID := uuid.New()
	birthTime, _ := time.Parse(format.DateFormat, "1992-11-05")
	userModel := &model.User{
		Uuid:         userUUID,
		FirstName:    "ModelF",
		LastName:     "ModelL",
		OIB:          "55443322110",
		Residence:    "Model Residence",
		BirthDate:    birthTime,
		Email:        "model.user@example.com",
		PasswordHash: "somehash", // Not included in NewUserDto
		Role:         model.RoleSuperAdmin,
	}

	expectedDto := &dto.NewUserDto{
		Uuid:      userUUID.String(),
		FirstName: "ModelF",
		LastName:  "ModelL",
		OIB:       "55443322110",
		Residence: "Model Residence",
		BirthDate: "1992-11-05",
		Email:     "model.user@example.com",
		Password:  "", // Password is not part of FromModel for NewUserDto
		Role:      fmt.Sprint(model.RoleSuperAdmin),
	}

	var d dto.NewUserDto
	gotDto := d.FromModel(userModel)

	assert.Equal(t, expectedDto, gotDto)
}
