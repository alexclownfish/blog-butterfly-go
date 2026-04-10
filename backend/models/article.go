package models

import "time"

type Article struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Title      string    `json:"title" gorm:"size:200;not null"`
	Content    string    `json:"content" gorm:"type:longtext"`
	Summary    string    `json:"summary" gorm:"size:500"`
	CoverImage string    `json:"cover_image" gorm:"size:500"`
	CategoryID uint      `json:"category_id"`
	Category   *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Tags       string    `json:"tags" gorm:"size:200"`
	IsTop      bool      `json:"is_top" gorm:"default:false"`
	Status     string    `json:"status" gorm:"size:20;default:'published'"`
	Views      int       `json:"views" gorm:"default:0"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Category struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"size:50;unique"`
}

type Tag struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"size:50;unique"`
}

type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"size:50;unique"`
	Password string `json:"-" gorm:"size:200"`
}
