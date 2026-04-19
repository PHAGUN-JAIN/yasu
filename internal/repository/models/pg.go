package models

import "time"

type ShortURL struct {
	Code        string
	OriginalURL string
	CreatedAt   time.Time
}
