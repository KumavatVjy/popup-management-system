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

type WebsiteResponse struct {
	ID          uint   `json:"id"`
	WebsiteName string `json:"website_name"`
	Domain      string `json:"domain"`
	Platform    string `json:"platform"`
	Status      bool   `json:"status"`
}

func MapWebsiteToResponse(w Website) WebsiteResponse {
	return WebsiteResponse{
		ID:          w.ID,
		WebsiteName: w.WebsiteName,
		Domain:      w.Domain,
		Platform:    w.Platform,
		Status:      w.Status,
	}
}

func MapWebsitesToResponse(websites []Website) []WebsiteResponse {
	var responses []WebsiteResponse
	for _, w := range websites {
		responses = append(responses, MapWebsiteToResponse(w))
	}
	return responses
}
