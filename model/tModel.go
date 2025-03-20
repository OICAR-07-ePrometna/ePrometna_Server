package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Tmodel struct {
	gorm.Model

	Name string
	age  int
	Uuid uuid.UUID
}
