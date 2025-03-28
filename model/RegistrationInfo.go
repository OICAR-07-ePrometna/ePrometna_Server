package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RegistrationInfo struct {
	gorm.Model
	Uuid             uuid.UUID `gorm:"type:uuid;unique;not null"`
	CarId            uint      `gorm:"type:uint;unique;not null"`
	PassTechnical    bool      `gorm:"type:bool;not null"`
	TraveledDistance int       `gorm:"type:int;not null"`
	TechnicalDate    time.Time `gorm:"type:date;not null"`
	Registration     string    `gorm:"type:varchar(20);not null"`
}
