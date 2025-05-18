package model

import (
	"ePrometna_Server/util/cerror"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleHAK        UserRole = "hak"
	RoleMupADMIN   UserRole = "mupadmin"
	RoleOsoba      UserRole = "osoba"
	RoleFirma      UserRole = "firma"
	RolePolicija   UserRole = "policija"
	RoleSuperAdmin UserRole = "superadmin"
)

func StoUserRole(text string) (UserRole, error) {
	switch text {
	case fmt.Sprint(RoleHAK):
		return RoleHAK, nil

	case fmt.Sprint(RoleMupADMIN):
		return RoleMupADMIN, nil

	case fmt.Sprint(RoleOsoba):
		return RoleOsoba, nil

	case fmt.Sprint(RoleFirma):
		return RoleFirma, nil

	case fmt.Sprint(RolePolicija):
		return RolePolicija, nil

	case fmt.Sprint(RoleSuperAdmin):
		return RoleSuperAdmin, nil

	default:
		return "", cerror.ErrUnknownRole
	}
}

type User struct {
	gorm.Model

	Uuid             uuid.UUID        `gorm:"type:uuid;unique;not null"`
	FirstName        string           `gorm:"type:varchar(100);not null"`
	LastName         string           `gorm:"type:varchar(100);not null"`
	OIB              string           `gorm:"type:char(11);unique;not null"`
	Residence        string           `gorm:"type:varchar(255);not null"`
	BirthDate        time.Time        `gorm:"type:date;not null"`
	Email            string           `gorm:"type:varchar(100);unique;not null"`
	PasswordHash     string           `gorm:"type:varchar(255);not null"`
	Role             UserRole         `gorm:"type:varchar(20);not null"`
	Vehicles         []Vehicle        `gorm:"foreignKey:UserId"`
	BorrowedVehicles []VehicleDrivers `gorm:"foreignKey:UserId"`
	VehicleHistory   []OwnerHistory   `gorm:"foreignKey:UserId"`
	RegisteredDevice *Mobile          `gorm:"foreignKey:UserId"`
	CreatedDevices   []Mobile         `gorm:"foreignKey:CreatorId"`
	TemporaryData    *TempData        `gorm:"foreignKey:DriverId"`
	License          *DriverLicense   `gorm:"foreignKey:UserId"`
	PoliceToken      *string          `gorm:"column:police_token;null"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	validRoles := map[UserRole]bool{
		RoleHAK: true, RoleMupADMIN: true, RoleOsoba: true,
		RoleFirma: true, RolePolicija: true, RoleSuperAdmin: true,
	}

	if _, ok := validRoles[u.Role]; !ok {
		return errors.New("invalid user role")
	}
	return nil
}

func (u *User) Update(user *User) *User {
	u.BirthDate = user.BirthDate
	u.FirstName = user.FirstName
	u.LastName = user.LastName
	u.OIB = user.OIB
	u.Residence = user.Residence
	u.Email = user.Email
	u.Role = user.Role
	u.PoliceToken = user.PoliceToken

	return u
}
