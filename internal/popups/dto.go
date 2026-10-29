package popups

import "time"

type CreatePopupRequest struct {
	WebsiteID  uint       `json:"website_id"`
	WebsiteIDs []uint     `json:"website_ids"`
	Title      string     `json:"title" binding:"required,min=2,max=255"`
	Content    string     `json:"content" binding:"required"`
	Position   string     `json:"position" binding:"required"`
	Status     bool       `json:"status"`
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`

	BannerImageURL        string     `json:"banner_image_url"`
	BannerLinkURL         string     `json:"banner_link_url"`
	BannerLinkTarget      string     `json:"banner_link_target"`
	BadgeText             string     `json:"badge_text"`
	ShowImage             bool       `json:"show_image"`
	ShowContent           bool       `json:"show_content"`
	MediaOverlayColor     string     `json:"media_overlay_color"`
	MediaOverlayOpacity   int        `json:"media_overlay_opacity"`
	MediaBorderColor      string     `json:"media_border_color"`
	MediaBorderWidth      int        `json:"media_border_width"`
	FooterBackgroundColor string     `json:"footer_background_color"`
	FooterBorderColor     string     `json:"footer_border_color"`
	FooterBorderWidth     int        `json:"footer_border_width"`
	MediaBoxShadow        string     `json:"media_box_shadow"`
	FooterBoxShadow       string     `json:"footer_box_shadow"`
	CountdownEnabled      bool       `json:"countdown_enabled"`
	CountdownEndTime      *time.Time `json:"countdown_end_time"`
	CountdownTitle        string     `json:"countdown_title"`

	TitleColor        string `json:"title_color"`
	BackgroundColor   string `json:"background_color"`
	BorderColor       string `json:"border_color"`
	BorderWidth       int    `json:"border_width"`
	BorderTopColor    string `json:"border_top_color"`
	BorderTopWidth    int    `json:"border_top_width"`
	BorderRightColor  string `json:"border_right_color"`
	BorderRightWidth  int    `json:"border_right_width"`
	BorderBottomColor string `json:"border_bottom_color"`
	BorderBottomWidth int    `json:"border_bottom_width"`
	BorderLeftColor   string `json:"border_left_color"`
	BorderLeftWidth   int    `json:"border_left_width"`
	BoxShadow         string `json:"box_shadow"`
	BorderRadius      int    `json:"border_radius"`
	PopupWidth        int    `json:"popup_width"`
	PopupHeight       int    `json:"popup_height"`
	ShowOnDesktop     bool   `json:"show_on_desktop"`
	ShowOnTablet      bool   `json:"show_on_tablet"`
	ShowOnMobile      bool   `json:"show_on_mobile"`

	CtaText      string `json:"cta_text"`
	CtaURL       string `json:"cta_url"`
	CtaTarget    string `json:"cta_target"`
	CtaTextColor string `json:"cta_text_color"`
	CtaBgColor   string `json:"cta_bg_color"`

	CloseButtonChoice   string `json:"close_button_choice"`
	CloseButtonPosition string `json:"close_button_position"`
	CloseButtonColor    string `json:"close_button_color"`

	TriggerType             string `json:"trigger_type"`
	TriggerDelaySeconds     int    `json:"trigger_delay_seconds"`
	TriggerScrollPercentage int    `json:"trigger_scroll_percentage"`
	CooldownMinutes         int    `json:"cooldown_minutes"`

	CustomCSS    string `json:"custom_css"`
	CustomJS     string `json:"custom_js"`
	TargetPages  string `json:"target_pages"`
	ExcludePages string `json:"exclude_pages"`
}

type UpdatePopupRequest struct {
	WebsiteID  uint       `json:"website_id"`
	WebsiteIDs []uint     `json:"website_ids"`
	Title      string     `json:"title" binding:"required,min=2,max=255"`
	Content    string     `json:"content" binding:"required"`
	Position   string     `json:"position" binding:"required"`
	Status     bool       `json:"status"`
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`

	BannerImageURL        string     `json:"banner_image_url"`
	BannerLinkURL         string     `json:"banner_link_url"`
	BannerLinkTarget      string     `json:"banner_link_target"`
	BadgeText             string     `json:"badge_text"`
	ShowImage             bool       `json:"show_image"`
	ShowContent           bool       `json:"show_content"`
	MediaOverlayColor     string     `json:"media_overlay_color"`
	MediaOverlayOpacity   int        `json:"media_overlay_opacity"`
	MediaBorderColor      string     `json:"media_border_color"`
	MediaBorderWidth      int        `json:"media_border_width"`
	FooterBackgroundColor string     `json:"footer_background_color"`
	FooterBorderColor     string     `json:"footer_border_color"`
	FooterBorderWidth     int        `json:"footer_border_width"`
	MediaBoxShadow        string     `json:"media_box_shadow"`
	FooterBoxShadow       string     `json:"footer_box_shadow"`
	CountdownEnabled      bool       `json:"countdown_enabled"`
	CountdownEndTime      *time.Time `json:"countdown_end_time"`
	CountdownTitle        string     `json:"countdown_title"`

	TitleColor        string `json:"title_color"`
	BackgroundColor   string `json:"background_color"`
	BorderColor       string `json:"border_color"`
	BorderWidth       int    `json:"border_width"`
	BorderTopColor    string `json:"border_top_color"`
	BorderTopWidth    int    `json:"border_top_width"`
	BorderRightColor  string `json:"border_right_color"`
	BorderRightWidth  int    `json:"border_right_width"`
	BorderBottomColor string `json:"border_bottom_color"`
	BorderBottomWidth int    `json:"border_bottom_width"`
	BorderLeftColor   string `json:"border_left_color"`
	BorderLeftWidth   int    `json:"border_left_width"`
	BoxShadow         string `json:"box_shadow"`
	BorderRadius      int    `json:"border_radius"`
	PopupWidth        int    `json:"popup_width"`
	PopupHeight       int    `json:"popup_height"`
	ShowOnDesktop     bool   `json:"show_on_desktop"`
	ShowOnTablet      bool   `json:"show_on_tablet"`
	ShowOnMobile      bool   `json:"show_on_mobile"`

	CtaText      string `json:"cta_text"`
	CtaURL       string `json:"cta_url"`
	CtaTarget    string `json:"cta_target"`
	CtaTextColor string `json:"cta_text_color"`
	CtaBgColor   string `json:"cta_bg_color"`

	CloseButtonChoice   string `json:"close_button_choice"`
	CloseButtonPosition string `json:"close_button_position"`
	CloseButtonColor    string `json:"close_button_color"`

	TriggerType             string `json:"trigger_type"`
	TriggerDelaySeconds     int    `json:"trigger_delay_seconds"`
	TriggerScrollPercentage int    `json:"trigger_scroll_percentage"`
	CooldownMinutes         int    `json:"cooldown_minutes"`

	CustomCSS    string `json:"custom_css"`
	CustomJS     string `json:"custom_js"`
	TargetPages  string `json:"target_pages"`
	ExcludePages string `json:"exclude_pages"`
}

type PopupResponse struct {
	ID         uint       `json:"id"`
	WebsiteID  uint       `json:"website_id"`
	WebsiteIDs []uint     `json:"website_ids"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	Position   string     `json:"position"`
	Status     bool       `json:"status"`
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`

	BannerImageURL        string     `json:"banner_image_url"`
	BannerLinkURL         string     `json:"banner_link_url"`
	BannerLinkTarget      string     `json:"banner_link_target"`
	BadgeText             string     `json:"badge_text"`
	ShowImage             bool       `json:"show_image"`
	ShowContent           bool       `json:"show_content"`
	MediaOverlayColor     string     `json:"media_overlay_color"`
	MediaOverlayOpacity   int        `json:"media_overlay_opacity"`
	MediaBorderColor      string     `json:"media_border_color"`
	MediaBorderWidth      int        `json:"media_border_width"`
	FooterBackgroundColor string     `json:"footer_background_color"`
	FooterBorderColor     string     `json:"footer_border_color"`
	FooterBorderWidth     int        `json:"footer_border_width"`
	MediaBoxShadow        string     `json:"media_box_shadow"`
	FooterBoxShadow       string     `json:"footer_box_shadow"`
	CountdownEnabled      bool       `json:"countdown_enabled"`
	CountdownEndTime      *time.Time `json:"countdown_end_time"`
	CountdownTitle        string     `json:"countdown_title"`

	TitleColor        string `json:"title_color"`
	BackgroundColor   string `json:"background_color"`
	BorderColor       string `json:"border_color"`
	BorderWidth       int    `json:"border_width"`
	BorderTopColor    string `json:"border_top_color"`
	BorderTopWidth    int    `json:"border_top_width"`
	BorderRightColor  string `json:"border_right_color"`
	BorderRightWidth  int    `json:"border_right_width"`
	BorderBottomColor string `json:"border_bottom_color"`
	BorderBottomWidth int    `json:"border_bottom_width"`
	BorderLeftColor   string `json:"border_left_color"`
	BorderLeftWidth   int    `json:"border_left_width"`
	BoxShadow         string `json:"box_shadow"`
	BorderRadius      int    `json:"border_radius"`
	PopupWidth        int    `json:"popup_width"`
	PopupHeight       int    `json:"popup_height"`
	ShowOnDesktop     bool   `json:"show_on_desktop"`
	ShowOnTablet      bool   `json:"show_on_tablet"`
	ShowOnMobile      bool   `json:"show_on_mobile"`

	CtaText      string `json:"cta_text"`
	CtaURL       string `json:"cta_url"`
	CtaTarget    string `json:"cta_target"`
	CtaTextColor string `json:"cta_text_color"`
	CtaBgColor   string `json:"cta_bg_color"`

	CloseButtonChoice   string `json:"close_button_choice"`
	CloseButtonPosition string `json:"close_button_position"`
	CloseButtonColor    string `json:"close_button_color"`

	TriggerType             string `json:"trigger_type"`
	TriggerDelaySeconds     int    `json:"trigger_delay_seconds"`
	TriggerScrollPercentage int    `json:"trigger_scroll_percentage"`
	CooldownMinutes         int    `json:"cooldown_minutes"`

	CustomCSS    string `json:"custom_css"`
	CustomJS     string `json:"custom_js"`
	TargetPages  string `json:"target_pages"`
	ExcludePages string `json:"exclude_pages"`
}

func MapPopupToResponse(p Popup) PopupResponse {
	websiteIDs := make([]uint, 0, len(p.Websites))
	for _, website := range p.Websites {
		websiteIDs = append(websiteIDs, website.ID)
	}
	if len(websiteIDs) == 0 && p.WebsiteID != 0 {
		websiteIDs = append(websiteIDs, p.WebsiteID)
	}
	return PopupResponse{
		ID:         p.ID,
		WebsiteID:  p.WebsiteID,
		WebsiteIDs: websiteIDs,
		Title:      p.Title,
		Content:    p.Content,
		Position:   p.Position,
		Status:     p.Status,
		StartTime:  p.StartTime,
		EndTime:    p.EndTime,

		BannerImageURL:        p.BannerImageURL,
		BannerLinkURL:         p.BannerLinkURL,
		BannerLinkTarget:      p.BannerLinkTarget,
		BadgeText:             p.BadgeText,
		ShowImage:             p.ShowImage,
		ShowContent:           p.ShowContent,
		MediaOverlayColor:     p.MediaOverlayColor,
		MediaOverlayOpacity:   p.MediaOverlayOpacity,
		MediaBorderColor:      p.MediaBorderColor,
		MediaBorderWidth:      p.MediaBorderWidth,
		FooterBackgroundColor: p.FooterBackgroundColor,
		FooterBorderColor:     p.FooterBorderColor,
		FooterBorderWidth:     p.FooterBorderWidth,
		MediaBoxShadow:        p.MediaBoxShadow,
		FooterBoxShadow:       p.FooterBoxShadow,
		CountdownEnabled:      p.CountdownEnabled,
		CountdownEndTime:      p.CountdownEndTime,
		CountdownTitle:        p.CountdownTitle,

		TitleColor:      p.TitleColor,
		BackgroundColor: p.BackgroundColor,
		BorderColor:     p.BorderColor,
		BorderWidth:     p.BorderWidth,
		BorderTopColor:  p.BorderTopColor, BorderTopWidth: p.BorderTopWidth,
		BorderRightColor: p.BorderRightColor, BorderRightWidth: p.BorderRightWidth,
		BorderBottomColor: p.BorderBottomColor, BorderBottomWidth: p.BorderBottomWidth,
		BorderLeftColor: p.BorderLeftColor, BorderLeftWidth: p.BorderLeftWidth,
		BoxShadow:     p.BoxShadow,
		BorderRadius:  p.BorderRadius,
		PopupWidth:    p.PopupWidth,
		PopupHeight:   p.PopupHeight,
		ShowOnDesktop: p.ShowOnDesktop,
		ShowOnTablet:  p.ShowOnTablet,
		ShowOnMobile:  p.ShowOnMobile,

		CtaText:      p.CtaText,
		CtaURL:       p.CtaURL,
		CtaTarget:    p.CtaTarget,
		CtaTextColor: p.CtaTextColor,
		CtaBgColor:   p.CtaBgColor,

		CloseButtonChoice:   p.CloseButtonChoice,
		CloseButtonPosition: p.CloseButtonPosition,
		CloseButtonColor:    p.CloseButtonColor,

		TriggerType:             p.TriggerType,
		TriggerDelaySeconds:     p.TriggerDelaySeconds,
		TriggerScrollPercentage: p.TriggerScrollPercentage,
		CooldownMinutes:         p.CooldownMinutes,

		CustomCSS:    p.CustomCSS,
		CustomJS:     p.CustomJS,
		TargetPages:  p.TargetPages,
		ExcludePages: p.ExcludePages,
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

	BannerImageURL        string     `json:"banner_image_url"`
	BannerLinkURL         string     `json:"banner_link_url"`
	BannerLinkTarget      string     `json:"banner_link_target"`
	BadgeText             string     `json:"badge_text"`
	ShowImage             bool       `json:"show_image"`
	ShowContent           bool       `json:"show_content"`
	MediaOverlayColor     string     `json:"media_overlay_color"`
	MediaOverlayOpacity   int        `json:"media_overlay_opacity"`
	MediaBorderColor      string     `json:"media_border_color"`
	MediaBorderWidth      int        `json:"media_border_width"`
	FooterBackgroundColor string     `json:"footer_background_color"`
	FooterBorderColor     string     `json:"footer_border_color"`
	FooterBorderWidth     int        `json:"footer_border_width"`
	MediaBoxShadow        string     `json:"media_box_shadow"`
	FooterBoxShadow       string     `json:"footer_box_shadow"`
	CountdownEnabled      bool       `json:"countdown_enabled"`
	CountdownEndTime      *time.Time `json:"countdown_end_time"`
	CountdownTitle        string     `json:"countdown_title"`

	TitleColor        string `json:"title_color"`
	BackgroundColor   string `json:"background_color"`
	BorderColor       string `json:"border_color"`
	BorderWidth       int    `json:"border_width"`
	BorderTopColor    string `json:"border_top_color"`
	BorderTopWidth    int    `json:"border_top_width"`
	BorderRightColor  string `json:"border_right_color"`
	BorderRightWidth  int    `json:"border_right_width"`
	BorderBottomColor string `json:"border_bottom_color"`
	BorderBottomWidth int    `json:"border_bottom_width"`
	BorderLeftColor   string `json:"border_left_color"`
	BorderLeftWidth   int    `json:"border_left_width"`
	BoxShadow         string `json:"box_shadow"`
	BorderRadius      int    `json:"border_radius"`
	PopupWidth        int    `json:"popup_width"`
	PopupHeight       int    `json:"popup_height"`
	ShowOnDesktop     bool   `json:"show_on_desktop"`
	ShowOnTablet      bool   `json:"show_on_tablet"`
	ShowOnMobile      bool   `json:"show_on_mobile"`

	CtaText      string `json:"cta_text"`
	CtaURL       string `json:"cta_url"`
	CtaTarget    string `json:"cta_target"`
	CtaTextColor string `json:"cta_text_color"`
	CtaBgColor   string `json:"cta_bg_color"`

	CloseButtonChoice   string `json:"close_button_choice"`
	CloseButtonPosition string `json:"close_button_position"`
	CloseButtonColor    string `json:"close_button_color"`

	TriggerType             string `json:"trigger_type"`
	TriggerDelaySeconds     int    `json:"trigger_delay_seconds"`
	TriggerScrollPercentage int    `json:"trigger_scroll_percentage"`
	CooldownMinutes         int    `json:"cooldown_minutes"`

	CustomCSS    string `json:"custom_css"`
	CustomJS     string `json:"custom_js"`
	TargetPages  string `json:"target_pages"`
	ExcludePages string `json:"exclude_pages"`
}

func MapPopupToPublicResponse(p Popup) PublicPopupResponse {
	return PublicPopupResponse{
		ID:        p.ID,
		Title:     p.Title,
		Content:   p.Content,
		Position:  p.Position,
		StartTime: p.StartTime,
		EndTime:   p.EndTime,

		BannerImageURL:        p.BannerImageURL,
		BannerLinkURL:         p.BannerLinkURL,
		BannerLinkTarget:      p.BannerLinkTarget,
		BadgeText:             p.BadgeText,
		ShowImage:             p.ShowImage,
		ShowContent:           p.ShowContent,
		MediaOverlayColor:     p.MediaOverlayColor,
		MediaOverlayOpacity:   p.MediaOverlayOpacity,
		MediaBorderColor:      p.MediaBorderColor,
		MediaBorderWidth:      p.MediaBorderWidth,
		FooterBackgroundColor: p.FooterBackgroundColor,
		FooterBorderColor:     p.FooterBorderColor,
		FooterBorderWidth:     p.FooterBorderWidth,
		MediaBoxShadow:        p.MediaBoxShadow,
		FooterBoxShadow:       p.FooterBoxShadow,
		CountdownEnabled:      p.CountdownEnabled,
		CountdownEndTime:      p.CountdownEndTime,
		CountdownTitle:        p.CountdownTitle,

		TitleColor:      p.TitleColor,
		BackgroundColor: p.BackgroundColor,
		BorderColor:     p.BorderColor,
		BorderWidth:     p.BorderWidth,
		BorderTopColor:  p.BorderTopColor, BorderTopWidth: p.BorderTopWidth,
		BorderRightColor: p.BorderRightColor, BorderRightWidth: p.BorderRightWidth,
		BorderBottomColor: p.BorderBottomColor, BorderBottomWidth: p.BorderBottomWidth,
		BorderLeftColor: p.BorderLeftColor, BorderLeftWidth: p.BorderLeftWidth,
		BoxShadow:     p.BoxShadow,
		BorderRadius:  p.BorderRadius,
		PopupWidth:    p.PopupWidth,
		PopupHeight:   p.PopupHeight,
		ShowOnDesktop: p.ShowOnDesktop,
		ShowOnTablet:  p.ShowOnTablet,
		ShowOnMobile:  p.ShowOnMobile,

		CtaText:      p.CtaText,
		CtaURL:       p.CtaURL,
		CtaTarget:    p.CtaTarget,
		CtaTextColor: p.CtaTextColor,
		CtaBgColor:   p.CtaBgColor,

		CloseButtonChoice:   p.CloseButtonChoice,
		CloseButtonPosition: p.CloseButtonPosition,
		CloseButtonColor:    p.CloseButtonColor,

		TriggerType:             p.TriggerType,
		TriggerDelaySeconds:     p.TriggerDelaySeconds,
		TriggerScrollPercentage: p.TriggerScrollPercentage,
		CooldownMinutes:         p.CooldownMinutes,

		CustomCSS:    p.CustomCSS,
		CustomJS:     p.CustomJS,
		TargetPages:  p.TargetPages,
		ExcludePages: p.ExcludePages,
	}
}

func MapPopupsToPublicResponse(popups []Popup) []PublicPopupResponse {
	responses := make([]PublicPopupResponse, 0, len(popups))
	for _, p := range popups {
		responses = append(responses, MapPopupToPublicResponse(p))
	}
	return responses
}
