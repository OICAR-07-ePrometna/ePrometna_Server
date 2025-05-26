package service

import (
	"ePrometna_Server/model"

	"gorm.io/gorm"
)

type ITempDataService interface {
	CreateTempData(tempData *model.TempData) error
	GetAndDeleteByUUID(uuid string) (*model.TempData, error)
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

func (s *TempDataService) GetAndDeleteByUUID(uuid string) (*model.TempData, error) {
	var tempData model.TempData
	err := s.db.Where("uuid = ?", uuid).First(&tempData).Error
	if err != nil {
		return nil, err
	}

	err = s.db.Delete(&tempData).Error
	if err != nil {
		return nil, err
	}

	return &tempData, nil
}
