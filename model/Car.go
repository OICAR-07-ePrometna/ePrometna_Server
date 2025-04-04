package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Car struct {
	gorm.Model
	Uuid           uuid.UUID         `gorm:"type:uuid;unique;not null"`
	CarType        string            `gorm:"type:varchar(50);not null"`
	CarModel       string            `gorm:"type:varchar(50);not null"`
	ProductionYear int               `gorm:"type:int;not null"`
	ChassisNumber  string            `gorm:"type:varchar(50);unique;not null"`
	UserId         uint              `gorm:"type:uint;not null"`
	Drivers        []CarDrivers      `gorm:"foreignKey:CarId"`
	PastOwners     []OwnerHistory    `gorm:"foreignKey:CarId"`
	TemporaryData  *TempData         `gorm:"foreignKey:CarId"`
	Registration   *RegistrationInfo `gorm:"foreignKey:CarId;not null"`
}
