package controller

import (
	"ePrometna_Server/app"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TestController struct {
	db *gorm.DB
}

func NewTestController() *TestController {
	var controller *TestController

	// Call dependency injection
	app.Invoke(func(db *gorm.DB) {
		// create controller
		controller = &TestController{
			db: db,
		}
	})

	return controller
}

func (c *TestController) RegisterEndpoints(api *gin.RouterGroup) {
	// create a group with the name of the router
	group := api.Group("/test")

	// register Endpoints
	group.GET("/", c.test)
}

func (c *TestController) test(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "Bokic")
}
