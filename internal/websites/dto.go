package websites

type CreateWebsiteRequest struct {
	WebsiteName string `json:"website_name" binding:"required,min=2,max=150"`

	Domain string `json:"domain" binding:"required"`

	Platform string `json:"platform" binding:"required"`
}

type UpdateWebsiteRequest struct {
	WebsiteName string `json:"website_name" binding:"required,min=2,max=150"`

	Domain string `json:"domain" binding:"required"`

	Platform string `json:"platform" binding:"required"`

	Status bool `json:"status"`
}
