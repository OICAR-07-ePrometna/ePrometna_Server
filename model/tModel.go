package model

import "gorm.io/gorm"

type Tmodel struct {
	gorm.Model
	Name string
}
