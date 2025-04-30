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

type LicenseController struct {
	LicenseService service.IDriverLicenseCrudService
	logger         *zap.SugaredLogger
}

func NewLicenseController() *LicenseController {
	var controller *LicenseController

	// Use the mock service for testing
	app.Invoke(func(licenseService service.IDriverLicenseCrudService, logger *zap.SugaredLogger) {
		// create controller
		controller = &LicenseController{
			LicenseService: licenseService,
			logger:         logger,
		}
	})

	return controller
}

func (c *LicenseController) RegisterEndpoints(api *gin.RouterGroup) {
	// create a group with the name of the router
	group := api.Group("/license")

	// register Endpoints
	group.POST("/", c.createLicense)
	group.GET("/:uuid", c.getLicense)
	group.GET("/", c.getAllLicenses)
	group.PUT("/:uuid", c.updateLicense)
	group.DELETE("/:uuid", c.deleteLicense)
}

// CreateLicense godoc
//
//	@Summary	Creates a new license
//	@Schemes
//	@Description	Create a new license with an owner
//	@Tags			license
//	@Produce		json
//	@Success		201	{object}	dto.DriverLicenseDto
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Param			model	body	dto.DriverLicenseDto	true	"License model"
//	@Router			/license [post]
func (c *LicenseController) createLicense(ctx *gin.Context) {
	var licenseDto dto.DriverLicenseDto
	if err := ctx.Bind(&licenseDto); err != nil {
		c.logger.Errorf("Failed to bind error = %+v", err)
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	_, claims, err := auth.ParseToken(ctx.Request.Header.Get("Authorization"))
	if err != nil {
		c.logger.Errorf("Failed to parse token: %v", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	ownerUuid, err := uuid.Parse(claims.Uuid)
	if err != nil {
		c.logger.Errorf("Failed to parse owner UUID: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid owner UUID"})
		return
	}
	license := licenseDto.ToModel()

	license, err = c.LicenseService.Create(license, ownerUuid)
	if err != nil {
		if errors.Is(err, cerror.ErrBadRole) {
			c.logger.Errorf("Role or user is invalid, err = %+v", err)
			ctx.AbortWithError(http.StatusNotFound, err)
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.logger.Errorf("Owner with uuid = %s not found", ownerUuid)
			ctx.AbortWithError(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var dto dto.DriverLicenseDto
	ctx.JSON(http.StatusCreated, dto.FromModel(license))
}

// GetLicense godoc
//
//	@Summary	Gets a license with uuid
//	@Schemes
//	@Tags		license
//	@Produce	json
//	@Success	200	{object}	dto.DriverLicenseDto
//	@Failure	400
//	@Failure	404
//	@Failure	500
//	@Param		uuid	path	string	true	"License UUID"
//	@Router		/license/{uuid} [get]
func (c *LicenseController) getLicense(ctx *gin.Context) {
	licenseUuid, err := uuid.Parse(ctx.Param("uuid"))
	if err != nil {
		c.logger.Errorf("error parsing uuid value = %s", ctx.Param("uuid"))
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	license, err := c.LicenseService.GetByUuid(licenseUuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.logger.Errorf("License with uuid = %s not found", licenseUuid)
			ctx.AbortWithError(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var licenseDto dto.DriverLicenseDto
	ctx.JSON(http.StatusOK, licenseDto.FromModel(license))
}

// MyLicenses godoc
//
//	@Summary	Gets your licenses
//	@Schemes
//	@Tags		license
//	@Produce	json
//	@Success	200	{object}	[]dto.DriverLicenseDto
//	@Failure	400
//	@Failure	404
//	@Failure	500
//	@Router		/license [get]
func (c *LicenseController) getAllLicenses(ctx *gin.Context) {
	_, claims, err := auth.ParseToken(ctx.Request.Header.Get("Authorization"))
	if err != nil {
		c.logger.Errorf("Failed to parse token: %v", err)
		ctx.AbortWithError(http.StatusUnauthorized, err)
		return
	}
	userUuid, err := uuid.Parse(claims.Uuid)
	if err != nil {
		c.logger.Errorf("Failed to parse uuid = %s, err + %+v", claims.Uuid, err)
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	licenses, err := c.LicenseService.GetAll()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.logger.Errorf("Licenses for user uuid = %s not found", userUuid)
			ctx.AbortWithError(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var licenseDtos []dto.DriverLicenseDto
	for _, license := range licenses {
		var licenseDto dto.DriverLicenseDto
		licenseDtos = append(licenseDtos, *licenseDto.FromModel(&license))
	}

	ctx.JSON(http.StatusOK, licenseDtos)
}

// UpdateLicense godoc
//
//	@Summary	Updates a license
//	@Schemes
//	@Tags		license
//	@Produce	json
//	@Success	200	{object}	dto.DriverLicenseDto
//	@Failure	400
//	@Failure	404
//	@Failure	500
//	@Param		uuid	path	string					true	"License UUID"
//	@Param		model	body	dto.DriverLicenseDto	true	"License model"
//	@Router		/license/{uuid} [put]
func (c *LicenseController) updateLicense(ctx *gin.Context) {
	licenseUuid, err := uuid.Parse(ctx.Param("uuid"))
	if err != nil {
		c.logger.Errorf("error parsing uuid value = %s", ctx.Param("uuid"))
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var updateDto dto.DriverLicenseDto
	if err := ctx.Bind(&updateDto); err != nil {
		c.logger.Errorf("Failed to bind error = %+v", err)
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	updatedLicense, err := c.LicenseService.Update(licenseUuid, updateDto.ToModel())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.logger.Errorf("License with uuid = %s not found", licenseUuid)
			ctx.AbortWithError(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, updateDto.FromModel(updatedLicense))
}

// DeleteLicense godoc
//
//	@Summary	Deletes a license
//	@Schemes
//	@Tags		license
//	@Success	204
//	@Failure	400
//	@Failure	404
//	@Failure	500
//	@Param		uuid	path	string	true	"License UUID"
//	@Router		/license/{uuid} [delete]
func (c *LicenseController) deleteLicense(ctx *gin.Context) {
	licenseUuid, err := uuid.Parse(ctx.Param("uuid"))
	if err != nil {
		c.logger.Errorf("error parsing uuid value = %s", ctx.Param("uuid"))
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = c.LicenseService.Delete(licenseUuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.logger.Errorf("License with uuid = %s not found", licenseUuid)
			ctx.AbortWithError(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.AbortWithStatus(http.StatusNoContent)
}
