package service

import (
	"ePrometna_Server/app"
	"ePrometna_Server/model"
	"ePrometna_Server/util/auth"
	"sort"
	"strings"

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
	rez := u.db.
		Where("uuid = ?", _uuid).
		Delete(&model.User{})

	u.logger.Debugf("Delete statment on uuid = %s, rez %+v", _uuid, rez)
	if rez.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return rez.Error
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

func (u *UserCrudService) Create(user *model.User, password string) (*model.User, error) {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = hash

	// Handle police token based on role
	if user.Role != model.RolePolicija {
		// For non-police users, ensure the token is nil
		user.PoliceToken = nil
	}

	// Debug logging
	tokenValue := "nil"
	if user.PoliceToken != nil {
		tokenValue = *user.PoliceToken
	}
	u.logger.Infof("Creating user: %s %s, role: %s, token: %s",
		user.FirstName, user.LastName, user.Role, tokenValue)

	// Create the user
	rez := u.db.Create(&user)
	if rez.Error != nil {
		return nil, rez.Error
	}

	// Verify the token was saved correctly for police officers
	if user.Role == model.RolePolicija && user.PoliceToken != nil && *user.PoliceToken != "" {
		var savedUser model.User
		if err := u.db.Where("uuid = ?", user.Uuid).First(&savedUser).Error; err != nil {
			u.logger.Errorf("Could not verify saved user: %v", err)
		} else {
			savedTokenValue := "nil"
			if savedUser.PoliceToken != nil {
				savedTokenValue = *savedUser.PoliceToken
			}
			u.logger.Infof("Saved user: %s %s, role: %s, token: %s",
				savedUser.FirstName, savedUser.LastName, savedUser.Role, savedTokenValue)

			// If token not saved correctly, try direct update
			if (savedUser.PoliceToken == nil || *savedUser.PoliceToken == "") &&
				user.PoliceToken != nil && *user.PoliceToken != "" {
				u.logger.Infof("Token not saved correctly, trying direct update")
				if err := u.db.Model(&savedUser).Update("police_token", user.PoliceToken).Error; err != nil {
					u.logger.Errorf("Failed to update token directly: %v", err)
				} else {
					u.logger.Infof("Token updated directly")
					// Re-fetch the user to confirm the update
					if err := u.db.Where("uuid = ?", user.Uuid).First(&savedUser).Error; err != nil {
						u.logger.Errorf("Could not fetch updated user: %v", err)
					} else {
						// Return the updated user
						return &savedUser, nil
					}
				}
			}
		}
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
