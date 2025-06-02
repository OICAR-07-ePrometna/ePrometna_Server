package service_test

import (
	"ePrometna_Server/app"
	"ePrometna_Server/config"
	"ePrometna_Server/model"
	"ePrometna_Server/service"
	"ePrometna_Server/util/auth"
	"ePrometna_Server/util/cerror"
	"ePrometna_Server/util/device"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// --- LoginService Test Suite ---
type LoginServiceTestSuite struct {
	suite.Suite
	db           *gorm.DB
	loginService service.ILoginService
	logger       *zap.SugaredLogger
	logObserver  *observer.ObservedLogs
}

// SetupSuite runs once before all tests in the suite
func (suite *LoginServiceTestSuite) SetupSuite() {
	core, obs := observer.New(zap.InfoLevel)
	suite.logger = zap.New(core).Sugar()
	suite.logObserver = obs
	zap.ReplaceGlobals(zap.New(core))

	db, err := gorm.Open(sqlite.Open("file:loginservice_test.db?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	suite.Require().NoError(err, "Failed to connect to SQLite")
	suite.db = db

	err = suite.db.AutoMigrate(model.GetAllModels()...)
	suite.Require().NoError(err, "Failed to migrate database schema")

	config.AppConfig = &config.AppConfiguration{
		Env:        config.Dev,
		AccessKey:  "login-service-test-access-key",
		RefreshKey: "login-service-test-refresh-key",
	}

	app.Test() // Initialize DIG container
	app.Provide(func() *gorm.DB { return suite.db })
	app.Provide(func() *zap.SugaredLogger { return suite.logger })
	suite.loginService = service.NewLoginService()
	suite.Require().NotNil(suite.loginService, "LoginService should be initialized by DIG")
}

// TearDownSuite runs once after all tests
func (suite *LoginServiceTestSuite) TearDownSuite() {
	if suite.logger != nil {
		_ = suite.logger.Sync()
	}
	if suite.db != nil {
		sqlDB, _ := suite.db.DB()
		err := sqlDB.Close()
		suite.Require().NoError(err)
	}
}

// clearTables helper
func (suite *LoginServiceTestSuite) clearTables(tables ...string) {
	suite.db.Exec("PRAGMA foreign_keys = OFF")
	defer suite.db.Exec("PRAGMA foreign_keys = ON")
	for _, table := range tables {
		var modelInstance interface{}
		switch table {
		case "users":
			modelInstance = &model.User{}
		case "mobiles":
			modelInstance = &model.Mobile{}
		default:
			suite.T().Fatalf("Unsupported table for clearing: %s", table)
		}
		err := suite.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(modelInstance).Error
		suite.Require().NoError(err, fmt.Sprintf("Failed to clear table %s", table))
	}
}

// SetupTest runs before each test
func (suite *LoginServiceTestSuite) SetupTest() {
	suite.logObserver.TakeAll()           // Clear observed logs
	suite.clearTables("users", "mobiles") // Clear relevant tables
}

// Helper to create a user with a hashed password
func (suite *LoginServiceTestSuite) createTestUser(email, plainPassword string, role model.UserRole) *model.User {
	hashedPassword, err := auth.HashPassword(plainPassword)
	suite.Require().NoError(err)
	user := &model.User{
		Uuid:         uuid.New(),
		FirstName:    "Test",
		LastName:     string(role),
		OIB:          uuid.New().String()[:11], // Unique OIB
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         role,
		BirthDate:    time.Now().AddDate(-25, 0, 0),
		Residence:    "Test Residence",
	}
	err = suite.db.Create(user).Error
	suite.Require().NoError(err, "Failed to create user for test")
	return user
}

// TestLoginServiceSuite runs the test suite
func TestLoginServiceSuite(t *testing.T) {
	suite.Run(t, new(LoginServiceTestSuite))
}

// --- Test Cases ---

func (suite *LoginServiceTestSuite) TestLogin_Success() {
	email := "testlogin@example.com"
	password := "password123"
	suite.createTestUser(email, password, model.RoleOsoba)

	accessToken, refreshToken, err := suite.loginService.Login(email, password)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), accessToken)
	assert.NotEmpty(suite.T(), refreshToken)
}

func (suite *LoginServiceTestSuite) TestLogin_UserNotFound() {
	accessToken, refreshToken, err := suite.loginService.Login("nonexistent@example.com", "password123")

	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, cerror.ErrInvalidCredentials))
	assert.Empty(suite.T(), accessToken)
	assert.Empty(suite.T(), refreshToken)
}

func (suite *LoginServiceTestSuite) TestLogin_IncorrectPassword() {
	email := "wrongpass@example.com"
	password := "password123"
	suite.createTestUser(email, password, model.RoleOsoba)

	accessToken, refreshToken, err := suite.loginService.Login(email, "wrongPassword")

	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, cerror.ErrInvalidCredentials))
	assert.Empty(suite.T(), accessToken)
	assert.Empty(suite.T(), refreshToken)
}

func (suite *LoginServiceTestSuite) TestRefreshTokens_Success() {
	user := suite.createTestUser("refreshuser@example.com", "password123", model.RoleFirma)

	accessToken, refreshToken, err := suite.loginService.RefreshTokens(user)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), accessToken)
	assert.NotEmpty(suite.T(), refreshToken)

	_, accessClaims, errAccess := auth.ParseToken("Bearer " + accessToken)
	suite.Require().NoError(errAccess)
	assert.Equal(suite.T(), user.Uuid.String(), accessClaims.Uuid)

	var refreshClaims auth.Claims
	_, errRefresh := jwt.ParseWithClaims(refreshToken, &refreshClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.RefreshKey), nil
	})
	suite.Require().NoError(errRefresh)
	assert.Equal(suite.T(), user.Uuid.String(), refreshClaims.Uuid)
}

func (suite *LoginServiceTestSuite) TestLoginMobile_Success_NewDevice() {
	email := "mobile.new@example.com"
	password := "mobilePass"
	user := suite.createTestUser(email, password, model.RoleOsoba)
	deviceInfo := device.DeviceInfo{
		Platform:  "Android",
		Brand:     "TestBrand",
		ModelName: "TestModel",
		DeviceID:  "newDeviceID123",
	}

	result, err := suite.loginService.LoginMobile(email, password, deviceInfo)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.NotEmpty(suite.T(), result.AccessToken)
	assert.NotEmpty(suite.T(), result.RefreshToken)
	assert.NotEmpty(suite.T(), result.DeviceToken)

	// Verify device was registered in DB
	var mobileRec model.Mobile
	dbErr := suite.db.Where("user_id = ?", user.ID).First(&mobileRec).Error
	assert.NoError(suite.T(), dbErr)
	assert.Contains(suite.T(), mobileRec.RegisteredDevice, deviceInfo.DeviceID)
	assert.Equal(suite.T(), result.DeviceToken, mobileRec.ActivationToken)
}

func (suite *LoginServiceTestSuite) TestLoginMobile_Success_ExistingDevice_SameUser() {
	email := "mobile.existing@example.com"
	password := "mobilePassExisting"
	user := suite.createTestUser(email, password, model.RoleOsoba)
	deviceInfo := device.DeviceInfo{
		Platform:  "iOS",
		Brand:     "Apple",
		ModelName: "iPhoneTest",
		DeviceID:  "existingDeviceID456",
	}
	deviceMgr := device.NewDeviceManager()
	formattedName := deviceMgr.FormatDeviceName(deviceInfo)

	// Pre-register the device
	initialDeviceToken, _ := auth.GenerateDeviceToken(user)
	initialMobile := model.Mobile{
		Uuid:             uuid.New(),
		UserId:           user.ID,
		CreatorId:        user.ID,
		RegisteredDevice: formattedName,
		ActivationToken:  initialDeviceToken,
	}
	errCreate := suite.db.Create(&initialMobile).Error
	suite.Require().NoError(errCreate)

	result, err := suite.loginService.LoginMobile(email, password, deviceInfo)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.NotEmpty(suite.T(), result.AccessToken)
	assert.NotEmpty(suite.T(), result.RefreshToken)
	assert.NotEmpty(suite.T(), result.DeviceToken)
	assert.Equal(suite.T(), initialDeviceToken, result.DeviceToken, "Device token should be updated")

	// Verify device token was updated in DB
	var mobileRec model.Mobile
	dbErr := suite.db.First(&mobileRec, initialMobile.ID).Error
	assert.NoError(suite.T(), dbErr)
	assert.Equal(suite.T(), result.DeviceToken, mobileRec.ActivationToken)
}

func (suite *LoginServiceTestSuite) TestLoginMobile_Error_DeviceRegisteredToAnotherUser() {
	emailUser1 := "mobile.user1@example.com"
	passwordUser1 := "passUser1"
	user1 := suite.createTestUser(emailUser1, passwordUser1, model.RoleOsoba)

	emailUser2 := "mobile.user2@example.com"
	passwordUser2 := "passUser2"
	suite.createTestUser(emailUser2, passwordUser2, model.RoleFirma) // User2 who will attempt login

	deviceInfo := device.DeviceInfo{
		Platform:  "Android",
		Brand:     "SharedBrand",
		ModelName: "SharedModel",
		DeviceID:  "sharedDeviceID789",
	}
	deviceMgr := device.NewDeviceManager()
	formattedName := deviceMgr.FormatDeviceName(deviceInfo)

	// Register device to user1
	deviceTokenUser1, _ := auth.GenerateDeviceToken(user1)
	mobileForUser1 := model.Mobile{
		Uuid:             uuid.New(),
		UserId:           user1.ID,
		CreatorId:        user1.ID,
		RegisteredDevice: formattedName,
		ActivationToken:  deviceTokenUser1,
	}
	errCreate := suite.db.Create(&mobileForUser1).Error
	suite.Require().NoError(errCreate)

	// User2 attempts to login with the same device
	result, err := suite.loginService.LoginMobile(emailUser2, passwordUser2, deviceInfo)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.EqualError(suite.T(), err, "this device is already registered to another user")
}

func (suite *LoginServiceTestSuite) TestLoginMobile_Error_UserAlreadyHasDifferentDevice() {
	email := "mobile.userwithdevice@example.com"
	password := "passWithDevice"
	user := suite.createTestUser(email, password, model.RoleOsoba)

	// Register an initial device for the user
	initialDeviceInfo := device.DeviceInfo{DeviceID: "initialUserDeviceABC"}
	deviceMgr := device.NewDeviceManager()
	formattedInitialName := deviceMgr.FormatDeviceName(initialDeviceInfo)
	initialDeviceToken, _ := auth.GenerateDeviceToken(user)
	initialMobile := model.Mobile{
		Uuid:             uuid.New(),
		UserId:           user.ID,
		CreatorId:        user.ID,
		RegisteredDevice: formattedInitialName,
		ActivationToken:  initialDeviceToken,
	}
	errCreate := suite.db.Create(&initialMobile).Error
	suite.Require().NoError(errCreate)

	// User attempts to login with a *new* different device
	newDeviceInfo := device.DeviceInfo{
		Platform:  "Web",
		Brand:     "BrowserBrand",
		ModelName: "BrowserModel",
		DeviceID:  "newDifferentDeviceIDXYZ",
	}
	result, err := suite.loginService.LoginMobile(email, password, newDeviceInfo)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.EqualError(suite.T(), err, "you already have a different device registered to your account")
}

func (suite *LoginServiceTestSuite) TestLoginMobile_LoginFailed() {
	email := "mobile.loginfail@example.com"
	password := "actualPassword"
	// User exists, but login will use wrong password
	suite.createTestUser(email, password, model.RoleOsoba)

	deviceInfo := device.DeviceInfo{DeviceID: "deviceForLoginFail"}

	result, err := suite.loginService.LoginMobile(email, "wrongPasswordForLoginFail", deviceInfo)

	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, cerror.ErrInvalidCredentials))
	assert.Nil(suite.T(), result)
}
