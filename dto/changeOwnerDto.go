package dto

type ChangeOwnerDto struct {
	VehicleUuid  string `json:"vehicleUuid" binding:"required,uuid"`
	NewOwnerUuid string `json:"newOwnerUuid" binding:"required,uuid"`
}
