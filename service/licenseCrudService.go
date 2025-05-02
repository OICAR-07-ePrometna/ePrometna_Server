package service

import (
	"ePrometna_Server/app"
	"ePrometna_Server/model"
	"ePrometna_Server/util/cerror"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IDriverLicenseCrudService interface {
	Create(license *model.DriverLicense, ownerUuid uuid.UUID) (*model.DriverLicense, error)
	GetByUuid(uuid uuid.UUID) (*model.DriverLicense, error)
	GetAll() ([]model.DriverLicense, error)
	Update(uuid uuid.UUID, updated *model.DriverLicense) (*model.DriverLicense, error)
	Delete(uuid uuid.UUID) error
}

type DriverLicenseCrudService struct {
	db          *gorm.DB
	userService IUserCrudService
	logger      *zap.SugaredLogger
}

func NewDriverLicenseService(db *gorm.DB) IDriverLicenseCrudService {
	var service IDriverLicenseCrudService
	app.Invoke(func(db *gorm.DB, logger *zap.SugaredLogger, uService IUserCrudService) {
		service = &DriverLicenseCrudService{
			db:          db,
			logger:      logger,
			userService: uService,
		}
	})
	return service
}

// Create implements IDriverLicenseService.
func (s *DriverLicenseCrudService) Create(license *model.DriverLicense, ownerUuid uuid.UUID) (*model.DriverLicense, error) {
	owner, err := s.userService.Read(ownerUuid)
	if err != nil {
		s.logger.Errorf("Error reading user with uuid = %s, err = %+v", ownerUuid, err)
		return nil, err
	}

	if owner.Role != model.RoleFirma && owner.Role != model.RoleOsoba {
		s.logger.Errorf("User with role %+v can't own a driver license", owner.Role)
		return nil, cerror.ErrBadRole
	}

	license.UserId = owner.ID

	s.logger.Debugf("Creating driver license: %v", license)
	if err := s.db.Create(&license).Error; err != nil {
		s.logger.Errorf("Error creating driver license: %v", err)
		return nil, err
	}
	return license, nil
}

// GetById implements IDriverLicenseService
func (s *DriverLicenseCrudService) GetByUuid(uuid uuid.UUID) (*model.DriverLicense, error) {
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
	license, err := s.GetByUuid(uuid)
	if err != nil {
		s.logger.Errorf("Error getting driver license: %+v", err)
		return nil, err
	}

	if updated.UserId != 0 && updated.UserId != license.UserId {
		s.logger.Errorf("Attempt to change license ownership denied")
		return nil, cerror.ErrBadRole
	}

	s.logger.Debugf("Updating driver license: %+v", license)
	updatedLicense, err := license.Update(updated)
	if err != nil {
		s.logger.Errorf("Error updating driver license: %+v", err)
		return nil, err
	}
	license = updatedLicense

	rez := s.db.Where("uuid = ?", uuid).Save(license)
	if rez.Error != nil {
		s.logger.Errorf("Error updating driver license: %+v", rez.Error)
		return nil, rez.Error
	}
	return license, nil
}

// Delete implements IDriverLicenseService.
func (s *DriverLicenseCrudService) Delete(uuid uuid.UUID) error {
	_, err := s.GetByUuid(uuid)
	if err != nil {
		s.logger.Errorf("Error getting driver license: %+v", err)
		return err
	}

	s.logger.Debugf("Deleting driver license with UUID: %s", uuid)
	if err := s.db.Where("uuid = ?", uuid).Delete(&model.DriverLicense{}).Error; err != nil {
		s.logger.Errorf("Error deleting driver license: %+v", err)
		return err
	}
	return nil
}
