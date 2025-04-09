package service

import (
	"ePrometna_Server/app"
	"ePrometna_Server/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IVehicleService interface {
	ReadAll() ([]model.Vehicle, error)
	Read(id uuid.UUID) (*model.Vehicle, error)
	Create(test *model.Vehicle) (*model.Vehicle, error)
	Update(test *model.Vehicle) (*model.Vehicle, error)
	Delete(id uuid.UUID) error
}

// TODO: implement service
type VehicleService struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

func NewVehicleService() IVehicleService {
	var service IVehicleService
	app.Invoke(func(db *gorm.DB, logger *zap.SugaredLogger) {
		service = &VehicleService{
			db:     db,
			logger: logger,
		}
	})
	return service
}

// Create implements IVehicleService.
func (v *VehicleService) Create(test *model.Vehicle) (*model.Vehicle, error) {
	panic("unimplemented")
}

// Delete implements IVehicleService.
func (v *VehicleService) Delete(id uuid.UUID) error {
	panic("unimplemented")
}

// Read implements IVehicleService.
func (v *VehicleService) Read(id uuid.UUID) (*model.Vehicle, error) {
	panic("unimplemented")
}

// ReadAll implements IVehicleService.
func (v *VehicleService) ReadAll() ([]model.Vehicle, error) {
	panic("unimplemented")
}

// Update implements IVehicleService.
func (v *VehicleService) Update(test *model.Vehicle) (*model.Vehicle, error) {
	panic("unimplemented")
}
