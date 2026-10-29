package popups

import (
	"net/url"
	"regexp"
	"strings"
	"time"

	"popup-manager-api/internal/common"
	"popup-manager-api/internal/websites"
)

type PopupService interface {
	Create(request CreatePopupRequest, createdBy uint) error
	GetAll(params common.QueryParams) ([]Popup, int64, error)
	GetByID(id uint) (*Popup, error)
	GetByWebsiteID(websiteID uint, params common.QueryParams) ([]Popup, int64, error)
	GetPublicPopups(websiteKey string) ([]Popup, error)
	Update(id uint, request UpdatePopupRequest) error
	Delete(id uint) error
}

type popupService struct {
	repository        PopupRepository
	websiteRepository websites.WebsiteRepository
}

func NewPopupService(repository PopupRepository, websiteRepository websites.WebsiteRepository) PopupService {
	return &popupService{
		repository:        repository,
		websiteRepository: websiteRepository,
	}
}

func isValidPosition(position string) bool {
	switch position {
	case PositionCenter, PositionTopBanner, PositionBottomLeft, PositionBottomRight,
		PositionTopLeft, PositionTopCenter, PositionTopRight,
		PositionMiddleLeft, PositionMiddleCenter, PositionMiddleRight,
		PositionBottomCenter, PositionTopBannerCenter, PositionBottomBannerCenter:
		return true
	default:
		return false
	}
}

func validateOptionalWebURL(value, field string, validation *common.ValidationError) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil {
		validation.Add(field, "must be an absolute HTTP or HTTPS URL")
	}
}

func validateLinkTarget(value, field string, validation *common.ValidationError) {
	if value != "" && value != "_blank" && value != "_self" {
		validation.Add(field, "must be _blank or _self")
	}
}

func validateTriggerSettings(triggerType string, delay, scroll, cooldown int, validation *common.ValidationError) {
	if triggerType != "" && triggerType != "immediate" && triggerType != "delay" && triggerType != "scroll" {
		validation.Add("trigger_type", "must be immediate, delay, or scroll")
	}
	if delay < 0 || delay > 300 {
		validation.Add("trigger_delay_seconds", "must be between 0 and 300")
	}
	if scroll < 0 || scroll > 100 {
		validation.Add("trigger_scroll_percentage", "must be between 0 and 100")
	}
	if cooldown < 0 || cooldown > 525600 {
		validation.Add("cooldown_minutes", "must be between 0 and 525600")
	}
}

func validatePopupSize(width, height int, validation *common.ValidationError) {
	if width != 0 && (width < 200 || width > 2000) {
		validation.Add("popup_width", "width must be between 200 and 2000 pixels, or 0 for automatic")
	}
	if height != 0 && (height < 200 || height > 2000) {
		validation.Add("popup_height", "height must be between 200 and 2000 pixels, or 0 for automatic")
	}
}

func normalizeTargetPath(value string) string {
	value = strings.TrimSpace(value)
	if index := strings.IndexAny(value, "?#"); index >= 0 {
		value = value[:index]
	}
	if value == "" {
		return "/"
	}
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	value = regexp.MustCompile(`/+`).ReplaceAllString(value, "/")
	if len(value) > 1 {
		value = strings.TrimRight(value, "/")
	}
	return value
}

func targetPathMatches(path, rule string) bool {
	path = normalizeTargetPath(path)
	rule = strings.TrimSpace(rule)
	if rule == "*" {
		return true
	}
	rule = normalizeTargetPath(rule)
	if strings.HasSuffix(rule, "/*") {
		prefix := strings.TrimRight(strings.TrimSuffix(rule, "/*"), "/")
		return path == prefix || strings.HasPrefix(path, prefix+"/")
	}
	return path == rule
}

// FilterPopupsByPagePath applies campaign page include/exclude rules to the
// current pathname. A blank path keeps the legacy all-pages delivery behavior.
func FilterPopupsByPagePath(items []Popup, pagePath string) []Popup {
	if strings.TrimSpace(pagePath) == "" {
		return items
	}
	path := normalizeTargetPath(pagePath)
	filtered := make([]Popup, 0, len(items))
	for _, popup := range items {
		exclusions := strings.FieldsFunc(popup.ExcludePages, func(r rune) bool { return r == ',' || r == '\n' || r == '\r' })
		excluded := false
		for _, rule := range exclusions {
			if targetPathMatches(path, rule) {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		targets := strings.FieldsFunc(popup.TargetPages, func(r rune) bool { return r == ',' || r == '\n' || r == '\r' })
		if len(targets) == 0 {
			filtered = append(filtered, popup)
			continue
		}
		matched := false
		for _, rule := range targets {
			if strings.EqualFold(strings.TrimSpace(rule), "all") || targetPathMatches(path, rule) {
				matched = true
				break
			}
		}
		if matched {
			filtered = append(filtered, popup)
		}
	}
	return filtered
}

func (s *popupService) Create(request CreatePopupRequest, createdBy uint) error {
	vErr := common.NewValidationError()

	title := strings.TrimSpace(request.Title)
	if title == "" {
		vErr.Add("title", "title cannot be empty")
	}

	content := strings.TrimSpace(request.Content)
	if content == "" {
		vErr.Add("content", "content cannot be empty")
	}

	if !isValidPosition(request.Position) {
		vErr.Add("position", "invalid popup position")
	}
	validateOptionalWebURL(request.BannerImageURL, "banner_image_url", vErr)
	validateOptionalWebURL(request.BannerLinkURL, "banner_link_url", vErr)
	validateOptionalWebURL(request.CtaURL, "cta_url", vErr)
	validateLinkTarget(request.BannerLinkTarget, "banner_link_target", vErr)
	validateLinkTarget(request.CtaTarget, "cta_target", vErr)
	validateTriggerSettings(request.TriggerType, request.TriggerDelaySeconds, request.TriggerScrollPercentage, request.CooldownMinutes, vErr)
	validatePopupSize(request.PopupWidth, request.PopupHeight, vErr)

	if request.StartTime != nil && request.EndTime != nil {
		if !request.StartTime.Before(*request.EndTime) {
			vErr.Add("start_time", "start time must be before end time")
		}
	}

	if vErr.HasErrors() {
		return vErr
	}
	showOnDesktop, showOnTablet, showOnMobile := request.ShowOnDesktop, request.ShowOnTablet, request.ShowOnMobile
	if !showOnDesktop && !showOnTablet && !showOnMobile {
		showOnDesktop, showOnTablet, showOnMobile = true, true, true
	}
	showImage, showContent := request.ShowImage, request.ShowContent
	if !showImage && !showContent {
		showImage, showContent = true, true
	}

	websiteIDs := request.WebsiteIDs
	if len(websiteIDs) == 0 && request.WebsiteID != 0 {
		websiteIDs = []uint{request.WebsiteID}
	}
	if len(websiteIDs) == 0 {
		vErr.Add("website_ids", "select at least one website")
		return vErr
	}
	selectedWebsites := make([]websites.Website, 0, len(websiteIDs))
	seen := map[uint]bool{}
	for _, websiteID := range websiteIDs {
		if websiteID == 0 || seen[websiteID] {
			continue
		}
		website, err := s.websiteRepository.GetByID(websiteID)
		if err != nil {
			return common.ErrNotFound
		}
		selectedWebsites = append(selectedWebsites, *website)
		seen[websiteID] = true
	}
	if len(selectedWebsites) == 0 {
		vErr.Add("website_ids", "select at least one website")
		return vErr
	}

	popup := Popup{
		WebsiteID: selectedWebsites[0].ID,
		Websites:  selectedWebsites,
		Title:     title,
		Content:   content,
		Position:  request.Position,
		Status:    request.Status,
		StartTime: request.StartTime,
		EndTime:   request.EndTime,

		BannerImageURL:        request.BannerImageURL,
		BannerLinkURL:         request.BannerLinkURL,
		BannerLinkTarget:      request.BannerLinkTarget,
		BadgeText:             request.BadgeText,
		ShowImage:             showImage,
		ShowContent:           showContent,
		MediaOverlayColor:     request.MediaOverlayColor,
		MediaOverlayOpacity:   max(0, min(100, request.MediaOverlayOpacity)),
		MediaBorderColor:      request.MediaBorderColor,
		MediaBorderWidth:      max(0, min(20, request.MediaBorderWidth)),
		FooterBackgroundColor: request.FooterBackgroundColor,
		FooterBorderColor:     request.FooterBorderColor,
		FooterBorderWidth:     max(0, min(20, request.FooterBorderWidth)),
		MediaBoxShadow:        request.MediaBoxShadow,
		FooterBoxShadow:       request.FooterBoxShadow,
		CountdownEnabled:      request.CountdownEnabled,
		CountdownEndTime:      request.CountdownEndTime,
		CountdownTitle:        request.CountdownTitle,

		TitleColor:      request.TitleColor,
		BackgroundColor: request.BackgroundColor,
		BorderColor:     request.BorderColor,
		BorderWidth:     request.BorderWidth,
		BorderTopColor:  request.BorderTopColor, BorderTopWidth: request.BorderTopWidth,
		BorderRightColor: request.BorderRightColor, BorderRightWidth: request.BorderRightWidth,
		BorderBottomColor: request.BorderBottomColor, BorderBottomWidth: request.BorderBottomWidth,
		BorderLeftColor: request.BorderLeftColor, BorderLeftWidth: request.BorderLeftWidth,
		BoxShadow:     request.BoxShadow,
		BorderRadius:  request.BorderRadius,
		PopupWidth:    request.PopupWidth,
		PopupHeight:   request.PopupHeight,
		ShowOnDesktop: showOnDesktop,
		ShowOnTablet:  showOnTablet,
		ShowOnMobile:  showOnMobile,

		CtaText:      request.CtaText,
		CtaURL:       request.CtaURL,
		CtaTarget:    request.CtaTarget,
		CtaTextColor: request.CtaTextColor,
		CtaBgColor:   request.CtaBgColor,

		CloseButtonChoice:   request.CloseButtonChoice,
		CloseButtonPosition: request.CloseButtonPosition,
		CloseButtonColor:    request.CloseButtonColor,

		TriggerType:             request.TriggerType,
		TriggerDelaySeconds:     request.TriggerDelaySeconds,
		TriggerScrollPercentage: request.TriggerScrollPercentage,
		CooldownMinutes:         request.CooldownMinutes,

		CustomCSS:    request.CustomCSS,
		CustomJS:     request.CustomJS,
		TargetPages:  request.TargetPages,
		ExcludePages: request.ExcludePages,
		CreatedBy:    createdBy,
	}

	return s.repository.Create(&popup)
}

func (s *popupService) GetAll(params common.QueryParams) ([]Popup, int64, error) {
	return s.repository.GetAll(params)
}

func (s *popupService) GetByID(id uint) (*Popup, error) {
	return s.repository.GetByID(id)
}

func (s *popupService) GetByWebsiteID(websiteID uint, params common.QueryParams) ([]Popup, int64, error) {
	return s.repository.GetByWebsiteID(websiteID, params)
}

func (s *popupService) GetPublicPopups(websiteKey string) ([]Popup, error) {
	websiteKey = strings.TrimSpace(websiteKey)
	if websiteKey == "" {
		return nil, common.ErrBadRequest
	}

	website, err := s.websiteRepository.GetByKey(websiteKey)
	if err != nil {
		return nil, err
	}

	if !website.Status {
		return nil, common.ErrNotFound
	}

	return s.repository.GetEligiblePublicByWebsite(website.ID, time.Now())
}

func (s *popupService) Update(id uint, request UpdatePopupRequest) error {
	popup, err := s.repository.GetByID(id)
	if err != nil {
		return err
	}

	vErr := common.NewValidationError()

	title := strings.TrimSpace(request.Title)
	if title == "" {
		vErr.Add("title", "title cannot be empty")
	}

	content := strings.TrimSpace(request.Content)
	if content == "" {
		vErr.Add("content", "content cannot be empty")
	}

	if !isValidPosition(request.Position) {
		vErr.Add("position", "invalid popup position")
	}
	validateOptionalWebURL(request.BannerImageURL, "banner_image_url", vErr)
	validateOptionalWebURL(request.BannerLinkURL, "banner_link_url", vErr)
	validateOptionalWebURL(request.CtaURL, "cta_url", vErr)
	validateLinkTarget(request.BannerLinkTarget, "banner_link_target", vErr)
	validateLinkTarget(request.CtaTarget, "cta_target", vErr)
	validateTriggerSettings(request.TriggerType, request.TriggerDelaySeconds, request.TriggerScrollPercentage, request.CooldownMinutes, vErr)
	validatePopupSize(request.PopupWidth, request.PopupHeight, vErr)

	if request.StartTime != nil && request.EndTime != nil {
		if !request.StartTime.Before(*request.EndTime) {
			vErr.Add("start_time", "start time must be before end time")
		}
	}

	if vErr.HasErrors() {
		return vErr
	}
	showOnDesktop, showOnTablet, showOnMobile := request.ShowOnDesktop, request.ShowOnTablet, request.ShowOnMobile
	if !showOnDesktop && !showOnTablet && !showOnMobile {
		showOnDesktop, showOnTablet, showOnMobile = true, true, true
	}
	showImage, showContent := request.ShowImage, request.ShowContent
	if !showImage && !showContent {
		showImage, showContent = true, true
	}

	if len(request.WebsiteIDs) > 0 || request.WebsiteID > 0 {
		websiteIDs := request.WebsiteIDs
		if len(websiteIDs) == 0 {
			websiteIDs = []uint{request.WebsiteID}
		}
		selectedWebsites := make([]websites.Website, 0, len(websiteIDs))
		seen := map[uint]bool{}
		for _, websiteID := range websiteIDs {
			if websiteID == 0 || seen[websiteID] {
				continue
			}
			website, err := s.websiteRepository.GetByID(websiteID)
			if err != nil {
				return common.ErrNotFound
			}
			selectedWebsites = append(selectedWebsites, *website)
			seen[websiteID] = true
		}
		if len(selectedWebsites) == 0 {
			vErr.Add("website_ids", "select at least one website")
			return vErr
		}
		popup.WebsiteID = selectedWebsites[0].ID
		popup.Websites = selectedWebsites
	}
	popup.Title = title
	popup.Content = content
	popup.Position = request.Position
	popup.Status = request.Status
	popup.StartTime = request.StartTime
	popup.EndTime = request.EndTime

	popup.BannerImageURL = request.BannerImageURL
	popup.BannerLinkURL = request.BannerLinkURL
	popup.BannerLinkTarget = request.BannerLinkTarget
	popup.BadgeText = request.BadgeText
	popup.ShowImage = showImage
	popup.ShowContent = showContent
	popup.MediaOverlayColor = request.MediaOverlayColor
	popup.MediaOverlayOpacity = max(0, min(100, request.MediaOverlayOpacity))
	popup.MediaBorderColor = request.MediaBorderColor
	popup.MediaBorderWidth = max(0, min(20, request.MediaBorderWidth))
	popup.FooterBackgroundColor = request.FooterBackgroundColor
	popup.FooterBorderColor = request.FooterBorderColor
	popup.FooterBorderWidth = max(0, min(20, request.FooterBorderWidth))
	popup.MediaBoxShadow = request.MediaBoxShadow
	popup.FooterBoxShadow = request.FooterBoxShadow
	popup.CountdownEnabled = request.CountdownEnabled
	popup.CountdownEndTime = request.CountdownEndTime
	popup.CountdownTitle = request.CountdownTitle

	popup.TitleColor = request.TitleColor
	popup.BackgroundColor = request.BackgroundColor
	popup.BorderColor = request.BorderColor
	popup.BorderWidth = request.BorderWidth
	popup.BorderTopColor, popup.BorderTopWidth = request.BorderTopColor, request.BorderTopWidth
	popup.BorderRightColor, popup.BorderRightWidth = request.BorderRightColor, request.BorderRightWidth
	popup.BorderBottomColor, popup.BorderBottomWidth = request.BorderBottomColor, request.BorderBottomWidth
	popup.BorderLeftColor, popup.BorderLeftWidth = request.BorderLeftColor, request.BorderLeftWidth
	popup.BoxShadow = request.BoxShadow
	popup.BorderRadius = request.BorderRadius
	popup.PopupWidth = request.PopupWidth
	popup.PopupHeight = request.PopupHeight
	popup.ShowOnDesktop = showOnDesktop
	popup.ShowOnTablet = showOnTablet
	popup.ShowOnMobile = showOnMobile

	popup.CtaText = request.CtaText
	popup.CtaURL = request.CtaURL
	popup.CtaTarget = request.CtaTarget
	popup.CtaTextColor = request.CtaTextColor
	popup.CtaBgColor = request.CtaBgColor

	popup.CloseButtonChoice = request.CloseButtonChoice
	popup.CloseButtonPosition = request.CloseButtonPosition
	popup.CloseButtonColor = request.CloseButtonColor

	popup.TriggerType = request.TriggerType
	popup.TriggerDelaySeconds = request.TriggerDelaySeconds
	popup.TriggerScrollPercentage = request.TriggerScrollPercentage
	popup.CooldownMinutes = request.CooldownMinutes

	popup.CustomCSS = request.CustomCSS
	popup.CustomJS = request.CustomJS
	popup.TargetPages = request.TargetPages
	popup.ExcludePages = request.ExcludePages

	return s.repository.Update(popup)
}

func (s *popupService) Delete(id uint) error {
	return s.repository.Delete(id)
}
