package service

import (
	"database/sql"
	"ePrometna_Server/model"
	"ePrometna_Server/util/cerror"
	"ePrometna_Server/util/mock"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Helper function (no changes needed)
func setupVehicleServiceTest(t *testing.T) (IVehicleService, *mock.MockUserCrudService, sqlmock.Sqlmock) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
		// SkipInitializeWithVersion: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	mockUserService := new(mock.MockUserCrudService)
	zapLogger := zap.NewNop().Sugar()

	vehicleService := &VehicleService{
		db:          gormDB,
		userService: mockUserService,
		logger:      zapLogger,
	}

	return vehicleService, mockUserService, mockDB
}

// --- Test Cases ---

func TestVehicleService_Create(t *testing.T) {
	ownerUUID := uuid.New()
	vehicleUUID := uuid.New()
	validOwner := &model.User{
		Model: gorm.Model{ID: 101},
		Uuid:  ownerUUID,
		Role:  model.RoleOsoba,
	}
	invalidRoleOwner := &model.User{
		Model: gorm.Model{ID: 102},
		Uuid:  ownerUUID,
		Role:  model.RoleAdmin,
	}
	newVehicleInput := &model.Vehicle{
		Uuid:           vehicleUUID,
		VehicleType:    "Car",
		VehicleModel:   "Škoda Octavia",
		ProductionYear: 2023,
		ChassisNumber:  "SKDCHASSIS987654PQRS",
		Drivers:        nil,
		PastOwners:     nil,
		TemporaryData:  nil,
		Registration:   nil,
	}

	// --- Success Case (No changes needed) ---
	t.Run("Success", func(t *testing.T) {
		service, mockUserSvc, dbMock := setupVehicleServiceTest(t)
		defer func() {
			sqlDb, _ := service.(*VehicleService).db.DB()
			sqlDb.Close()
		}()

		mockUserSvc.On("Read", ownerUUID).Return(validOwner, nil).Once()

		dbMock.ExpectBegin()
		expectedSQL := `INSERT INTO "vehicles" ("created_at","updated_at","deleted_at","uuid","vehicle_type","vehicle_model","production_year","chassis_number","user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING "id"`
		dbMock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
			WithArgs(
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				nil,
				newVehicleInput.Uuid,
				newVehicleInput.VehicleType,
				newVehicleInput.VehicleModel,
				newVehicleInput.ProductionYear,
				newVehicleInput.ChassisNumber,
				validOwner.ID,
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		dbMock.ExpectCommit()

		createdVehicle, err := service.Create(newVehicleInput, ownerUUID)

		require.NoError(t, err)
		require.NotNil(t, createdVehicle)
		assert.Equal(t, vehicleUUID, createdVehicle.Uuid)
		assert.Equal(t, validOwner.ID, createdVehicle.UserId)
		assert.Equal(t, newVehicleInput.VehicleType, createdVehicle.VehicleType)
		assert.Equal(t, newVehicleInput.VehicleModel, createdVehicle.VehicleModel)
		assert.NotZero(t, createdVehicle.ID)
		mockUserSvc.AssertExpectations(t)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	// --- Error_UserNotFound Case (No changes needed) ---
	t.Run("Error_UserNotFound", func(t *testing.T) {
		service, mockUserSvc, dbMock := setupVehicleServiceTest(t)
		defer func() {
			sqlDb, _ := service.(*VehicleService).db.DB()
			sqlDb.Close()
		}()

		mockUserSvc.On("Read", ownerUUID).Return(nil, gorm.ErrRecordNotFound).Once()
		createdVehicle, err := service.Create(newVehicleInput, ownerUUID)

		require.Error(t, err)
		assert.Nil(t, createdVehicle)
		assert.ErrorIs(t, err, cerror.ErrUserIsNil) // Or gorm.ErrRecordNotFound if propagated directly
		mockUserSvc.AssertExpectations(t)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	// --- Error_UserBadRole Case (No changes needed) ---
	t.Run("Error_UserBadRole", func(t *testing.T) {
		service, mockUserSvc, dbMock := setupVehicleServiceTest(t)
		defer func() {
			sqlDb, _ := service.(*VehicleService).db.DB()
			sqlDb.Close()
		}()

		mockUserSvc.On("Read", ownerUUID).Return(invalidRoleOwner, nil).Once()
		createdVehicle, err := service.Create(newVehicleInput, ownerUUID)

		require.Error(t, err)
		assert.Nil(t, createdVehicle)
		assert.ErrorIs(t, err, cerror.ErrBadRole)
		mockUserSvc.AssertExpectations(t)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	// --- Error_DatabaseFailure Case ---
	t.Run("Error_DatabaseFailure", func(t *testing.T) {
		service, mockUserSvc, dbMock := setupVehicleServiceTest(t)
		defer func() {
			sqlDb, _ := service.(*VehicleService).db.DB()
			sqlDb.Close()
		}()

		dbError := sql.ErrConnDone // Simulate a generic DB error

		mockUserSvc.On("Read", ownerUUID).Return(validOwner, nil).Once()

		dbMock.ExpectBegin()
		expectedSQL := `INSERT INTO "vehicles" ("created_at","updated_at","deleted_at","uuid","vehicle_type","vehicle_model","production_year","chassis_number","user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING "id"`
		// **Correction**: Ensure the mock returns the specific error we want to test for
		dbMock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, newVehicleInput.Uuid, newVehicleInput.VehicleType, newVehicleInput.VehicleModel, newVehicleInput.ProductionYear, newVehicleInput.ChassisNumber, validOwner.ID).
			WillReturnError(dbError) // Mock returns the specific dbError
		dbMock.ExpectRollback()

		createdVehicle, err := service.Create(newVehicleInput, ownerUUID)

		require.Error(t, err)
		assert.Nil(t, createdVehicle)
		// **Correction**: Assert that the *specific* dbError (or a wrapped version) is returned
		assert.ErrorContains(t, err, dbError.Error())
		mockUserSvc.AssertExpectations(t)
		// **Correction**: Verify expectations *after* asserting the error
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestVehicleService_Delete(t *testing.T) {
	vehicleUUID := uuid.New()

	// --- Success Case (No changes needed) ---
	t.Run("Success", func(t *testing.T) {
		service, _, dbMock := setupVehicleServiceTest(t)
		defer func() {
			sqlDb, _ := service.(*VehicleService).db.DB()
			sqlDb.Close()
		}()

		dbMock.ExpectBegin()
		expectedSQL := `UPDATE "vehicles" SET "deleted_at"=$1 WHERE uuid = $2 AND "vehicles"."deleted_at" IS NULL`
		dbMock.ExpectExec(regexp.QuoteMeta(expectedSQL)).
			WithArgs(sqlmock.AnyArg(), vehicleUUID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		dbMock.ExpectCommit()

		err := service.Delete(vehicleUUID)

		require.NoError(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	// --- Error_NotFound Case (No changes needed) ---
	t.Run("Error_NotFound", func(t *testing.T) {
		service, _, dbMock := setupVehicleServiceTest(t)
		defer func() {
			sqlDb, _ := service.(*VehicleService).db.DB()
			sqlDb.Close()
		}()

		dbMock.ExpectBegin()
		expectedSQL := `UPDATE "vehicles" SET "deleted_at"=$1 WHERE uuid = $2 AND "vehicles"."deleted_at" IS NULL`
		dbMock.ExpectExec(regexp.QuoteMeta(expectedSQL)).
			WithArgs(sqlmock.AnyArg(), vehicleUUID).
			WillReturnResult(sqlmock.NewResult(0, 0)) // RowsAffected=0
		dbMock.ExpectCommit() // GORM might still commit

		err := service.Delete(vehicleUUID)

		require.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound) // Expect GORM's standard not found error
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	// --- Error_DatabaseFailure Case ---
	t.Run("Error_DatabaseFailure", func(t *testing.T) {
		service, _, dbMock := setupVehicleServiceTest(t)
		defer func() {
			sqlDb, _ := service.(*VehicleService).db.DB()
			sqlDb.Close()
		}()

		dbError := sql.ErrConnDone // Simulate DB error

		dbMock.ExpectBegin()
		expectedSQL := `UPDATE "vehicles" SET "deleted_at"=$1 WHERE uuid = $2 AND "vehicles"."deleted_at" IS NULL`
		// **Correction**: Ensure mock returns the specific error
		dbMock.ExpectExec(regexp.QuoteMeta(expectedSQL)).
			WithArgs(sqlmock.AnyArg(), vehicleUUID).
			WillReturnError(dbError) // Mock returns the specific dbError
		dbMock.ExpectRollback()

		err := service.Delete(vehicleUUID)

		require.Error(t, err)
		// **Correction**: Assert the specific error is returned/contained
		assert.ErrorContains(t, err, dbError.Error())
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestVehicleService_Read(t *testing.T) {
	vehicleUUID := uuid.New()
	ownerUserID := uint(101)
	now := time.Now()
	registrationUUID := uuid.New() // Example UUID for registration

	// **Correction**: Define columns matching the *actual* JOIN query output
	// Need columns from vehicles AND registration_infos (with aliases)
	// Example columns - **ADJUST THESE TO MATCH YOUR ACTUAL RegistrationInfo MODEL**
	readCols := []string{
		"id", "created_at", "updated_at", "deleted_at", "uuid", "vehicle_type", "vehicle_model", "production_year", "chassis_number", "user_id", // Vehicle cols
		"Registration__id", "Registration__created_at", "Registration__updated_at", "Registration__deleted_at", "Registration__uuid", "Registration__vehicle_id", "Registration__pass_technical", "Registration__traveled_distance", "Registration__technical_date", "Registration__registration", // Aliased RegistrationInfo cols
	}

	// **Correction**: Mock row data matching the combined columns
	// Example data - **ADJUST THESE TO MATCH YOUR ACTUAL RegistrationInfo MODEL**
	mockRow := sqlmock.NewRows(readCols).
		AddRow(1, now, now, nil, vehicleUUID, "Car", "Škoda Octavia", 2023, "SKDCHASSIS987654PQRS", ownerUserID, // Vehicle data
			10, now, now, nil, registrationUUID, 1, true, 15000.5, now.AddDate(0, -1, 0), "ZG1234AB") // Registration data

	// **Correction**: Define the expected SQL regex matching the actual JOIN query
	// Use .* for flexibility in selected columns if they might change slightly
	expectedSQL := `SELECT .* FROM "vehicles" INNER JOIN "registration_infos" "Registration" ON "vehicles"."id" = "Registration"."vehicle_id" AND "Registration"."deleted_at" IS NULL WHERE vehicles.uuid = \$1 AND "vehicles"."deleted_at" IS NULL ORDER BY "vehicles"."id" LIMIT \$2`
	// Use a more flexible regex if the exact column list is long or might change
	// expectedSQLRegex := `SELECT .* FROM "vehicles" INNER JOIN "registration_infos" "Registration" ON .* WHERE vehicles.uuid = \$1 .* LIMIT \$2`

	t.Run("Success", func(t *testing.T) {
		service, _, dbMock := setupVehicleServiceTest(t)
		defer func() {
			sqlDb, _ := service.(*VehicleService).db.DB()
			sqlDb.Close()
		}()

		// **Correction**: Use the corrected SQL regex and expect LIMIT $2 (GORM often adds limit arg)
		dbMock.ExpectQuery(expectedSQL). // Or use expectedSQLRegex
							WithArgs(vehicleUUID, 1). // Expect UUID and Limit argument
							WillReturnRows(mockRow)

		vehicle, err := service.Read(vehicleUUID)

		require.NoError(t, err)
		require.NotNil(t, vehicle)
		assert.Equal(t, vehicleUUID, vehicle.Uuid)
		assert.Equal(t, "Škoda Octavia", vehicle.VehicleModel)
		assert.Equal(t, ownerUserID, vehicle.UserId)
		require.NotNil(t, vehicle.Registration) // Assert Registration is loaded
		assert.Equal(t, registrationUUID, vehicle.Registration.Uuid)
		assert.Equal(t, "ZG1234AB", vehicle.Registration.Registration) // Check a field from registration

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("Error_NotFound", func(t *testing.T) {
		service, _, dbMock := setupVehicleServiceTest(t)
		defer func() {
			sqlDb, _ := service.(*VehicleService).db.DB()
			sqlDb.Close()
		}()

		// **Correction**: Use the corrected SQL regex
		dbMock.ExpectQuery(expectedSQL). // Or use expectedSQLRegex
							WithArgs(vehicleUUID, 1).
							WillReturnError(gorm.ErrRecordNotFound) // Mock returns not found

		vehicle, err := service.Read(vehicleUUID)

		require.Error(t, err)
		assert.Nil(t, vehicle)
		// **Correction**: Assert the specific error returned by the service
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("Error_DatabaseFailure", func(t *testing.T) {
		service, _, dbMock := setupVehicleServiceTest(t)
		defer func() {
			sqlDb, _ := service.(*VehicleService).db.DB()
			sqlDb.Close()
		}()

		dbError := sql.ErrConnDone // Simulate DB error

		// **Correction**: Use the corrected SQL regex
		dbMock.ExpectQuery(expectedSQL). // Or use expectedSQLRegex
							WithArgs(vehicleUUID, 1).
							WillReturnError(dbError) // Mock returns the specific error

		vehicle, err := service.Read(vehicleUUID)

		require.Error(t, err)
		assert.Nil(t, vehicle)
		// **Correction**: Assert the specific error is contained
		assert.ErrorContains(t, err, dbError.Error())
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestVehicleService_ReadAll(t *testing.T) {
	userUUID := uuid.New()
	userId := uint(101)
	vehicleUUID1 := uuid.New()
	vehicleUUID2 := uuid.New()
	regUUID1 := uuid.New()
	regUUID2 := uuid.New()
	now := time.Now()

	// **Correction**: Define columns matching the *actual* double JOIN query output
	// Example columns - **ADJUST THESE TO MATCH YOUR ACTUAL RegistrationInfo MODEL**
	readAllCols := []string{
		"id", "created_at", "updated_at", "deleted_at", "uuid", "vehicle_type", "vehicle_model", "production_year", "chassis_number", "user_id", // Vehicle cols
		"Registration__id", "Registration__created_at", "Registration__updated_at", "Registration__deleted_at", "Registration__uuid", "Registration__vehicle_id", "Registration__pass_technical", "Registration__traveled_distance", "Registration__technical_date", "Registration__registration", // Aliased RegistrationInfo cols
	}

	// **Correction**: Mock rows matching the combined columns
	// Example data - **ADJUST THESE TO MATCH YOUR ACTUAL RegistrationInfo MODEL**
	mockRows := sqlmock.NewRows(readAllCols).
		AddRow(1, now, now, nil, vehicleUUID1, "Car", "BMW X5", 2021, "BMWX5...", userId, // Vehicle 1 data
													10, now, now, nil, regUUID1, 1, true, 25000, now.AddDate(0, -2, 0), "ZG555BMW"). // Registration 1 data
		AddRow(2, now, now, nil, vehicleUUID2, "Car", "Audi A4", 2020, "AUDIA4...", userId, // Vehicle 2 data
			11, now, now, nil, regUUID2, 2, false, 35000, now.AddDate(0, -3, 0), "ZG777AUD") // Registration 2 data

	// **Correction**: Define the expected SQL regex matching the actual double JOIN query
	// Use .* for flexibility
	expectedSQL := `SELECT .* FROM "vehicles" INNER JOIN "registration_infos" "Registration" ON "vehicles"."id" = "Registration"."vehicle_id" AND "Registration"."deleted_at" IS NULL inner join users on vehicles.user_id = users.id WHERE users.uuid = \$1 AND "vehicles"."deleted_at" IS NULL`
	// More specific regex if needed:
	// expectedSQLSpecific := `SELECT "vehicles"."id",.*,"Registration"."id" AS "Registration__id",.* FROM "vehicles" INNER JOIN "registration_infos" "Registration" ON "vehicles"."id" = "Registration"."vehicle_id" AND "Registration"."deleted_at" IS NULL inner join users on vehicles.user_id = users.id WHERE users.uuid = \$1 AND "vehicles"."deleted_at" IS NULL`

	t.Run("Success_FoundMultiple", func(t *testing.T) {
		service, _, dbMock := setupVehicleServiceTest(t)
		defer func() {
			sqlDb, _ := service.(*VehicleService).db.DB()
			sqlDb.Close()
		}()

		// **Correction**: Use the corrected SQL regex
		dbMock.ExpectQuery(expectedSQL). // Use the flexible regex
							WithArgs(userUUID).
							WillReturnRows(mockRows)

		vehicles, err := service.ReadAll(userUUID)

		require.NoError(t, err)
		require.NotNil(t, vehicles)
		assert.Len(t, vehicles, 2)
		// Check Vehicle 1
		assert.Equal(t, vehicleUUID1, vehicles[0].Uuid)
		assert.Equal(t, "BMW X5", vehicles[0].VehicleModel)
		require.NotNil(t, vehicles[0].Registration)
		assert.Equal(t, regUUID1, vehicles[0].Registration.Uuid)
		assert.Equal(t, "ZG555BMW", vehicles[0].Registration.Registration)
		// Check Vehicle 2
		assert.Equal(t, vehicleUUID2, vehicles[1].Uuid)
		assert.Equal(t, "Audi A4", vehicles[1].VehicleModel)
		require.NotNil(t, vehicles[1].Registration)
		assert.Equal(t, regUUID2, vehicles[1].Registration.Uuid)
		assert.Equal(t, "ZG777AUD", vehicles[1].Registration.Registration)

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("Success_FoundNone", func(t *testing.T) {
		service, _, dbMock := setupVehicleServiceTest(t)
		defer func() {
			sqlDb, _ := service.(*VehicleService).db.DB()
			sqlDb.Close()
		}()

		// **Correction**: Use the corrected SQL regex and return empty rows matching the *correct* columns
		dbMock.ExpectQuery(expectedSQL). // Use the flexible regex
							WithArgs(userUUID).
							WillReturnRows(sqlmock.NewRows(readAllCols)) // Empty rows with correct columns

		vehicles, err := service.ReadAll(userUUID)

		require.NoError(t, err)
		require.NotNil(t, vehicles) // Expect empty slice
		assert.Len(t, vehicles, 0)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("Error_DatabaseFailure", func(t *testing.T) {
		service, _, dbMock := setupVehicleServiceTest(t)
		defer func() {
			sqlDb, _ := service.(*VehicleService).db.DB()
			sqlDb.Close()
		}()

		dbError := sql.ErrConnDone // Simulate DB error

		// **Correction**: Use the corrected SQL regex
		dbMock.ExpectQuery(expectedSQL). // Use the flexible regex
							WithArgs(userUUID).
							WillReturnError(dbError) // Mock returns the specific error

		vehicles, err := service.ReadAll(userUUID)

		require.Error(t, err)
		assert.Nil(t, vehicles) // Expect nil slice on DB error
		// **Correction**: Assert the specific error is contained
		assert.ErrorContains(t, err, dbError.Error())
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}
