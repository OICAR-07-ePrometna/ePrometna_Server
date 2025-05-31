package dto

type MobileDto struct {
	RegisteredDevice string `json:"registeredDevice" binding:"required"`
	CreatedAt        string `json:"createdAt" binding:"required"`
}
