package popups

import "time"

type CreatePopupRequest struct {
	WebsiteID uint       `json:"website_id" binding:"required"`
	Title     string     `json:"title" binding:"required,min=2,max=255"`
	Content   string     `json:"content" binding:"required"`
	Position  string     `json:"position" binding:"required"`
	Status    bool       `json:"status"`
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
}

type UpdatePopupRequest struct {
	Title     string     `json:"title" binding:"required,min=2,max=255"`
	Content   string     `json:"content" binding:"required"`
	Position  string     `json:"position" binding:"required"`
	Status    bool       `json:"status"`
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
}

type PopupResponse struct {
	ID        uint       `json:"id"`
	WebsiteID uint       `json:"website_id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Position  string     `json:"position"`
	Status    bool       `json:"status"`
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
}

func MapPopupToResponse(p Popup) PopupResponse {
	return PopupResponse{
		ID:        p.ID,
		WebsiteID: p.WebsiteID,
		Title:     p.Title,
		Content:   p.Content,
		Position:  p.Position,
		Status:    p.Status,
		StartTime: p.StartTime,
		EndTime:   p.EndTime,
	}
}

func MapPopupsToResponse(popups []Popup) []PopupResponse {
	var responses []PopupResponse
	for _, p := range popups {
		responses = append(responses, MapPopupToResponse(p))
	}
	return responses
}
