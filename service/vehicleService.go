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
	ReadByVin(vin string) (*model.Vehicle, error)
	Create(newVehicle *model.Vehicle, ownerUuid uuid.UUID) (*model.Vehicle, error)
	Delete(uuid uuid.UUID) error
	ChangeOwner(vehicle uuid.UUID, newOwner uuid.UUID) error
	Registration(vehicleUuid uuid.UUID, model model.RegistrationInfo) error
	Update(vehicleUuid uuid.UUID, model model.Vehicle) (*model.Vehicle, error)
	Deregister(vehicleUuid uuid.UUID) error
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
	var vehicle model.Vehicle
	// TODO: see what to do with other objects

	rez := v.db.
		InnerJoins("Registration").
		Preload("Owner").
		Where("vehicles.uuid = ?", _uuid).
		First(&vehicle)

	if rez.Error != nil {
		return nil, rez.Error
	}

	if err := v.loadRegistreation(&vehicle); err != nil {
		return nil, err
	}

	return &vehicle, nil
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

	for _, vehicle := range vehicles {
		if err := v.loadRegistreation(&vehicle); err != nil {
			return nil, err
		}
	}

	return vehicles, nil
}

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
			UserId:    *vehicle.UserId, // Old owner
		}
		if err := v.db.Create(&pastOwnerEntry).Error; err != nil {
			return err
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

		zap.S().Info("Vehicle UUID %s: Registration ID: %d", vehicle.Uuid, vehicle.RegistrationID)

		if err := tx.Save(&vehicle).Error; err != nil {
			v.logger.Errorf("Failed to save vehicle (ID: %d) with updated current registration (RegistrationInfo ID: %d): %+v", vehicle.ID, newRegInfo.ID, err)
			return err
		}

		v.logger.Infof("Successfully updated vehicle UUID %s (ID: %d) to set new current registration (RegistrationInfo ID: %d, UUID: %s).", vehicle.Uuid, vehicle.ID, newRegInfo.ID, newRegInfo.Uuid)
		return nil
	})
}

// ReadByVin implements IVehicleService.
func (v *VehicleService) ReadByVin(vin string) (*model.Vehicle, error) {
	v.logger.Debugf("Attempting to read vehicle with vin = %s ", vin)
	var vehicle model.Vehicle

	rez := v.db.
		InnerJoins("Registration").
		Preload("Owner").
		Where("vehicles.chassis_number = ?", vin).
		First(&vehicle)

	if rez.Error != nil {
		return nil, rez.Error
	}

	if err := v.loadRegistreation(&vehicle); err != nil {
		return nil, err
	}

	v.logger.Debugf("Vehicle reg id = %+v", *vehicle.RegistrationID)
	v.logger.Debugf("Vehicle reg = %+v", vehicle.Registration.Registration)
	return &vehicle, nil
}

// Deregister implements IVehicleService.
func (v *VehicleService) Deregister(vehicleUuid uuid.UUID) error {
	v.logger.Debugf("Attempting to deregister vehicle with UUID: %s", vehicleUuid)

	return v.db.Transaction(func(tx *gorm.DB) error {
		var vehicle model.Vehicle
		if err := tx.
			Preload("Registration").
			Where("uuid = ?", vehicleUuid).
			First(&vehicle).
			Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				v.logger.Warnf("Vehicle with UUID = %s not found for deregistration.", vehicleUuid)
				return gorm.ErrRecordNotFound
			}
			v.logger.Errorf("Failed to find vehicle with UUID = %s: %+v", vehicleUuid, err)
			return err
		}

		v.logger.Debugf("Found vehicle (ID: %d) for deregistration.", vehicle.ID)

		if vehicle.Registration != nil {
			v.logger.Infof("Vehicle UUID %s (ID: %d) has an active registration (RegistrationInfo ID: %d). This registration will be moved to past registrations.", vehicle.Uuid, vehicle.ID, vehicle.Registration.ID)
			if err := tx.Model(&vehicle).Omit("RegistrationID").
				Association("PastRegistration").Append(vehicle.Registration); err != nil {
				return err
			}
		}

		// Set RegistrationID to nil to deregister the vehicle
		vehicle.RegistrationID = nil
		vehicle.Registration = nil

		if err := tx.Save(&vehicle).Error; err != nil {
			v.logger.Errorf("Failed to save vehicle (ID: %d) with null registration: %+v", vehicle.ID, err)
			return err
		}

		v.logger.Infof("Successfully deregistered vehicle UUID %s (ID: %d).", vehicle.Uuid, vehicle.ID)
		return nil
	})
}

func (v *VehicleService) Update(vehicleUuid uuid.UUID, newVehicle model.Vehicle) (*model.Vehicle, error) {
	v.logger.Debugf("Attempting to update vehicle with UUID: %s", vehicleUuid)

	var existingVehicle model.Vehicle
	if err := v.db.Where("uuid = ?", vehicleUuid).First(&existingVehicle).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			v.logger.Warnf("Vehicle with UUID = %s not found for update.", vehicleUuid)
			return nil, gorm.ErrRecordNotFound
		}
		v.logger.Errorf("Failed to find vehicle with UUID = %s: %+v", vehicleUuid, err)
		return nil, err
	}

	v.logger.Debugf("Found vehicle (ID: %d) for update. Current data: %+v", existingVehicle.ID, existingVehicle)
	v.logger.Debugf("New data for update: %+v", newVehicle)

	existingVehicle.Update(newVehicle)

	if err := v.db.Save(&existingVehicle).Error; err != nil {
		v.logger.Errorf("Failed to save updated vehicle (ID: %d, UUID: %s): %+v", existingVehicle.ID, existingVehicle.Uuid, err)
		return nil, err
	}

	v.logger.Infof("Successfully updated vehicle UUID %s (ID: %d).", existingVehicle.Uuid, existingVehicle.ID)

	return &existingVehicle, nil
}

func (v *VehicleService) loadRegistreation(vehicle *model.Vehicle) error {
	if vehicle.RegistrationID != nil {
		return nil
	}

	vehicle.Registration.ID = *vehicle.RegistrationID
	rez := v.db.
		Where("id = ?", *vehicle.RegistrationID).
		First(&vehicle.Registration)

	return rez.Error
}
