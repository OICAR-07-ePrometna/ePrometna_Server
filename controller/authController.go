package controller

import (
	"ePrometna_Server/utils"

	"github.com/gin-gonic/gin"
)

type AuthController struct{}

func NewAuthController() *AuthController {
	return &AuthController{}
}

func (a *AuthController) RegisterEndpoints(api *gin.RouterGroup) {
	group := api.Group("")

	group.POST("/refresh", a.RefreshToken)
}

// Refresh godoc
//
//	@Summary		Refresh Access Token
//	@Description	Generates a new access token using a valid refresh token
//	@Tags			auth
//	@Accept			x-www-form-urlencoded
//	@Produce		json
//	@Param			refresh_token	formData	string				true	"Refresh Token"
//	@Success		200				{object}	map[string]string	"access_token"
//	@Failure		401				{object}	map[string]string	"error"
//	@Router			/refresh [post]
func (a *AuthController) RefreshToken(c *gin.Context) {
	utils.HandleRefresh(c)
}
