package controller

import (
	"ePrometna_Server/app"
	"ePrometna_Server/dto"
	"ePrometna_Server/model"
	"ePrometna_Server/service"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TempDataController struct {
	TempDataService service.ITempDataService
	VehicleService  service.IVehicleService
	UserService     service.IUserCrudService
	logger          *zap.SugaredLogger
}

func NewTempDataController() *TempDataController {
	var controller *TempDataController
	app.Invoke(func(tempDataService service.ITempDataService, vehicleService service.IVehicleService, userService service.IUserCrudService, logger *zap.SugaredLogger) {
		controller = &TempDataController{
			TempDataService: tempDataService,
			VehicleService:  vehicleService,
			UserService:     userService,
			logger:          logger,
		}
	})
	return controller
}

func (c *TempDataController) RegisterEndpoints(api *gin.RouterGroup) {
	// create a group with the name of the router
	group := api.Group("/tempdata")

	// register Endpoints
	group.POST("/", c.createTempData)
	group.PUT("/:uuid", c.getAndDeleteTempData)
}

// createTempData godoc
// @Summary		Creates a new temporary data entry
// @Schemes
// @Description	Create a new temporary data entry with vehicle and user information
// @Tags			tempdata
// @Produce		json
// @Success		201	{object}	string
// @Failure		400
// @Failure		500
// @Router		/tempdata/ [post]
func (c *TempDataController) createTempData(ctx *gin.Context) {
	var tempData dto.TempData

	if err := ctx.ShouldBindJSON(&tempData); err != nil {
		c.logger.Errorf("Invalid request: %+v", err)
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	vehicleUUID, err := uuid.Parse(tempData.VehicleId)
	if err != nil {
		c.logger.Errorf("Invalid vehicleId: %s", tempData.VehicleId)
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	driverUUID, err := uuid.Parse(tempData.DriverId)
	if err != nil {
		c.logger.Errorf("Invalid driverId: %s", tempData.DriverId)
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	vehicle, err := c.VehicleService.Read(vehicleUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.logger.Errorf("Vehicle not found: %s", vehicleUUID)
			ctx.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.logger.Errorf("Failed to read vehicle: %+v", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	driver, err := c.UserService.Read(driverUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.logger.Errorf("Driver not found: %s", driverUUID)
			ctx.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.logger.Errorf("Failed to read driver: %+v", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newTempData := &model.TempData{
		Uuid:      uuid.New(),
		VehicleId: vehicle.ID,
		DriverId:  driver.ID,
		Expiring:  time.Now().Add(1 * time.Minute),
	}

	if err := c.TempDataService.CreateTempData(newTempData); err != nil {
		c.logger.Errorf("Failed to create temp data: %+v", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, newTempData.Uuid)
}

// getAndDeleteTempData godoc
// @Summary		Retrieves and deletes temporary data by UUID
// @Schemes
// @Description	Retrieve temporary data by UUID and delete it
// @Tags			tempdata
// @Produce		json
// @Param		uuid	path	string	true	"UUID of the temporary data"
// @Success		200	{object}	model.TempData
// @Failure		400
// @Failure		404
// @Failure		500
// @Router		/tempdata/{uuid} [put]
func (c *TempDataController) getAndDeleteTempData(ctx *gin.Context) {
	uuidStr := ctx.Param("uuid")
	if uuidStr == "" {
		c.logger.Error("UUID is required")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, "UUID is required")
		return
	}

	_, err := uuid.Parse(uuidStr)
	if err != nil {
		c.logger.Errorf("Invalid UUID: %s", uuidStr)
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	tempData, err := c.TempDataService.GetAndDeleteByUUID(uuidStr)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.logger.Errorf("Temp data not found: %s", uuidStr)
			ctx.AbortWithStatusJSON(http.StatusNotFound, "Temp data not found")
			return
		}
		c.logger.Errorf("Failed to get/delete temp data: %+v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, "Failed to get/delete temp data")
		return
	}

	ctx.JSON(http.StatusOK, tempData)
}
