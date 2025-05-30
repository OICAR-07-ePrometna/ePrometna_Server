package controller_test

import (
	"bytes"
	"ePrometna_Server/app"
	"ePrometna_Server/config"
	"ePrometna_Server/controller"
	"ePrometna_Server/dto"
	"ePrometna_Server/model"
	"ePrometna_Server/service"
	"ePrometna_Server/util/auth"
	"ePrometna_Server/util/cerror"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

// --- Mock DriverLicenseCrudService ---
type MockDriverLicenseCrudService struct {
	mock.Mock
}

func (m *MockDriverLicenseCrudService) Create(license *model.DriverLicense, ownerUuid uuid.UUID) (*model.DriverLicense, error) {
	args := m.Called(license, ownerUuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.DriverLicense), args.Error(1)
}

func (m *MockDriverLicenseCrudService) GetByUuid(id uuid.UUID) (*model.DriverLicense, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.DriverLicense), args.Error(1)
}

func (m *MockDriverLicenseCrudService) GetAll() ([]model.DriverLicense, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.DriverLicense), args.Error(1)
}

func (m *MockDriverLicenseCrudService) Update(id uuid.UUID, updated *model.DriverLicense) (*model.DriverLicense, error) {
	args := m.Called(id, updated)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.DriverLicense), args.Error(1)
}

func (m *MockDriverLicenseCrudService) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// --- LicenseController Test Suite ---
type LicenseControllerTestSuite struct {
	suite.Suite
	router             *gin.Engine
	mockLicenseService *MockDriverLicenseCrudService
	sugar              *zap.SugaredLogger
}

// SetupSuite runs once before all tests
func (suite *LicenseControllerTestSuite) SetupSuite() {
	loggerCfg := zap.NewDevelopmentConfig()
	loggerCfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	zapLogger, _ := loggerCfg.Build()
	suite.sugar = zapLogger.Sugar()
	zap.ReplaceGlobals(zapLogger)

	gin.SetMode(gin.TestMode)

	config.AppConfig = &config.AppConfiguration{
		IsDevelopment: true,
		AccessKey:     "license-ctrl-test-access-key",
		RefreshKey:    "license-ctrl-test-refresh-key",
	}

	suite.mockLicenseService = new(MockDriverLicenseCrudService)

	app.Test()
	app.Provide(func() *zap.SugaredLogger { return suite.sugar })
	app.Provide(func() service.IDriverLicenseCrudService { return suite.mockLicenseService })

	suite.router = gin.Default()
	apiGroup := suite.router.Group("/api")

	licenseCtrl := controller.NewLicenseController()
	licenseCtrl.RegisterEndpoints(apiGroup)
}

// TearDownSuite runs once after all tests
func (suite *LicenseControllerTestSuite) TearDownSuite() {
	if suite.sugar != nil {
		_ = suite.sugar.Sync()
	}
}

// SetupTest runs before each test
func (suite *LicenseControllerTestSuite) SetupTest() {
	suite.mockLicenseService.ExpectedCalls = nil
	suite.mockLicenseService.Calls = nil
}

// Helper to generate a token for a test user (can be shared or specific)
func generateLicenseTestToken(userID uuid.UUID, userEmail string, userRole model.UserRole) string {
	token, _, _ := auth.GenerateTokens(&model.User{
		Uuid:  userID,
		Email: userEmail,
		Role:  userRole,
	})
	return token
}

// TestLicenseController runs the test suite
func TestLicenseController(t *testing.T) {
	suite.Run(t, new(LicenseControllerTestSuite))
}

// --- Test Cases ---

func (suite *LicenseControllerTestSuite) TestCreateLicense_Success() {
	ownerUUID := uuid.New()
	token := generateLicenseTestToken(ownerUUID, "owner@example.com", model.RoleOsoba)

	licenseDto := dto.DriverLicenseDto{
		LicenseNumber: "DL12345", IssueDate: "2023-01-01", ExpiringDate: "2033-01-01", Category: "B",
	}
	expectedLicenseModel, errConv := licenseDto.ToModel()
	assert.NoError(suite.T(), errConv)

	// Mock service Create
	// The service will receive the ownerUUID from the token, not from the DTO.
	suite.mockLicenseService.On("Create", mock.MatchedBy(func(l *model.DriverLicense) bool {
		return l.LicenseNumber == licenseDto.LicenseNumber && l.Category == licenseDto.Category
	}), ownerUUID).Return(expectedLicenseModel, nil).Once()

	jsonValue, _ := json.Marshal(licenseDto)
	req, _ := http.NewRequest(http.MethodPost, "/api/license/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	var responseDto dto.DriverLicenseDto
	err := json.Unmarshal(w.Body.Bytes(), &responseDto)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedLicenseModel.Uuid.String(), responseDto.Uuid)
	assert.Equal(suite.T(), licenseDto.LicenseNumber, responseDto.LicenseNumber)
	suite.mockLicenseService.AssertExpectations(suite.T())
}

func (suite *LicenseControllerTestSuite) TestCreateLicense_BindingError() {
	token := generateLicenseTestToken(uuid.New(), "owner@example.com", model.RoleOsoba)
	req, _ := http.NewRequest(http.MethodPost, "/api/license/", strings.NewReader(`{"licenseNumber": "DL123",`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *LicenseControllerTestSuite) TestCreateLicense_ServiceError_OwnerNotFound() {
	ownerUUID := uuid.New()
	token := generateLicenseTestToken(ownerUUID, "nonexistent.owner@example.com", model.RoleOsoba)
	licenseDto := dto.DriverLicenseDto{LicenseNumber: "DLX01", Category: "X", IssueDate: "2023-01-01", ExpiringDate: "2033-01-01"}

	suite.mockLicenseService.On("Create", mock.AnythingOfType("*model.DriverLicense"), ownerUUID).
		Return(nil, gorm.ErrRecordNotFound).Once()

	jsonValue, _ := json.Marshal(licenseDto)
	req, _ := http.NewRequest(http.MethodPost, "/api/license/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	suite.mockLicenseService.AssertExpectations(suite.T())
}

func (suite *LicenseControllerTestSuite) TestCreateLicense_ServiceError_BadRole() {
	ownerUUID := uuid.New()
	token := generateLicenseTestToken(ownerUUID, "badrole.owner@example.com", model.RolePolicija) // Policija cannot own license
	licenseDto := dto.DriverLicenseDto{LicenseNumber: "DLX02", Category: "Y", IssueDate: "2023-01-01", ExpiringDate: "2033-01-01"}

	// Mock service to return cerror.ErrBadRole
	suite.mockLicenseService.On("Create", mock.AnythingOfType("*model.DriverLicense"), ownerUUID).
		Return(nil, cerror.ErrBadRole).Once()

	jsonValue, _ := json.Marshal(licenseDto)
	req, _ := http.NewRequest(http.MethodPost, "/api/license/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	suite.mockLicenseService.AssertExpectations(suite.T())
}

func (suite *LicenseControllerTestSuite) TestGetLicense_Success() {
	// Any authenticated user can try to get a license by UUID.
	// Authorization for *which* license can be fetched might be handled by service or here.
	token := generateLicenseTestToken(uuid.New(), "viewer@example.com", model.RoleOsoba)
	targetLicenseUUID := uuid.New()
	expectedLicense := &model.DriverLicense{
		Uuid: targetLicenseUUID, LicenseNumber: "GETDL001", Category: "A",
		IssueDate: time.Now().AddDate(-2, 0, 0), ExpiringDate: time.Now().AddDate(8, 0, 0),
	}
	suite.mockLicenseService.On("GetByUuid", targetLicenseUUID).Return(expectedLicense, nil).Once()

	req, _ := http.NewRequest(http.MethodGet, "/api/license/"+targetLicenseUUID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var responseDto dto.DriverLicenseDto
	err := json.Unmarshal(w.Body.Bytes(), &responseDto)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), targetLicenseUUID.String(), responseDto.Uuid)
	assert.Equal(suite.T(), expectedLicense.LicenseNumber, responseDto.LicenseNumber)
	suite.mockLicenseService.AssertExpectations(suite.T())
}

func (suite *LicenseControllerTestSuite) TestGetLicense_NotFound() {
	token := generateLicenseTestToken(uuid.New(), "viewer@example.com", model.RoleOsoba)
	targetLicenseUUID := uuid.New()
	suite.mockLicenseService.On("GetByUuid", targetLicenseUUID).Return(nil, gorm.ErrRecordNotFound).Once()

	req, _ := http.NewRequest(http.MethodGet, "/api/license/"+targetLicenseUUID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	suite.mockLicenseService.AssertExpectations(suite.T())
}

func (suite *LicenseControllerTestSuite) TestGetAllLicenses_Success() {
	// The controller extracts user UUID from token to fetch *their* licenses.
	userUUID := uuid.New()
	token := generateLicenseTestToken(userUUID, "lic.owner@example.com", model.RoleFirma)

	licenses := []model.DriverLicense{
		{Uuid: uuid.New(), LicenseNumber: "L1", Category: "B", IssueDate: time.Now(), ExpiringDate: time.Now().AddDate(5, 0, 0)},
		{Uuid: uuid.New(), LicenseNumber: "L2", Category: "C", IssueDate: time.Now(), ExpiringDate: time.Now().AddDate(3, 0, 0)},
	}
	suite.mockLicenseService.On("GetAll").Return(licenses, nil).Once()

	req, _ := http.NewRequest(http.MethodGet, "/api/license/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var responseDtos []dto.DriverLicenseDto
	err := json.Unmarshal(w.Body.Bytes(), &responseDtos)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), responseDtos, 2)
	assert.Equal(suite.T(), licenses[0].LicenseNumber, responseDtos[0].LicenseNumber)
	suite.mockLicenseService.AssertExpectations(suite.T())
}

func (suite *LicenseControllerTestSuite) TestUpdateLicense_Success() {
	userUUID := uuid.New()
	token := generateLicenseTestToken(userUUID, "updater@example.com", model.RoleOsoba)
	targetLicenseUUID := uuid.New()

	updateDto := dto.DriverLicenseDto{
		Uuid: targetLicenseUUID.String(), LicenseNumber: "UPDATED-DL", Category: "B,C",
		IssueDate: "2022-02-02", ExpiringDate: "2032-02-02",
	}
	updatedLicenseModel, _ := updateDto.ToModel()
	updatedLicenseModel.Uuid = targetLicenseUUID

	suite.mockLicenseService.On("Update", targetLicenseUUID, mock.MatchedBy(func(l *model.DriverLicense) bool {
		return l.LicenseNumber == updateDto.LicenseNumber && l.Category == updateDto.Category
	})).Return(updatedLicenseModel, nil).Once()

	jsonValue, _ := json.Marshal(updateDto)
	req, _ := http.NewRequest(http.MethodPut, "/api/license/"+targetLicenseUUID.String(), bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var responseDto dto.DriverLicenseDto
	err := json.Unmarshal(w.Body.Bytes(), &responseDto)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), updateDto.LicenseNumber, responseDto.LicenseNumber)
	suite.mockLicenseService.AssertExpectations(suite.T())
}

func (suite *LicenseControllerTestSuite) TestDeleteLicense_Success() {
	userUUID := uuid.New() // User performing the delete
	token := generateLicenseTestToken(userUUID, "deleter@example.com", model.RoleOsoba)
	targetLicenseUUID := uuid.New()

	suite.mockLicenseService.On("Delete", targetLicenseUUID).Return(nil).Once()

	req, _ := http.NewRequest(http.MethodDelete, "/api/license/"+targetLicenseUUID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)
	suite.mockLicenseService.AssertExpectations(suite.T())
}

func (suite *LicenseControllerTestSuite) TestDeleteLicense_NotFound() {
	userUUID := uuid.New()
	token := generateLicenseTestToken(userUUID, "deleter@example.com", model.RoleOsoba)
	targetLicenseUUID := uuid.New()

	suite.mockLicenseService.On("Delete", targetLicenseUUID).Return(gorm.ErrRecordNotFound).Once()

	req, _ := http.NewRequest(http.MethodDelete, "/api/license/"+targetLicenseUUID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	suite.mockLicenseService.AssertExpectations(suite.T())
}
