package websites

import (
	"net/http"

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
		common.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	common.Success(c, "Website created successfully", nil)
}

// Get all websites
func (wc *WebsiteController) List(c *gin.Context) {

	websites, err := wc.service.GetAll()

	if err != nil {
		common.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	common.Success(c, "Websites fetched successfully", websites)
}