package model

import (
	"errors"
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

func (dl *DriverLicense) Update(license *DriverLicense) (*DriverLicense, error) {
	if license.ExpiringDate.Before(license.IssueDate) {
		return nil, errors.New("expiring date must be after issue date")
	}

	dl.LicenseNumber = license.LicenseNumber
	dl.Category = license.Category
	dl.IssueDate = license.IssueDate
	dl.ExpiringDate = license.ExpiringDate
	return dl, nil
}
