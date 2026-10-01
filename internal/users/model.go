package users

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID uint `gorm:"primaryKey"`

	Name string `gorm:"type:varchar(100);not null"`

	Email string `gorm:"type:varchar(150);uniqueIndex;not null"`

	Password string `gorm:"type:varchar(255);not null"`

	Role string `gorm:"type:varchar(30);default:'admin'"`

	Status bool `gorm:"default:true"`

	CreatedAt time.Time

	UpdatedAt time.Time

	DeletedAt gorm.DeletedAt `gorm:"index"`
}
