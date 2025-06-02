package device_test

import (
	"ePrometna_Server/app"
	"ePrometna_Server/config"
	"ePrometna_Server/model"
	"ePrometna_Server/util/auth" //
	"ePrometna_Server/util/device"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// --- Global Test Variables ---
var (
	testUserGlobal = &model.User{
		Uuid:         uuid.New(),
		FirstName:    "Global",
		LastName:     "Tester",
		OIB:          "12345678901",
		Email:        "global.tester@example.com",
		PasswordHash: "hashedpassword",
		Role:         model.RoleOsoba,
		BirthDate:    time.Now().AddDate(-30, 0, 0),
		Residence:    "123 Test St",
	}
	testDeviceTokenGlobal = "fixed-test-device-token"
)

// --- Test Suite Definition ---
type DeviceManagerTestSuite struct {
	suite.Suite
	db          *gorm.DB
	deviceMgr   *device.DeviceManager
	logger      *zap.SugaredLogger
	logObserver *observer.ObservedLogs
}

// SetupSuite runs once before all tests in the suite
func (suite *DeviceManagerTestSuite) SetupSuite() {
	// Setup Zap logger
	core, obs := observer.New(zap.InfoLevel) // Capture logs for assertions if needed
	suite.logger = zap.New(core).Sugar()
	suite.logObserver = obs

	db, err := gorm.Open(sqlite.Open("file:devicemgr_test.db?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	suite.Require().NoError(err, "Failed to connect to SQLite")
	suite.db = db

	// Migrate schema
	err = suite.db.AutoMigrate(model.GetAllModels()...)
	suite.Require().NoError(err, "Failed to migrate database schema")

	// Configure AppConfig for auth.GenerateDeviceToken
	config.AppConfig = &config.AppConfiguration{
		AccessKey:  "test-access-key-device-sqlite",
		RefreshKey: "test-refresh-key-device-sqlite",
	}

	app.Test()
	app.Provide(func() *gorm.DB { return suite.db })
	app.Provide(func() *zap.SugaredLogger { return suite.logger })
	// Initialize DeviceManager with the real (in-memory) DB
	suite.deviceMgr = device.NewDeviceManager()
}

// TearDownSuite runs once after all tests in the suite
func (suite *DeviceManagerTestSuite) TearDownSuite() {
	if suite.logger != nil {
		_ = suite.logger.Sync()
	}
	if suite.db != nil {
		sqlDB, _ := suite.db.DB()
		err := sqlDB.Close()
		suite.Require().NoError(err)
	}
}

// clearTables removes all data from specified tables.
func (suite *DeviceManagerTestSuite) clearTables(tables ...string) {
	suite.db.Exec("PRAGMA foreign_keys = OFF")
	defer suite.db.Exec("PRAGMA foreign_keys = ON")

	for _, table := range tables {
		if table == "mobiles" {
			err := suite.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&model.Mobile{}).Error
			suite.Require().NoError(err, fmt.Sprintf("Failed to clear table %s", table))
		}
		if table == "users" {
			err := suite.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&model.User{}).Error
			suite.Require().NoError(err, fmt.Sprintf("Failed to clear table %s", table))
		}
	}
}

// SetupTest runs before each test in the suite
func (suite *DeviceManagerTestSuite) SetupTest() {
	suite.logObserver.TakeAll()
	suite.clearTables("mobiles", "users")

	userToCreate := *testUserGlobal
	userToCreate.OIB = uuid.New().String()[:11]
	userToCreate.Email = fmt.Sprintf("testuser_%s@example.com", uuid.New().String()[:6])

	err := suite.db.Create(&userToCreate).Error
	suite.Require().NoError(err, "Failed to create global test user for SetupTest")
	testUserGlobal.ID = userToCreate.ID
}

// Helper to create a user directly in the DB for tests
func (suite *DeviceManagerTestSuite) createTestUserInDB(email string, role model.UserRole) *model.User {
	user := &model.User{
		Uuid:         uuid.New(),
		FirstName:    "Test",
		LastName:     string(role),
		OIB:          uuid.New().String()[:11],
		Email:        email,
		Role:         role,
		BirthDate:    time.Now().AddDate(-25, 0, 0),
		Residence:    "DB Test Residence",
		PasswordHash: "db-dummy-hash",
	}
	err := suite.db.Create(user).Error
	suite.Require().NoError(err, "Failed to create user in DB for test. User: %+v", user)
	suite.Require().NotZero(user.ID, "User ID should not be zero after DB creation")
	return user
}

// TestDeviceManagerSuite runs the test suite
func TestDeviceManagerSuite(t *testing.T) {
	suite.Run(t, new(DeviceManagerTestSuite))
}

// --- Test Cases ---

func (suite *DeviceManagerTestSuite) TestNewDeviceManager() {
	assert.NotNil(suite.T(), suite.deviceMgr)
	assert.Equal(suite.T(), suite.db, suite.deviceMgr.DB)
	assert.Equal(suite.T(), suite.logger, suite.deviceMgr.Logger)
}

func (suite *DeviceManagerTestSuite) TestFormatDeviceName() {
	tests := []struct {
		name       string
		deviceInfo device.DeviceInfo
		expected   string
	}{
		{
			name: "Standard device info",
			deviceInfo: device.DeviceInfo{
				Platform:  "Android",
				Brand:     "Google",
				ModelName: "Pixel 8",
				DeviceID:  "testPixel8",
			},
			expected: "Google Pixel 8 (Android) [ID:testPixel8]",
		},
		{
			name: "Device info with empty fields",
			deviceInfo: device.DeviceInfo{
				Platform:  "iOS",
				Brand:     "",
				ModelName: "iPhone 15",
				DeviceID:  "testIphone15",
			},
			expected: " iPhone 15 (iOS) [ID:testIphone15]",
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			formattedName := suite.deviceMgr.FormatDeviceName(tc.deviceInfo)
			assert.Equal(t, tc.expected, formattedName)
		})
	}
}

func (suite *DeviceManagerTestSuite) TestFindDeviceByID_Found() {
	deviceID := "uniqueDevice123"
	expectedMobile := &model.Mobile{
		Uuid:             uuid.New(),
		UserId:           testUserGlobal.ID,
		CreatorId:        testUserGlobal.ID,
		RegisteredDevice: "SomeBrand SomeModel (Platform) [ID:" + deviceID + "]",
		ActivationToken:  "some-token",
	}
	err := suite.db.Create(expectedMobile).Error
	suite.Require().NoError(err)

	mobile, err := suite.deviceMgr.FindDeviceByID(deviceID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), mobile)
	assert.Equal(suite.T(), expectedMobile.ID, mobile.ID)
	assert.Contains(suite.T(), mobile.RegisteredDevice, deviceID)
}

func (suite *DeviceManagerTestSuite) TestFindDeviceByID_NotFound() {
	deviceID := "nonExistentDevice456"
	mobile, err := suite.deviceMgr.FindDeviceByID(deviceID)
	assert.ErrorIs(suite.T(), err, gorm.ErrRecordNotFound)
	assert.Nil(suite.T(), mobile)
}

func (suite *DeviceManagerTestSuite) TestFindUserDevice_Found() {
	expectedMobile := &model.Mobile{
		Uuid:             uuid.New(),
		UserId:           testUserGlobal.ID,
		CreatorId:        testUserGlobal.ID,
		RegisteredDevice: "User's Registered Device [ID:userDev1]",
		ActivationToken:  "user-device-token",
	}
	err := suite.db.Create(expectedMobile).Error
	suite.Require().NoError(err)

	mobile, err := suite.deviceMgr.FindUserDevice(testUserGlobal.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), mobile)
	assert.Equal(suite.T(), expectedMobile.ID, mobile.ID)
	assert.Equal(suite.T(), testUserGlobal.ID, mobile.UserId)
}

func (suite *DeviceManagerTestSuite) TestFindUserDevice_NotFound() {
	nonExistentUserID := uint(999999)
	mobile, err := suite.deviceMgr.FindUserDevice(nonExistentUserID)
	assert.ErrorIs(suite.T(), err, gorm.ErrRecordNotFound)
	assert.Nil(suite.T(), mobile)
}

func (suite *DeviceManagerTestSuite) TestUpdateDeviceToken_Success() {
	existingDevice := &model.Mobile{
		Uuid:             uuid.New(),
		UserId:           testUserGlobal.ID,
		CreatorId:        testUserGlobal.ID,
		ActivationToken:  "old-token",
		RegisteredDevice: "DeviceToUpdate [ID:updDev1]",
	}
	err := suite.db.Create(existingDevice).Error
	suite.Require().NoError(err)

	token, err := suite.deviceMgr.UpdateDeviceToken(existingDevice, testUserGlobal)
	assert.NoError(suite.T(), err)

	// Check that the token returned by GenerateDeviceToken is what we expect
	expectedToken, _ := auth.GenerateDeviceToken(testUserGlobal) // Generate expected token
	assert.Equal(suite.T(), expectedToken, token)

	// Verify in DB
	var updatedDevice model.Mobile
	err = suite.db.First(&updatedDevice, existingDevice.ID).Error
	suite.Require().NoError(err)
	assert.Equal(suite.T(), expectedToken, updatedDevice.ActivationToken)
}

func (suite *DeviceManagerTestSuite) TestRegisterNewDevice_Success() {
	deviceName := "NewBrand NewModel (Platform) [ID:newDevReg1]"

	token, err := suite.deviceMgr.RegisterNewDevice(testUserGlobal, deviceName)
	assert.NoError(suite.T(), err)

	expectedToken, _ := auth.GenerateDeviceToken(testUserGlobal)
	assert.Equal(suite.T(), expectedToken, token)

	// Verify in DB
	var newDevice model.Mobile
	err = suite.db.Where("registered_device = ?", deviceName).First(&newDevice).Error
	suite.Require().NoError(err)
	assert.Equal(suite.T(), testUserGlobal.ID, newDevice.UserId)
	assert.Equal(suite.T(), expectedToken, newDevice.ActivationToken)
}

// --- Tests for ValidateDeviceRegistration ---

func (suite *DeviceManagerTestSuite) TestValidateDeviceRegistration_NewDevice_UserHasNoDevice() {
	currentUser := suite.createTestUserInDB("newdeviceuser@example.com", model.RoleOsoba)
	deviceInfo := device.DeviceInfo{Platform: "TestOS", Brand: "TestBrand", ModelName: "TestModelX", DeviceID: "validateNewDev1"}

	token, isNew, err := suite.deviceMgr.ValidateDeviceRegistration(currentUser, deviceInfo)

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), isNew)
	expectedToken, _ := auth.GenerateDeviceToken(currentUser)
	assert.Equal(suite.T(), expectedToken, token)

	// Verify in DB
	var mobileDevice model.Mobile
	err = suite.db.Where("user_id = ? AND registered_device LIKE ?", currentUser.ID, "%"+deviceInfo.DeviceID+"%").First(&mobileDevice).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedToken, mobileDevice.ActivationToken)
}

func (suite *DeviceManagerTestSuite) TestValidateDeviceRegistration_ExistingDevice_SameUser() {
	currentUser := suite.createTestUserInDB("existingdevice_sameuser@example.com", model.RoleOsoba)
	deviceInfo := device.DeviceInfo{Platform: "TestOS", Brand: "TestBrand", ModelName: "TestModelY", DeviceID: "validateExistingDevSameUser"}
	formattedName := suite.deviceMgr.FormatDeviceName(deviceInfo)

	// Pre-register the device for this user
	initialDevice := &model.Mobile{
		Uuid:             uuid.New(),
		UserId:           currentUser.ID,
		CreatorId:        currentUser.ID,
		RegisteredDevice: formattedName,
		ActivationToken:  "initial-old-token",
	}
	err := suite.db.Create(initialDevice).Error
	suite.Require().NoError(err)

	token, isNew, err := suite.deviceMgr.ValidateDeviceRegistration(currentUser, deviceInfo)

	assert.NoError(suite.T(), err)
	assert.False(suite.T(), isNew, "Should not be a new registration")
	expectedToken, _ := auth.GenerateDeviceToken(currentUser)
	assert.Equal(suite.T(), expectedToken, token, "Token should be updated")

	// Verify in DB
	var mobileDevice model.Mobile
	err = suite.db.First(&mobileDevice, initialDevice.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedToken, mobileDevice.ActivationToken)
}

func (suite *DeviceManagerTestSuite) TestValidateDeviceRegistration_ExistingDevice_DifferentUser() {
	user1 := suite.createTestUserInDB("user1_diff@example.com", model.RoleOsoba)
	user2 := suite.createTestUserInDB("user2_diff@example.com", model.RoleFirma)

	deviceInfo := device.DeviceInfo{Platform: "TestOS", Brand: "TestBrand", ModelName: "TestModelZ", DeviceID: "validateExistingDevDiffUser"}
	formattedName := suite.deviceMgr.FormatDeviceName(deviceInfo)

	// Pre-register the device for user1
	initialDevice := &model.Mobile{
		Uuid:             uuid.New(),
		UserId:           user1.ID,
		CreatorId:        user1.ID,
		RegisteredDevice: formattedName,
		ActivationToken:  "user1-token",
	}
	err := suite.db.Create(initialDevice).Error
	suite.Require().NoError(err)

	// user2 tries to validate with this device
	token, isNew, err := suite.deviceMgr.ValidateDeviceRegistration(user2, deviceInfo)

	assert.Error(suite.T(), err)
	assert.False(suite.T(), isNew)
	assert.Empty(suite.T(), token)
	assert.EqualError(suite.T(), err, "this device is already registered to another user")
}

func (suite *DeviceManagerTestSuite) TestValidateDeviceRegistration_NewDevice_UserAlreadyHasADifferentDevice() {
	currentUser := suite.createTestUserInDB("user_has_other_device@example.com", model.RoleOsoba)
	deviceInfoNew := device.DeviceInfo{Platform: "NewOS", Brand: "NewBrand", ModelName: "NewModel", DeviceID: "newDeviceIDForUserWithOld"}

	// Pre-register a different device for this user
	oldDevice := &model.Mobile{
		Uuid:             uuid.New(),
		UserId:           currentUser.ID,
		CreatorId:        currentUser.ID,
		RegisteredDevice: "OldBrand OldModel (OldOS) [ID:userOldDeviceID]",
		ActivationToken:  "user-old-device-token",
	}
	err := suite.db.Create(oldDevice).Error
	suite.Require().NoError(err)

	// Current user tries to validate the new device
	token, isNew, err := suite.deviceMgr.ValidateDeviceRegistration(currentUser, deviceInfoNew)

	assert.Error(suite.T(), err)
	assert.False(suite.T(), isNew)
	assert.Empty(suite.T(), token)
	assert.EqualError(suite.T(), err, "you already have a different device registered to your account")

	// Verify the old device is still registered and unchanged
	var dbOldDevice model.Mobile
	err = suite.db.First(&dbOldDevice, oldDevice.ID).Error
	suite.Require().NoError(err)
	assert.Equal(suite.T(), "user-old-device-token", dbOldDevice.ActivationToken)

	// Verify the new device was not registered
	var dbNewDevice model.Mobile
	err = suite.db.Where("registered_device LIKE ?", "%"+deviceInfoNew.DeviceID+"%").First(&dbNewDevice).Error
	assert.True(suite.T(), errors.Is(err, gorm.ErrRecordNotFound), "New device should not have been registered")
}

// Test for concurrent ValidateDeviceRegistration calls trying to register the same new device for the same new user.
func (suite *DeviceManagerTestSuite) TestValidateDeviceRegistration_Concurrent_NewDevice_SameUser() {
	concurrentUser := suite.createTestUserInDB("concurrent_user@example.com", model.RoleOsoba)
	deviceInfo := device.DeviceInfo{Platform: "ConcOS", Brand: "ConcBrand", ModelName: "ConcModel", DeviceID: "concurrentDeviceID"}

	numGoroutines := 2
	errs := make(chan error, numGoroutines)
	successes := make(chan bool, numGoroutines)

	expectedTokenForConcurrent, _ := auth.GenerateDeviceToken(concurrentUser)

	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			token, isNew, err := suite.deviceMgr.ValidateDeviceRegistration(concurrentUser, deviceInfo)
			if err != nil {
				suite.logger.Errorf("Goroutine %d error: %v", routineID, err)
				errs <- err
				return
			}
			if isNew && token == expectedTokenForConcurrent {
				successes <- true
			} else if !isNew && token == expectedTokenForConcurrent {
				successes <- true
			} else {
				errs <- fmt.Errorf("goroutine %d: unexpected outcome (isNew: %v, token: %s)", routineID, isNew, token)
			}
		}(i)
	}

	successfulRegistrations := 0
	errorsEncountered := 0

	for i := 0; i < numGoroutines; i++ {
		select {
		case <-successes:
			successfulRegistrations++
		case err := <-errs:
			suite.logger.Warnf("Concurrent registration attempt failed with: %v", err)
			errorsEncountered++
		case <-time.After(5 * time.Second): // Timeout to prevent test hanging
			suite.T().Fatal("Timeout waiting for concurrent registrations to complete")
		}
	}

	assert.GreaterOrEqual(suite.T(), successfulRegistrations, 1, "At least one goroutine should successfully register or update the device token")
	assert.LessOrEqual(suite.T(), errorsEncountered, numGoroutines-1, "Not all goroutines should error out if one succeeds")

	// Verify DB state: only one device entry for this user and deviceID
	var count int64
	err := suite.db.Model(&model.Mobile{}).Where("user_id = ? AND registered_device LIKE ?", concurrentUser.ID, "%"+deviceInfo.DeviceID+"%").Count(&count).Error
	suite.Require().NoError(err)
	assert.Equal(suite.T(), int64(1), count, "Should be exactly one device registered for the user with this device ID after concurrent attempts")

	var finalDevice model.Mobile
	err = suite.db.Where("user_id = ? AND registered_device LIKE ?", concurrentUser.ID, "%"+deviceInfo.DeviceID+"%").First(&finalDevice).Error
	suite.Require().NoError(err)
	assert.Equal(suite.T(), expectedTokenForConcurrent, finalDevice.ActivationToken)
}
