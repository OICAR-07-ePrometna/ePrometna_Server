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
	app.Invoke(func(vehicleService service.IVehicleService, logger *zap.SugaredLogger) {
		controller = &VehicleController{
			VehicleService: vehicleService,
			logger:         logger,
		}
	})
	return controller
}

func (c *VehicleController) RegisterEndpoints(api *gin.RouterGroup) {
	group := api.Group("/vehicle")

	// Publicly accessible or role-specific GETs
	group.GET("/:uuid", middleware.Protect(model.RoleHAK, model.RoleFirma, model.RoleOsoba), c.get)
	group.GET("/", middleware.Protect(model.RoleFirma, model.RoleOsoba), c.myVehicles)
	group.GET("/vin/:vin", middleware.Protect(model.RoleHAK, model.RoleFirma, model.RoleOsoba), c.getByVin)

	// Endpoints requiring HAK role
	// Create a new sub-group for HAK specific middleware
	hakGroup := group.Group("")
	hakGroup.Use(middleware.Protect(model.RoleHAK))
	{
		hakGroup.POST("/", c.create)
		hakGroup.PUT("/:uuid", c.update)
		hakGroup.DELETE("/:uuid", c.delete)
		hakGroup.PUT("/change-owner", c.changeOwner)
		hakGroup.PUT("/registration/:uuid", c.registration)
		hakGroup.PUT("/deregister/:uuid", c.deregister)
	}
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
		v.logger.Errorf("Failed to delete vehicle %s: %+v", vehicleUuid, err)
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
		v.logger.Errorf("Failed to create model from DTO error = %+v", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	ownerUuid, err := uuid.Parse(newDto.OwnerUuid)
	if err != nil {
		v.logger.Errorf("Failed to parse owner uuid = %s, err + %+v", newDto.OwnerUuid, err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	createdVehicle, err := v.VehicleService.Create(vehicle, ownerUuid)
	if err != nil {
		if errors.Is(err, cerror.ErrBadRole) {
			v.logger.Errorf("Role or user is invalid for owning a vehicle, err = %+v", err)
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			v.logger.Errorf("User (owner) with uuid = %s not found", newDto.OwnerUuid)
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		v.logger.Errorf("Failed to create vehicle: %+v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var respDto dto.VehicleDto
	c.JSON(http.StatusCreated, respDto.FromModel(createdVehicle))
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
		v.logger.Errorf("Failed to read vehicle %s: %+v", vehicleUuid, err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var detailsDto dto.VehicleDetailsDto
	c.JSON(http.StatusOK, detailsDto.FromModel(vehicle))
}

// myVehicle godoc
//
//	@Summary	Gets your vehicles
//	@Schemes
//	@Tags		vehicle
//	@Produce	json
//	@Success	200	{object}	[]dto.VehicleDto
//	@Failure	400
//	@Failure	401
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
	userUuid, err := uuid.Parse(claims.Uuid)
	if err != nil {
		v.logger.Errorf("Failed to parse uuid from token claims = %s, err + %+v", claims.Uuid, err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	vehicles, err := v.VehicleService.ReadAll(userUuid)
	if err != nil {
		v.logger.Errorf("Failed to read vehicles for user %s: %+v", userUuid, err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	var dtos dto.VehiclesDto
	c.JSON(http.StatusOK, dtos.FromModel(vehicles))
}

// changeOwner godoc
//
//	@Summary	changes owner to new owner with uuid
//	@Schemes
//	@Tags		vehicle
//	@Success	200
//	@Failure	400
//	@Failure	404
//	@Failure	500
//	@Param		changeOwnerDto	body	dto.ChangeOwnerDto	true	"Dto for changing ownership"
//	@Router		/vehicle/change-owner [put]
func (v *VehicleController) changeOwner(c *gin.Context) {
	var cowner dto.ChangeOwnerDto
	if err := c.Bind(&cowner); err != nil {
		v.logger.Errorf("Failed to bind ChangeOwnerDto: %+v", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	vehicleUuid, err := uuid.Parse(cowner.VehicleUuid)
	if err != nil {
		v.logger.Errorf("Failed to parse vehicle uuid from DTO (after binding) error = %+v", err)
		c.AbortWithError(http.StatusBadRequest, errors.New("invalid vehicle UUID format in DTO"))
		return
	}

	ownerUuid, err := uuid.Parse(cowner.NewOwnerUuid)
	if err != nil {
		v.logger.Errorf("Failed to parse new owner uuid from DTO (after binding) error = %+v", err)
		c.AbortWithError(http.StatusBadRequest, errors.New("invalid new owner UUID format in DTO"))
		return
	}

	err = v.VehicleService.ChangeOwner(vehicleUuid, ownerUuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			v.logger.Errorf("ChangeOwner failed (record not found) for vehicle %s to owner %s: %+v", vehicleUuid, ownerUuid, err)
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		if errors.Is(err, cerror.ErrBadRole) {
			v.logger.Errorf("ChangeOwner failed (bad role) for vehicle %s to owner %s: %+v", vehicleUuid, ownerUuid, err)
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		v.logger.Errorf("Failed to change owner for vehicle %s to %s: %+v", vehicleUuid, ownerUuid, err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	v.logger.Debugf("Vehicle with uuid = %s changed owner (uuid = %s)", vehicleUuid, ownerUuid)
	c.AbortWithStatus(http.StatusNoContent)
}

// registerVehicle godoc
//
//	@Summary	Tehnicki pregled
//	@Schemes
//	@Description	Performs a technical inspection and registers a vehicle.
//	@Tags			vehicle
//	@Accept			json
//	@Produce		json
//	@Success		200					"Successfully registered"
//	@Failure		400					{object}	object{error=string}	"Invalid request (bad UUID, binding error)"
//	@Failure		404					{object}	object{error=string}	"Vehicle not found"
//	@Failure		500					{object}	object{error=string}	"Internal server error"
//	@Param			uuid				path		string					true	"Vehicle UUID"	Format(uuid)
//	@Param			registrationData	body		dto.RegistrationDto		true	"Data for vehicle registration"
//	@Router			/vehicle/registration/{uuid} [put]
func (v *VehicleController) registration(c *gin.Context) {
	vehicleUuidString := c.Param("uuid")
	vehicleUuid, err := uuid.Parse(vehicleUuidString)
	if err != nil {
		v.logger.Errorf("Failed to parse vehicle uuid from path: %s, error = %+v", vehicleUuidString, err)
		c.AbortWithError(http.StatusBadRequest, errors.New("invalid vehicle UUID format in path"))
		return
	}

	var regDto dto.RegistrationDto
	if err := c.Bind(&regDto); err != nil {
		v.logger.Errorf("Failed to bind RegistrationDto: %+v", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	regModel, err := regDto.ToModel()
	if err != nil {
		v.logger.Errorf("Failed to map RegistrationDto to model: %+v", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = v.VehicleService.Registration(vehicleUuid, regModel)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			v.logger.Errorf("Vehicle with uuid = %s not found for registration", vehicleUuid)
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		v.logger.Errorf("Error during vehicle registration for uuid = %s: %+v", vehicleUuid, err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	v.logger.Infof("Vehicle %s registered successfully.", vehicleUuid)
	c.AbortWithStatus(http.StatusNoContent)
}

// GetVehicleByVin godoc
//
//	@Summary	Gets a vehicle by VIN number
//	@Schemes
//	@Tags		vehicle
//	@Produce	json
//	@Success	200	{object}	dto.VehicleDetailsDto
//	@Failure	400	{object}	object{error=string}	"Invalid request"
//	@Failure	404	{object}	object{error=string}	"Vehicle not found"
//	@Failure	500	{object}	object{error=string}	"Internal server error"
//	@Param		vin	path		string					true	"Vehicle VIN number"
//	@Router		/vehicle/vin/{vin} [get]
func (v *VehicleController) getByVin(c *gin.Context) {
	vin := c.Param("vin")
	if vin == "" {
		v.logger.Error("Empty VIN provided")
		c.AbortWithError(http.StatusBadRequest, errors.New("VIN number is required"))
		return
	}

	vehicle, err := v.VehicleService.ReadByVin(vin)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			v.logger.Errorf("Vehicle with VIN = %s not found", vin)
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		v.logger.Errorf("Failed to read vehicle with VIN %s: %+v", vin, err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var detailsDto dto.VehicleDetailsDto
	c.JSON(http.StatusOK, detailsDto.FromModel(vehicle))
}

// DeregisterVehicle godoc
//
//	@Summary	Deregister a vehicle by setting its license plate to null
//	@Schemes
//	@Description	Sets the vehicle's license plate to null
//	@Tags			vehicle
//	@Success		200
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Param			uuid	path	string	true	"Vehicle UUID"
//	@Router			/vehicle/deregister/{uuid} [put]
func (v *VehicleController) deregister(c *gin.Context) {
	vehicleUuid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		v.logger.Errorf("error parsing uuid value = %s", c.Param("uuid"))
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = v.VehicleService.Deregister(vehicleUuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			v.logger.Errorf("Vehicle with uuid = %s not found", vehicleUuid)
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		v.logger.Errorf("Failed to deregister vehicle %s: %+v", vehicleUuid, err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

// Upodate vehicle godoc
//
//	@Summary	Update vehicle data
//	@Schemes
//	@Description	Will allow to update some wehicle data
//	@Tags			vehicle
//	@Success		200
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Param			uuid	path	string	true	"Vehicle UUID"
//	@Router			/vehicle/{uuid} [put]
func (v *VehicleController) update(c *gin.Context) {
	var newDto dto.VehicleDto
	if err := c.Bind(&newDto); err != nil {
		v.logger.Errorf("Failed to bind error = %+v", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	vehicle, err := newDto.ToModel()
	if err != nil {
		v.logger.Errorf("Failed to create model from DTO error = %+v", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	vehicleUuid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		v.logger.Errorf("error parsing uuid value = %s", c.Param("uuid"))
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	createdVehicle, err := v.VehicleService.Update(vehicleUuid, *vehicle)
	if err != nil {
		if errors.Is(err, cerror.ErrBadRole) {
			v.logger.Errorf("Role or user is invalid for owning a vehicle, err = %+v", err)
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			v.logger.Errorf("Vehicle with uuid = %s not found", vehicleUuid)
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		v.logger.Errorf("Failed to create vehicle: %+v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var respDto dto.VehicleDto
	c.JSON(http.StatusCreated, respDto.FromModel(createdVehicle))
}
