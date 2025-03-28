package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DriverLicense struct {
	gorm.Model
	Uuid          uuid.UUID `gorm:"type:uuid;unique;not null"`
	UserId        uint      `gorm:"type:uint;unique;not null"`
	DriverId      uint      `gorm:"type:uint;unique;not null"`
	LicenseNumber string    `gorm:"type:string;unique;not null"`
	IssueDate     time.Time `gorm:"type:date;not null"`
	ExpiringDate  time.Time `gorm:"type:date;not null"`
	Category      string    `gorm:"type:string;not null"`
	User          User      `gorm:"foreignKey:UserId;references:ID"`
}
