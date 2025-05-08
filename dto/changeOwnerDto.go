package dto

type ChangeOwnerDto struct {
	VehicleUuid  string `json:"vehicleUuid" binding:"required,vehicleUuid"`
	NewOwnerUuid string `json:"newOwnerUuid" binding:"required,newOwnerUuid"`
}
