package auth_test

import (
	"ePrometna_Server/util/auth"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestVerifyPassword(t *testing.T) {
	hashedCorrect, _ := bcrypt.GenerateFromPassword([]byte("securePassword"), bcrypt.DefaultCost)
	hashedWrong, _ := bcrypt.GenerateFromPassword([]byte("anotherPassword"), bcrypt.DefaultCost)
	hashedEmpty, _ := bcrypt.GenerateFromPassword([]byte(""), bcrypt.DefaultCost)

	tests := []struct {
		name           string
		hashedPassword string
		plainPassword  string
		want           bool
	}{
		{
			name:           "Correct password",
			hashedPassword: string(hashedCorrect),
			plainPassword:  "securePassword",
			want:           true,
		},
		{
			name:           "Incorrect password",
			hashedPassword: string(hashedCorrect),
			plainPassword:  "wrongPassword",
			want:           false,
		},
		{
			name:           "Empty plain password, non-empty hash",
			hashedPassword: string(hashedCorrect),
			plainPassword:  "",
			want:           false,
		},
		{
			name:           "Non-empty plain password, empty hash",
			hashedPassword: string(hashedEmpty),
			plainPassword:  "notEmpty",
			want:           false,
		},
		{
			name:           "Empty plain password, empty hash",
			hashedPassword: string(hashedEmpty),
			plainPassword:  "",
			want:           true,
		},
		{
			name:           "Different correct passwords",
			hashedPassword: string(hashedWrong),
			plainPassword:  "securePassword",
			want:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := auth.VerifyPassword(tt.hashedPassword, tt.plainPassword)
			if got != tt.want {
				t.Errorf("VerifyPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name         string
		password     string
		wantNonEmpty bool
		wantErr      bool
	}{
		{
			name:         "Valid password",
			password:     "testPassword",
			wantNonEmpty: true,
			wantErr:      false,
		},
		{
			name:         "Empty password",
			password:     "",
			wantNonEmpty: true, // bcrypt will hash an empty string
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := auth.HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantNonEmpty && got == "" {
				t.Errorf("HashPassword() returned an empty hash for password: %s", tt.password)
			}
			if !tt.wantNonEmpty && got != "" {
				t.Errorf("HashPassword() returned a non-empty hash for password: %s", tt.password)
			}
			if got != "" && err == nil {
				// Basic verification that the hash works (not exhaustive)
				err := bcrypt.CompareHashAndPassword([]byte(got), []byte(tt.password))
				if err != nil {
					t.Errorf("HashPassword() generated an invalid hash for password: %s, error: %v", tt.password, err)
				}
			}
		})
	}
}
