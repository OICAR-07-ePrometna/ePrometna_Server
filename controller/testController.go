package controller

import (
	"ePrometna_Server/app"
	"ePrometna_Server/model"
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

// @BasePath /api/test

func (c *TestController) RegisterEndpoints(api *gin.RouterGroup) {
	// create a group with the name of the router
	group := api.Group("/test")

	// register Endpoints
	group.GET("/", c.test)
	group.POST("/", c.insert)
}

// PingExample godoc
// @Summary ping example
// @Schemes
// @Description do ping
// @Tags example
// @Accept json
// @Produce json
// @Success 200
// @Router / [get]
func (c *TestController) test(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "Bokic")
}

func (c *TestController) insert(ctx *gin.Context) {
	t := model.Tmodel{Name: "Test insert"}
	c.db.Create(&t)

	ctx.JSON(http.StatusOK, t)
}
