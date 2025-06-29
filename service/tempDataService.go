package service

import (
	"ePrometna_Server/model"
	"ePrometna_Server/util/cerror"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ITempDataService interface {
	CreateTempData(tempData *model.TempData) error
	GetAndDeleteByUUID(uuid uuid.UUID) (string, string, error)
}

type TempDataService struct {
	db *gorm.DB
}

func NewTempDataService(db *gorm.DB) ITempDataService {
	return &TempDataService{
		db: db,
	}
}

func (s *TempDataService) CreateTempData(tempData *model.TempData) error {
	// First check if there's existing temp data for this user
	var existingTempData model.TempData
	err := s.db.Where("driver_id = ?", tempData.DriverId).First(&existingTempData).Error

	// If we found existing data, delete it
	if err == nil {
		if err := s.db.Unscoped().Delete(&existingTempData).Error; err != nil {
			return err
		}
	} else if err != gorm.ErrRecordNotFound {
		// If error is not "record not found", return it
		return err
	}

	// Proceed with creating new temp data
	return s.db.Create(tempData).Error
}

func (s *TempDataService) GetAndDeleteByUUID(uuid uuid.UUID) (string, string, error) {
	var tempData model.TempData
	err := s.db.Where("uuid = ?", uuid).First(&tempData).Error
	if err != nil {
		return "", "", err
	}

	if time.Now().After(tempData.Expiring) {
		s.db.Unscoped().Delete(&tempData)
		return "", "", cerror.ErrOutdated
	}

	var user model.User
	rez := s.db.First(&user, "id = ?", tempData.DriverId)
	if rez.Error != nil {
		return "", "", rez.Error
	}

	var vehicle model.Vehicle
	rez = s.db.First(&vehicle, "id = ?", tempData.VehicleId)
	if rez.Error != nil {
		return "", "", rez.Error
	}

	err = s.db.Unscoped().Delete(&tempData).Error
	if err != nil {
		return "", "", err
	}

	return user.Uuid.String(), vehicle.Uuid.String(), nil
}
