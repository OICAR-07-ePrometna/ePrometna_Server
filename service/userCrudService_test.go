package service_test

import (
	"ePrometna_Server/app"
	"ePrometna_Server/config"
	"ePrometna_Server/dto"
	"ePrometna_Server/model"
	"ePrometna_Server/service"
	"ePrometna_Server/util/auth"
	"ePrometna_Server/util/cerror"
	"errors"
	"strings"
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

// --- UserCrudService Test Suite ---
type UserCrudServiceTestSuite struct {
	suite.Suite
	db               *gorm.DB
	userCrudService  service.IUserCrudService
	logger           *zap.SugaredLogger
	logObserver      *observer.ObservedLogs
	createdUserIDs   []uint // To store IDs of users created during tests for cleanup or verification
	createdUserUUIDs []uuid.UUID
}

// SetupSuite runs once before all tests in the suite
func (suite *UserCrudServiceTestSuite) SetupSuite() {
	core, obs := observer.New(zap.InfoLevel)
	suite.logger = zap.New(core).Sugar()
	suite.logObserver = obs
	zap.ReplaceGlobals(zap.New(core))

	db, err := gorm.Open(sqlite.Open("file:usercrudservice_test.db?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	suite.Require().NoError(err, "Failed to connect to SQLite for UserCrudService tests")
	suite.db = db

	err = suite.db.AutoMigrate(model.GetAllModels()...)
	suite.Require().NoError(err, "Failed to migrate database schema for UserCrudService tests")

	config.AppConfig = &config.AppConfiguration{ // Basic config for any auth utilities if used indirectly
		IsDevelopment: true,
		AccessKey:     "usercrud-service-test-access-key",
		RefreshKey:    "usercrud-service-test-refresh-key",
	}

	app.Test()
	app.Provide(func() *gorm.DB { return suite.db })
	app.Provide(func() *zap.SugaredLogger { return suite.logger })
	suite.userCrudService = service.NewUserCrudService()
	suite.Require().NotNil(suite.userCrudService, "UserCrudService should be initialized by DIG")
}

// TearDownSuite runs once after all tests
func (suite *UserCrudServiceTestSuite) TearDownSuite() {
	if suite.logger != nil {
		_ = suite.logger.Sync()
	}
	if suite.db != nil {
		sqlDB, _ := suite.db.DB()
		err := sqlDB.Close()
		suite.Require().NoError(err)
	}
}

// clearUserTables helper
func (suite *UserCrudServiceTestSuite) clearUserTables() {
	suite.db.Exec("PRAGMA foreign_keys = OFF")
	defer suite.db.Exec("PRAGMA foreign_keys = ON")
	err := suite.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&model.User{}).Error
	suite.Require().NoError(err, "Failed to clear users table")
}

// SetupTest runs before each test
func (suite *UserCrudServiceTestSuite) SetupTest() {
	suite.logObserver.TakeAll()
	suite.clearUserTables()
	suite.createdUserIDs = []uint{}
	suite.createdUserUUIDs = []uuid.UUID{}
}

// Helper to create a user directly for testing purposes (without going through service.Create)
func (suite *UserCrudServiceTestSuite) seedUser(email string, role model.UserRole, oib, firstName string) *model.User {
	hashedPassword, _ := auth.HashPassword("seedpassword")
	user := &model.User{
		Uuid:         uuid.New(),
		FirstName:    firstName,
		LastName:     string(role),
		OIB:          oib,
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         role,
		BirthDate:    time.Now().AddDate(-30, 0, 0),
		Residence:    "Seed Residence",
	}
	err := suite.db.Create(user).Error
	suite.Require().NoError(err)
	suite.createdUserIDs = append(suite.createdUserIDs, user.ID)
	suite.createdUserUUIDs = append(suite.createdUserUUIDs, user.Uuid)
	return user
}

func TestUserCrudServiceSuite(t *testing.T) {
	suite.Run(t, new(UserCrudServiceTestSuite))
}

// --- Test Cases ---

func (suite *UserCrudServiceTestSuite) TestCreateUser_Success() {
	newUser := &model.User{
		Uuid:      uuid.New(),
		FirstName: "John",
		LastName:  "Doe",
		OIB:       "11122233344",
		Email:     "john.doe.create@example.com",
		Role:      model.RoleOsoba,
		BirthDate: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Residence: "123 Main St",
	}
	plainPassword := "password123"

	createdUser, err := suite.userCrudService.Create(newUser, plainPassword)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), createdUser)
	assert.NotZero(suite.T(), createdUser.ID)
	assert.Equal(suite.T(), newUser.Email, createdUser.Email)
	assert.NotEmpty(suite.T(), createdUser.PasswordHash)
	assert.True(suite.T(), auth.VerifyPassword(createdUser.PasswordHash, plainPassword))

	// Verify in DB
	var dbUser model.User
	errDb := suite.db.First(&dbUser, createdUser.ID).Error
	assert.NoError(suite.T(), errDb)
	assert.Equal(suite.T(), newUser.Email, dbUser.Email)
}

func (suite *UserCrudServiceTestSuite) TestCreateUser_DuplicateEmail() {
	email := "duplicate.email@example.com"
	oib1 := "12345678900"
	oib2 := "00987654321"
	suite.seedUser(email, model.RoleFirma, oib1, "seed")

	newUser := &model.User{
		Uuid:      uuid.New(),
		FirstName: "Jane",
		LastName:  "Doe",
		OIB:       oib2,  // Different OIB
		Email:     email, // Same email
		Role:      model.RoleOsoba,
		BirthDate: time.Date(1992, 2, 2, 0, 0, 0, 0, time.UTC),
		Residence: "456 Oak St",
	}
	_, err := suite.userCrudService.Create(newUser, "newPassword")

	assert.Error(suite.T(), err) // Expecting a unique constraint violation
	// GORM/SQLite error for unique constraint is typically "UNIQUE constraint failed: users.email"
	assert.Contains(suite.T(), strings.ToLower(err.Error()), "unique constraint failed")
}

func (suite *UserCrudServiceTestSuite) TestCreateUser_DuplicateOIB() {
	oib := "99988877766"
	email1 := "user1.oib@example.com"
	email2 := "user2.oib@example.com"
	suite.seedUser(email1, model.RoleFirma, oib, "seed")

	newUser := &model.User{
		Uuid:      uuid.New(),
		FirstName: "Peter",
		LastName:  "Pan",
		OIB:       oib,    // Same OIB
		Email:     email2, // Different email
		Role:      model.RoleOsoba,
		BirthDate: time.Date(1995, 5, 5, 0, 0, 0, 0, time.UTC),
		Residence: "Neverland",
	}
	_, err := suite.userCrudService.Create(newUser, "lostboy")

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), strings.ToLower(err.Error()), "unique constraint failed")
}

func (suite *UserCrudServiceTestSuite) TestReadUser_Success() {
	seededUser := suite.seedUser("read.user@example.com", model.RoleHAK, "88877766655", "seed")

	user, err := suite.userCrudService.Read(seededUser.Uuid)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), seededUser.ID, user.ID)
	assert.Equal(suite.T(), seededUser.Email, user.Email)
}

func (suite *UserCrudServiceTestSuite) TestReadUser_NotFound() {
	nonExistentUUID := uuid.New()
	user, err := suite.userCrudService.Read(nonExistentUUID)

	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, gorm.ErrRecordNotFound))
	assert.Nil(suite.T(), user)
}

func (suite *UserCrudServiceTestSuite) TestReadAllUsers_Success() {
	suite.seedUser("user1.readall@example.com", model.RoleOsoba, "10000000001", "seed")
	suite.seedUser("user2.readall@example.com", model.RoleFirma, "10000000002", "seed")
	suite.seedUser("super.readall@example.com", model.RoleSuperAdmin, "10000000003", "seed") // This one should be excluded by GetAllUsers

	users, err := suite.userCrudService.GetAllUsers() // GetAllUsers specifically excludes superadmin

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), users, 2) // Only user1 and user2
	for _, u := range users {
		assert.NotEqual(suite.T(), model.RoleSuperAdmin, u.Role)
	}
}

func (suite *UserCrudServiceTestSuite) TestUpdateUser_Success() {
	seededUser := suite.seedUser("update.user@example.com", model.RoleOsoba, "77766655544", "seed")
	updateData := &model.User{
		FirstName: "UpdatedJohn",
		LastName:  "UpdatedDoe",
		OIB:       seededUser.OIB,                 // OIB typically shouldn't change or needs careful handling
		Email:     "updated.john.doe@example.com", // Email can change if unique
		Role:      model.RoleFirma,                // Role can change
		BirthDate: time.Date(1985, 5, 15, 0, 0, 0, 0, time.UTC),
		Residence: "Updated Address",
	}

	updatedUser, err := suite.userCrudService.Update(seededUser.Uuid, updateData)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updatedUser)
	assert.Equal(suite.T(), seededUser.ID, updatedUser.ID) // ID should remain the same
	assert.Equal(suite.T(), updateData.FirstName, updatedUser.FirstName)
	assert.Equal(suite.T(), updateData.Email, updatedUser.Email)
	assert.Equal(suite.T(), updateData.Role, updatedUser.Role)

	// Verify in DB
	var dbUser model.User
	suite.db.First(&dbUser, seededUser.ID)
	assert.Equal(suite.T(), updateData.FirstName, dbUser.FirstName)
}

func (suite *UserCrudServiceTestSuite) TestUpdateUser_NotFound() {
	nonExistentUUID := uuid.New()
	updateData := &model.User{FirstName: "NoOne"}
	_, err := suite.userCrudService.Update(nonExistentUUID, updateData)

	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, gorm.ErrRecordNotFound))
}

func (suite *UserCrudServiceTestSuite) TestDeleteUser_Success() {
	seededUser := suite.seedUser("delete.user@example.com", model.RoleOsoba, "66655544433", "seed")
	err := suite.userCrudService.Delete(seededUser.Uuid)
	assert.NoError(suite.T(), err)

	// Verify soft delete (or hard delete if GORM default is changed)
	var dbUser model.User
	errDb := suite.db.Unscoped().First(&dbUser, seededUser.ID).Error
	assert.NoError(suite.T(), errDb)
	assert.NotNil(suite.T(), dbUser.DeletedAt)
}

func (suite *UserCrudServiceTestSuite) TestDeleteUser_NotFound() {
	nonExistentUUID := uuid.New()
	err := suite.userCrudService.Delete(nonExistentUUID)
	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, gorm.ErrRecordNotFound))
}

func (suite *UserCrudServiceTestSuite) TestGetAllPoliceOfficers_Success() {
	suite.seedUser("officer1@example.com", model.RolePolicija, "20000000001", "seed")
	suite.seedUser("officer2@example.com", model.RolePolicija, "20000000002", "seed")
	suite.seedUser("notanofficer@example.com", model.RoleOsoba, "20000000003", "seed")

	officers, err := suite.userCrudService.GetAllPoliceOfficers()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), officers, 2)
	for _, officer := range officers {
		assert.Equal(suite.T(), model.RolePolicija, officer.Role)
	}
}

func (suite *UserCrudServiceTestSuite) TestSearchUsersByName_Found() {
	suite.seedUser("john.search@example.com", model.RoleOsoba, "30000000001", "John")
	suite.seedUser("jane.search@example.com", model.RoleFirma, "30000000002", "Jane")
	suite.seedUser("jonathan.search@example.com", model.RoleHAK, "30000000003", "Jonathan")

	// Test exact match
	users, err := suite.userCrudService.SearchUsersByName("John")
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), users)
	foundJohn := false
	for _, u := range users {
		if u.FirstName == "John" {
			foundJohn = true
			break
		}
	}
	assert.True(suite.T(), foundJohn, "Should find John")

	// Test partial match (Jaro-Winkler should pick up Jonathan for Jon)
	users, err = suite.userCrudService.SearchUsersByName("Jon")
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), users)
	foundJonathanOrJohn := false
	for _, u := range users {
		if strings.HasPrefix(u.FirstName, "Jon") {
			foundJonathanOrJohn = true
			break
		}
	}
	assert.True(suite.T(), foundJonathanOrJohn, "Should find Jonathan or John for query 'Jon'")
}

func (suite *UserCrudServiceTestSuite) TestSearchUsersByName_NotFound() {
	suite.seedUser("no.match@example.com", model.RoleOsoba, "30000000004", "Zzz")
	users, err := suite.userCrudService.SearchUsersByName("NonExistentName")
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), users)
}

func (suite *UserCrudServiceTestSuite) TestGetUserByOIB_Success() {
	oib := "40000000001"
	seededUser := suite.seedUser("oib.user@example.com", model.RoleOsoba, oib, "seed")

	user, err := suite.userCrudService.GetUserByOIB(oib)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), seededUser.ID, user.ID)
	assert.Equal(suite.T(), oib, user.OIB)
}

func (suite *UserCrudServiceTestSuite) TestGetUserByOIB_NotFound() {
	oib := "00000000000" // Non-existent
	user, err := suite.userCrudService.GetUserByOIB(oib)
	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, gorm.ErrRecordNotFound))
	assert.Nil(suite.T(), user)
}

func (suite *UserCrudServiceTestSuite) TestCreateUser_InvalidRoleInModel() {
	invalidRoleUser := &model.User{
		Uuid:         uuid.New(),
		FirstName:    "Bad",
		LastName:     "RoleUser",
		OIB:          "50000000001",
		Email:        "bad.role.model@example.com",
		Role:         model.UserRole("nonexistentrole"), // Invalid role string
		BirthDate:    time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Residence:    "123 Error St",
		PasswordHash: "somehash",
	}

	err := suite.db.Create(invalidRoleUser).Error
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid user role")

	dtoWithBadRole := dto.NewUserDto{
		FirstName: "DtoBad", LastName: "RoleDto", OIB: "50000000002",
		Email: "bad.role.dto@example.com", Role: "verybadrole", Password: "password",
		BirthDate: "1990-01-01", Residence: "Dto Error St",
	}
	_, errDto := dtoWithBadRole.ToModel() // This calls model.StoUserRole
	assert.Error(suite.T(), errDto)
	assert.True(suite.T(), errors.Is(errDto, cerror.ErrUnknownRole))
}

func (suite *UserCrudServiceTestSuite) TestCreateUser_PoliceRole_WithToken() {
	policeTokenValue := "POLICETOKEN123"
	newUser := &model.User{
		Uuid:        uuid.New(),
		FirstName:   "Officer",
		LastName:    "WithToken",
		OIB:         "POLICE00001",
		Email:       "officer.withtoken@example.com",
		Role:        model.RolePolicija,
		BirthDate:   time.Date(1988, 8, 8, 0, 0, 0, 0, time.UTC),
		Residence:   "Police Station 1",
		PoliceToken: &policeTokenValue, // Explicitly set token for police role
	}
	plainPassword := "officerPass"

	createdUser, err := suite.userCrudService.Create(newUser, plainPassword)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), createdUser)
	assert.Equal(suite.T(), model.RolePolicija, createdUser.Role)
	assert.NotNil(suite.T(), createdUser.PoliceToken, "PoliceToken should be set for police officer")
	if createdUser.PoliceToken != nil {
		assert.Equal(suite.T(), policeTokenValue, *createdUser.PoliceToken)
	}

	// Verify in DB
	var dbUser model.User
	errDb := suite.db.First(&dbUser, createdUser.ID).Error
	assert.NoError(suite.T(), errDb)
	assert.Equal(suite.T(), model.RolePolicija, dbUser.Role)
	assert.NotNil(suite.T(), dbUser.PoliceToken)
	if dbUser.PoliceToken != nil {
		assert.Equal(suite.T(), policeTokenValue, *dbUser.PoliceToken)
	}
}

func (suite *UserCrudServiceTestSuite) TestCreateUser_PoliceRole_WithoutTokenInModel() {
	// Test case where the PoliceToken field in the input model.User is nil,
	// but the role is policija. The service should not automatically generate one here;
	// token generation/setting is a separate admin action via UserController.
	newUser := &model.User{
		Uuid:        uuid.New(),
		FirstName:   "Officer",
		LastName:    "NoTokenInModel",
		OIB:         "POLICE00002",
		Email:       "officer.notokenmodel@example.com",
		Role:        model.RolePolicija,
		BirthDate:   time.Date(1989, 9, 9, 0, 0, 0, 0, time.UTC),
		Residence:   "Police Station 2",
		PoliceToken: nil, // Token is nil in the model being passed to Create
	}
	plainPassword := "officerPassNoToken"

	createdUser, err := suite.userCrudService.Create(newUser, plainPassword)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), createdUser)
	assert.Equal(suite.T(), model.RolePolicija, createdUser.Role)
	// If the input DTO (and thus the model passed to Create) had no token,
	// the created user's PoliceToken should also be nil or empty string after creation.
	// The service's Create method will set it to nil if the role is not policija,
	// or use the provided value if it is policija. If nil is provided for policija, it stays nil.
	assert.Nil(suite.T(), createdUser.PoliceToken, "PoliceToken should be nil if not provided for police officer during creation")

	// Verify in DB
	var dbUser model.User
	errDb := suite.db.First(&dbUser, createdUser.ID).Error
	assert.NoError(suite.T(), errDb)
	assert.Equal(suite.T(), model.RolePolicija, dbUser.Role)
	assert.Nil(suite.T(), dbUser.PoliceToken)
}

func (suite *UserCrudServiceTestSuite) TestCreateUser_NonPoliceRole_IgnoresPoliceTokenField() {
	// If a non-police role user is created, any value in PoliceToken field of the input model
	// should be ignored and set to nil by the service.
	ignoredTokenValue := "SHOULD BE IGNORED"
	newUser := &model.User{
		Uuid:        uuid.New(),
		FirstName:   "Civilian",
		LastName:    "WithSuperfluousToken",
		OIB:         "CIVIL00001",
		Email:       "civilian.withtoken@example.com",
		Role:        model.RoleOsoba, // Non-police role
		BirthDate:   time.Date(1991, 1, 1, 0, 0, 0, 0, time.UTC),
		Residence:   "Civilian Residence",
		PoliceToken: &ignoredTokenValue, // Provide a token, but it should be ignored
	}
	plainPassword := "civilianPass"

	createdUser, err := suite.userCrudService.Create(newUser, plainPassword)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), createdUser)
	assert.Equal(suite.T(), model.RoleOsoba, createdUser.Role)
	assert.Nil(suite.T(), createdUser.PoliceToken, "PoliceToken should be nil for non-police officer, even if provided")

	// Verify in DB
	var dbUser model.User
	errDb := suite.db.First(&dbUser, createdUser.ID).Error
	assert.NoError(suite.T(), errDb)
	assert.Equal(suite.T(), model.RoleOsoba, dbUser.Role)
	assert.Nil(suite.T(), dbUser.PoliceToken)
}
