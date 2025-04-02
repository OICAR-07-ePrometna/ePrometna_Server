package service

import (
	"ePrometna_Server/app"
	"ePrometna_Server/model"
	"ePrometna_Server/util/auth"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IUserCrudService interface {
	Create(user *model.User, password string) (*model.User, error)
	Read(uuid uuid.UUID) (*model.User, error)
	ReadAll() ([]model.User, error)
	Update(uuid uuid.UUID, user *model.User) (*model.User, error)
	Delete(uuid uuid.UUID) error
}

type UserCrudSerrvice struct {
	db *gorm.DB
}

func NewUserCrudService() IUserCrudService {
	var service IUserCrudService
	app.Invoke(func(db *gorm.DB) {
		service = UserCrudSerrvice{
			db: db,
		}
	})

	return service
}

// ReadAll implements IUserCrudService.
func (u UserCrudSerrvice) ReadAll() ([]model.User, error) {
	var users []model.User
	rez := u.db.Find(&users)
	if rez.Error != nil {
		return nil, rez.Error
	}
	return users, nil
}

// Delete implements IUserCrudService.
func (u UserCrudSerrvice) Delete(_uuid uuid.UUID) error {
	rez := u.db.Delete(&model.Tmodel{}, "uuid = ?", _uuid)
	zap.S().Debugf("Delete statment on uuid = %s, rez %+v", _uuid, rez)
	if rez.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return rez.Error
}

// Read implements IUserCrudService.
func (u UserCrudSerrvice) Read(_uuid uuid.UUID) (*model.User, error) {
	var user model.User
	rez := u.db.Where("uuid = ?", _uuid).First(&user)
	if rez.Error != nil {
		return nil, rez.Error
	}
	return &user, nil
}

// Update implements IUserCrudService.
func (u UserCrudSerrvice) Update(_uuid uuid.UUID, user *model.User) (*model.User, error) {
	// TODO: preserve password
	rez := u.db.Where("uuid = ?", _uuid).Save(user)
	if rez.Error != nil {
		return nil, rez.Error
	}
	return user, nil
}

// Create implements IUserCrudService.
func (u UserCrudSerrvice) Create(user *model.User, password string) (*model.User, error) {
	// TODO: hash password
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = hash
	rez := u.db.Create(&user)
	if rez.Error != nil {
		return nil, rez.Error
	}
	return user, nil
}
