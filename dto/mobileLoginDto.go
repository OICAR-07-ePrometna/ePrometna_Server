package dto

import "ePrometna_Server/util/device"

type MobileLoginDto struct {
	Email      string            `json:"email" binding:"required"`
	Password   string            `json:"password" binding:"required"`
	DeviceInfo device.DeviceInfo `json:"deviceInfo" binding:"required"`
}

type MobileLoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	DeviceToken  string `json:"deviceToken"`
}
