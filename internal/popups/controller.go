package popups

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"popup-manager-api/internal/common"
)

type PopupController struct {
	service PopupService
}

func NewPopupController(service PopupService) *PopupController {
	return &PopupController{
		service: service,
	}
}

// Create Popup
func (pc *PopupController) Create(c *gin.Context) {
	var request CreatePopupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	createdBy := c.GetUint("user_id")

	err := pc.service.Create(request, createdBy)
	if err != nil {
		common.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	common.Success(c, "Popup created successfully", nil)
}

// Get all popups
func (pc *PopupController) List(c *gin.Context) {
	popups, err := pc.service.GetAll()
	if err != nil {
		common.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	common.Success(c, "Popups fetched successfully", MapPopupsToResponse(popups))
}

// Get popup by ID
func (pc *PopupController) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.Error(c, http.StatusBadRequest, "Invalid popup ID")
		return
	}

	popup, err := pc.service.GetByID(uint(id))
	if err != nil {
		common.Error(c, http.StatusNotFound, "Popup not found")
		return
	}

	common.Success(c, "Popup fetched successfully", MapPopupToResponse(*popup))
}

// Get popups by website ID
func (pc *PopupController) GetByWebsiteID(c *gin.Context) {
	websiteID, err := strconv.ParseUint(c.Param("website_id"), 10, 64)
	if err != nil {
		common.Error(c, http.StatusBadRequest, "Invalid website ID")
		return
	}

	popups, err := pc.service.GetByWebsiteID(uint(websiteID))
	if err != nil {
		common.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	common.Success(c, "Popups fetched successfully", MapPopupsToResponse(popups))
}

// Update popup by ID
func (pc *PopupController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.Error(c, http.StatusBadRequest, "Invalid popup ID")
		return
	}

	var request UpdatePopupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err = pc.service.Update(uint(id), request)
	if err != nil {
		common.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	common.Success(c, "Popup updated successfully", nil)
}

// Delete popup by ID
func (pc *PopupController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.Error(c, http.StatusBadRequest, "Invalid popup ID")
		return
	}

	err = pc.service.Delete(uint(id))
	if err != nil {
		common.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	common.Success(c, "Popup deleted successfully", nil)
}
