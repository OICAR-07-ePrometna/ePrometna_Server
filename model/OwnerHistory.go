package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OwnerHistory struct {
	gorm.Model
	Uuid   uuid.UUID `gorm:"type:uuid;unique;not null"`
	CarId  uint      `gorm:"type:uint;not null"`
	UserId uint      `gorm:"type:uint;not null"`
}
