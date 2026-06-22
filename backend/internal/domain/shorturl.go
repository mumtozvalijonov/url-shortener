package domain

import (
	"net/url"
	"time"
)

type ShortURL struct {
	ID        int
	TargetURL url.URL
	ShortCode string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ShortURLStatistics struct {
	ShortURL
	AccessCount int
}
