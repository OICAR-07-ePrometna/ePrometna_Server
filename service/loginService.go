package service

import (
	"ePrometna_Server/app"
	"ePrometna_Server/model"
	"ePrometna_Server/utils"
	"errors"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ILoginService interface {
	Login(email, password string) (string, string, error)
	GenerateToken(user *model.User) (string, string, error)
}

type LoginService struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewLoginService() ILoginService {
	var service ILoginService

	app.Invoke(func(db *gorm.DB, logger *zap.Logger) {
		service = &LoginService{
			db:     db,
			logger: logger,
		}
	})

	return service
}

func VerifyPassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

func (s *LoginService) Login(email, password string) (string, string, error) {
	var user model.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zap.S().Warn("User not found", zap.String("email", email))
			return "", "", errors.New("invalid email or password")
		}
		zap.S().Error("Database error", zap.Error(err))
		return "", "", err
	}

	if !VerifyPassword(user.PasswordHash, password) {
		zap.S().Warn("Invalid password", zap.String("email", email))
		return "", "", errors.New("invalid email or password")
	}

	accessToken, refreshToken, err := s.GenerateToken(&user)
	if err != nil {
		zap.S().Error("Failed to generate token", zap.Error(err))
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *LoginService) GenerateToken(user *model.User) (string, string, error) {
	accessToken, refreshToken, err := utils.GenerateTokens(*user)
	if err != nil {
		zap.S().Error("Failed to generate tokens", zap.Error(err))
		return "", "", err
	}
	return accessToken, refreshToken, nil
}
