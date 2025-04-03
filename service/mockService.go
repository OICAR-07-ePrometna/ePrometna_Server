package service

import (
	"ePrometna_Server/model"
	"ePrometna_Server/utils"
	"errors"

	"github.com/google/uuid"
)

type MockLoginService struct {
	mockUsers map[string]string // Mock users with email as key and password as value
}

func NewMockLoginService() *MockLoginService {
	return &MockLoginService{
		mockUsers: map[string]string{
			"user1@example.com": "password1",
			"user2@example.com": "password2",
		},
	}
}

func NewMockLoginServiceWithUsers(users map[string]string) *MockLoginService {
	return &MockLoginService{
		mockUsers: users,
	}
}
func (m *MockLoginService) AddMockUser(email, password string) {
	m.mockUsers[email] = password
}

func (m *MockLoginService) Login(email, password string) (string, string, error) {
	if mockPassword, exists := m.mockUsers[email]; exists {
		if mockPassword == password {

			user := model.User{
				Email: email,
				Uuid:  uuid.New(),
				Role:  model.RoleSuperAdmin,
			}

			accessToken, refreshToken, err := utils.GenerateTokens(user)
			if err != nil {
				return "", "", err
			}

			return accessToken, refreshToken, nil
		}
		return "", "", errors.New("invalid password")
	}
	return "", "", errors.New("user not found")
}
