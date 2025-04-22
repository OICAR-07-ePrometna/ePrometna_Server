package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DriverLicense struct {
	gorm.Model
	Uuid          uuid.UUID `gorm:"type:uuid;unique;not null"`
	UserId        uint      `gorm:"type:uint;unique;not null"`
	LicenseNumber string    `gorm:"type:varchar(50);unique;not null"`
	IssueDate     time.Time `gorm:"type:date;not null"`
	ExpiringDate  time.Time `gorm:"type:date;not null"`
	Category      string    `gorm:"type:varchar(50);not null"`
}

func (u *DriverLicense) Update(license *DriverLicense) *DriverLicense {
	u.LicenseNumber = license.LicenseNumber
	u.Category = license.Category
	u.IssueDate = license.IssueDate
	u.ExpiringDate = license.ExpiringDate
	return u
}
