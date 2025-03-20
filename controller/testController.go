package controller

import (
	"ePrometna_Server/app"
	"ePrometna_Server/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	group.POST("/", c.insert)
}

// PingExample godoc
// @Summary ping example
// @Schemes
// @Description do ping
// @Tags test
// @Accept json
// @Produce json
// @Success 200
// @Router /test/ [get]
func (c *TestController) test(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "Bokic")
}

// PingExample godoc
// @Summary Insert new test struct
// @Schemes
// @Description do a insert into databse with test user and returns inserted struct
// @Tags test
// @Accept json
// @Produce json
// @Success 200
// @Router /test/ [post]
func (c *TestController) insert(ctx *gin.Context) {
	t := model.Tmodel{Name: "Test insert", Uuid: uuid.New()}
	c.db.Create(&t)

	ctx.JSON(http.StatusOK, t)
}
