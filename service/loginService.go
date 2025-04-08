package service

import (
	"ePrometna_Server/app"
	"ePrometna_Server/model"
	"ePrometna_Server/util/auth"
	"ePrometna_Server/util/cerror"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ILoginService interface {
	Login(email, password string) (string, string, error)
	RefreshTokens(user *model.User) (string, string, error)
}

type LoginService struct {
	db *gorm.DB
}

func NewLoginService() ILoginService {
	var service ILoginService

	app.Invoke(func(db *gorm.DB) {
		service = &LoginService{
			db: db,
		}
	})

	return service
}

func (s *LoginService) Login(email, password string) (string, string, error) {
	var user model.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zap.S().Debugf("User not found Email = %s", email)
			return "", "", cerror.ErrInvalidCredentials
		}

		zap.S().Errorf("Failed to query user, error = %+v", err)
		return "", "", err
	}

	if !auth.VerifyPassword(user.PasswordHash, password) {
		zap.S().Debugf("Invalid password for user Email: %s, uuid: %s", user.Email, user.Uuid)
		return "", "", cerror.ErrInvalidCredentials
	}

	token, refresh, err := auth.GenerateTokens(&user)
	if err != nil {
		zap.S().Errorf("Failed to generate token error = %+v", err)
		return "", "", err
	}

	return token, refresh, nil
}

// Refresh implements ILoginService.
func (s *LoginService) RefreshTokens(user *model.User) (string, string, error) {
	return auth.GenerateTokens(user)
}
