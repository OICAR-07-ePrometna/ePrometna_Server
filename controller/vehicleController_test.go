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
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

// --- Mock VehicleService ---
type MockVehicleService struct {
	mock.Mock
}

// Update implements service.IVehicleService.
func (m *MockVehicleService) Update(vehicleUuid uuid.UUID, newWehicle model.Vehicle) (*model.Vehicle, error) {
	args := m.Called(vehicleUuid, newWehicle)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Vehicle), args.Error(1)
}

func (m *MockVehicleService) Create(newVehicle *model.Vehicle, ownerUuid uuid.UUID) (*model.Vehicle, error) {
	args := m.Called(newVehicle, ownerUuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Vehicle), args.Error(1)
}

func (m *MockVehicleService) Read(id uuid.UUID) (*model.Vehicle, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Vehicle), args.Error(1)
}

func (m *MockVehicleService) ReadAll(driverUuid uuid.UUID) ([]model.Vehicle, error) {
	args := m.Called(driverUuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Vehicle), args.Error(1)
}

func (m *MockVehicleService) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockVehicleService) ChangeOwner(vehicle uuid.UUID, newOwner uuid.UUID) error {
	args := m.Called(vehicle, newOwner)
	return args.Error(0)
}

func (m *MockVehicleService) Registration(vehicleUuid uuid.UUID, regModel model.RegistrationInfo) error {
	args := m.Called(vehicleUuid, regModel)
	return args.Error(0)
}

func (m *MockVehicleService) ReadByVin(vin string) (*model.Vehicle, error) {
	args := m.Called(vin)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Vehicle), args.Error(1)
}

func (m *MockVehicleService) Deregister(vehicleUuid uuid.UUID) error {
	args := m.Called(vehicleUuid)
	return args.Error(0)
}

// --- Test Setup ---
var (
	testSugarLogger    *zap.SugaredLogger
	mockVehicleService *MockVehicleService
	testRouter         *gin.Engine
)

func setupTestEnvironment() {
	// Setup Zap logger
	loggerCfg := zap.NewDevelopmentConfig()
	loggerCfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	zapLogger, _ := loggerCfg.Build()
	testSugarLogger = zapLogger.Sugar()
	zap.ReplaceGlobals(zapLogger)

	gin.SetMode(gin.TestMode)

	// Setup config
	config.AppConfig = &config.AppConfiguration{
		Env:        config.Dev,
		AccessKey:  "controller-test-access-key",
		RefreshKey: "controller-test-refresh-key",
		Port:       8080,
	}

	mockVehicleService = new(MockVehicleService)

	// Setup DIG
	app.Test() // Initialize the container
	app.Provide(func() *zap.SugaredLogger { return testSugarLogger })
	app.Provide(func() service.IVehicleService { return mockVehicleService }) // Provide the mock

	// Create router and register controller
	testRouter = gin.Default()
	apiGroup := testRouter.Group("/api")

	vehicleCtrl := controller.NewVehicleController()
	vehicleCtrl.RegisterEndpoints(apiGroup)
}

func teardownTestEnvironment() {
	testSugarLogger.Sync()
	// Reset mocks for next test if running multiple TestX functions in one package
	mockVehicleService = new(MockVehicleService)
}

// Helper to generate a token for a test user
func generateTestToken(userID uuid.UUID, userEmail string, userRole model.UserRole) string {
	token, _, _ := auth.GenerateTokens(&model.User{
		Uuid:  userID,
		Email: userEmail,
		Role:  userRole,
	})
	return token
}

// TestMain runs before and after all tests in the package
func TestMain(m *testing.M) {
	setupTestEnvironment()
	exitVal := m.Run()
	teardownTestEnvironment()
	os.Exit(exitVal)
}

func TestCreateVehicle_Controller_Success(t *testing.T) {
	mockVehicleService.ExpectedCalls = nil
	mockVehicleService.Calls = nil

	ownerUUID := uuid.New()
	vehicleUUID := uuid.New()
	token := generateTestToken(uuid.New(), "hakuser@example.com", model.RoleHAK)

	newVehicleDto := dto.NewVehicleDto{
		OwnerUuid:    ownerUUID.String(),
		Registration: "ZG-CTRL-01",
		Summary: dto.VehicleSummary{
			Model:       "Controller Test S",
			VehicleType: "TestCar",
		},
	}

	expectedVehicleModel := &model.Vehicle{
		Uuid:         vehicleUUID,
		VehicleModel: newVehicleDto.Summary.Model,
		VehicleType:  newVehicleDto.Summary.VehicleType,
		Registration: &model.RegistrationInfo{Registration: newVehicleDto.Registration},
		UserId:       func(id uint) *uint { return &id }(1), // Dummy owner ID
	}

	// Mock the service call
	mockVehicleService.On("Create",
		mock.MatchedBy(func(v *model.Vehicle) bool {
			return v.VehicleModel == newVehicleDto.Summary.Model &&
				v.VehicleType == newVehicleDto.Summary.VehicleType &&
				v.Registration != nil &&
				v.Registration.Registration == newVehicleDto.Registration
		}), ownerUUID).Return(expectedVehicleModel, nil).Once()

	jsonValue, _ := json.Marshal(newVehicleDto)
	req, _ := http.NewRequest(http.MethodPost, "/api/vehicle/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var responseDto dto.VehicleDto
	err := json.Unmarshal(w.Body.Bytes(), &responseDto)
	assert.NoError(t, err)
	assert.Equal(t, vehicleUUID.String(), responseDto.Uuid)
	assert.Equal(t, newVehicleDto.Summary.Model, responseDto.Model)

	mockVehicleService.AssertExpectations(t)
}

func TestCreateVehicle_Controller_Unauthorized(t *testing.T) {
	newVehicleDto := dto.NewVehicleDto{OwnerUuid: uuid.New().String()}
	jsonValue, _ := json.Marshal(newVehicleDto)
	req, _ := http.NewRequest(http.MethodPost, "/api/vehicle/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateVehicle_Controller_Forbidden(t *testing.T) {
	token := generateTestToken(uuid.New(), "nonhakuser@example.com", model.RoleOsoba)
	newVehicleDto := dto.NewVehicleDto{OwnerUuid: uuid.New().String()}
	jsonValue, _ := json.Marshal(newVehicleDto)
	req, _ := http.NewRequest(http.MethodPost, "/api/vehicle/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreateVehicle_Controller_BindingError(t *testing.T) {
	token := generateTestToken(uuid.New(), "hakuser@example.com", model.RoleHAK)
	// Malformed JSON
	req, _ := http.NewRequest(http.MethodPost, "/api/vehicle/", strings.NewReader(`{"ownerUuid": "not-a-uuid", malformed`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateVehicle_Controller_OwnerUUIDParseError(t *testing.T) {
	token := generateTestToken(uuid.New(), "hakuser@example.com", model.RoleHAK)
	newVehicleDto := dto.NewVehicleDto{OwnerUuid: "not-a-valid-uuid"}
	jsonValue, _ := json.Marshal(newVehicleDto)
	req, _ := http.NewRequest(http.MethodPost, "/api/vehicle/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateVehicle_Controller_ServiceError_OwnerNotFound(t *testing.T) {
	mockVehicleService.ExpectedCalls = nil
	mockVehicleService.Calls = nil
	ownerUUID := uuid.New()
	token := generateTestToken(uuid.New(), "hakuser@example.com", model.RoleHAK)

	newVehicleDto := dto.NewVehicleDto{OwnerUuid: ownerUUID.String(), Summary: dto.VehicleSummary{Model: "X"}}
	mockVehicleService.On("Create", mock.AnythingOfType("*model.Vehicle"), ownerUUID).Return(nil, gorm.ErrRecordNotFound).Once()

	jsonValue, _ := json.Marshal(newVehicleDto)
	req, _ := http.NewRequest(http.MethodPost, "/api/vehicle/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockVehicleService.AssertExpectations(t)
}

func TestCreateVehicle_Controller_ServiceError_BadRole(t *testing.T) {
	mockVehicleService.ExpectedCalls = nil
	mockVehicleService.Calls = nil
	ownerUUID := uuid.New()
	token := generateTestToken(uuid.New(), "hakuser@example.com", model.RoleHAK)

	newVehicleDto := dto.NewVehicleDto{OwnerUuid: ownerUUID.String(), Summary: dto.VehicleSummary{Model: "Y"}}
	mockVehicleService.On("Create", mock.AnythingOfType("*model.Vehicle"), ownerUUID).Return(nil, cerror.ErrBadRole).Once()

	jsonValue, _ := json.Marshal(newVehicleDto)
	req, _ := http.NewRequest(http.MethodPost, "/api/vehicle/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockVehicleService.AssertExpectations(t)
}

func TestGetVehicle_Controller_Success(t *testing.T) {
	mockVehicleService.ExpectedCalls = nil
	mockVehicleService.Calls = nil
	vehicleUUID := uuid.New()
	tokenUserUUID := uuid.New()
	token := generateTestToken(tokenUserUUID, "testuser@example.com", model.RoleOsoba)

	expectedVehicle := &model.Vehicle{
		Uuid:           vehicleUUID,
		VehicleModel:   "Tesla Model Y",
		VehicleType:    "Car",
		Owner:          &model.User{Uuid: tokenUserUUID, FirstName: "Test"},
		Registration:   &model.RegistrationInfo{Registration: "ZG-GET-01"},
		RegistrationID: func(id uint) *uint { return &id }(1),
	}
	mockVehicleService.On("Read", vehicleUUID).Return(expectedVehicle, nil).Once()

	req, _ := http.NewRequest(http.MethodGet, "/api/vehicle/"+vehicleUUID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var respDto dto.VehicleDetailsDto
	err := json.Unmarshal(w.Body.Bytes(), &respDto)
	assert.NoError(t, err)
	assert.Equal(t, vehicleUUID.String(), respDto.Uuid)
	assert.Equal(t, expectedVehicle.VehicleModel, respDto.Summary.Model)
	assert.Equal(t, expectedVehicle.Registration.Registration, respDto.Registration)
	assert.Equal(t, tokenUserUUID.String(), respDto.Owner.Uuid)

	mockVehicleService.AssertExpectations(t)
}

func TestGetVehicle_Controller_NotFound(t *testing.T) {
	mockVehicleService.ExpectedCalls = nil
	mockVehicleService.Calls = nil
	vehicleUUID := uuid.New()
	token := generateTestToken(uuid.New(), "testuser@example.com", model.RoleOsoba)

	mockVehicleService.On("Read", vehicleUUID).Return(nil, gorm.ErrRecordNotFound).Once()

	req, _ := http.NewRequest(http.MethodGet, "/api/vehicle/"+vehicleUUID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockVehicleService.AssertExpectations(t)
}

func TestGetVehicle_Controller_InvalidUUID(t *testing.T) {
	token := generateTestToken(uuid.New(), "testuser@example.com", model.RoleOsoba)
	req, _ := http.NewRequest(http.MethodGet, "/api/vehicle/not-a-uuid", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMyVehicles_Controller_Success(t *testing.T) {
	mockVehicleService.ExpectedCalls = nil
	mockVehicleService.Calls = nil
	userUUID := uuid.New()
	token := generateTestToken(userUUID, "myvehicles@example.com", model.RoleFirma)

	v1UUID := uuid.New()
	expectedVehicles := []model.Vehicle{
		{Uuid: v1UUID, VehicleModel: "Civic", Registration: &model.RegistrationInfo{Registration: "ZG-MY-01"}},
		{Uuid: uuid.New(), VehicleModel: "Accord", Registration: &model.RegistrationInfo{Registration: "ZG-MY-02"}},
	}
	mockVehicleService.On("ReadAll", userUUID).Return(expectedVehicles, nil).Once()

	req, _ := http.NewRequest(http.MethodGet, "/api/vehicle/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var respDtos dto.VehiclesDto
	err := json.Unmarshal(w.Body.Bytes(), &respDtos)
	assert.NoError(t, err)
	assert.Len(t, respDtos, 2)
	assert.Equal(t, v1UUID.String(), respDtos[0].Uuid)
	assert.Equal(t, "ZG-MY-01", respDtos[0].Registration)

	mockVehicleService.AssertExpectations(t)
}

func TestMyVehicles_Controller_ServiceError(t *testing.T) {
	mockVehicleService.ExpectedCalls = nil
	mockVehicleService.Calls = nil
	userUUID := uuid.New()
	token := generateTestToken(userUUID, "myvehicles@example.com", model.RoleFirma)

	mockVehicleService.On("ReadAll", userUUID).Return(nil, errors.New("some internal service error")).Once()

	req, _ := http.NewRequest(http.MethodGet, "/api/vehicle/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockVehicleService.AssertExpectations(t)
}

func TestDeleteVehicle_Controller_Success(t *testing.T) {
	mockVehicleService.ExpectedCalls = nil
	mockVehicleService.Calls = nil
	vehicleUUID := uuid.New()
	token := generateTestToken(uuid.New(), "hakdeleter@example.com", model.RoleHAK)

	mockVehicleService.On("Delete", vehicleUUID).Return(nil).Once()

	req, _ := http.NewRequest(http.MethodDelete, "/api/vehicle/"+vehicleUUID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockVehicleService.AssertExpectations(t)
}

func TestDeleteVehicle_Controller_NotFound(t *testing.T) {
	mockVehicleService.ExpectedCalls = nil
	mockVehicleService.Calls = nil
	vehicleUUID := uuid.New()
	token := generateTestToken(uuid.New(), "hakdeleter@example.com", model.RoleHAK)

	mockVehicleService.On("Delete", vehicleUUID).Return(gorm.ErrRecordNotFound).Once()

	req, _ := http.NewRequest(http.MethodDelete, "/api/vehicle/"+vehicleUUID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockVehicleService.AssertExpectations(t)
}

func TestDeleteVehicle_Controller_Forbidden(t *testing.T) {
	vehicleUUID := uuid.New()
	token := generateTestToken(uuid.New(), "userdeleter@example.com", model.RoleOsoba)

	req, _ := http.NewRequest(http.MethodDelete, "/api/vehicle/"+vehicleUUID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestChangeOwner_Controller_Success(t *testing.T) {
	mockVehicleService.ExpectedCalls = nil
	mockVehicleService.Calls = nil
	vehicleUUID := uuid.New()
	newOwnerUUID := uuid.New()
	token := generateTestToken(uuid.New(), "hakchanger@example.com", model.RoleHAK)

	changeDto := dto.ChangeOwnerDto{
		VehicleUuid:  vehicleUUID.String(),
		NewOwnerUuid: newOwnerUUID.String(),
	}
	mockVehicleService.On("ChangeOwner", vehicleUUID, newOwnerUUID).Return(nil).Once()

	jsonValue, _ := json.Marshal(changeDto)
	req, _ := http.NewRequest(http.MethodPut, "/api/vehicle/change-owner", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockVehicleService.AssertExpectations(t)
}

func TestChangeOwner_Controller_ServiceError_NotFound(t *testing.T) {
	mockVehicleService.ExpectedCalls = nil
	mockVehicleService.Calls = nil
	vehicleUUID := uuid.New()
	newOwnerUUID := uuid.New()
	token := generateTestToken(uuid.New(), "hakchanger@example.com", model.RoleHAK)

	changeDto := dto.ChangeOwnerDto{
		VehicleUuid:  vehicleUUID.String(),
		NewOwnerUuid: newOwnerUUID.String(),
	}
	mockVehicleService.On("ChangeOwner", vehicleUUID, newOwnerUUID).Return(gorm.ErrRecordNotFound).Once()

	jsonValue, _ := json.Marshal(changeDto)
	req, _ := http.NewRequest(http.MethodPut, "/api/vehicle/change-owner", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockVehicleService.AssertExpectations(t)
}

func TestChangeOwner_Controller_BindingError(t *testing.T) {
	token := generateTestToken(uuid.New(), "hakuser@example.com", model.RoleHAK)
	req, _ := http.NewRequest(http.MethodPut, "/api/vehicle/change-owner", strings.NewReader(`{malformed`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegistration_Controller_Success(t *testing.T) {
	mockVehicleService.ExpectedCalls = nil
	mockVehicleService.Calls = nil
	vehicleUUID := uuid.New()
	token := generateTestToken(uuid.New(), "hakregistrar@example.com", model.RoleHAK)

	regDto := dto.RegistrationDto{
		PassTechnical:    true,
		TraveledDistance: 50000,
		Registration:     "ZG-REG-CTRL",
		Note:             "Controller test registration",
	}

	mockVehicleService.On("Registration", vehicleUUID, mock.MatchedBy(func(m model.RegistrationInfo) bool {
		return m.Registration == regDto.Registration && m.TraveledDistance == regDto.TraveledDistance
	})).Return(nil).Once()

	jsonValue, _ := json.Marshal(regDto)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/vehicle/registration/%s", vehicleUUID.String()), bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockVehicleService.AssertExpectations(t)
}

func TestRegistration_Controller_VehicleNotFound(t *testing.T) {
	mockVehicleService.ExpectedCalls = nil
	mockVehicleService.Calls = nil
	vehicleUUID := uuid.New()
	token := generateTestToken(uuid.New(), "hakregistrar@example.com", model.RoleHAK)

	regDto := dto.RegistrationDto{Registration: "ZG-REG-FAIL"}
	mockVehicleService.On("Registration", vehicleUUID, mock.AnythingOfType("model.RegistrationInfo")).Return(gorm.ErrRecordNotFound).Once()

	jsonValue, _ := json.Marshal(regDto)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/vehicle/registration/%s", vehicleUUID.String()), bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockVehicleService.AssertExpectations(t)
}

func TestRegistration_Controller_InvalidPathUUID(t *testing.T) {
	token := generateTestToken(uuid.New(), "hakregistrar@example.com", model.RoleHAK)
	regDto := dto.RegistrationDto{Registration: "ZG-REG-FAIL"}
	jsonValue, _ := json.Marshal(regDto)

	req, _ := http.NewRequest(http.MethodPut, "/api/vehicle/registration/not-a-valid-uuid", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegistration_Controller_BindingError(t *testing.T) {
	vehicleUUID := uuid.New()
	token := generateTestToken(uuid.New(), "hakregistrar@example.com", model.RoleHAK)

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/vehicle/registration/%s", vehicleUUID.String()), strings.NewReader(`{malformed`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func (suite *UserControllerTestSuite) TestDeregisterVehicle_Controller_Success() {
	mockVehicleService.ExpectedCalls = nil
	mockVehicleService.Calls = nil
	vehicleUUID := uuid.New()
	token := generateTestToken(uuid.New(), "hakderegistrar@example.com", model.RoleHAK)

	mockVehicleService.On("Deregister", vehicleUUID).Return(nil).Once()

	req, _ := http.NewRequest(http.MethodPut, "/api/vehicle/deregister/"+vehicleUUID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	mockVehicleService.AssertExpectations(suite.T())
}

func (suite *UserControllerTestSuite) TestDeregisterVehicle_Controller_NotFound() {
	mockVehicleService.ExpectedCalls = nil
	mockVehicleService.Calls = nil
	vehicleUUID := uuid.New()
	token := generateTestToken(uuid.New(), "hakderegistrar@example.com", model.RoleHAK)

	mockVehicleService.On("Deregister", vehicleUUID).Return(gorm.ErrRecordNotFound).Once()

	req, _ := http.NewRequest(http.MethodPut, "/api/vehicle/deregister/"+vehicleUUID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	mockVehicleService.AssertExpectations(suite.T())
}

func (suite *UserControllerTestSuite) TestDeregisterVehicle_Controller_Forbidden() {
	vehicleUUID := uuid.New()
	token := generateTestToken(uuid.New(), "userderegistrar@example.com", model.RoleOsoba)

	req, _ := http.NewRequest(http.MethodPut, "/api/vehicle/deregister/"+vehicleUUID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
	mockVehicleService.AssertNotCalled(suite.T(), "Deregister", mock.Anything)
}

func (suite *UserControllerTestSuite) TestGetVehicleByVin_Controller_Success() {
	mockVehicleService.ExpectedCalls = nil
	mockVehicleService.Calls = nil
	targetVIN := "TESTVIN1234567890"
	vehicleUUID := uuid.New()
	token := generateTestToken(uuid.New(), "vingetter@example.com", model.RoleHAK)

	expectedVehicle := &model.Vehicle{
		Uuid:          vehicleUUID,
		VehicleModel:  "VIN Test Model",
		ChassisNumber: targetVIN,
		Registration:  &model.RegistrationInfo{Registration: "ZG-VIN-01"},
		Owner:         &model.User{Uuid: uuid.New(), FirstName: "VIN Owner"},
	}
	mockVehicleService.On("ReadByVin", targetVIN).Return(expectedVehicle, nil).Once()

	req, _ := http.NewRequest(http.MethodGet, "/api/vehicle/vin/"+targetVIN, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var respDto dto.VehicleDetailsDto
	err := json.Unmarshal(w.Body.Bytes(), &respDto)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), vehicleUUID.String(), respDto.Uuid)
	assert.Equal(suite.T(), targetVIN, respDto.Summary.ChassisNumber)
	mockVehicleService.AssertExpectations(suite.T())
}

func (suite *UserControllerTestSuite) TestGetVehicleByVin_Controller_NotFound() {
	mockVehicleService.ExpectedCalls = nil
	mockVehicleService.Calls = nil
	targetVIN := "NONEXISTENTVIN00"
	token := generateTestToken(uuid.New(), "vingetter@example.com", model.RoleHAK)

	mockVehicleService.On("ReadByVin", targetVIN).Return(nil, gorm.ErrRecordNotFound).Once()

	req, _ := http.NewRequest(http.MethodGet, "/api/vehicle/vin/"+targetVIN, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	mockVehicleService.AssertExpectations(suite.T())
}

func (suite *UserControllerTestSuite) TestGetVehicleByVin_Controller_EmptyVIN() {
	token := generateTestToken(uuid.New(), "vingetter@example.com", model.RoleHAK)

	req, _ := http.NewRequest(http.MethodGet, "/api/vehicle/vin/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		suite.T().Log("Gin router returned 404 for empty VIN path segment, skipping controller logic check for empty VIN.")
	} else {
		assert.Equal(suite.T(), http.StatusMovedPermanently, w.Code)
		assert.Contains(suite.T(), w.Body.String(), "Moved Permanently")
	}
}

func (suite *UserControllerTestSuite) TestUpdateVehicle_Controller_Success() {
	mockVehicleService.ExpectedCalls = nil
	mockVehicleService.Calls = nil
	vehicleUUID := uuid.New()
	token := generateTestToken(uuid.New(), "hakupdater@example.com", model.RoleHAK)

	updateDto := dto.VehicleDetailsDto{
		Uuid: vehicleUUID.String(),
		Summary: dto.VehicleSummary{
			Model:       "Updated Model X",
			VehicleType: "UpdatedCar",
		},
	}
	expectedVehicleModel, _ := updateDto.ToModel()
	expectedVehicleModel.Uuid = vehicleUUID

	mockVehicleService.On("Update", vehicleUUID,
		mock.MatchedBy(func(v model.Vehicle) bool {
			return v.VehicleModel == updateDto.Summary.Model &&
				v.VehicleType == updateDto.Summary.VehicleType
		})).Return(expectedVehicleModel, nil).Once()

	jsonValue, _ := json.Marshal(updateDto)
	req, _ := http.NewRequest(http.MethodPut, "/api/vehicle/"+vehicleUUID.String(), bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	var responseDto dto.VehicleDto
	err := json.Unmarshal(w.Body.Bytes(), &responseDto)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), vehicleUUID.String(), responseDto.Uuid)
	assert.Equal(suite.T(), updateDto.Summary.Model, responseDto.Model)
	mockVehicleService.AssertExpectations(suite.T())
}

func (suite *UserControllerTestSuite) TestUpdateVehicle_Controller_NotFound() {
	mockVehicleService.ExpectedCalls = nil
	mockVehicleService.Calls = nil
	vehicleUUID := uuid.New()
	token := generateTestToken(uuid.New(), "hakupdater@example.com", model.RoleHAK)

	updateDto := dto.VehicleDetailsDto{Summary: dto.VehicleSummary{Model: "NonExistentUpdate"}}

	updateModel, _ := updateDto.ToModel()

	mockVehicleService.On("Update", vehicleUUID, *updateModel).Return(nil, gorm.ErrRecordNotFound).Once()

	jsonValue, _ := json.Marshal(updateDto)
	req, _ := http.NewRequest(http.MethodPut, "/api/vehicle/"+vehicleUUID.String(), bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	mockVehicleService.AssertExpectations(suite.T())
}

func (suite *UserControllerTestSuite) TestUpdateVehicle_Controller_BindingError() {
	vehicleUUID := uuid.New()
	token := generateTestToken(uuid.New(), "hakupdater@example.com", model.RoleHAK)

	// Malformed JSON
	req, _ := http.NewRequest(http.MethodPut, "/api/vehicle/"+vehicleUUID.String(), strings.NewReader(`{"summary": {"model":`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	mockVehicleService.AssertNotCalled(suite.T(), "Update", mock.Anything, mock.Anything)
}

func (suite *UserControllerTestSuite) TestUpdateVehicle_Controller_ServiceError_BadRole() {
	mockVehicleService.ExpectedCalls = nil
	mockVehicleService.Calls = nil
	vehicleUUID := uuid.New()
	token := generateTestToken(uuid.New(), "hakupdater@example.com", model.RoleHAK)

	updateDto := dto.VehicleDetailsDto{Summary: dto.VehicleSummary{Model: "BadRoleUpdate"}}
	updateModel, _ := updateDto.ToModel()

	mockVehicleService.On("Update", vehicleUUID, *updateModel).Return(nil, cerror.ErrBadRole).Once()

	jsonValue, _ := json.Marshal(updateDto)
	req, _ := http.NewRequest(http.MethodPut, "/api/vehicle/"+vehicleUUID.String(), bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	mockVehicleService.AssertExpectations(suite.T())
}
