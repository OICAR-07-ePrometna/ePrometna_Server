package auth

import (
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func VerifyPassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		zap.S().Debugf("Failed to hash password err = %+v", err)
		return "", err
	}

	return string(hash), nil
}
