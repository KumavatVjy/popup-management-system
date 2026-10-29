package popups

import (
	"time"

	"gorm.io/gorm"
	"popup-manager-api/internal/websites"
)

const (
	// Legacy / Common Positions
	PositionCenter      = "center"
	PositionTopBanner   = "top-banner"
	PositionBottomLeft  = "bottom-left"
	PositionBottomRight = "bottom-right"

	// 9 Standard Placement Grid Positions
	PositionTopLeft            = "top-left"
	PositionTopCenter          = "top-center"
	PositionTopRight           = "top-right"
	PositionMiddleLeft         = "middle-left"
	PositionMiddleCenter       = "middle-center"
	PositionMiddleRight        = "middle-right"
	PositionBottomCenter       = "bottom-center"
	PositionTopBannerCenter    = "top-banner-center"
	PositionBottomBannerCenter = "bottom-banner-center"
)

type Popup struct {
	gorm.Model

	WebsiteID uint               `gorm:"not null" json:"website_id"`
	Websites  []websites.Website `gorm:"many2many:popup_websites;" json:"websites,omitempty"`
	Title     string             `gorm:"size:255;not null" json:"title"`
	Content   string             `gorm:"type:text;not null" json:"content"`
	Position  string             `gorm:"size:50;not null" json:"position"`
	Status    bool               `gorm:"default:true" json:"status"`

	// 1. Auto-deactivate schedule
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"` // Deactivation date/time

	// 2. Banner image with link
	BannerImageURL        string `gorm:"size:1024" json:"banner_image_url"`
	BannerLinkURL         string `gorm:"size:1024" json:"banner_link_url"`
	BannerLinkTarget      string `gorm:"size:20;default:'_blank'" json:"banner_link_target"`
	BadgeText             string `gorm:"size:100" json:"badge_text"`
	ShowImage             bool   `gorm:"not null" json:"show_image"`
	ShowContent           bool   `gorm:"not null" json:"show_content"`
	MediaOverlayColor     string `gorm:"size:50" json:"media_overlay_color"`
	MediaOverlayOpacity   int    `gorm:"default:0" json:"media_overlay_opacity"`
	MediaBorderColor      string `gorm:"size:50" json:"media_border_color"`
	MediaBorderWidth      int    `gorm:"default:0" json:"media_border_width"`
	FooterBackgroundColor string `gorm:"size:50" json:"footer_background_color"`
	FooterBorderColor     string `gorm:"size:50" json:"footer_border_color"`
	FooterBorderWidth     int    `gorm:"default:0" json:"footer_border_width"`
	MediaBoxShadow        string `gorm:"type:text" json:"media_box_shadow"`
	FooterBoxShadow       string `gorm:"type:text" json:"footer_box_shadow"`

	// 3. Coming soon / event countdown timer
	CountdownEnabled bool       `gorm:"default:false" json:"countdown_enabled"`
	CountdownEndTime *time.Time `json:"countdown_end_time"`
	CountdownTitle   string     `gorm:"size:255" json:"countdown_title"`

	// 4. Color pattern on CTA, Title, Border, Background
	TitleColor        string `gorm:"size:50" json:"title_color"`
	BackgroundColor   string `gorm:"size:50" json:"background_color"`
	BorderColor       string `gorm:"size:50" json:"border_color"`
	BorderWidth       int    `gorm:"default:1" json:"border_width"`
	BorderTopColor    string `gorm:"size:50" json:"border_top_color"`
	BorderTopWidth    int    `gorm:"default:-1" json:"border_top_width"`
	BorderRightColor  string `gorm:"size:50" json:"border_right_color"`
	BorderRightWidth  int    `gorm:"default:-1" json:"border_right_width"`
	BorderBottomColor string `gorm:"size:50" json:"border_bottom_color"`
	BorderBottomWidth int    `gorm:"default:-1" json:"border_bottom_width"`
	BorderLeftColor   string `gorm:"size:50" json:"border_left_color"`
	BorderLeftWidth   int    `gorm:"default:-1" json:"border_left_width"`
	BoxShadow         string `gorm:"type:text" json:"box_shadow"`
	BorderRadius      int    `gorm:"default:12" json:"border_radius"`
	PopupWidth        int    `json:"popup_width"`
	PopupHeight       int    `json:"popup_height"`
	ShowOnDesktop     bool   `gorm:"not null" json:"show_on_desktop"`
	ShowOnTablet      bool   `gorm:"not null" json:"show_on_tablet"`
	ShowOnMobile      bool   `gorm:"not null" json:"show_on_mobile"`

	// 5. CTA with URL and parameters
	CtaText      string `gorm:"size:255" json:"cta_text"`
	CtaURL       string `gorm:"size:1024" json:"cta_url"`
	CtaTarget    string `gorm:"size:20;default:'_blank'" json:"cta_target"`
	CtaTextColor string `gorm:"size:50" json:"cta_text_color"`
	CtaBgColor   string `gorm:"size:50" json:"cta_bg_color"`

	// 6. Close button choices
	CloseButtonChoice   string `gorm:"size:50;default:'default'" json:"close_button_choice"`
	CloseButtonPosition string `gorm:"size:50;default:'top-right'" json:"close_button_position"`
	CloseButtonColor    string `gorm:"size:50" json:"close_button_color"`

	// 7. On scroll setting when it will display & triggers
	TriggerType             string `gorm:"size:50;default:'immediate'" json:"trigger_type"`
	TriggerDelaySeconds     int    `gorm:"default:0" json:"trigger_delay_seconds"`
	TriggerScrollPercentage int    `gorm:"default:0" json:"trigger_scroll_percentage"`

	// 8. Reappearance cooldown after close (in minutes)
	CooldownMinutes int `gorm:"default:0" json:"cooldown_minutes"`

	// 9. Custom CSS
	CustomCSS    string `gorm:"type:text" json:"custom_css"`
	CustomJS     string `gorm:"type:text" json:"custom_js"`
	TargetPages  string `gorm:"type:text" json:"target_pages"`
	ExcludePages string `gorm:"type:text" json:"exclude_pages"`

	CreatedBy uint `json:"created_by"`
}
