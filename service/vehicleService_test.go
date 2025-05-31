package service_test

import (
	"ePrometna_Server/app"
	"ePrometna_Server/config"
	"ePrometna_Server/model"
	"ePrometna_Server/service"
	"ePrometna_Server/util/cerror"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// --- Mock UserCrudService (remains the same) ---
type MockUserCrudService struct {
	mock.Mock
}

// DeleteUserDevice implements service.IUserCrudService.
func (m *MockUserCrudService) DeleteUserDevice(userUUID uuid.UUID) error {
	panic("unimplemented")
}

// GetUserDevice implements service.IUserCrudService.
func (m *MockUserCrudService) GetUserDevice(userId uint) (*model.Mobile, error) {
	panic("unimplemented")
}

// GetUserByOIB implements service.IUserCrudService.
func (m *MockUserCrudService) GetUserByOIB(oib string) (*model.User, error) {
	args := m.Called(oib)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserCrudService) Create(user *model.User, password string) (*model.User, error) {
	args := m.Called(user, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserCrudService) Read(id uuid.UUID) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserCrudService) ReadAll() ([]model.User, error) {
	args := m.Called()
	if len(args) < 2 || args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserCrudService) Update(id uuid.UUID, user *model.User) (*model.User, error) {
	args := m.Called(id, user)
	if len(args) < 2 || args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserCrudService) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserCrudService) GetAllUsers() ([]model.User, error) {
	args := m.Called()
	if len(args) < 2 || args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserCrudService) GetAllPoliceOfficers() ([]model.User, error) {
	args := m.Called()
	if len(args) < 2 || args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserCrudService) SearchUsersByName(query string) ([]model.User, error) {
	args := m.Called(query)
	if len(args) < 2 || args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.User), args.Error(1)
}

// --- Test Suite ---
type VehicleServiceTestSuite struct {
	suite.Suite
	db             *gorm.DB
	vehicleService service.IVehicleService
	mockUserSvc    *MockUserCrudService
	sugar          *zap.SugaredLogger
}

// SetupSuite runs once before all tests in the suite
func (suite *VehicleServiceTestSuite) SetupSuite() {
	loggerCfg := zap.NewDevelopmentConfig()
	loggerCfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	loggerCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zapLogger, _ := loggerCfg.Build()
	suite.sugar = zapLogger.Sugar()

	config.AppConfig = &config.AppConfiguration{
		Env:          config.Dev,
		AccessKey:    "test-access-key",
		RefreshKey:   "test-refresh-key",
		DbConnection: "",
	}

	app.Test()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		suite.T().Fatalf("Failed to connect to SQLite: %v", err)
	}
	suite.db = db

	err = suite.db.AutoMigrate(model.GetAllModels()...)
	if err != nil {
		suite.T().Fatalf("Failed to migrate SQLite: %v", err)
	}

	suite.mockUserSvc = new(MockUserCrudService)

	app.Provide(func() *gorm.DB { return suite.db })
	app.Provide(func() *zap.SugaredLogger { return suite.sugar })
	app.Provide(func() service.IUserCrudService { return suite.mockUserSvc })
	suite.vehicleService = service.NewVehicleService()
}

func (suite *VehicleServiceTestSuite) TearDownSuite() {
	if suite.db != nil {
		sqlDB, _ := suite.db.DB()
		sqlDB.Close()
	}
	if suite.sugar != nil {
		suite.sugar.Sync()
	}
}

func (suite *VehicleServiceTestSuite) SetupTest() {
	suite.db.Exec("PRAGMA foreign_keys = OFF")
	defer suite.db.Exec("PRAGMA foreign_keys = ON")
	tables := []string{
		"owner_histories", "registration_infos", "vehicle_drivers", "temp_data",
		"vehicles", "driver_licenses", "mobiles", "users",
	}
	for _, table := range tables {
		err := suite.db.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error
		if err != nil && !strings.Contains(err.Error(), "no such table") {
			suite.sugar.Warnf("Could not clean table %s: %v", table, err)
		}
	}
	suite.mockUserSvc.ExpectedCalls = nil
	suite.mockUserSvc.Calls = nil
}

// Helper to create a user directly in the DB for testing FK constraints
func createTestUserInDB(db *gorm.DB, s *suite.Suite, role model.UserRole, userUUID uuid.UUID) *model.User {
	oibSuffix := time.Now().UnixNano() % 1000000
	emailSuffix := userUUID.String()[:8]

	user := &model.User{
		Uuid:         userUUID,
		FirstName:    "Test",
		LastName:     string(role),
		OIB:          fmt.Sprintf("FN%07d", oibSuffix),
		Email:        fmt.Sprintf("db.test.%s.%s@example.com", string(role), emailSuffix),
		Role:         role,
		BirthDate:    time.Now().AddDate(-25, 0, 0),
		Residence:    "DB Test Residence",
		PasswordHash: "db-dummy-hash-bcrypt-valid",
	}
	err := db.Create(user).Error
	s.Require().NoError(err, "Failed to create user in DB for test. User: %+v", user)
	s.Require().NotZero(user.ID, "User ID should not be zero after DB creation")
	return user
}

// Helper to create a vehicle with an initial registration, properly linked in the DB
func createTestVehicleWithInitialReg(db *gorm.DB, s *suite.Suite, ownerID uint, vehicleUUID uuid.UUID, initialRegPlate string) *model.Vehicle {
	vehicle := &model.Vehicle{
		Uuid:          vehicleUUID,
		UserId:        &ownerID,
		VehicleModel:  "Test Model S",
		VehicleType:   "Car",
		ChassisNumber: "CHASSIS" + vehicleUUID.String()[:8],
	}
	err := db.Create(vehicle).Error
	s.Require().NoError(err, "Failed to create vehicle in DB. Vehicle: %+v", vehicle)
	s.Require().NotZero(vehicle.ID)

	initialReg := &model.RegistrationInfo{
		Uuid:             uuid.New(),
		VehicleId:        vehicle.ID,
		PassTechnical:    true,
		TraveledDistance: 10000,
		TechnicalDate:    time.Now().AddDate(0, -6, 0),
		Registration:     initialRegPlate,
	}
	err = db.Create(initialReg).Error
	s.Require().NoError(err, "Failed to create initial registration info in DB. RegInfo: %+v", initialReg)
	s.Require().NotZero(initialReg.ID)

	vehicle.RegistrationID = &initialReg.ID
	err = db.Model(&model.Vehicle{}).Where("id = ?", vehicle.ID).Update("registration_id", initialReg.ID).Error
	s.Require().NoError(err, "Failed to link initial registration to vehicle")

	var reloadedVehicle model.Vehicle
	err = db.Preload("Registration").Preload("Owner").First(&reloadedVehicle, vehicle.ID).Error
	s.Require().NoError(err)
	return &reloadedVehicle
}

// --- Test Cases (Copied from the artifact, ensure they align with the new DI) ---

func (suite *VehicleServiceTestSuite) TestCreateVehicle_Success() {
	ownerUUID := uuid.New()
	dbOwner := createTestUserInDB(suite.db, &suite.Suite, model.RoleOsoba, ownerUUID)
	suite.mockUserSvc.On("Read", ownerUUID).Return(dbOwner, nil)

	newVehicleData := &model.Vehicle{
		Uuid:          uuid.New(),
		VehicleModel:  "Test Create",
		VehicleType:   "Car",
		ChassisNumber: "CHASSIS_CREATE" + uuid.NewString()[:4],
		Registration: &model.RegistrationInfo{
			Uuid:             uuid.New(),
			PassTechnical:    true,
			TraveledDistance: 50,
			TechnicalDate:    time.Now(),
			Registration:     "ZG-SVC-CREATE",
		},
	}

	createdVehicle, err := suite.vehicleService.Create(newVehicleData, ownerUUID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), createdVehicle)
	assert.NotZero(suite.T(), createdVehicle.ID)
	assert.Equal(suite.T(), dbOwner.ID, *createdVehicle.UserId)
	assert.NotEqual(suite.T(), uuid.Nil, createdVehicle.Uuid)

	assert.NotNil(suite.T(), createdVehicle.Registration)
	assert.NotZero(suite.T(), createdVehicle.Registration.ID)
	assert.NotEqual(suite.T(), uuid.Nil, createdVehicle.Registration.Uuid)
	assert.Equal(suite.T(), "ZG-SVC-CREATE", createdVehicle.Registration.Registration)
	assert.WithinDuration(suite.T(), time.Now(), createdVehicle.Registration.TechnicalDate, 5*time.Second)

	var dbVehicle model.Vehicle
	err = suite.db.Preload("Registration").First(&dbVehicle, "uuid = ?", createdVehicle.Uuid).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), dbOwner.ID, *dbVehicle.UserId)
	assert.NotNil(suite.T(), dbVehicle.Registration)
	assert.Equal(suite.T(), "ZG-SVC-CREATE", dbVehicle.Registration.Registration)

	suite.mockUserSvc.AssertExpectations(suite.T())
}

func (suite *VehicleServiceTestSuite) TestCreateVehicle_OwnerNotFound() {
	ownerUUID := uuid.New()
	suite.mockUserSvc.On("Read", ownerUUID).Return(nil, gorm.ErrRecordNotFound)

	newVehicle := &model.Vehicle{Uuid: uuid.New(), VehicleModel: "FailCar"}
	_, err := suite.vehicleService.Create(newVehicle, ownerUUID)

	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, gorm.ErrRecordNotFound))
	suite.mockUserSvc.AssertExpectations(suite.T())
}

func (suite *VehicleServiceTestSuite) TestCreateVehicle_OwnerBadRole() {
	ownerUUID := uuid.New()
	badRoleOwner := &model.User{Uuid: ownerUUID, Role: model.RolePolicija}
	badRoleOwner.ID = 99
	suite.mockUserSvc.On("Read", ownerUUID).Return(badRoleOwner, nil)

	newVehicle := &model.Vehicle{Uuid: uuid.New(), VehicleModel: "FailCarRole"}
	_, err := suite.vehicleService.Create(newVehicle, ownerUUID)

	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, cerror.ErrBadRole))
	suite.mockUserSvc.AssertExpectations(suite.T())
}

func (suite *VehicleServiceTestSuite) TestReadVehicle_Success() {
	owner := createTestUserInDB(suite.db, &suite.Suite, model.RoleOsoba, uuid.New())
	vehicle := createTestVehicleWithInitialReg(suite.db, &suite.Suite, owner.ID, uuid.New(), "ZG-READ-01")

	readVehicle, err := suite.vehicleService.Read(vehicle.Uuid)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), readVehicle)
	assert.Equal(suite.T(), vehicle.Uuid, readVehicle.Uuid)
	assert.NotNil(suite.T(), readVehicle.Registration)
	assert.Equal(suite.T(), "ZG-READ-01", readVehicle.Registration.Registration)
	assert.NotNil(suite.T(), readVehicle.Owner)
	assert.Equal(suite.T(), owner.ID, readVehicle.Owner.ID)
}

func (suite *VehicleServiceTestSuite) TestChangeOwner_Success() {
	oldOwner := createTestUserInDB(suite.db, &suite.Suite, model.RoleOsoba, uuid.New())
	newOwner := createTestUserInDB(suite.db, &suite.Suite, model.RoleFirma, uuid.New())
	vehicle := createTestVehicleWithInitialReg(suite.db, &suite.Suite, oldOwner.ID, uuid.New(), "ZG-CHOWN-01")

	err := suite.vehicleService.ChangeOwner(vehicle.Uuid, newOwner.Uuid)
	assert.NoError(suite.T(), err)

	var dbVehicle model.Vehicle
	// Preload Owner to verify the owner details are correctly updated/fetched.
	err = suite.db.Preload("Owner").First(&dbVehicle, "uuid = ?", vehicle.Uuid).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newOwner.ID, *dbVehicle.UserId)
	assert.NotNil(suite.T(), dbVehicle.Owner, "Owner should be preloaded and not nil")
	if dbVehicle.Owner != nil {
		assert.Equal(suite.T(), newOwner.ID, dbVehicle.Owner.ID)
		assert.Equal(suite.T(), newOwner.Uuid, dbVehicle.Owner.Uuid)
	}

	var ownerHistory []model.OwnerHistory
	err = suite.db.Where("vehicle_id = ? AND user_id = ?", dbVehicle.ID, oldOwner.ID).Find(&ownerHistory).Error
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), ownerHistory, 1, "Should have one history record for the old owner")
	if len(ownerHistory) > 0 {
		assert.Equal(suite.T(), oldOwner.ID, ownerHistory[0].UserId)
	}
}

func (suite *VehicleServiceTestSuite) TestChangeOwner_NewOwnerNotFound() {
	oldOwner := createTestUserInDB(suite.db, &suite.Suite, model.RoleOsoba, uuid.New())
	vehicle := createTestVehicleWithInitialReg(suite.db, &suite.Suite, oldOwner.ID, uuid.New(), "ZG-CHOWN-02")
	nonExistentOwnerUUID := uuid.New()

	err := suite.vehicleService.ChangeOwner(vehicle.Uuid, nonExistentOwnerUUID)
	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, gorm.ErrRecordNotFound), "Expected gorm.ErrRecordNotFound for new owner")
}

func (suite *VehicleServiceTestSuite) TestChangeOwner_NewOwnerBadRole() {
	oldOwner := createTestUserInDB(suite.db, &suite.Suite, model.RoleOsoba, uuid.New())
	newOwnerBadRole := createTestUserInDB(suite.db, &suite.Suite, model.RolePolicija, uuid.New()) // Policija cannot own
	vehicle := createTestVehicleWithInitialReg(suite.db, &suite.Suite, oldOwner.ID, uuid.New(), "ZG-CHOWN-03")

	err := suite.vehicleService.ChangeOwner(vehicle.Uuid, newOwnerBadRole.Uuid)
	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, cerror.ErrBadRole))
}

func (suite *VehicleServiceTestSuite) TestRegistration_SupersedeExisting() {
	owner := createTestUserInDB(suite.db, &suite.Suite, model.RoleOsoba, uuid.New())
	vehicle := createTestVehicleWithInitialReg(suite.db, &suite.Suite, owner.ID, uuid.New(), "ZG-OLD-REG")
	initialRegID := vehicle.Registration.ID // ID of the first registration

	newRegInfo := model.RegistrationInfo{
		PassTechnical:    true,
		TraveledDistance: 25000,
		Registration:     "ZG-FRESH-REG",
	}

	err := suite.vehicleService.Registration(vehicle.Uuid, newRegInfo)
	assert.NoError(suite.T(), err)

	var dbVehicle model.Vehicle
	err = suite.db.Preload("Registration").First(&dbVehicle, "uuid = ?", vehicle.Uuid).Error
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), dbVehicle.Registration)
	assert.Equal(suite.T(), "ZG-FRESH-REG", dbVehicle.Registration.Registration)
	assert.Equal(suite.T(), 25000, dbVehicle.Registration.TraveledDistance)
	assert.NotEqual(suite.T(), initialRegID, dbVehicle.Registration.ID, "New registration should have a different ID from the initial one")

	var allRegInfos []model.RegistrationInfo
	err = suite.db.Where("vehicle_id = ?", dbVehicle.ID).Order("technical_date asc").Find(&allRegInfos).Error
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), allRegInfos, 2)

	if len(allRegInfos) == 2 {
		assert.Equal(suite.T(), "ZG-OLD-REG", allRegInfos[0].Registration)
		assert.Equal(suite.T(), 10000, allRegInfos[0].TraveledDistance)
		assert.Equal(suite.T(), "ZG-FRESH-REG", allRegInfos[1].Registration)
		assert.Equal(suite.T(), 25000, allRegInfos[1].TraveledDistance)
		assert.Equal(suite.T(), allRegInfos[1].ID, *dbVehicle.RegistrationID)
	}
}

// TestDeleteVehicle_Success tests the successful soft deletion of a vehicle.
func (suite *VehicleServiceTestSuite) TestDeleteVehicle_Success() {
	owner := createTestUserInDB(suite.db, &suite.Suite, model.RoleOsoba, uuid.New())
	vehicleToTest := createTestVehicleWithInitialReg(suite.db, &suite.Suite, owner.ID, uuid.New(), "ZG-DEL-01")
	suite.T().Logf("Vehicle created for deletion test: ID=%d, UUID=%s, OwnerID=%d", vehicleToTest.ID, vehicleToTest.Uuid, *vehicleToTest.UserId)

	err := suite.vehicleService.Delete(vehicleToTest.Uuid)
	assert.NoError(suite.T(), err)

	var dbVehicle model.Vehicle
	// Use Unscoped to retrieve soft-deleted records
	err = suite.db.Unscoped().Preload("Registration").First(&dbVehicle, "uuid = ?", vehicleToTest.Uuid).Error
	assert.NoError(suite.T(), err, "Should be able to find the vehicle (even if soft-deleted) by UUID")

	assert.Nil(suite.T(), dbVehicle.UserId, "Vehicle's UserId should be nil after deletion")
	assert.True(suite.T(), dbVehicle.DeletedAt.Valid, "Vehicle's DeletedAt should be valid (not NULL)")
	assert.NotNil(suite.T(), dbVehicle.DeletedAt.Time, "Vehicle's DeletedAt time should be set")
	assert.WithinDuration(suite.T(), time.Now(), dbVehicle.DeletedAt.Time, 5*time.Second, "DeletedAt should be recent")

	if vehicleToTest.Registration != nil {
		var regInfo model.RegistrationInfo
		errReg := suite.db.First(&regInfo, "id = ?", vehicleToTest.Registration.ID).Error
		assert.NoError(suite.T(), errReg, "Registration info should still exist if not cascade deleted")
	}
}

// TestDeleteVehicle_NotFound tests deleting a non-existent vehicle.
func (suite *VehicleServiceTestSuite) TestDeleteVehicle_NotFound() {
	nonExistentUUID := uuid.New()
	err := suite.vehicleService.Delete(nonExistentUUID)
	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, gorm.ErrRecordNotFound), "Expected gorm.ErrRecordNotFound for non-existent vehicle")
}

// TestReadAllVehicles_OwnerHasVehicles tests retrieving vehicles for an owner who has them.
func (suite *VehicleServiceTestSuite) TestReadAllVehicles_OwnerHasVehicles() {
	ownerUser := createTestUserInDB(suite.db, &suite.Suite, model.RoleOsoba, uuid.New())

	// Create a couple of vehicles for this owner
	vehicle1 := createTestVehicleWithInitialReg(suite.db, &suite.Suite, ownerUser.ID, uuid.New(), "ZG-ALL-01")
	vehicle2 := createTestVehicleWithInitialReg(suite.db, &suite.Suite, ownerUser.ID, uuid.New(), "ZG-ALL-02")

	// Create another owner and their vehicle, to ensure we only get vehicles for the specified owner
	otherOwner := createTestUserInDB(suite.db, &suite.Suite, model.RoleFirma, uuid.New())
	_ = createTestVehicleWithInitialReg(suite.db, &suite.Suite, otherOwner.ID, uuid.New(), "KA-OTHER-01")

	retrievedVehicles, err := suite.vehicleService.ReadAll(ownerUser.Uuid)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedVehicles)
	assert.Len(suite.T(), retrievedVehicles, 2, "Should retrieve two vehicles for the owner")

	// Check if the correct vehicles are returned (can check UUIDs or other unique properties)
	var foundV1, foundV2 bool
	for _, v := range retrievedVehicles {
		assert.NotNil(suite.T(), v.Registration, "Vehicle in list should have registration preloaded")
		if v.Uuid == vehicle1.Uuid {
			foundV1 = true
			assert.Equal(suite.T(), "ZG-ALL-01", v.Registration.Registration)
		}
		if v.Uuid == vehicle2.Uuid {
			foundV2 = true
			assert.Equal(suite.T(), "ZG-ALL-02", v.Registration.Registration)
		}
	}
	assert.True(suite.T(), foundV1, "Vehicle 1 not found in retrieved list")
	assert.True(suite.T(), foundV2, "Vehicle 2 not found in retrieved list")
}

// TestReadAllVehicles_OwnerHasNoVehicles tests retrieving vehicles for an owner who has none.
func (suite *VehicleServiceTestSuite) TestReadAllVehicles_OwnerHasNoVehicles() {
	ownerWithNoVehicles := createTestUserInDB(suite.db, &suite.Suite, model.RoleOsoba, uuid.New())

	// Create a vehicle for a different owner to ensure the DB isn't empty
	otherOwner := createTestUserInDB(suite.db, &suite.Suite, model.RoleFirma, uuid.New())
	_ = createTestVehicleWithInitialReg(suite.db, &suite.Suite, otherOwner.ID, uuid.New(), "KA-OTHER-02")

	retrievedVehicles, err := suite.vehicleService.ReadAll(ownerWithNoVehicles.Uuid)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedVehicles)
	assert.Len(suite.T(), retrievedVehicles, 0, "Should retrieve an empty list for an owner with no vehicles")
}

func (suite *VehicleServiceTestSuite) TestReadByVin_Success() {
	owner := createTestUserInDB(suite.db, &suite.Suite, model.RoleOsoba, uuid.New())
	vinToFind := "VINSUCCESS123"
	vehicle := createTestVehicleWithInitialReg(suite.db, &suite.Suite, owner.ID, uuid.New(), "ZG-VIN-S01")
	vehicle.ChassisNumber = vinToFind // Set the VIN for this test
	errUpdate := suite.db.Save(vehicle).Error
	suite.Require().NoError(errUpdate)

	readVehicle, err := suite.vehicleService.ReadByVin(vinToFind)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), readVehicle)
	assert.Equal(suite.T(), vehicle.Uuid, readVehicle.Uuid)
	assert.Equal(suite.T(), vinToFind, readVehicle.ChassisNumber)
	assert.NotNil(suite.T(), readVehicle.Registration)
	assert.Equal(suite.T(), "ZG-VIN-S01", readVehicle.Registration.Registration)
	assert.NotNil(suite.T(), readVehicle.Owner)
	assert.Equal(suite.T(), owner.ID, readVehicle.Owner.ID)
}

func (suite *VehicleServiceTestSuite) TestReadByVin_NotFound() {
	nonExistentVIN := "VINNOTFOUND000"
	_, err := suite.vehicleService.ReadByVin(nonExistentVIN)

	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, gorm.ErrRecordNotFound), "Expected gorm.ErrRecordNotFound for non-existent VIN")
}

func (suite *VehicleServiceTestSuite) TestDeregister_Success() {
	owner := createTestUserInDB(suite.db, &suite.Suite, model.RoleOsoba, uuid.New())
	vehicle := createTestVehicleWithInitialReg(suite.db, &suite.Suite, owner.ID, uuid.New(), "ZG-DEREG-01")
	initialRegID := vehicle.Registration.ID

	err := suite.vehicleService.Deregister(vehicle.Uuid)
	assert.NoError(suite.T(), err)

	var dbVehicle model.Vehicle
	err = suite.db.Preload("PastRegistration").First(&dbVehicle, "uuid = ?", vehicle.Uuid).Error
	assert.NoError(suite.T(), err)

	assert.Nil(suite.T(), dbVehicle.RegistrationID, "RegistrationID should be nil after deregistration")

	// Check if the old registration is now in PastRegistration
	foundInPast := false
	for _, pastReg := range dbVehicle.PastRegistration {
		if pastReg.ID == initialRegID {
			foundInPast = true
			assert.Equal(suite.T(), "ZG-DEREG-01", pastReg.Registration)
			break
		}
	}
	assert.True(suite.T(), foundInPast, "Initial registration should be moved to past registrations")
}

func (suite *VehicleServiceTestSuite) TestDeregister_VehicleNotFound() {
	nonExistentUUID := uuid.New()
	err := suite.vehicleService.Deregister(nonExistentUUID)
	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, gorm.ErrRecordNotFound), "Expected gorm.ErrRecordNotFound for non-existent vehicle")
}

func (suite *VehicleServiceTestSuite) TestDeregister_VehicleAlreadyDeregistered() {
	owner := createTestUserInDB(suite.db, &suite.Suite, model.RoleOsoba, uuid.New())
	vehicle := createTestVehicleWithInitialReg(suite.db, &suite.Suite, owner.ID, uuid.New(), "ZG-DEREG-ALR")
	initialRegID := vehicle.Registration.ID

	// First deregistration
	err := suite.vehicleService.Deregister(vehicle.Uuid)
	assert.NoError(suite.T(), err)

	// Attempt to deregister again
	err = suite.vehicleService.Deregister(vehicle.Uuid)
	assert.NoError(suite.T(), err, "Deregistering an already deregistered vehicle should not error (idempotent)")

	var dbVehicle model.Vehicle
	err = suite.db.Preload("PastRegistration").First(&dbVehicle, "uuid = ?", vehicle.Uuid).Error
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), dbVehicle.RegistrationID)

	// Ensure the original registration is still in past registrations and not duplicated
	countPastRegs := 0
	for _, pastReg := range dbVehicle.PastRegistration {
		if pastReg.ID == initialRegID {
			countPastRegs++
		}
	}
	assert.Equal(suite.T(), 1, countPastRegs, "Initial registration should appear only once in past registrations")
}

func (suite *VehicleServiceTestSuite) TestUpdateVehicle_Service_Success() {
	owner := createTestUserInDB(suite.db, &suite.Suite, model.RoleOsoba, uuid.New())
	vehicleToUpdate := createTestVehicleWithInitialReg(suite.db, &suite.Suite, owner.ID, uuid.New(), "ZG-UPDATE-01")

	updateData := model.Vehicle{
		// Note: VehicleService.Update only updates fields present in model.Vehicle.Update method
		HomologationType: "UPDATED_HOMO_TYPE",
		BodyShape:        "Updated Coupe",
		ColourOfVehicle:  "Deep Blue",
		EnginePower:      "250kW",
		// Other fields that are updatable by vehicle.Update()
	}

	updatedVehicle, err := suite.vehicleService.Update(vehicleToUpdate.Uuid, updateData)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updatedVehicle)

	assert.Equal(suite.T(), vehicleToUpdate.Uuid, updatedVehicle.Uuid) // UUID should not change
	assert.Equal(suite.T(), updateData.HomologationType, updatedVehicle.HomologationType)
	assert.Equal(suite.T(), updateData.BodyShape, updatedVehicle.BodyShape)
	assert.Equal(suite.T(), updateData.ColourOfVehicle, updatedVehicle.ColourOfVehicle)
	assert.Equal(suite.T(), updateData.EnginePower, updatedVehicle.EnginePower)

	// Verify in DB
	var dbVehicle model.Vehicle
	errDb := suite.db.First(&dbVehicle, "uuid = ?", vehicleToUpdate.Uuid).Error
	assert.NoError(suite.T(), errDb)
	assert.Equal(suite.T(), updateData.HomologationType, dbVehicle.HomologationType)
	assert.Equal(suite.T(), updateData.BodyShape, dbVehicle.BodyShape)

	// Fields not in vehicle.Update() should remain unchanged from original
	assert.Equal(suite.T(), vehicleToUpdate.VehicleModel, dbVehicle.VehicleModel)
	assert.Equal(suite.T(), vehicleToUpdate.ChassisNumber, dbVehicle.ChassisNumber)
}

func (suite *VehicleServiceTestSuite) TestUpdateVehicle_Service_NotFound() {
	nonExistentUUID := uuid.New()
	updateData := model.Vehicle{VehicleModel: "NonExistentUpdate"}

	_, err := suite.vehicleService.Update(nonExistentUUID, updateData)
	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, gorm.ErrRecordNotFound), "Expected gorm.ErrRecordNotFound for non-existent vehicle")
}

// --- Run Test Suite ---
func TestVehicleServiceSuite(t *testing.T) {
	suite.Run(t, new(VehicleServiceTestSuite))
}
