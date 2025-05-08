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
	ChangeOwner(vehicle uuid.UUID, newOwner uuid.UUID) error
	Registration(model.RegistrationInfo) error
}

// TODO: implement service
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
	// TODO: Create other objects

	owner, err := v.userService.Read(ownerUuid)
	if err != nil {
		v.logger.Errorf("Error reading user with uuid = %s, err = %+v", ownerUuid, err)
		return nil, err
	}

	// NOTE: Users that are not roles Firma or Osoba are now allowed to own a car
	if owner.Role != model.RoleFirma && owner.Role != model.RoleOsoba {
		v.logger.Errorf("User with role %+v can't own a car", owner.Role)
		return nil, cerror.ErrBadRole
	}

	vehicle.UserId = &owner.ID

	v.logger.Debugf("Creating new vehicle %+v", vehicle)
	rez := v.db.Create(&vehicle)
	if rez.Error != nil {
		return nil, rez.Error
	}

	return vehicle, nil
}

// Delete implements IVehicleService.
func (v *VehicleService) Delete(_uuid uuid.UUID) error {
	vehicle := model.Vehicle{}

	rez := v.db.
		Where("uuid = ?", _uuid).
		First(&vehicle)

	if rez.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	if rez.Error != nil {
		return rez.Error
	}

	// TODO: write a service for removing users and putting them into past woners
	vehicle.UserId = nil

	rez = v.db.
		Save(&vehicle)

	v.logger.Debugf("Update statment on uuid = %s, rez %+v", _uuid, rez)
	if rez.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	if rez.Error != nil {
		return rez.Error
	}

	rez = v.db.
		Delete(&vehicle)

	v.logger.Debugf("Delete statment on uuid = %s, rez %+v", _uuid, rez)
	if rez.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return rez.Error
}

// Read implements IVehicleService.
func (v *VehicleService) Read(_uuid uuid.UUID) (*model.Vehicle, error) {
	var user model.Vehicle
	// TODO: see what to do with other objects

	rez := v.db.
		InnerJoins("Registration").
		Preload("Owner").
		Where("vehicles.uuid = ?", _uuid).
		First(&user)

	if rez.Error != nil {
		return nil, rez.Error
	}

	return &user, nil
}

// ReadAll implements IVehicleService.
func (v *VehicleService) ReadAll(driverUuid uuid.UUID) ([]model.Vehicle, error) {
	vehicles := make([]model.Vehicle, 0)

	// TODO: read vehicles that other people borrowd you
	rez := v.db.
		InnerJoins("Registration").
		Joins("inner join users on vehicles.user_id = users.id").
		Where("users.uuid = ?", driverUuid).
		Find(&vehicles)

	if rez.Error != nil {
		return nil, rez.Error
	}
	return vehicles, nil
}

// TODO: test
// ChangeOwner implements IVehicleService.
func (v *VehicleService) ChangeOwner(vehicleUUID uuid.UUID, newOwner uuid.UUID) error {
	var user model.User
	rez := v.db.
		Where("uuid = ?", newOwner).
		First(&user)

	if rez.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	if rez.Error != nil {
		return rez.Error
	}

	var vehicle model.Vehicle
	rez = v.db.
		Preload("Owner").
		Preload("PaswOwners").
		Where("uuid = ?", vehicleUUID).
		First(&vehicle)

	if rez.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	if rez.Error != nil {
		return rez.Error
	}

	// Moving current user to past user
	oldUser := model.OwnerHistory{}
	oldUser.FromUser(*vehicle.Owner)
	vehicle.PastOwners = append(vehicle.PastOwners, oldUser)

	// assingn new user
	vehicle.UserId = &user.ID

	rez = v.db.
		Save(&vehicle)
	if rez.Error != nil {
		return rez.Error
	}

	return nil
}

// Registration implements IVehicleService.
func (v *VehicleService) Registration(reg model.RegistrationInfo) error {
	return nil
}
