package controller

import (
	"ePrometna_Server/app"
	"ePrometna_Server/dto"
	"ePrometna_Server/model"
	"ePrometna_Server/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TestController struct {
	testService service.ITestService
}

func NewTestController() *TestController {
	var controller *TestController

	// Call dependency injection
	app.Invoke(func(testService service.ITestService) {
		// create controller
		controller = &TestController{
			testService: testService,
		}
	})

	return controller
}

func (c *TestController) RegisterEndpoints(api *gin.RouterGroup) {
	// create a group with the name of the router
	group := api.Group("/test")

	// register Endpoints
	group.GET("/", c.test)
	group.PUT("/", c.insert)
	group.POST("/", c.create)
	group.DELETE("/:uuid", c.delete)
}

// PingExample godoc
//
//	@Summary	ping example
//	@Schemes
//	@Description	do ping
//	@Tags			test
//	@Accept			json
//	@Produce		json
//	@Success		200
//	@Router			/test [get]
func (c *TestController) test(ctx *gin.Context) {
	tmodel, err := c.testService.ReadAll()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, tmodel)
}

// PingExample godoc
//
//	@Summary	Insert new test struct
//	@Schemes
//	@Description	do a insert into databse with test user and returns inserted struct
//	@Tags			test
//	@Accept			json
//	@Produce		json
//	@Success		200
//	@Router			/test [put]
func (c *TestController) insert(ctx *gin.Context) {
	t := model.Tmodel{Name: "Test insert", Uuid: uuid.New()}
	tmodel, err := c.testService.Create(&t)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, tmodel)
}

// DeleteExamle godoc
//
//	@Summary	Deletes test item
//	@Schemes
//	@Description	Deletes an item with uuid
//	@Tags			test
//	@Accept			json
//	@Produce		json
//	@Success		200
//	@Param			uuid	path	string	true	"Test model UUID"
//	@Router			/test/{uuid} [delete]
func (c *TestController) delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("uuid"))
	if err != nil {
		zap.S().Errorf("error parsing uuid value = %s", ctx.Param("uuid"))
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = c.testService.Delete(id)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

// CreateExample godoc
//
//	@Summary	Creates test item
//	@Schemes
//	@Description	Create a test model
//	@Tags			test
//	@Accept			json
//	@Produce		json
//	@Success		201
//	@Param			model	body	dto.TmodelDto	true	"Test model"
//	@Router			/test [post]
func (c *TestController) create(ctx *gin.Context) {
	// TODO: should use dto not Tmodel
	var md dto.TmodelDto
	if err := ctx.Bind(&md); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
	}

	cmod, err := c.testService.Create(md.ToModel())
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}

	// ctx.JSON(http.StatusOK, cmod)
	ctx.JSON(http.StatusCreated, md.FromModel(cmod))
}
