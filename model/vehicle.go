package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Vehicle struct {
	gorm.Model
	Uuid           uuid.UUID         `gorm:"type:uuid;unique;not null"`
	VehicleType    string            `gorm:"type:varchar(50);not null"`
	VehicleModel   string            `gorm:"type:varchar(50);not null"`
	ProductionYear int               `gorm:"type:int;not null"`
	ChassisNumber  string            `gorm:"type:varchar(50);unique;not null"`
	UserId         uint              `gorm:"type:uint;not null"`
	Owner          *User             `gorm:"foreignKey:UserId;OnDelete:SET NULL"`
	Drivers        []VehicleDrivers  `gorm:"foreignKey:VehicleId"`
	PastOwners     []OwnerHistory    `gorm:"foreignKey:VehicleId"`
	TemporaryData  *TempData         `gorm:"foreignKey:VehicleId"`
	Registration   *RegistrationInfo `gorm:"foreignKey:VehicleId;null"`
}
