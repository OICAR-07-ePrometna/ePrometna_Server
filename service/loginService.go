package service

import (
	"ePrometna_Server/app"
	"ePrometna_Server/model"
	"ePrometna_Server/util/auth"
	"ePrometna_Server/util/cerror"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DeviceInfo struct {
	Platform  string `json:"platform"`
	Brand     string `json:"brand"`
	ModelName string `json:"modelName"`
	DeviceID  string `json:"deviceId"`
}

type MobileLoginResult struct {
	AccessToken  string
	RefreshToken string
	DeviceToken  string
}

type ILoginService interface {
	Login(email, password string) (string, string, error)
	RefreshTokens(user *model.User) (string, string, error)
	LoginMobile(email, password string, deviceInfo DeviceInfo) (*MobileLoginResult, error)
}

type LoginService struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

func NewLoginService() ILoginService {
	var service ILoginService

	app.Invoke(func(db *gorm.DB, logger *zap.SugaredLogger) {
		service = &LoginService{
			db:     db,
			logger: logger,
		}
	})

	return service
}

func (s *LoginService) Login(email, password string) (string, string, error) {
	var user model.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Debugf("User not found Email = %s", email)
			return "", "", cerror.ErrInvalidCredentials
		}

		s.logger.Errorf("Failed to query user, error = %+v", err)
		return "", "", err
	}

	if !auth.VerifyPassword(user.PasswordHash, password) {
		s.logger.Debugf("Invalid password for user Email: %s, uuid: %s", user.Email, user.Uuid)
		return "", "", cerror.ErrInvalidCredentials
	}

	token, refresh, err := auth.GenerateTokens(&user)
	if err != nil {
		s.logger.Errorf("Failed to generate token error = %+v", err)
		return "", "", err
	}

	return token, refresh, nil
}

// formatDeviceName creates a consistent device name format
func (s *LoginService) formatDeviceName(deviceInfo DeviceInfo) string {
	deviceName := deviceInfo.Brand + " " + deviceInfo.ModelName + " (" + deviceInfo.Platform + ")"
	return deviceName + " [ID:" + deviceInfo.DeviceID + "]"
}

// findDeviceByID checks if a device with the given ID is registered to any user
func (s *LoginService) findDeviceByID(deviceID string) (*model.Mobile, error) {
	var device model.Mobile
	result := s.db.Where("registered_device LIKE ?", "%"+deviceID+"%").First(&device)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Device not found, not an error
		}
		return nil, result.Error // Database error
	}

	return &device, nil
}

// findUserDevice checks if a user has any registered device
func (s *LoginService) findUserDevice(userID uint) (*model.Mobile, error) {
	var device model.Mobile
	result := s.db.Where("user_id = ?", userID).First(&device)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // User has no device, not an error
		}
		return nil, result.Error // Database error
	}

	return &device, nil
}

// updateDeviceToken updates the activation token for an existing device
func (s *LoginService) updateDeviceToken(device *model.Mobile, user *model.User) (string, error) {
	deviceToken, err := auth.GenerateDeviceToken(user)
	if err != nil {
		return "", err
	}

	device.ActivationToken = deviceToken
	if err := s.db.Save(device).Error; err != nil {
		return "", err
	}

	return deviceToken, nil
}

// registerNewDevice creates a new device registration for a user
func (s *LoginService) registerNewDevice(user *model.User, deviceName string) (string, error) {
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

	if err := s.db.Create(&newDevice).Error; err != nil {
		return "", err
	}

	return deviceToken, nil
}

// LoginMobile authenticates a user and manages their device registration
func (s *LoginService) LoginMobile(email, password string, deviceInfo DeviceInfo) (*MobileLoginResult, error) {
	// Authenticate user
	accessToken, refreshToken, err := s.Login(email, password)
	if err != nil {
		return nil, err
	}

	// Get user from database
	var user model.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		s.logger.Errorf("Failed to find user after login, error = %+v", err)
		return nil, err
	}

	// Format device info
	deviceIDStr := deviceInfo.DeviceID
	deviceName := s.formatDeviceName(deviceInfo)

	// STEP 1: Check if this device is already registered to ANY user
	existingDevice, err := s.findDeviceByID(deviceIDStr)
	if err != nil {
		s.logger.Errorf("Error checking device registration: %v", err)
		return nil, err
	}

	if existingDevice != nil {
		// Device is already registered
		if existingDevice.UserId == user.ID {
			// Registered to THIS user, update token
			deviceToken, err := s.updateDeviceToken(existingDevice, &user)
			if err != nil {
				s.logger.Errorf("Failed to update device token: %v", err)
				return nil, err
			}

			s.logger.Infof("Device already registered to user %s, updated token", user.Email)
			return &MobileLoginResult{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				DeviceToken:  deviceToken,
			}, nil
		} else {
			// Registered to ANOTHER user
			s.logger.Warnf("Device with ID %s is already registered to another user", deviceIDStr)
			return nil, fmt.Errorf("this device is already registered to another user")
		}
	}

	// STEP 2: Check if this user already has ANY registered device
	userDevice, err := s.findUserDevice(user.ID)
	if err != nil {
		s.logger.Errorf("Error checking user device: %v", err)
		return nil, err
	}

	if userDevice != nil {
		// User already has a registered device
		s.logger.Warnf("User %s already has a different registered device", user.Email)
		return nil, fmt.Errorf("you already have a different device registered to your account")
	}

	// STEP 3: Register new device
	s.logger.Infof("Registering new device for user Email: %s, Device: %s", user.Email, deviceName)
	deviceToken, err := s.registerNewDevice(&user, deviceName)
	if err != nil {
		s.logger.Errorf("Failed to register device: %v", err)
		return nil, err
	}

	// Return all tokens
	return &MobileLoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		DeviceToken:  deviceToken,
	}, nil
}

func (s *LoginService) RefreshTokens(user *model.User) (string, string, error) {
	return auth.GenerateTokens(user)
}
