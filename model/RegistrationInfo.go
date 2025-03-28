package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RegistrationInfo struct {
	gorm.Model
	Uuid          uuid.UUID `gorm:"type:uuid;unique;not null"`
	CarId         uint      `gorm:"type:uint;unique;not null"`
	PassTehnical  bool      `gorm:"type:bool;not null"`
	TraveledMiles int       `gorm:"type:int;not null"`
	TehnicalDate  time.Time `gorm:"type:date;not null"`
	Registration  string    `gorm:"type:varchar(20);not null"`
}
