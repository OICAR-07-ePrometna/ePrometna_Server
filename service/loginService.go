package service

import (
	"ePrometna_Server/app"
	"ePrometna_Server/model"
	"errors"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ILoginService interface {
	Login(username, password string) (*model.User, error)
	GenerateToken(user *model.User) (string, error)
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

func verifyPassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

func (s *LoginService) Login(username, password string) (*model.User, error) {
	var user model.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zap.S().Warn("User not found", zap.String("username", username))
			return nil, errors.New("Invalid username or password")
		}
		zap.S().Error("Failed to query user", zap.Error(err))
		return nil, err
	}

	if !verifyPassword(user.PasswordHash, password) {
		zap.S().Warn("Invalid password", zap.String("username", username))
		return nil, errors.New("Invalid username or password")
	}

	return &user, nil
}

func (s *LoginService) GenerateToken(user *model.User) (string, error) {
	token, err := app.GenerateTokens(user.ID, user.Email)
	if err != nil {
		zap.S().Error("Failed to generate token", zap.Error(err))
		return "", err
	}
	return token, nil
}
