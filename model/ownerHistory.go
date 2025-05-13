package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OwnerHistory struct {
	gorm.Model
	Uuid      uuid.UUID `gorm:"type:uuid;unique;not null"`
	VehicleId uint      `gorm:"type:uint;not null"`
	UserId    uint      `gorm:"type:uint;not null"`
	User      User      `gorm:"foreignKey:UserId"`
}

func (m *OwnerHistory) FromUser(user User) *OwnerHistory {
	m.UserId = user.ID
	m.Uuid = uuid.New()
	return m
}
