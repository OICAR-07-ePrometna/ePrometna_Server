package controller

import (
	"ePrometna_Server/app"
	"ePrometna_Server/dto"
	"ePrometna_Server/model"
	"ePrometna_Server/service"
	"ePrometna_Server/util/auth"
	"ePrometna_Server/util/cerror"
	"ePrometna_Server/util/middleware"
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
	group.POST("/:uuid", middleware.Protect(model.RoleFirma, model.RoleOsoba), c.createTempData)
	group.PUT("/:uuid", middleware.Protect(model.RolePolicija), c.getAndDeleteTempData)
}

// createTempData godoc
//
//	@Summary	Creates a new temporary data entry
//	@Schemes
//	@Description	Create a new temporary data entry with vehicle and user information
//	@Tags			tempdata
//	@Param			uuid	path	string	true	"UUID of vehicle"
//	@Produce		json
//	@Success		201	{object}	string
//	@Failure		400
//	@Failure		500
//	@Router			/tempdata/{uuid} [post]
func (c *TempDataController) createTempData(ctx *gin.Context) {
	vehicleUuidStr := ctx.Param("uuid")
	if vehicleUuidStr == "" {
		c.logger.Error("Empty UUID provided")
		ctx.AbortWithError(http.StatusBadRequest, errors.New("vehicle UUID is required"))
		return
	}

	vehicleUuid, err := uuid.Parse(vehicleUuidStr)
	if err != nil {
		c.logger.Errorf("Error parsing UUID = %s", vehicleUuidStr)
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// Parse JWT and get driver UUID
	_, claims, err := auth.ParseToken(ctx.Request.Header.Get("Authorization"))
	if err != nil {
		c.logger.Errorf("Failed to parse token: %v", err)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, "Invalid token")
		return
	}

	driverUUID, err := uuid.Parse(claims.Uuid)
	if err != nil {
		c.logger.Errorf("Failed to parse uuid from token claims = %s, err = %+v", claims.Uuid, err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, "Invalid user UUID in token")
		return
	}

	vehicle, err := c.VehicleService.Read(vehicleUuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.logger.Errorf("vehicle not found: %s", vehicleUuid)
			ctx.AbortWithStatusJSON(http.StatusNotFound, "vehicle not found")
			return
		}
		c.logger.Errorf("Failed to read vehicle: %+v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, "Failed to read vehicle")
		return
	}

	driver, err := c.UserService.Read(driverUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.logger.Errorf("Driver not found: %s", driverUUID)
			ctx.AbortWithStatusJSON(http.StatusNotFound, "Driver not found")
			return
		}
		c.logger.Errorf("Failed to read driver: %+v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, "Failed to read driver")
		return
	}

	newTempData := &model.TempData{
		Uuid:      uuid.New(),
		VehicleId: vehicle.ID,
		DriverId:  driver.ID,
		Expiring:  time.Now().Add(5 * time.Minute),
	}

	if err := c.TempDataService.CreateTempData(newTempData); err != nil {
		c.logger.Errorf("Failed to create temp data: %+v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, "Failed to create temp data")
		return
	}

	ctx.JSON(http.StatusCreated, newTempData.Uuid)
}

// getAndDeleteTempData godoc
//
//	@Summary	Retrieves and deletes temporary data by UUID
//	@Schemes
//	@Description	Retrieve temporary data by UUID and delete it
//	@Tags			tempdata
//	@Produce		json
//	@Param			uuid	path		string	true	"UUID of the temporary data"
//	@Success		200		{object}	dto.TempDataDto
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/tempdata/{uuid} [put]
func (c *TempDataController) getAndDeleteTempData(ctx *gin.Context) {
	uuidStr := ctx.Param("uuid")
	if uuidStr == "" {
		c.logger.Error("UUID is required")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, "UUID is required")
		return
	}

	uuid, err := uuid.Parse(uuidStr)
	if err != nil {
		c.logger.Errorf("Invalid UUID: %s", uuidStr)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, "Invalid UUID")
		return
	}

	driverUuid, vehicleUuid, err := c.TempDataService.GetAndDeleteByUUID(uuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.logger.Errorf("Temp data not found: %s", uuid)
			ctx.AbortWithStatusJSON(http.StatusNotFound, "Temp data not found")
			return
		}
		if errors.Is(err, cerror.ErrOutdated) {
			c.logger.Errorf("Temp data expired: %s", uuid)
			ctx.AbortWithStatusJSON(http.StatusGone, "Temporary data has expired")
			return
		}

		c.logger.Errorf("Failed to get/delete temp data: %+v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, "Failed to get/delete temp data")
		return
	}

	result := dto.TempDataDto{
		VehicleUuid: vehicleUuid,
		DriverUuid:  driverUuid,
	}

	ctx.JSON(http.StatusOK, result)
}
