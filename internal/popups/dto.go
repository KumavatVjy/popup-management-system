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
