package controller

import (
	"ePrometna_Server/app"
	"ePrometna_Server/dto"
	"ePrometna_Server/model"
	"ePrometna_Server/service"
	"ePrometna_Server/util/auth"
	"ePrometna_Server/util/middleware"
	"errors"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserController struct {
	UserCrud service.IUserCrudService
	logger   *zap.SugaredLogger
}

func NewUserController() *UserController {
	var controller *UserController

	// Call dependency injection
	app.Invoke(func(UserService service.IUserCrudService, logger *zap.SugaredLogger) {
		// create controller
		controller = &UserController{
			UserCrud: UserService,
			logger:   logger,
		}
	})

	return controller
}

func (u *UserController) RegisterEndpoints(api *gin.RouterGroup) {
	// create a group with the name of the router
	group := api.Group("/user")

	// Protected endpint
	group.GET("/my-data", middleware.Protect(), u.getLoggedInUser)

	// Mup admin endpiont TODO: see pagination or search and or bothe
	group.GET("/police-officers", middleware.Protect(model.RoleMupADMIN), u.getAllPoliceOfficers)

	// HAK can search users by OIB for vehicle registration
	group.GET("/oib/:oib", middleware.Protect(model.RoleHAK), u.getUserByOib)

	// Police token endpoints
	group.POST("/:uuid/generate-token", middleware.Protect(model.RoleSuperAdmin, model.RoleMupADMIN), u.generatePoliceToken)
	group.PATCH("/:uuid/police-token", middleware.Protect(model.RoleSuperAdmin, model.RoleMupADMIN), u.setPoliceToken)

	// register Endpoints
	group.Use(middleware.Protect(model.RoleSuperAdmin, model.RoleMupADMIN))
	group.POST("/", u.create)
	group.GET("/:uuid", u.get)
	group.PUT("/:uuid", u.update)
	group.DELETE("/:uuid", u.delete)
	group.GET("/all-users", u.getAllUsersForSuperAdmin)
	group.GET("/search", u.searchUsersByName)
}

// UserExample godoc
//
//	@Summary		get user with uuid
//	@Description	get a user with uuid
//	@Tags			user
//	@Produce		json
//	@Success		200	{object}	dto.UserDto
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Param			uuid	path	string	true	"user uuid"
//	@Router			/user/{uuid} [get]
func (u *UserController) get(c *gin.Context) {
	userUuid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		u.logger.Errorf("error parsing uuid value = %s", c.Param("uuid"))
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := u.UserCrud.Read(userUuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u.logger.Errorf("User with uuid = %s not found", userUuid)
			c.AbortWithError(http.StatusNotFound, err)
			return
		}

		u.logger.Errorf("Failed to get user with uuid = %s", userUuid)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	dto := dto.UserDto{}
	c.JSON(http.StatusOK, dto.FromModel(user))
}

// UserExample godoc
//
//	@Summary	Create new user
//	@Tags		user
//	@Produce	json
//	@Success	201	{object}	dto.UserDto
//	@Failure	400
//	@Failure	404
//	@Failure	500
//	@Param		model	body	dto.NewUserDto	true	"Data for new user"
//	@Router		/user [post]
func (u *UserController) create(c *gin.Context) {
	var dto dto.NewUserDto
	if err := c.BindJSON(&dto); err != nil {
		u.logger.Errorf("Failed to bind error = %+v", err)
		return
	}

	newUser, err := dto.ToModel()
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// Handle police token based on role
	if newUser.Role == model.RolePolicija {
		// For police officers, ensure token is set if provided
		if dto.PoliceToken != "" {
			tokenValue := dto.PoliceToken
			newUser.PoliceToken = &tokenValue
		}
	} else {
		// For non-police users, explicitly set to nil
		newUser.PoliceToken = nil
	}

	user, err := u.UserCrud.Create(newUser, dto.Password)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// For police officers, reload to ensure we have the latest state
	if user.Role == model.RolePolicija {
		refreshedUser, err := u.UserCrud.Read(user.Uuid)
		if err == nil {
			user = refreshedUser
		}
	}

	// Log the response for debugging
	responseDto := dto.FromModel(user)
	tokenValue := ""
	if user.PoliceToken != nil {
		tokenValue = *user.PoliceToken
	}
	u.logger.Infof("Response DTO: %+v", responseDto)
	u.logger.Infof("Response PoliceToken: %q", tokenValue)

	c.JSON(http.StatusCreated, responseDto)
}

// UserExample godoc
//
//	@Summary	Update user with new dat
//	@Tags		user
//	@Produce	json
//	@Success	200	{object}	dto.UserDto
//	@Failure	400
//	@Failure	404
//	@Failure	500
//	@Param		uuid	path	string		true	"uuid of user to be updated"
//	@Param		model	body	dto.UserDto	true	"Data for updating user"
//	@Router		/user/{uuid} [put]
func (u *UserController) update(c *gin.Context) {
	userUuid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		u.logger.Errorf("Error parsing UUID = %s", c.Param("uuid"))
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var dto dto.UserDto
	if err := c.BindJSON(&dto); err != nil {
		u.logger.Errorf("Failed to bind error = %+v", err)
		return
	}

	newUser, err := dto.ToModel()
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := u.UserCrud.Update(userUuid, newUser)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, dto.FromModel(user))
}

// UserExample  godoc
//
//	@Summary		delete user with uuid
//	@Description	delete a user with uuid
//	@Tags			user
//	@Produce		json
//	@Success		204
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Param			uuid	path	string	true	"user uuid"
//	@Router			/user/{uuid} [delete]
func (u *UserController) delete(c *gin.Context) {
	userUuid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		u.logger.Errorf("error parsing uuid value = %s", c.Param("uuid"))
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = u.UserCrud.Delete(userUuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u.logger.Errorf("User with uuid = %s not found", userUuid)
			c.AbortWithError(http.StatusNotFound, err)
			return
		}

		u.logger.Errorf("Failed to delete user with uuid = %s", userUuid)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetLoggedInUser godoc
//
//	@Summary		Get logged-in user data
//	@Description	Fetches the currently logged-in user's data based on the JWT token
//	@Tags			user
//	@Produce		json
//	@Success		200	{object}	dto.UserDto
//	@Failure		400
//	@Failure		401
//	@Failure		404
//	@Failure		500
//	@Router			/user/my-data [get]
func (u *UserController) getLoggedInUser(c *gin.Context) {
	_, claims, err := auth.ParseToken(c.Request.Header.Get("Authorization"))
	if err != nil {
		u.logger.Errorf("Failed to parse token: %v", err)
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	userUuid, err := uuid.Parse(claims.Uuid)
	if err != nil {
		u.logger.Errorf("Error parsing UUID = %s", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := u.UserCrud.Read(userUuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u.logger.Errorf("User with uuid = %s not found", userUuid)
			c.AbortWithError(http.StatusNotFound, err)
			return
		}

		u.logger.Errorf("Failed to fetch user with uuid = %s: %v", userUuid, err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	dto := dto.UserDto{}
	c.JSON(http.StatusOK, dto.FromModel(user))
}

// GetAllUsers godoc
//
//	@Summary		Get all users for superadmin
//	@Description	Fetches all users for superadmin
//	@Tags			user
//	@Produce		json
//	@Success		200	{array}	dto.UserDto
//	@Failure		401
//	@Failure		403
//	@Failure		500
//	@Router			/user/all-users [get]
func (u *UserController) getAllUsersForSuperAdmin(c *gin.Context) {
	_, claims, err := auth.ParseToken(c.Request.Header.Get("Authorization"))
	if err != nil {
		u.logger.Errorf("Failed to parse token: %v", err)
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	if claims.Role != model.RoleSuperAdmin {
		u.logger.Warnf("Unauthorized access attempt by user with role: %s", claims.Role)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	users, err := u.UserCrud.GetAllUsers()
	if err != nil {
		u.logger.Errorf("Failed to fetch users: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	userDtos := make([]dto.UserDto, 0, len(users))
	for _, user := range users {
		dto := dto.UserDto{}
		userDtos = append(userDtos, dto.FromModel(&user))
	}

	c.JSON(http.StatusOK, userDtos)
}

// GetAllPoliceOfficers godoc
//
//	@Summary		Get all police officers for MUP Admin
//	@Description	Fetches all police officers for MUP Admin
//	@Tags			user
//	@Produce		json
//	@Success		200	{array}	dto.UserDto
//	@Failure		401
//	@Failure		403
//	@Failure		500
//	@Router			/user/police-officers [get]
func (u *UserController) getAllPoliceOfficers(c *gin.Context) {
	_, claims, err := auth.ParseToken(c.Request.Header.Get("Authorization"))
	if err != nil {
		u.logger.Errorf("Failed to parse token: %v", err)
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	if claims.Role != model.RoleMupADMIN {
		u.logger.Warnf("Unauthorized access attempt by user with role: %s", claims.Role)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	users, err := u.UserCrud.GetAllPoliceOfficers()
	if err != nil {
		u.logger.Errorf("Failed to fetch police officers: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	userDtos := make([]dto.UserDto, 0, len(users))
	for _, user := range users {
		dto := dto.UserDto{}
		userDtos = append(userDtos, dto.FromModel(&user))
	}

	c.JSON(http.StatusOK, userDtos)
}

// SearchUsersByName godoc
//
//	@Summary		Search users by name
//	@Description	Performs a fuzzy search for users by first name, last name, or full name with similarity matching
//	@Tags			user
//	@Produce		json
//	@Param			query	query	string	true	"Search query"
//	@Success		200		{array}	dto.UserDto
//	@Failure		400
//	@Failure		500
//	@Router			/user/search [get]
func (u *UserController) searchUsersByName(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		u.logger.Warn("Search query is empty")
		c.JSON(http.StatusBadRequest, "Search query is required")
		return
	}

	u.logger.Infof("Searching users with query: %s", query)

	users, err := u.UserCrud.SearchUsersByName(query)
	if err != nil {
		u.logger.Errorf("Failed to search users: %v", err)
		c.JSON(http.StatusInternalServerError, "Failed to search users")
		return
	}

	userDtos := make([]dto.UserDto, 0, len(users))
	for _, user := range users {
		dto := dto.UserDto{}
		userDtos = append(userDtos, dto.FromModel(&user))
	}

	c.JSON(http.StatusOK, userDtos)
}

// GetUserByOIB godoc
//
//	@Summary		get user with oib
//	@Description	get a user with oib
//	@Tags			user
//	@Produce		json
//	@Success		200	{object}	dto.UserDto
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Param			oib	path	string	true	"user oib"
//	@Router			/user/oib/{oib} [get]
func (u *UserController) getUserByOib(c *gin.Context) {
	oib := c.Param("oib")
	if oib == "" {
		u.logger.Errorf("OIB parameter is empty")
		c.AbortWithError(http.StatusBadRequest, errors.New("OIB parameter is required"))
		return
	}

	user, err := u.UserCrud.GetUserByOIB(oib)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u.logger.Errorf("User with OIB = %s not found", oib)
			c.AbortWithError(http.StatusNotFound, err)
			return
		}

		u.logger.Errorf("Failed to get user with OIB = %s", oib)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	dto := dto.UserDto{}
	c.JSON(http.StatusOK, dto.FromModel(user))
}

// Backend - Add to controller/userController.go

// GeneratePoliceToken godoc
//
//	@Summary	Generate a new police token for a user
//	@Tags		user
//	@Produce	json
//	@Success	200
//	@Failure	400
//	@Failure	404
//	@Failure	500
//	@Param		uuid	path	string	true	"User UUID"
//	@Router		/user/{uuid}/generate-token [post]
func (u *UserController) generatePoliceToken(c *gin.Context) {
	userUuid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		u.logger.Errorf("Error parsing UUID = %s", c.Param("uuid"))
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := u.UserCrud.Read(userUuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u.logger.Errorf("User with UUID = %s not found", userUuid)
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Only allow generating police token for police officers
	if user.Role != model.RolePolicija {
		u.logger.Warnf("Attempted to generate police token for non-police user: %s", user.Role)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Police token can only be generated for police officers"})
		return
	}

	token := u.generateToken()
	user.PoliceToken = &token

	// Save the updated user
	_, err = u.UserCrud.Update(userUuid, user)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Return the generated token to the client
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// SetPoliceToken godoc
//
//	@Summary	Set police token for a user
//	@Tags		user
//	@Produce	json
//	@Success	200
//	@Failure	400
//	@Failure	404
//	@Failure	500
//	@Param		uuid	path	string	true	"User UUID"
//	@Param		model	body	object	true	"Police token"
//	@Router		/user/{uuid}/police-token [patch]
func (u *UserController) setPoliceToken(c *gin.Context) {
	userUuid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		u.logger.Errorf("Error parsing UUID = %s", c.Param("uuid"))
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var tokenRequest struct {
		PoliceToken string `json:"police_token" binding:"required"`
	}

	if err := c.BindJSON(&tokenRequest); err != nil {
		u.logger.Errorf("Failed to bind error = %+v", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := u.UserCrud.Read(userUuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u.logger.Errorf("User with UUID = %s not found", userUuid)
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Only allow setting police token for police officers
	if user.Role != model.RolePolicija {
		u.logger.Warnf("Attempted to set police token for non-police user: %s", user.Role)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Police token can only be set for police officers"})
		return
	}

	// Update user with token
	tokenValue := tokenRequest.PoliceToken
	user.PoliceToken = &tokenValue

	_, err = u.UserCrud.Update(userUuid, user)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

// Generate police token
func (u *UserController) generateToken() string {
	digits := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	token := ""
	for i := 0; i < 8; i++ {
		token += string(digits[rand.Intn(len(digits))])
	}
	return token
}
