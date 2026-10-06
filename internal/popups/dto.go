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

// PublicPopupResponse represents the sanitized public popup payload for the JavaScript SDK.
// It deliberately excludes admin metadata (website_id, created_by, status, created_at, updated_at, deleted_at).
type PublicPopupResponse struct {
	ID        uint       `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Position  string     `json:"position"`
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
}

func MapPopupToPublicResponse(p Popup) PublicPopupResponse {
	return PublicPopupResponse{
		ID:        p.ID,
		Title:     p.Title,
		Content:   p.Content,
		Position:  p.Position,
		StartTime: p.StartTime,
		EndTime:   p.EndTime,
	}
}

func MapPopupsToPublicResponse(popups []Popup) []PublicPopupResponse {
	responses := make([]PublicPopupResponse, 0, len(popups))
	for _, p := range popups {
		responses = append(responses, MapPopupToPublicResponse(p))
	}
	return responses
}
