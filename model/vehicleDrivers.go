package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VehicleDrivers struct {
	gorm.Model

	Uuid      uuid.UUID `gorm:"type:uuid;unique;not null"`
	VehicleId uint      `gorm:"type:uint;not null"`
	UserId    uint      `gorm:"type:uint;not null"`
	User      User      `gorm:"foreignKey:UserId"`
	Given     time.Time `gorm:"type:date;not null"`
	Until     time.Time `gorm:"type:date;null"`
}

func (cd *VehicleDrivers) BeforeCreate(tx *gorm.DB) error {
	if cd.Until.IsZero() {
		return nil
	}

	if cd.Given.After(cd.Until) || cd.Given.Equal(cd.Until) {
		return errors.New("given date must be before until date")
	}
	return nil
}
