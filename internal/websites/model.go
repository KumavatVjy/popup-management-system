package websites

import "gorm.io/gorm"

const (
	PlatformWordPress = "WordPress"
	PlatformReact     = "React"
	PlatformPython    = "Python"
	PlatformLaravel   = "Laravel"
	PlatformJava      = "Java"
	PlatformVue       = "VueJS"
	PlatformGo        = "Go"
	PlatformNodeJS    = "NodeJS"
	PlatformHTML      = "HTML"
	PlatformOther     = "Other"
)

type Website struct {
	gorm.Model

	WebsiteName string `gorm:"size:150;not null" json:"website_name"`

	Domain string `gorm:"size:255;unique;not null" json:"domain"`

	Platform string `gorm:"size:50;not null" json:"platform"`

	WebsiteKey string `gorm:"size:64;unique;not null" json:"website_key"`

	Status bool `gorm:"default:true" json:"status"`

	CreatedBy uint `json:"created_by"`
}
