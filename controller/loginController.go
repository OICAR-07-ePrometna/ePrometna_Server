package controller

import (
	"ePrometna_Server/app"
	"ePrometna_Server/config"
	"ePrometna_Server/dto"
	"ePrometna_Server/model"
	"ePrometna_Server/service"
	"ePrometna_Server/util/auth"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type LoginController struct {
	loginService service.ILoginService
	logger       *zap.SugaredLogger
}

func NewLoginController() *LoginController {
	var controller *LoginController

	// Use the mock service for testing
	app.Invoke(func(loginService service.ILoginService, logger *zap.SugaredLogger) {
		// create controller
		controller = &LoginController{
			loginService: loginService,
			logger:       logger,
		}
	})

	return controller
}

func (c *LoginController) RegisterEndpoints(api *gin.RouterGroup) {
	// create a group with the name of the router
	group := api.Group("/auth")

	// register Endpoints
	group.POST("/login", c.login)
	group.POST("/refresh", c.RefreshToken)

	// Add mobile-specific endpoints
	group.POST("/login-mobile", c.loginMobile)
}

// Login godoc
//
//	@Summary		User login
//	@Description	Authenticates a user and returns access and refresh tokens
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			loginDto	body		dto.LoginDto	true	"Login credentials"
//	@Success		200			{object}	dto.TokenDto
//	@Router			/auth/login [post]
func (l *LoginController) login(c *gin.Context) {
	var loginDto dto.LoginDto

	if err := c.BindJSON(&loginDto); err != nil {
		l.logger.Errorf("Invalid login request err = %+v", err)
		return
	}

	accessToken, refreshToken, err := l.loginService.Login(loginDto.Email, loginDto.Password)
	if err != nil {
		l.logger.Errorf("Login failed err = %+v", err)
		c.JSON(http.StatusUnauthorized, err.Error())
		return
	}

	c.JSON(http.StatusOK, dto.TokenDto{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// LoginMobile godoc
//
//	@Summary		Mobile login with device registration
//	@Description	Authenticates a user on a mobile device and registers the device
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			mobileLoginDto	body		dto.MobileLoginDto	true	"Mobile login credentials"
//	@Success		200				{object}	dto.MobileLoginResponse
//	@Router			/auth/login-mobile [post]
func (l *LoginController) loginMobile(c *gin.Context) {
	var loginDto dto.MobileLoginDto

	if err := c.BindJSON(&loginDto); err != nil {
		l.logger.Errorf("Invalid mobile login request err = %+v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request format"})
		return
	}

	// Use the service to handle mobile login
	result, err := l.loginService.LoginMobile(
		loginDto.Email,
		loginDto.Password,
		loginDto.DeviceInfo,
	)
	if err != nil {
		l.logger.Errorf("Mobile login failed err = %+v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	// Return all tokens
	c.JSON(http.StatusOK, dto.MobileLoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		DeviceToken:  result.DeviceToken,
	})
}

// Refresh godoc
//
//	@Summary		Refresh Access Token
//	@Description	Generates a new access token using a valid refresh token
//	@Tags			auth
//	@Produce		json
//	@Param			refreshToken	body		string	true	"Refresh Token"
//	@Success		200				{object}	dto.TokenDto
//	@Router			/auth/refresh [post]
func (l *LoginController) RefreshToken(c *gin.Context) {
	var rToken dto.RefreshDto
	if err := c.BindJSON(&rToken); err != nil {
		l.logger.Errorf("Failed to bind refresh token JSON, err %+v", err)
		return
	}
	l.logger.Debugf("Parsed token from body token = %+v", rToken)

	var claims auth.Claims

	_, err := jwt.ParseWithClaims(rToken.RefreshToken, &claims, func(token *jwt.Token) (any, error) {
		return []byte(config.AppConfig.RefreshKey), nil
	})
	if err != nil {
		l.logger.Errorf("Error Parsing clames err = %+v", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	userUuid, err := uuid.Parse(claims.Uuid)
	if err != nil {
		l.logger.Errorf("Error Parsing uuid err = %+v", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	token, refreshNew, err := l.loginService.RefreshTokens(&model.User{
		Uuid:  userUuid,
		Email: claims.Email,
		Role:  claims.Role,
	})
	if err != nil {
		l.logger.Error("Refresh failed err = %+v", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, dto.TokenDto{
		AccessToken:  token,
		RefreshToken: refreshNew,
	})
}
