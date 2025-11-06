package db

import (
	"github.com/google/uuid"
	"time"
)

type Post struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null"`
	Title       string     `gorm:"type:text;not null"`
	Description *string    `gorm:"type:text"`
	Extension   string     `gorm:"type:char(4);not null"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `gorm:"default:null"`

	Boards []*Board `gorm:"many2many:board_posts;"`
	Tags   []*Tag   `gorm:"many2many:post_tags;"`
}

type Board struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null"`
	Name        string     `gorm:"type:text;not null"`
	Description *string    `gorm:"type:text"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	DeletedAt   *time.Time `gorm:"default:null"`

	Posts []*Post `gorm:"many2many:board_posts;"`
}

type BoardPost struct {
	BoardID uuid.UUID `gorm:"type:uuid;primaryKey"`
	PostID  uuid.UUID `gorm:"type:uuid;primaryKey"`
}

type Tag struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name string    `gorm:"type:text;unique;not null"`

	Posts []*Post `gorm:"many2many:post_tags"`
}

type PostTag struct {
	PostID uuid.UUID `gorm:"type:uuid;primaryKey"`
	TagID  uuid.UUID `gorm:"type:uuid;primaryKey"`
}
