package popups

import (
	"testing"
)

func TestPositionValidation_All9Positions(t *testing.T) {
	validPositions := []string{
		// Legacy aliases
		PositionCenter,
		PositionTopBanner,
		PositionBottomLeft,
		PositionBottomRight,

		// 9-position standard grid
		PositionTopLeft,
		PositionTopCenter,
		PositionTopRight,
		PositionMiddleLeft,
		PositionMiddleCenter,
		PositionMiddleRight,
		PositionBottomLeft,
		PositionBottomCenter,
		PositionBottomRight,
	}

	for _, pos := range validPositions {
		if !isValidPosition(pos) {
			t.Errorf("expected position %q to be valid", pos)
		}
	}

	invalidPositions := []string{
		"",
		"top",
		"bottom",
		"middle",
		"left",
		"right",
		"floating",
		"random-invalid",
	}

	for _, pos := range invalidPositions {
		if isValidPosition(pos) {
			t.Errorf("expected position %q to be invalid", pos)
		}
	}
}

func TestPopupMapping_ExtendedFields(t *testing.T) {
	p := Popup{
		WebsiteID: 42,
		Title:     "Summer Sale",
		Content:   "<p>Big discounts</p>",
		Position:  PositionTopCenter,
		Status:    true,

		BannerImageURL:   "https://example.com/banner.png",
		BannerLinkURL:    "https://example.com/shop?promo=summer",
		BannerLinkTarget: "_blank",
		CountdownEnabled: true,
		CountdownTitle:   "Event Starts In:",

		TitleColor:      "#0f172a",
		BackgroundColor: "#ffffff",
		BorderColor:     "#e2e8f0",
		BorderWidth:     2,
		BorderRadius:    16,

		CtaText:      "Claim 50% Off",
		CtaURL:       "https://example.com/claim?promo=summer&utm_source=popup",
		CtaTarget:    "_blank",
		CtaTextColor: "#ffffff",
		CtaBgColor:   "#2563eb",

		CloseButtonChoice:   "circle",
		CloseButtonPosition: "top-right",
		CloseButtonColor:    "#64748b",

		TriggerType:             "scroll",
		TriggerDelaySeconds:     5,
		TriggerScrollPercentage: 50,
		CooldownMinutes:         1440,

		CustomCSS: "#popup-card { box-shadow: 0 8px 30px rgba(0,0,0,0.12); }",
	}
	p.ID = 101

	publicResp := MapPopupToPublicResponse(p)

	if publicResp.ID != 101 {
		t.Fatalf("expected ID 101, got %d", publicResp.ID)
	}
	if publicResp.Position != PositionTopCenter {
		t.Fatalf("expected Position %q, got %q", PositionTopCenter, publicResp.Position)
	}
	if publicResp.BannerImageURL != "https://example.com/banner.png" {
		t.Fatalf("expected BannerImageURL to match")
	}
	if publicResp.BannerLinkURL != "https://example.com/shop?promo=summer" {
		t.Fatalf("expected BannerLinkURL to match")
	}
	if !publicResp.CountdownEnabled || publicResp.CountdownTitle != "Event Starts In:" {
		t.Fatalf("expected Countdown fields to match")
	}
	if publicResp.CtaText != "Claim 50% Off" || publicResp.CtaURL != "https://example.com/claim?promo=summer&utm_source=popup" {
		t.Fatalf("expected CTA fields to match")
	}
	if publicResp.CloseButtonChoice != "circle" || publicResp.CloseButtonPosition != "top-right" {
		t.Fatalf("expected CloseButton fields to match")
	}
	if publicResp.TriggerType != "scroll" || publicResp.TriggerScrollPercentage != 50 {
		t.Fatalf("expected Trigger fields to match")
	}
	if publicResp.CooldownMinutes != 1440 {
		t.Fatalf("expected CooldownMinutes 1440, got %d", publicResp.CooldownMinutes)
	}
	if publicResp.CustomCSS != p.CustomCSS {
		t.Fatalf("expected CustomCSS to match")
	}
}
