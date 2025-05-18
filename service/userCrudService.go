package service

import (
	"ePrometna_Server/app"
	"ePrometna_Server/model"
	"ePrometna_Server/util/auth"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/xrash/smetrics"

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
	GetAllUsers() ([]model.User, error)
	GetAllPoliceOfficers() ([]model.User, error)
	SearchUsersByName(query string) ([]model.User, error)
	GetUserByOIB(oib string) (*model.User, error)
}

type UserCrudService struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

type UserWithScore struct {
	User  model.User
	Score float64
}

func NewUserCrudService() IUserCrudService {
	var service IUserCrudService
	app.Invoke(func(db *gorm.DB, logger *zap.SugaredLogger) {
		service = &UserCrudService{
			db:     db,
			logger: logger,
		}
	})

	return service
}

// ReadAll implements IUserCrudService.
func (u *UserCrudService) ReadAll() ([]model.User, error) {
	var users []model.User
	rez := u.db.Find(&users)
	if rez.Error != nil {
		return nil, rez.Error
	}
	return users, nil
}

// Delete implements IUserCrudService.
func (u *UserCrudService) Delete(_uuid uuid.UUID) error {
	var user model.User
	rez := u.db.Preload("License").Where("uuid = ?", _uuid).First(&user)
	if rez.Error != nil {
		if rez.RowsAffected == 0 {
			u.logger.Debugf("User with UUID %s not found", _uuid)
			return gorm.ErrRecordNotFound
		}
		u.logger.Errorf("Error finding user with UUID %s: %v", _uuid, rez.Error)
		return rez.Error
	}

	if user.License != nil {
		u.logger.Debugf("Deleting driver license with UUID %s", user.License.Uuid)
		if err := u.db.Where("uuid = ?", user.License.Uuid).Delete(&user.License).Error; err != nil {
			u.logger.Errorf("Error deleting driver license with UUID %s: %v", user.License.Uuid, err)
			return err
		}
	}

	user.FirstName = "Deleted"
	user.LastName = "User"
	user.OIB = fmt.Sprintf("000000%05d", user.ID)
	user.BirthDate = time.Time{}
	user.Residence = "Anonymized"
	user.Email = fmt.Sprintf("deleted_%s@example.com", _uuid.String())
	user.PasswordHash = ""

	saveRez := u.db.Save(&user)
	if saveRez.Error != nil {
		u.logger.Errorf("Error saving anonymized user with UUID %s: %v", _uuid, saveRez.Error)
		return saveRez.Error
	}

	u.logger.Debugf("User with UUID %s anonymized successfully", _uuid)
	return nil
}

// Read implements IUserCrudService.
func (u *UserCrudService) Read(_uuid uuid.UUID) (*model.User, error) {
	var user model.User
	rez := u.db.
		Where("uuid = ?", _uuid).
		First(&user)
	if rez.Error != nil {
		return nil, rez.Error
	}

	return &user, nil
}

// Update implements IUserCrudService.
func (u *UserCrudService) Update(_uuid uuid.UUID, user *model.User) (*model.User, error) {
	userOld, err := u.Read(_uuid)
	if err != nil {
		return nil, err
	}

	u.logger.Debugf("Updating user %+v", userOld)
	userOld = userOld.Update(user)

	rez := u.db.
		Where("uuid = ?", _uuid).
		Save(userOld)

	if rez.Error != nil {
		return nil, rez.Error
	}
	return userOld, nil
}

// Create implements IUserCrudService.
func (u *UserCrudService) Create(user *model.User, password string) (*model.User, error) {
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

// Gets all users except super admin
func (u *UserCrudService) GetAllUsers() ([]model.User, error) {
	var users []model.User
	rez := u.db.
		Where("role != ?", model.RoleSuperAdmin).
		Find(&users)
	if rez.Error != nil {
		return nil, rez.Error
	}
	return users, nil
}

// Gets all police officers users
func (u *UserCrudService) GetAllPoliceOfficers() ([]model.User, error) {
	var users []model.User
	rez := u.db.
		Where("role = ?", model.RolePolicija).
		Find(&users)
	if rez.Error != nil {
		return nil, rez.Error
	}
	return users, nil
}

// SearchUsersByName searches for users by name and surname
func (u *UserCrudService) SearchUsersByName(query string) ([]model.User, error) {
	normalizedQuery := strings.ToLower(strings.TrimSpace(query))

	var users []model.User
	err := u.db.Find(&users).Error
	if err != nil {
		return nil, err
	}

	var scoredUsers []UserWithScore
	for _, user := range users {
		fullName := strings.ToLower(user.FirstName + " " + user.LastName)
		score := smetrics.JaroWinkler(normalizedQuery, fullName, 0.7, 4)
		scoredUsers = append(scoredUsers, UserWithScore{User: user, Score: score})
	}

	sort.Slice(scoredUsers, func(i, j int) bool {
		return scoredUsers[i].Score > scoredUsers[j].Score
	})

	var filteredUsers []model.User
	for _, scoredUser := range scoredUsers {
		if scoredUser.Score >= 0.8 {
			filteredUsers = append(filteredUsers, scoredUser.User)
		}
	}

	return filteredUsers, nil
}

// GetUserByOIB implements IUserCrudService.
func (u *UserCrudService) GetUserByOIB(oib string) (*model.User, error) {
	var user model.User
	rez := u.db.Where("oib = ?", oib).First(&user)
	if rez.Error != nil {
		return nil, rez.Error
	}
	return &user, nil
}
