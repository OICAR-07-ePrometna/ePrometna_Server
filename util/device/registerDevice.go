package device

import (
	"database/sql"
	"ePrometna_Server/app"
	"ePrometna_Server/model"
	"ePrometna_Server/util/auth"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DeviceInfo contains information about a mobile device
type DeviceInfo struct {
	Platform  string `json:"platform"`
	Brand     string `json:"brand"`
	ModelName string `json:"modelName"`
	DeviceID  string `json:"deviceId"`
}

// DeviceManager handles device registration operations
type DeviceManager struct {
	DB     *gorm.DB
	Logger *zap.SugaredLogger
}

// NewDeviceManager creates a new device manager
func NewDeviceManager() *DeviceManager {
	var service *DeviceManager

	app.Invoke(func(db *gorm.DB, logger *zap.SugaredLogger) {
		// Initialize the device manager
		service = &DeviceManager{
			DB:     db,
			Logger: logger,
		}
	})
	return service
}

// FormatDeviceName creates a consistent device name format
func (dm *DeviceManager) FormatDeviceName(deviceInfo DeviceInfo) string {
	deviceName := deviceInfo.Brand + " " + deviceInfo.ModelName + " (" + deviceInfo.Platform + ")"
	return deviceName + " [ID:" + deviceInfo.DeviceID + "]"
}

// FindDeviceByID checks if a device with the given ID is registered to any user
func (dm *DeviceManager) FindDeviceByID(deviceID string) (*model.Mobile, error) {
	var device model.Mobile
	result := dm.DB.Where("registered_device LIKE ?", "%"+deviceID+"%").First(&device)

	if result.Error != nil {
		return nil, result.Error
	}

	return &device, nil
}

// FindUserDevice checks if a user has any registered device
func (dm *DeviceManager) FindUserDevice(userID uint) (*model.Mobile, error) {
	var device model.Mobile
	result := dm.DB.Where("user_id = ?", userID).First(&device)

	if result.Error != nil {
		return nil, result.Error
	}

	return &device, nil
}

// UpdateDeviceToken updates the activation token for an existing device
func (dm *DeviceManager) UpdateDeviceToken(device *model.Mobile, user *model.User) (string, error) {
	deviceToken, err := auth.GenerateDeviceToken(user)
	if err != nil {
		return "", err
	}

	device.ActivationToken = deviceToken
	if err := dm.DB.Save(device).Error; err != nil {
		return "", err
	}

	return deviceToken, nil
}

// RegisterNewDevice creates a new device registration for a user
func (dm *DeviceManager) RegisterNewDevice(user *model.User, deviceName string) (string, error) {
	deviceToken, err := auth.GenerateDeviceToken(user)
	if err != nil {
		return "", err
	}

	newDevice := model.Mobile{
		Uuid:             uuid.New(),
		UserId:           user.ID,
		CreatorId:        user.ID,
		RegisteredDevice: deviceName,
		ActivationToken:  deviceToken,
	}

	if err := dm.DB.Create(&newDevice).Error; err != nil {
		return "", err
	}

	return deviceToken, nil
}

// ValidateDeviceRegistration checks if the device can be registered or authenticated
// Returns:
// - deviceToken: the token if device is authenticated
// - isNewRegistration: true if device was newly registered
// - error: any error that occurred during the process
func (dm *DeviceManager) ValidateDeviceRegistration(user *model.User, deviceInfo DeviceInfo) (string, bool, error) {
	deviceIDStr := deviceInfo.DeviceID
	deviceName := dm.FormatDeviceName(deviceInfo)

	var deviceToken string
	var isNewRegistration bool

	// Start a transaction with SERIALIZABLE isolation level to prevent race conditions
	err := dm.DB.Transaction(func(tx *gorm.DB) error {
		// STEP 1: Check if this device is already registered to ANY user
		var existingDevice model.Mobile
		result := tx.Where("registered_device LIKE ?", "%"+deviceIDStr+"%").First(&existingDevice)

		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			dm.Logger.Errorf("Error checking device registration: %v", result.Error)
			return result.Error
		}

		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Device is already registered
			if existingDevice.UserId == user.ID {
				// Registered to THIS user, update token
				var err error
				deviceTokenTemp, err := auth.GenerateDeviceToken(user)
				if err != nil {
					dm.Logger.Errorf("Failed to generate device token: %v", err)
					return err
				}

				existingDevice.ActivationToken = deviceTokenTemp
				if err := tx.Save(&existingDevice).Error; err != nil {
					dm.Logger.Errorf("Failed to update device token: %v", err)
					return err
				}

				deviceToken = deviceTokenTemp
				isNewRegistration = false
				dm.Logger.Infof("Device already registered to user %s, updated token", user.Email)
				return nil
			} else {
				// Registered to ANOTHER user
				dm.Logger.Warnf("Device with ID %s is already registered to another user", deviceIDStr)
				return fmt.Errorf("this device is already registered to another user")
			}
		}

		// STEP 2: Check if this user already has ANY registered device
		var userDevice model.Mobile
		result = tx.Where("user_id = ?", user.ID).First(&userDevice)

		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			dm.Logger.Errorf("Error checking user device: %v", result.Error)
			return result.Error
		}

		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// User already has a registered device
			dm.Logger.Warnf("User %s already has a different registered device", user.Email)
			return fmt.Errorf("you already have a different device registered to your account")
		}

		// STEP 3: Register new device
		dm.Logger.Infof("Registering new device for user Email: %s, Device: %s", user.Email, deviceName)

		deviceTokenTemp, err := auth.GenerateDeviceToken(user)
		if err != nil {
			return err
		}

		newDevice := model.Mobile{
			Uuid:             uuid.New(),
			UserId:           user.ID,
			CreatorId:        user.ID,
			RegisteredDevice: deviceName,
			ActivationToken:  deviceTokenTemp,
		}

		if err := tx.Create(&newDevice).Error; err != nil {
			return err
		}

		deviceToken = deviceTokenTemp
		isNewRegistration = true
		return nil
	}, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return "", false, err
	}

	return deviceToken, isNewRegistration, nil
}
