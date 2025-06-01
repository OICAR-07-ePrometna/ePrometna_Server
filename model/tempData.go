package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TempData struct {
	gorm.Model
	Uuid      uuid.UUID `gorm:"type:uuid;unique;not null"`
	VehicleId uint      `gorm:"type:uint;unique;not null"`
	DriverId  uint      `gorm:"type:uint;unique;not null"`
	Expiring  time.Time `gorm:"type:timestamp;not null"`
}
