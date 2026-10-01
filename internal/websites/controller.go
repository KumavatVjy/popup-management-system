package websites

import (
	"net/http"

	"strconv"

	"github.com/gin-gonic/gin"

	"popup-manager-api/internal/common"
)

type WebsiteController struct {
	service WebsiteService
}

func NewWebsiteController(service WebsiteService) *WebsiteController {
	return &WebsiteController{
		service: service,
	}
}

// Create Website
func (wc *WebsiteController) Create(c *gin.Context) {

	var request CreateWebsiteRequest

	// Validate Request
	if err := c.ShouldBindJSON(&request); err != nil {
		common.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// Logged-in User ID from JWT Middleware
	createdBy := c.GetUint("user_id")

	// Call Service
	err := wc.service.Create(request, createdBy)

	if err != nil {
		common.HandleError(c, err)
		return
	}

	common.Success(c, "Website created successfully", nil)
}

// Get all websites
func (wc *WebsiteController) List(c *gin.Context) {

	params := common.GetQueryParams(c)

	websites, total, err := wc.service.GetAll(params)

	if err != nil {
		common.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	pagination := common.BuildPagination(total, params)

	common.SuccessWithPagination(c, "Websites fetched successfully", MapWebsitesToResponse(websites), pagination)
}

// Get website by id
func (wc *WebsiteController) GetByID(c *gin.Context) {

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)

	if err != nil {
		common.Error(c, http.StatusBadRequest, "Invalid website ID")
		return
	}

	website, err := wc.service.GetByID(uint(id))

	if err != nil {
		common.Error(c, http.StatusNotFound, "Website not found")
		return
	}

	common.Success(c, "Website fetched successfully", MapWebsiteToResponse(*website))
}

// Update website by ID
func (wc *WebsiteController) Update(c *gin.Context) {

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)

	if err != nil {
		common.Error(c, http.StatusBadRequest, "Invalid website ID")
		return
	}

	var request UpdateWebsiteRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		common.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err = wc.service.Update(uint(id), request)

	if err != nil {
		common.HandleError(c, err)
		return
	}

	common.Success(c, "Website updated successfully", nil)
}

// Delete website by ID
func (wc *WebsiteController) Delete(c *gin.Context) {

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)

	if err != nil {
		common.Error(c, http.StatusBadRequest, "Invalid website ID")
		return
	}

	err = wc.service.Delete(uint(id))

	if err != nil {
		common.HandleError(c, err)
		return
	}

	common.Success(c, "Website deleted successfully", nil)

}
