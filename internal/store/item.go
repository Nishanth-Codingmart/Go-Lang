package store

import "time"

type Item struct {
	Value     string
	ExpiresAt *time.Time
}

func (i *Item) IsExpired() bool {
	if i.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*i.ExpiresAt)
}
