package mock

import (
	"ePrometna_Server/model" // Adjust import path if necessary

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockUserCrudService is a mock type for IUserCrudService
type MockUserCrudService struct {
	mock.Mock
}

// Read mocks the Read method
func (m *MockUserCrudService) Read(id uuid.UUID) (*model.User, error) {
	args := m.Called(id)
	// Need to handle potential nil pointer if user is not found
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// --- Mock other methods if IUserCrudService defines them ---

func (m *MockUserCrudService) Create(user *model.User, password string) (*model.User, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserCrudService) Update(id uuid.UUID, user *model.User) (*model.User, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserCrudService) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserCrudService) ReadAll() ([]model.User, error) {
	args := m.Called()
	return args.Get(0).([]model.User), args.Error(1)
}

// Ensure MockUserCrudService satisfies the interface (compile-time check)
// Adjust the import path for service.IUserCrudService
// var _ service.IUserCrudService = (*MockUserCrudService)(nil)
