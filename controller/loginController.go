package controller

import (
	"ePrometna_Server/dto"
	"ePrometna_Server/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type LoginController struct {
	loginService service.ILoginService
}

// func NewLoginController() *LoginController {
// 	var controller *LoginController

// 	// Call dependency injection
// 	app.Invoke(func(loginService service.ILoginService) {
// 		// create controller
// 		controller = &LoginController{
// 			loginService: loginService,
// 		}
// 	})

// 	return controller
// }

func NewLoginController() *LoginController {
	var controller *LoginController

	// Use the mock service for testing
	mockService := service.NewMockLoginService()
	controller = &LoginController{
		loginService: mockService,
	}

	return controller
}

func (c *LoginController) RegisterEndpoints(api *gin.RouterGroup) {
	// create a group with the name of the router
	group := api.Group("")

	// register Endpoints
	group.POST("/login", c.login)
}

// Login godoc
//
//	@Summary		User login
//	@Description	Authenticates a user and returns access and refresh tokens
//	@Tags			login
//	@Accept			json
//	@Produce		json
//	@Param			loginDto	body	dto.LoginDto	true	"Login credentials"
//	@Success		200
//	@Router			/login [post]
func (c *LoginController) login(ctx *gin.Context) {
	var loginDto dto.LoginDto

	if err := ctx.ShouldBindJSON(&loginDto); err != nil {
		zap.S().Error("Invalid login request", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	accessToken, refreshToken, err := c.loginService.Login(loginDto.Email, loginDto.Password)
	if err != nil {
		zap.S().Error("Login failed", zap.Error(err))
		ctx.JSON(http.StatusUnauthorized, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})

}
