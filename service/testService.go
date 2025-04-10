package service

import (
	"ePrometna_Server/app"
	"ePrometna_Server/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ITestService interface {
	ReadAll() ([]model.Tmodel, error)
	Read(id uuid.UUID) (*model.Tmodel, error)
	Create(test *model.Tmodel) (*model.Tmodel, error)
	Update(test *model.Tmodel) (*model.Tmodel, error)
	Delete(id uuid.UUID) error
}

type TestService struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

func NewTestService() ITestService {
	var service ITestService

	app.Invoke(func(db *gorm.DB, logger *zap.SugaredLogger) {
		service = TestService{
			db:     db,
			logger: logger,
		}
	})

	return service
}

// Delete implements ITestService.
func (t TestService) Delete(id uuid.UUID) error {
	// NOTE: this doesn't actually delete the entry rather sets deleted at field to now
	rez := t.db.Delete(&model.Tmodel{}, "uuid = ?", id)
	t.logger.Debugf("Delete statment on uuid = %s, rez %+v", id, rez)
	return rez.Error
}

// Create implements ITestService.
func (t TestService) Create(test *model.Tmodel) (*model.Tmodel, error) {
	rez := t.db.Create(&test)
	if rez.Error != nil {
		return nil, rez.Error
	}
	return test, nil
}

// Read implements ITestService.
func (t TestService) Read(id uuid.UUID) (*model.Tmodel, error) {
	var tmodel model.Tmodel
	rez := t.db.Where("uuid = ?", id).First(&tmodel)
	if rez.Error != nil {
		return nil, rez.Error
	}
	return &tmodel, nil
}

// ReadAll implements ITestService.
func (t TestService) ReadAll() ([]model.Tmodel, error) {
	var tmodels []model.Tmodel
	rez := t.db.Find(&tmodels)
	if rez.Error != nil {
		return []model.Tmodel{}, rez.Error
	}

	return tmodels, nil
}

// Update implements ITestService.
func (t TestService) Update(test *model.Tmodel) (*model.Tmodel, error) {
	panic("unimplemented Update")
}
