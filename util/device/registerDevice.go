package device

import (
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
func NewDeviceManager(db *gorm.DB, logger *zap.SugaredLogger) *DeviceManager {
	return &DeviceManager{
		DB:     db,
		Logger: logger,
	}
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
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Device not found, not an error
		}
		return nil, result.Error // Database error
	}

	return &device, nil
}

// FindUserDevice checks if a user has any registered device
func (dm *DeviceManager) FindUserDevice(userID uint) (*model.Mobile, error) {
	var device model.Mobile
	result := dm.DB.Where("user_id = ?", userID).First(&device)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // User has no device, not an error
		}
		return nil, result.Error // Database error
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

	// STEP 1: Check if this device is already registered to ANY user
	existingDevice, err := dm.FindDeviceByID(deviceIDStr)
	if err != nil {
		dm.Logger.Errorf("Error checking device registration: %v", err)
		return "", false, err
	}

	if existingDevice != nil {
		// Device is already registered
		if existingDevice.UserId == user.ID {
			// Registered to THIS user, update token
			deviceToken, err := dm.UpdateDeviceToken(existingDevice, user)
			if err != nil {
				dm.Logger.Errorf("Failed to update device token: %v", err)
				return "", false, err
			}

			dm.Logger.Infof("Device already registered to user %s, updated token", user.Email)
			return deviceToken, false, nil
		} else {
			// Registered to ANOTHER user
			dm.Logger.Warnf("Device with ID %s is already registered to another user", deviceIDStr)
			return "", false, fmt.Errorf("this device is already registered to another user")
		}
	}

	// STEP 2: Check if this user already has ANY registered device
	userDevice, err := dm.FindUserDevice(user.ID)
	if err != nil {
		dm.Logger.Errorf("Error checking user device: %v", err)
		return "", false, err
	}

	if userDevice != nil {
		// User already has a registered device
		dm.Logger.Warnf("User %s already has a different registered device", user.Email)
		return "", false, fmt.Errorf("you already have a different device registered to your account")
	}

	// STEP 3: Register new device
	dm.Logger.Infof("Registering new device for user Email: %s, Device: %s", user.Email, deviceName)
	deviceToken, err := dm.RegisterNewDevice(user, deviceName)
	if err != nil {
		dm.Logger.Errorf("Failed to register device: %v", err)
		return "", false, err
	}

	return deviceToken, true, nil
}
