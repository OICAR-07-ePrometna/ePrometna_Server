package service

import (
	"ePrometna_Server/app"

	"ePrometna_Server/dto"
	"ePrometna_Server/model"
	"ePrometna_Server/util/auth"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ILoginService interface {
	Login(username, password string) (*dto.TokenResponse, error)
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

func (s *LoginService) Login(email, password string) (*dto.TokenResponse, error) {
	var user model.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zap.S().Warnf("User not found Email = %s", user.Email)
			return nil, errors.New("invalid username or password")
		}

		zap.S().Errorf("Failed to query user, error = %+v", err)
		return nil, err
	}

	if !auth.VerifyPassword(user.PasswordHash, password) {
		zap.S().Warnf("Invalid password for user Email: %s, uuid: %s", user.Email, user.Uuid)
		return nil, errors.New("invalid username or password")
	}

	token, refresh, err := auth.GenerateTokens(&user)
	if err != nil {
		zap.S().Errorf("Failed to generate token error = %+v", err)
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  token,
		RefreshToken: refresh,
	}, nil
}
