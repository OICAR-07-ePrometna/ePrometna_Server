package service_test

import (
	"ePrometna_Server/app"
	"ePrometna_Server/config"
	"ePrometna_Server/model"
	"ePrometna_Server/service"
	"ePrometna_Server/util/auth"
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
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// --- Mock UserCrudService (Dependency for DriverLicenseCrudService) ---
type MockUserCrudServiceForLicense struct {
	mock.Mock
}

// DeleteUserDevice implements service.IUserCrudService.
func (m *MockUserCrudServiceForLicense) DeleteUserDevice(userUUID uuid.UUID) error {
	panic("unimplemented")
}

// GetUserDevice implements service.IUserCrudService.
func (m *MockUserCrudServiceForLicense) GetUserDevice(userId uint) (*model.Mobile, error) {
	panic("unimplemented")
}

func (m *MockUserCrudServiceForLicense) Create(user *model.User, password string) (*model.User, error) {
	args := m.Called(user, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserCrudServiceForLicense) Read(id uuid.UUID) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserCrudServiceForLicense) ReadAll() ([]model.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserCrudServiceForLicense) Update(id uuid.UUID, user *model.User) (*model.User, error) {
	args := m.Called(id, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserCrudServiceForLicense) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserCrudServiceForLicense) GetAllUsers() ([]model.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserCrudServiceForLicense) GetAllPoliceOfficers() ([]model.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserCrudServiceForLicense) SearchUsersByName(query string) ([]model.User, error) {
	args := m.Called(query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserCrudServiceForLicense) GetUserByOIB(oib string) (*model.User, error) {
	args := m.Called(oib)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// --- DriverLicenseCrudService Test Suite ---
type DriverLicenseCrudServiceTestSuite struct {
	suite.Suite
	db              *gorm.DB
	licenseService  service.IDriverLicenseCrudService
	mockUserSvc     *MockUserCrudServiceForLicense
	logger          *zap.SugaredLogger
	logObserver     *observer.ObservedLogs
	seededOwnerUser *model.User
}

// SetupSuite runs once before all tests
func (suite *DriverLicenseCrudServiceTestSuite) SetupSuite() {
	core, obs := observer.New(zap.InfoLevel)
	suite.logger = zap.New(core).Sugar()
	suite.logObserver = obs
	zap.ReplaceGlobals(zap.New(core))

	db, err := gorm.Open(sqlite.Open("file:licenseservice_test.db?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	suite.Require().NoError(err, "Failed to connect to SQLite for LicenseService tests")
	suite.db = db

	err = suite.db.AutoMigrate(model.GetAllModels()...)
	suite.Require().NoError(err, "Failed to migrate database schema for LicenseService tests")

	config.AppConfig = &config.AppConfiguration{
		Env:       config.Dev,
		AccessKey: "license-service-test-access-key",
	}

	suite.mockUserSvc = new(MockUserCrudServiceForLicense)

	app.Test()
	app.Provide(func() *gorm.DB { return suite.db })
	app.Provide(func() *zap.SugaredLogger { return suite.logger })
	app.Provide(func() service.IUserCrudService { return suite.mockUserSvc })
	suite.licenseService = service.NewDriverLicenseService(suite.db)
	suite.Require().NotNil(suite.licenseService, "LicenseService should be initialized")
}

// TearDownSuite runs once after all tests
func (suite *DriverLicenseCrudServiceTestSuite) TearDownSuite() {
	if suite.logger != nil {
		_ = suite.logger.Sync()
	}
	if suite.db != nil {
		sqlDB, _ := suite.db.DB()
		err := sqlDB.Close()
		suite.Require().NoError(err)
	}
}

// clearLicenseTables helper
func (suite *DriverLicenseCrudServiceTestSuite) clearLicenseTables() {
	suite.db.Exec("PRAGMA foreign_keys = OFF")
	defer suite.db.Exec("PRAGMA foreign_keys = ON")
	tables := []string{"driver_licenses", "users"}
	for _, table := range tables {
		var modelInstance interface{}
		switch table {
		case "users":
			modelInstance = &model.User{}
		case "driver_licenses":
			modelInstance = &model.DriverLicense{}
		default:
			suite.T().Logf("Skipping unknown table for clearing: %s", table)
			continue
		}
		err := suite.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(modelInstance).Error
		if err != nil && !strings.Contains(err.Error(), "no such table") {
			suite.Require().NoError(err, fmt.Sprintf("Failed to clear table %s", table))
		}
	}
}

// SetupTest runs before each test
func (suite *DriverLicenseCrudServiceTestSuite) SetupTest() {
	suite.logObserver.TakeAll()
	suite.clearLicenseTables()
	suite.mockUserSvc.ExpectedCalls = nil
	suite.mockUserSvc.Calls = nil

	// Seed a user and store it
	suite.seededOwnerUser = suite.seedUserForLicenseTest(
		"owner.seed@example.com",
		model.RoleOsoba,
		"SEEDLIC",
	)
}

// Helper to create a user directly in the DB for FK constraints
func (suite *DriverLicenseCrudServiceTestSuite) seedUserForLicenseTest(email string, role model.UserRole, oibPrefix string) *model.User {
	hashedPassword, _ := auth.HashPassword("licensetestpass")
	user := &model.User{
		Uuid:         uuid.New(),
		FirstName:    "LicOwner",
		LastName:     string(role),
		OIB:          oibPrefix + uuid.New().String()[:(11-len(oibPrefix))],
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         role,
		BirthDate:    time.Now().AddDate(-30, 0, 0),
		Residence:    "License Test Residence",
	}
	err := suite.db.Create(user).Error
	suite.Require().NoError(err)
	return user
}

func TestDriverLicenseCrudServiceSuite(t *testing.T) {
	suite.Run(t, new(DriverLicenseCrudServiceTestSuite))
}

// --- Test Cases ---

func (suite *DriverLicenseCrudServiceTestSuite) TestCreateLicense_Success() {
	// Use the seeded user's UUID from SetupTest
	ownerUUID := suite.seededOwnerUser.Uuid

	newLicenseData := &model.DriverLicense{
		Uuid:          uuid.New(),
		LicenseNumber: "DL12345",
		Category:      "B",
		IssueDate:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		ExpiringDate:  time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	// Mock service Create call's dependency (userService.Read)
	// Return the actual seeded user object with its correct DB ID
	suite.mockUserSvc.On("Read", ownerUUID).Return(suite.seededOwnerUser, nil).Once()

	createdLicense, err := suite.licenseService.Create(newLicenseData, ownerUUID)

	// ... assertions ...
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), createdLicense)
	assert.NotZero(suite.T(), createdLicense.ID)
	assert.Equal(suite.T(), suite.seededOwnerUser.ID, createdLicense.UserId)
	assert.Equal(suite.T(), newLicenseData.LicenseNumber, createdLicense.LicenseNumber)

	// Verify in DB by ID
	var dbLicense model.DriverLicense
	errDb := suite.db.First(&dbLicense, createdLicense.ID).Error
	assert.NoError(suite.T(), errDb)
	assert.Equal(suite.T(), suite.seededOwnerUser.ID, dbLicense.UserId)
	suite.mockUserSvc.AssertExpectations(suite.T())
}

func (suite *DriverLicenseCrudServiceTestSuite) TestCreateLicense_OwnerNotFound() {
	ownerUUID := uuid.New()
	suite.mockUserSvc.On("Read", ownerUUID).Return(nil, gorm.ErrRecordNotFound).Once()

	newLicenseData := &model.DriverLicense{LicenseNumber: "DLNFAIL1"}
	_, err := suite.licenseService.Create(newLicenseData, ownerUUID)

	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, gorm.ErrRecordNotFound))
	suite.mockUserSvc.AssertExpectations(suite.T())
}

func (suite *DriverLicenseCrudServiceTestSuite) TestCreateLicense_OwnerBadRole() {
	ownerUser := &model.User{Uuid: uuid.New(), Role: model.RolePolicija} // Policija cannot own license
	ownerUser.ID = 2
	suite.mockUserSvc.On("Read", ownerUser.Uuid).Return(ownerUser, nil).Once()

	newLicenseData := &model.DriverLicense{LicenseNumber: "DLNFAIL2"}
	_, err := suite.licenseService.Create(newLicenseData, ownerUser.Uuid)

	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, cerror.ErrBadRole))
	suite.mockUserSvc.AssertExpectations(suite.T())
}

func (suite *DriverLicenseCrudServiceTestSuite) TestGetLicenseByUuid_Success() {
	owner := suite.seedUserForLicenseTest("owner.getlic@example.com", model.RoleOsoba, "LICGET")
	licenseToCreate := &model.DriverLicense{
		Uuid:          uuid.New(),
		UserId:        owner.ID,
		LicenseNumber: "LICGET001",
		Category:      "C",
		IssueDate:     time.Now().AddDate(-1, 0, 0),
		ExpiringDate:  time.Now().AddDate(4, 0, 0),
	}
	err := suite.db.Create(licenseToCreate).Error
	suite.Require().NoError(err)

	fetchedLicense, err := suite.licenseService.GetByUuid(licenseToCreate.Uuid)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), fetchedLicense)
	assert.Equal(suite.T(), licenseToCreate.ID, fetchedLicense.ID)
	assert.Equal(suite.T(), licenseToCreate.LicenseNumber, fetchedLicense.LicenseNumber)
}

func (suite *DriverLicenseCrudServiceTestSuite) TestGetLicenseByUuid_NotFound() {
	nonExistentUUID := uuid.New()
	_, err := suite.licenseService.GetByUuid(nonExistentUUID)
	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, gorm.ErrRecordNotFound))
}

func (suite *DriverLicenseCrudServiceTestSuite) TestGetAllLicenses_Success() {
	owner1 := suite.seedUserForLicenseTest("owner1.getall@example.com", model.RoleOsoba, "LICALL1")
	owner2 := suite.seedUserForLicenseTest("owner2.getall@example.com", model.RoleFirma, "LICALL2")
	suite.db.Create(&model.DriverLicense{Uuid: uuid.New(), UserId: owner1.ID, LicenseNumber: "LICA1", Category: "A"})
	suite.db.Create(&model.DriverLicense{Uuid: uuid.New(), UserId: owner2.ID, LicenseNumber: "LICB1", Category: "B"})

	licenses, err := suite.licenseService.GetAll()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), licenses, 2)
}

func (suite *DriverLicenseCrudServiceTestSuite) TestUpdateLicense_Success() {
	owner := suite.seedUserForLicenseTest("owner.updlic@example.com", model.RoleOsoba, "LICUPD")
	licenseToUpdate := &model.DriverLicense{
		Uuid:          uuid.New(),
		UserId:        owner.ID,
		LicenseNumber: "LICUPD001",
		Category:      "D",
		IssueDate:     time.Now().AddDate(-3, 0, 0),
		ExpiringDate:  time.Now().AddDate(2, 0, 0),
	}
	err := suite.db.Create(licenseToUpdate).Error
	suite.Require().NoError(err)

	updateData := &model.DriverLicense{
		// Uuid and UserId should not be in updateData if we are not changing owner
		LicenseNumber: "LICUPD001-MODIFIED",
		Category:      "D, E",
		IssueDate:     licenseToUpdate.IssueDate, // Keep issue date same or update validly
		ExpiringDate:  time.Now().AddDate(7, 0, 0),
	}

	updatedLicense, err := suite.licenseService.Update(licenseToUpdate.Uuid, updateData)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updatedLicense)
	assert.Equal(suite.T(), "LICUPD001-MODIFIED", updatedLicense.LicenseNumber)
	assert.Equal(suite.T(), "D, E", updatedLicense.Category)
	assert.Equal(suite.T(), owner.ID, updatedLicense.UserId) // Ensure owner didn't change

	var dbLicense model.DriverLicense
	suite.db.First(&dbLicense, licenseToUpdate.ID)
	assert.Equal(suite.T(), "LICUPD001-MODIFIED", dbLicense.LicenseNumber)
}

func (suite *DriverLicenseCrudServiceTestSuite) TestUpdateLicense_AttemptToChangeOwnerDenied() {
	owner1 := suite.seedUserForLicenseTest("owner1.updlicfail@example.com", model.RoleOsoba, "LICUFL1")
	owner2 := suite.seedUserForLicenseTest("owner2.updlicfail@example.com", model.RoleFirma, "LICUFL2") // Different owner

	licenseToUpdate := &model.DriverLicense{
		Uuid:          uuid.New(),
		UserId:        owner1.ID, // Belongs to owner1
		LicenseNumber: "LICFAILUPD",
	}
	err := suite.db.Create(licenseToUpdate).Error
	suite.Require().NoError(err)

	updateDataAttemptingOwnerChange := &model.DriverLicense{
		UserId:        owner2.ID, // Attempting to change UserId
		LicenseNumber: "LICFAILUPD-MOD",
	}

	_, err = suite.licenseService.Update(licenseToUpdate.Uuid, updateDataAttemptingOwnerChange)
	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, cerror.ErrBadRole), "Expected error for trying to change owner via update")
}

func (suite *DriverLicenseCrudServiceTestSuite) TestDeleteLicense_Success() {
	owner := suite.seedUserForLicenseTest("owner.dellic@example.com", model.RoleOsoba, "LICDEL")
	licenseToDelete := &model.DriverLicense{
		Uuid:          uuid.New(),
		UserId:        owner.ID,
		LicenseNumber: "LICDEL001",
	}
	err := suite.db.Create(licenseToDelete).Error
	suite.Require().NoError(err)

	err = suite.licenseService.Delete(licenseToDelete.Uuid)
	assert.NoError(suite.T(), err)

	var dbLicense model.DriverLicense
	errDb := suite.db.First(&dbLicense, "uuid = ?", licenseToDelete.Uuid).Error
	assert.Error(suite.T(), errDb)
	assert.True(suite.T(), errors.Is(errDb, gorm.ErrRecordNotFound))
}

func (suite *DriverLicenseCrudServiceTestSuite) TestDeleteLicense_NotFound() {
	nonExistentUUID := uuid.New()
	err := suite.licenseService.Delete(nonExistentUUID)
	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, gorm.ErrRecordNotFound))
}
