package service

import (
	"ePrometna_Server/app"
	"ePrometna_Server/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IDriverLicenseCrudService interface {
	Create(license *model.DriverLicense) (*model.DriverLicense, error)
	GetById(uuid uuid.UUID) (*model.DriverLicense, error)
	GetAll() ([]model.DriverLicense, error)
	Update(uuid uuid.UUID, updated *model.DriverLicense) (*model.DriverLicense, error)
	Delete(uuid uuid.UUID) error
}

type DriverLicenseCrudService struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

func NewDriverLicenseService(db *gorm.DB) IDriverLicenseCrudService {
	var service IDriverLicenseCrudService
	app.Invoke(func(db *gorm.DB, logger *zap.SugaredLogger) {
		service = &DriverLicenseCrudService{
			db:     db,
			logger: logger,
		}
	})
	return service
}

// Create implements IDriverLicenseService.
func (s *DriverLicenseCrudService) Create(license *model.DriverLicense) (*model.DriverLicense, error) {
	s.logger.Debugf("Createing driver license: %v", license)
	if err := s.db.Create(&license).Error; err != nil {
		s.logger.Errorf("Error creating driver license: %v", err)
		return nil, err
	}
	return license, nil
}

// GetById implements IDriverLicenseService
func (s *DriverLicenseCrudService) GetById(uuid uuid.UUID) (*model.DriverLicense, error) {
	var license model.DriverLicense
	s.logger.Debugf("Getting driver license with UUID: %s", uuid)
	if err := s.db.Where("uuid = ?", uuid).First(&license).Error; err != nil {
		s.logger.Errorf("Error getting driver license: %+v", err)
		return nil, err
	}
	return &license, nil
}

// GetAll implements IDriverLicenseService.
func (s *DriverLicenseCrudService) GetAll() ([]model.DriverLicense, error) {
	var licenses []model.DriverLicense
	s.logger.Debug("Getting all driver licenses")
	if err := s.db.Find(&licenses).Error; err != nil {
		s.logger.Errorf("Error getting all driver licenses: %+v", err)
		return nil, err
	}
	return licenses, nil
}

// Update implements IDriverLicenseService.
func (s *DriverLicenseCrudService) Update(uuid uuid.UUID, updated *model.DriverLicense) (*model.DriverLicense, error) {
	var license model.DriverLicense
	s.logger.Debugf("Updating driver license with UUID: %s", uuid)
	if err := s.db.Where("uuid = ?", uuid).First(&license).Error; err != nil {
		s.logger.Errorf("Error finding driver license: %+v", err)
		return nil, err
	}

	license.LicenseNumber = updated.LicenseNumber
	license.Category = updated.Category
	license.IssueDate = updated.IssueDate
	license.ExpiringDate = updated.ExpiringDate

	if err := s.db.Save(&license).Error; err != nil {
		s.logger.Errorf("Error updating driver license: %+v", err)
		return nil, err
	}
	return &license, nil
}

// Delete implements IDriverLicenseService.
func (s *DriverLicenseCrudService) Delete(uuid uuid.UUID) error {
	s.logger.Debugf("Deleting driver license with UUID: %s", uuid)
	if err := s.db.Where("uuid = ?", uuid).Delete(&model.DriverLicense{}).Error; err != nil {
		s.logger.Errorf("Error deleting driver license: %+v", err)
		return err
	}
	return nil
}
