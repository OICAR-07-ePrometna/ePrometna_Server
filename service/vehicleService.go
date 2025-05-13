package service

import (
	"ePrometna_Server/app"
	"ePrometna_Server/model"
	"ePrometna_Server/util/cerror"
	"errors"
	"time"

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
	Registration(vehicleUuid uuid.UUID, model model.RegistrationInfo) error
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
	return v.db.Transaction(
		func(tx *gorm.DB) error {
			vehicle := model.Vehicle{}

			rez := tx.
				Where("uuid = ?", _uuid).
				First(&vehicle)

			if rez.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
			if rez.Error != nil {
				return rez.Error
			}

			// TODO: write a service for removing users and putting them into past oners
			vehicle.UserId = nil

			rez = tx.Save(&vehicle)
			v.logger.Debugf("Update statment on uuid = %s, rez %+v", _uuid, rez)
			if rez.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
			if rez.Error != nil {
				return rez.Error
			}

			rez = tx.Delete(&vehicle)
			v.logger.Debugf("Delete statment on uuid = %s, rez %+v", _uuid, rez)
			if rez.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}

			return rez.Error
		})
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
func (v *VehicleService) ChangeOwner(vehicleUUID uuid.UUID, newOwnerUuid uuid.UUID) error {
	var newOwner model.User
	rez := v.db.
		Where("uuid = ?", newOwnerUuid).
		First(&newOwner)

	if rez.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	if rez.Error != nil {
		return rez.Error
	}
	if newOwner.Role != model.RoleFirma && newOwner.Role != model.RoleOsoba {
		v.logger.Errorf("New owner (UUID: %s) with role '%s' cannot own a vehicle", newOwnerUuid, newOwner.Role)
		return cerror.ErrBadRole
	}

	var vehicle model.Vehicle
	rez = v.db.
		Preload("Owner").
		Preload("PastOwners").
		Where("uuid = ?", vehicleUUID).
		First(&vehicle)

	if rez.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	if rez.Error != nil {
		return rez.Error
	}

	if vehicle.Owner != nil && vehicle.UserId != nil {
		pastOwnerEntry := model.OwnerHistory{
			Uuid:      uuid.New(),
			VehicleId: vehicle.ID,
			UserId:    *vehicle.UserId, // This should be oldOwner.ID
		}
		if err := v.db.Create(&pastOwnerEntry).Error; err != nil { /* handle error */
		}
	}
	vehicle.UserId = &newOwner.ID
	vehicle.Owner = &newOwner

	rez = v.db.
		Save(&vehicle)
	if rez.Error != nil {
		return rez.Error
	}
	return nil
}

// Registration implements IVehicleService.
func (v *VehicleService) Registration(vehicleUuid uuid.UUID, newRegInfo model.RegistrationInfo) error {
	v.logger.Debugf("Attempting to register vehicle with UUID: %s", vehicleUuid)

	return v.db.Transaction(func(tx *gorm.DB) error {
		var vehicle model.Vehicle
		if err := tx.
			Preload("Registration").
			Where("uuid = ?", vehicleUuid).
			First(&vehicle).
			Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				v.logger.Warnf("Vehicle with UUID = %s not found for registration.", vehicleUuid)
				return gorm.ErrRecordNotFound
			}
			v.logger.Errorf("Failed to find vehicle with UUID = %s: %+v", vehicleUuid, err)
			return err
		}

		v.logger.Debugf("Found vehicle (ID: %d) for registration.", vehicle.ID)

		if vehicle.Registration != nil {
			v.logger.Infof("Vehicle UUID %s (ID: %d) already has an active registration (RegistrationInfo ID: %d). This registration will be superseded by the new one.", vehicle.Uuid, vehicle.ID, vehicle.Registration.ID)
			if err := tx.Model(&vehicle).Omit("RegistrationID").
				Association("PastRegistration").Append(vehicle.Registration); err != nil {
				return err
			}
		}

		newRegInfo.VehicleId = vehicle.ID
		newRegInfo.TechnicalDate = time.Now()

		if err := tx.Create(&newRegInfo).Error; err != nil {
			v.logger.Errorf("Failed to create new RegistrationInfo for vehicle ID %d (UUID: %s): %+v", vehicle.ID, newRegInfo.Uuid, err)
			return err
		}
		v.logger.Debugf("Successfully created new RegistrationInfo (ID: %d, UUID: %s) for vehicle ID %d.", newRegInfo.ID, newRegInfo.Uuid, vehicle.ID)

		vehicle.RegistrationID = &newRegInfo.ID
		vehicle.Registration = &newRegInfo

		if err := tx.Save(&vehicle).Error; err != nil {
			v.logger.Errorf("Failed to save vehicle (ID: %d) with updated current registration (RegistrationInfo ID: %d): %+v", vehicle.ID, newRegInfo.ID, err)
			return err
		}

		v.logger.Infof("Successfully updated vehicle UUID %s (ID: %d) to set new current registration (RegistrationInfo ID: %d, UUID: %s).", vehicle.Uuid, vehicle.ID, newRegInfo.ID, newRegInfo.Uuid)
		return nil
	})
}
