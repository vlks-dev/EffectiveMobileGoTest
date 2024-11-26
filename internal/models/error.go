package models

import "errors"

var (
	NothingToUpdate = errors.New("nothing to update")
	SongNotFound    = errors.New("song not found")
	NoTextFound     = errors.New("no text found")
)
