package service

import (
	"ePrometna_Server/app"
	"ePrometna_Server/model"
	"ePrometna_Server/util/auth"
	"ePrometna_Server/util/cerror"
	"ePrometna_Server/util/device"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MobileLoginResult contains tokens for a mobile login
type MobileLoginResult struct {
	AccessToken  string
	RefreshToken string
	DeviceToken  string
}

type ILoginService interface {
	Login(email, password string) (string, string, error)
	RefreshTokens(user *model.User) (string, string, error)
	LoginMobile(email, password string, deviceInfo device.DeviceInfo) (*MobileLoginResult, error)
	RegisterPolice(code string, deviceInfo device.DeviceInfo) (*MobileLoginResult, error)
}

type LoginService struct {
	db            *gorm.DB
	logger        *zap.SugaredLogger
	deviceManager *device.DeviceManager
}

func NewLoginService() ILoginService {
	var service ILoginService

	app.Invoke(func(db *gorm.DB, logger *zap.SugaredLogger) {
		// Initialize the device manager
		deviceManager := device.NewDeviceManager()

		service = &LoginService{
			db:            db,
			logger:        logger,
			deviceManager: deviceManager,
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

// LoginMobile authenticates a user and manages their device registration
func (s *LoginService) LoginMobile(email, password string, deviceInfo device.DeviceInfo) (*MobileLoginResult, error) {
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

	// Handle device registration through the device manager
	deviceToken, _, err := s.deviceManager.ValidateDeviceRegistration(&user, deviceInfo)
	if err != nil {
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

func (s *LoginService) RegisterPolice(code string, deviceInfo device.DeviceInfo) (*MobileLoginResult, error) {
	// Get user from database
	var user model.User
	if err := s.db.
		Where("police_token = ?", code).
		First(&user).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Debugf("Failed to register officer", err)
			return nil, cerror.ErrInvalidCredentials
		}

		s.logger.Errorf("Error finding police officer, error = %+v", err)
		return nil, err
	}

	// Hash the police token after successful authentication
	hashedToken, err := auth.HashPassword(code)
	if err != nil {
		s.logger.Errorf("Failed to hash police token, error = %+v", err)
		return nil, err
	}

	// Update the police token with the hashed version
	if err := s.db.Model(&user).Update("police_token", hashedToken).Error; err != nil {
		s.logger.Errorf("Failed to update police token, error = %+v", err)
		return nil, err
	}

	accessToken, refreshToken, err := auth.GenerateTokens(&user)
	if err != nil {
		s.logger.Errorf("Failed to generate token error = %+v", err)
		return nil, err
	}

	// Handle device registration through the device manager
	deviceToken, _, err := s.deviceManager.ValidateDeviceRegistration(&user, deviceInfo)
	if err != nil {
		return nil, err
	}

	// Return all tokens
	return &MobileLoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		DeviceToken:  deviceToken,
	}, nil
}
