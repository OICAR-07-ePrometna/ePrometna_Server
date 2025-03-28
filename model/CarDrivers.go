package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CarDrivers struct {
	gorm.Model

	Uuid   uuid.UUID `gorm:"type:uuid;unique;not null"`
	CarId  uint      `gorm:"type:uint;unique;not null"`
	UserId uint      `gorm:"type:uint;unique;not null"`
	Given  time.Time `gorm:"type:date;not null"`
	Until  time.Time `gorm:"type:date;not null"`
}
