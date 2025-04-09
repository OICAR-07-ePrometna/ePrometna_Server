package model

// NOTE: Here register all models to be used in migration

// TODO: see about design or different name

// GetAllModels returns an array of all registered models
func GetAllModels() []any {
	return []any{
		&Tmodel{},
		&User{},
		&Vehicle{},
		&VehicleDrivers{},
		&DriverLicense{},
		&OwnerHistory{},
		&Mobile{},
		&RegistrationInfo{},
		&TempData{},
	}
}
