package services

import (
	"context"
	"github.com/vlks-dev/EffectiveMobileGoTest/internal/models"
	"github.com/vlks-dev/EffectiveMobileGoTest/internal/repositories"
	"github.com/vlks-dev/EffectiveMobileGoTest/utils/pagination"
	"strings"
)

type SongService struct {
	storage repositories.SongStorage
}

func NewSongService(storage repositories.SongStorage) *SongService {
	return &SongService{storage: storage}
}

func (s *SongService) GetSongs(ctx context.Context, filters map[string]interface{}, page, limit string) ([]models.Song, error) {
	return s.storage.GetSongs(ctx, filters, page, limit)
}

func (s *SongService) GetSongText(ctx context.Context, id, page string) (string, error) {
	text, err := s.storage.GetSongText(ctx, id)
	if err != nil {
		return "", err
	}

	verses := pagination.SplitTextIntoVerses(text)
	pagedVerses, err := pagination.PaginateVerses(verses, page)
	if err != nil {
		return "", err
	}
	return strings.Join(pagedVerses, "\n\n"), nil
}

func (s *SongService) DeleteSong(ctx context.Context, id string) error {
	return s.storage.DeleteSong(ctx, id)
}

func (s *SongService) AddSong(ctx context.Context, req models.AddSong) error {
	return s.storage.AddSong(ctx, req)
}

func (s *SongService) UpdateSong(ctx context.Context, id string, song models.Song) error {
	return s.storage.UpdateSong(ctx, id, song)
}
