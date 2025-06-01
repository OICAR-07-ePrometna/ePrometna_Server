package service

import (
	"ePrometna_Server/model"

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
	err := s.db.Create(tempData).Error
	if err != nil {
		return err
	}
	return nil
}

func (s *TempDataService) GetAndDeleteByUUID(uuid uuid.UUID) (string, string, error) {
	var tempData model.TempData
	err := s.db.Where("uuid = ?", uuid).First(&tempData).Error
	if err != nil {
		return "", "", err
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
