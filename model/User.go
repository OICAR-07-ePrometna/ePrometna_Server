package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleHAK        UserRole = "hak"
	RoleAdmin      UserRole = "admin"
	RoleOsoba      UserRole = "osoba"
	RoleFirma      UserRole = "firma"
	RolePolicija   UserRole = "policija"
	RoleSuperAdmin UserRole = "superadmin"
)

type User struct {
	gorm.Model

	Uuid             uuid.UUID      `gorm:"type:uuid;unique;not null"`
	FirstName        string         `gorm:"type:varchar(100);not null"`
	LastName         string         `gorm:"type:varchar(100);not null"`
	OIB              string         `gorm:"type:char(11);unique;not null"`
	Residence        string         `gorm:"type:varchar(255);not null"`
	BirthDate        time.Time      `gorm:"type:date;not null"`
	Email            string         `gorm:"type:varchar(100);unique;not null"`
	PasswordHash     string         `gorm:"type:varchar(255);not null"`
	Role             UserRole       `gorm:"type:varchar(20);not null"`
	Cars             []Car          `gorm:"foreignKey:UserId"`
	BorrowedCars     []CarDrivers   `gorm:"foreignKey:UserId"`
	CarHistory       []OwnerHistory `gorm:"foreignKey:UserId"`
	RegisteredDevice Mobile         `gorm:"foreignKey:UserId"`
	CreatedDevices   []Mobile       `gorm:"foreignKey:CreatorId"`
	TemporaryData    TempData       `gorm:"foreignKey:DriverId"`
	License          DriverLicense  `gorm:"foreignKey:UserId"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	validRoles := map[UserRole]bool{
		RoleHAK: true, RoleAdmin: true, RoleOsoba: true,
		RoleFirma: true, RolePolicija: true, RoleSuperAdmin: true,
	}

	if _, ok := validRoles[u.Role]; !ok {
		return errors.New("invalid user role")
	}
	return nil
}
