package models

import "time"

type AddSong struct {
	Group string `json:"group"`
	Song  string `json:"song"`
}

type Song struct {
	Group       *string    `json:"group,omitempty"`
	Song        *string    `json:"song,omitempty"`
	ReleaseDate *time.Time `json:"release_date,omitempty"`
	Text        *string    `json:"text,omitempty"`
	Link        *string    `json:"link,omitempty"`
}
