package service

import (
	"ePrometna_Server/app"
	"ePrometna_Server/model"
	"ePrometna_Server/util/cerror"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IVehicleService interface {
	ReadAll(driverUuid uuid.UUID) ([]model.Vehicle, error)
	Read(uuid uuid.UUID) (*model.Vehicle, error)
	Create(newVehicle *model.Vehicle, ownerUuid uuid.UUID) (*model.Vehicle, error)
	Delete(uuid uuid.UUID) error
}

type VehicleService struct {
	db          *gorm.DB
	userService IUserCrudService
	logger      *zap.SugaredLogger
}

func NewVehicleService() IVehicleService {
	var service IVehicleService
	app.Invoke(func(db *gorm.DB, logger *zap.SugaredLogger, uService IUserCrudService) {
		service = &VehicleService{
			db:          db,
			logger:      logger,
			userService: uService,
		}
	})
	return service
}

// Create implements IVehicleService.
func (v *VehicleService) Create(vehicle *model.Vehicle, ownerUuid uuid.UUID) (*model.Vehicle, error) {
	owner, err := v.userService.Read(ownerUuid)
	if err != nil {
		v.logger.Errorf("Error reading user with uuid = %s, err = %+v", ownerUuid, err)
		return nil, cerror.ErrUserIsNil
	}

	//  NOTE: Users that are not roles Firma or Osoba are now allowed to own a car
	if owner.Role != model.RoleFirma && owner.Role != model.RoleOsoba {
		v.logger.Errorf("User with role %+v can't own a car", owner.Role)
		return nil, cerror.ErrBadRole
	}

	vehicle.UserId = owner.ID

	v.logger.Debugf("Creating new registration = %+v", vehicle.Registration)
	v.logger.Debugf("Creating new vehicle = %+v", vehicle)
	rez := v.db.Create(&vehicle)
	if rez.Error != nil {
		v.logger.Errorf("Failed creating new vehicle = %+v", vehicle)
		return nil, rez.Error
	}

	return vehicle, nil
}

// Delete implements IVehicleService.
func (v *VehicleService) Delete(_uuid uuid.UUID) error {
	rez := v.db.Where("uuid = ?", _uuid).Delete(&model.Vehicle{})
	v.logger.Debugf("Delete statement on uuid = %s, rez = %+v", _uuid, rez)
	if rez.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return rez.Error
}

// Read implements IVehicleService.
func (v *VehicleService) Read(_uuid uuid.UUID) (*model.Vehicle, error) {
	var user model.Vehicle
	// TODO: see what to do with other objects
	rez := v.db.InnerJoins("Registration").Where("vehicles.uuid = ?", _uuid).First(&user)
	if rez.Error != nil {
		return nil, rez.Error
	}

	return &user, nil
}

// ReadAll implements IVehicleService.
func (v *VehicleService) ReadAll(driverUuid uuid.UUID) ([]model.Vehicle, error) {
	vehicles := make([]model.Vehicle, 0)

	// TODO: read vehicles that other people borrowed you
	rez := v.db.InnerJoins("Registration").Joins("inner join users on vehicles.user_id = users.id").Where("users.uuid = ?", driverUuid).Find(&vehicles)
	if rez.Error != nil {
		return nil, rez.Error
	}
	return vehicles, nil
}
