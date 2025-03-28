package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Mobile struct {
	gorm.Model
	Uuid             uuid.UUID `gorm:"type:uuid;unique;not null"`
	UserId           uint      `gorm:"type:uint;unique;null"`
	CreatorId        uint      `gorm:"type:uint;not null"`
	RegisteredDevice string    `gorm:"type:varchar(50);null"`
	ActivationToken  string    `gorm:"type:varchar(50);unique;not null"`
}
