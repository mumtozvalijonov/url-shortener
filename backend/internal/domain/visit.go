package domain

import "time"

type ShortURLVisited struct {
	ShortCode string
	VisitedAt time.Time
}
