package popups

import (
	"time"

	"gorm.io/gorm"
)

const (
	PositionCenter      = "center"
	PositionTopBanner   = "top-banner"
	PositionBottomLeft  = "bottom-left"
	PositionBottomRight = "bottom-right"
)

type Popup struct {
	gorm.Model

	WebsiteID uint `gorm:"not null" json:"website_id"`

	Title string `gorm:"size:255;not null" json:"title"`

	Content string `gorm:"type:text;not null" json:"content"`

	Position string `gorm:"size:50;not null" json:"position"`

	Status bool `gorm:"default:true" json:"status"`

	StartTime *time.Time `json:"start_time"`

	EndTime *time.Time `json:"end_time"`

	CreatedBy uint `json:"created_by"`
}
