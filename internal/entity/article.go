package entity

import (
	"time"
)

// DefaultTeamId default for field Article -> TeamID.
const DefaultTeamId = "t94"

// Article it's a full article entity.
type Article struct {
	ID          string    `bson:"id" json:"id"`
	TeamID      string    `bson:"teamId" json:"teamId"`
	ExternalId  int       `bson:"externalId" json:"-"`
	OptaMatchID *string   `bson:"optaMatchId,omitempty" json:"optaMatchId,omitempty"`
	Title       string    `bson:"title" json:"title"`
	Type        []string  `bson:"type" json:"type"`
	Teaser      string    `bson:"teaser" json:"teaser"`
	Content     string    `bson:"content" json:"content"`
	URL         string    `bson:"url" json:"url"`
	ImageURL    string    `bson:"imageUrl" json:"imageUrl"`
	GalleryUrls any       `bson:"galleryUrls,omitempty" json:"galleryUrls"`
	VideoURL    any       `bson:"videoUrl,omitempty" json:"videoUrl"`
	Published   time.Time `bson:"published" json:"published"`
}
