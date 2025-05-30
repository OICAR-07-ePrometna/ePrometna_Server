package dto

import "ePrometna_Server/util/device"

type MobileRegisterDto struct {
	Email      string            `json:"email" binding:"required"`
	Password   string            `json:"password" binding:"required"`
	DeviceInfo device.DeviceInfo `json:"deviceInfo" binding:"required"`
}

type PoliceRegisterDto struct {
	Code       string            `json:"code"`
	DeviceInfo device.DeviceInfo `json:"deviceInfo" binding:"required"`
}

type MobileLoginDto struct {
	DeviceToken string `json:"deviceToken"`
}

type DeviceLoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	DeviceToken  string `json:"deviceToken"`
}
