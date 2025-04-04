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
}

func NewLoginController() *LoginController {
	var controller *LoginController

	// Use the mock service for testing
	app.Invoke(func(loginService service.ILoginService) {
		// create controller
		controller = &LoginController{
			loginService: loginService,
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
}

// Login godoc
//
//	@Summary		User login
//	@Description	Authenticates a user and returns access and refresh tokens
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			loginDto	body	dto.LoginDto	true	"Login credentials"
//	@Success		200
//	@Router			/auth/login [post]
func (l *LoginController) login(c *gin.Context) {
	var loginDto dto.LoginDto

	if err := c.ShouldBindJSON(&loginDto); err != nil {
		zap.S().Error("Invalid login request err = %+v", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	accessToken, refreshToken, err := l.loginService.Login(loginDto.Email, loginDto.Password)
	if err != nil {
		zap.S().Errorf("Login failed err = %+v", err)
		c.JSON(http.StatusUnauthorized, err.Error())
		return
	}

	c.JSON(http.StatusOK, dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// Refresh godoc
//
//	@Summary		Refresh Access Token
//	@Description	Generates a new access token using a valid refresh token
//	@Tags			auth
//	@Produce		json
//	@Param			refreshToken	body	string	true	"Refresh Token"
//	@Success		200
//	@Router			/auth/refresh [post]
func (l *LoginController) RefreshToken(c *gin.Context) {
	var rToken dto.RefreshDto
	if err := c.ShouldBindJSON(&rToken); err != nil {
		zap.S().Errorf("Failed to bind refresh token JSON, err %+v", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	zap.S().Debugf("Parsed token from body token = %+v", rToken)

	var claims auth.Claims

	_, err := jwt.ParseWithClaims(rToken.RefreshToken, &claims, func(token *jwt.Token) (any, error) {
		return []byte(config.AppConfig.RefreshKey), nil
	})
	if err != nil {
		zap.S().Errorf("Error Parsing clames err = %+v", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	userUuid, err := uuid.Parse(claims.Uuid)
	if err != nil {
		zap.S().Errorf("Error Parsing uuid err = %+v", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	token, refreshNew, err := l.loginService.RefreshTokens(&model.User{
		Uuid:  userUuid,
		Email: claims.Email,
		Role:  claims.Role,
	})
	if err != nil {
		zap.S().Error("Refresh failed err = %+v", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, dto.TokenResponse{
		AccessToken:  token,
		RefreshToken: refreshNew,
	})
}
