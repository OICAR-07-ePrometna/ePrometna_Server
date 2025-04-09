package controller

import (
	"ePrometna_Server/app"
	"ePrometna_Server/dto"
	"ePrometna_Server/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type VehicleController struct {
	VehicleService service.IVehicleService
	logger         *zap.SugaredLogger
}

func NewVehicleController() *VehicleController {
	var controller *VehicleController

	// Call dependency injection
	app.Invoke(func(testService service.IVehicleService, logger *zap.SugaredLogger) {
		// create controller
		controller = &VehicleController{
			VehicleService: testService,
			logger:         logger,
		}
	})

	return controller
}

func (c *VehicleController) RegisterEndpoints(api *gin.RouterGroup) {
	// create a group with the name of the router
	group := api.Group("/vehicle")

	// TODO: uncomment after testing
	// group.Use(middleware.Protect(model.RoleHAK))

	// register Endpoints
	group.POST("/", c.create)
	group.DELETE("/:uuid", c.delete)
}

// DeleteVehicle godoc
//
//	@Summary	Soft delete on vehicle
//	@Schemes
//	@Description	Preforms a soft delete
//	@Tags			vehicle
//	@Success		204
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Param			uuid	path	string	true	"Vehicle UUID"
//	@Router			/vehicle/{uuid} [delete]
func (v *VehicleController) delete(c *gin.Context) {
	vehicleUuid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		v.logger.Errorf("error parsing uuid value = %s", c.Param("uuid"))
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = v.VehicleService.Delete(vehicleUuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			v.logger.Errorf("Vehicle with uuid = %s not found", vehicleUuid)
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, nil)
}

// CreateVehicle godoc
//
//	@Summary	Creates new vehicle
//	@Schemes
//	@Description	Create new vehicle with an owner
//	@Tags			vehicle
//	@Produce		json
//	@Success		201 {object} dto.VehicleDto
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Param			model	body	dto.NewVehicleDto	true	"Vehicle model"
//	@Router			/vehicle [post]
func (v *VehicleController) create(c *gin.Context) {
	var newDto dto.NewVehicleDto
	if err := c.Bind(&newDto); err != nil {
		v.logger.Errorf("Failed to bind error = %+v", err)
		c.AbortWithError(http.StatusBadRequest, err)
	}

	vehicle, err := newDto.ToModel()
	if err != nil {
		v.logger.Errorf("Failed to create model error = %+v", err)
		c.AbortWithError(http.StatusBadRequest, err)
	}

	vehicle, err = v.VehicleService.Create(vehicle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			v.logger.Errorf("User with uuid = %s not found", newDto.OwnerUuid)
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	var dto dto.VehicleDto
	c.JSON(http.StatusCreated, dto.FromModel(vehicle))
}
