package controller

import (
	"ePrometna_Server/app"
	"ePrometna_Server/dto"
	"ePrometna_Server/service"
	"ePrometna_Server/util/auth"
	"ePrometna_Server/util/cerror"
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
	app.Invoke(func(vehicleService service.IVehicleService, logger *zap.SugaredLogger) {
		// create controller
		controller = &VehicleController{
			VehicleService: vehicleService,
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
	group.GET("/:uuid", c.get)
	group.GET("/", c.myVehicles)
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

	c.AbortWithStatus(http.StatusNoContent)
}

// CreateVehicle godoc
//
//	@Summary	Creates new vehicle
//	@Schemes
//	@Description	Create new vehicle with an owner
//	@Tags			vehicle
//	@Produce		json
//	@Success		201	{object}	dto.VehicleDto
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
		return
	}

	vehicle, err := newDto.ToModel()
	if err != nil {
		v.logger.Errorf("Failed to create model error = %+v", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	ownerUuid, err := uuid.Parse(newDto.OwnerUuid)
	if err != nil {
		v.logger.Errorf("Failed to parse uuid = %s, err + %+v", newDto.OwnerUuid, err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	vehicle, err = v.VehicleService.Create(vehicle, ownerUuid)
	if err != nil {
		if errors.Is(err, cerror.ErrBadRole) {
			v.logger.Errorf("Role or user is invalid, err = %+v", err)
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			v.logger.Errorf("User with uuid = %s not found", newDto.OwnerUuid)
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var dto dto.VehicleDto
	c.JSON(http.StatusCreated, dto.FromModel(vehicle))
}

// GetVehicle godoc
//
//	@Summary	Gets a vehicle with uuid
//	@Schemes
//	@Tags		vehicle
//	@Produce	json
//	@Success	200	{object}	dto.VehicleDetailsDto
//	@Failure	400
//	@Failure	404
//	@Failure	500
//	@Param		uuid	path	string	true	"Vehicle UUID"
//	@Router		/vehicle/{uuid} [get]
func (v *VehicleController) get(c *gin.Context) {
	vehicleUuid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		v.logger.Errorf("error parsing uuid value = %s", c.Param("uuid"))
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	vehicle, err := v.VehicleService.Read(vehicleUuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			v.logger.Errorf("Vehicle with uuid = %s not found", vehicleUuid)
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var dto dto.VehicleDetailsDto
	// var dto dto.VehicleDto
	c.JSON(http.StatusOK, dto.FromModel(vehicle))
}

// myVehicle godoc
//
//	@Summary	Gets your vehicles
//	@Schemes
//	@Tags		vehicle
//	@Produce	json
//	@Success	200	{object}	[]dto.VehicleDto
//	@Failure	400
//	@Failure	404
//	@Failure	500
//	@Router		/vehicle [get]
func (v *VehicleController) myVehicles(c *gin.Context) {
	_, claims, err := auth.ParseToken(c.Request.Header.Get("Authorization"))
	if err != nil {
		v.logger.Errorf("Failed to parse token: %v", err)
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}
	uuid, err := uuid.Parse(claims.Uuid)
	if err != nil {
		v.logger.Errorf("Failed to parse uuid = %s, err + %+v", claims.Uuid, err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	vehicles, err := v.VehicleService.ReadAll(uuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			v.logger.Errorf("Vehicle with owner uuid = %s not found", uuid)
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	var dtos dto.VehiclesDto

	c.JSON(http.StatusOK, dtos.FromModel(vehicles))
}
