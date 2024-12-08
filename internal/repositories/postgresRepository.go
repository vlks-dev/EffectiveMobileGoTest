package repositories

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlks-dev/EffectiveMobileGoTest/internal/models"
	"github.com/vlks-dev/EffectiveMobileGoTest/utils/dbutil"
	"log/slog"
	"strconv"
)

type SongRepository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewPostgresRepository(pool *pgxpool.Pool, logger *slog.Logger) *SongRepository {
	return &SongRepository{pool,
		logger}
}

type SongStorage interface {
	GetSongs(ctx context.Context, filters map[string]interface{}, page, limit string) ([]models.Song, error)
	GetSongText(ctx context.Context, id string) (string, error)
	DeleteSong(ctx context.Context, id string) error
	AddSong(ctx context.Context, req models.AddSong) error
	UpdateSong(ctx context.Context, id string, song models.Song) error
}

// GetSongs Получение списка песен с фильтрацией и пагинацией
func (r *SongRepository) GetSongs(ctx context.Context, filters map[string]interface{}, page, limit string) ([]models.Song, error) {
	queryBuilder := dbutil.NewQueryBuilder("SELECT group_name, song_name, release_date, text, link FROM songs")

	for field, value := range filters {
		if value == "" {
			continue
		}

		if field == "release_date" {
			queryBuilder.AddFilter(field, "=", value)
		} else {
			queryBuilder.AddFilter(field, "ILIKE", "%"+value.(string)+"%")
		}
	}

	l, _ := strconv.Atoi(limit)
	p, _ := strconv.Atoi(page)

	queryBuilder.SetPagination(l, (p-1)*l)

	query, args := queryBuilder.Build()

	r.logger.Debug("Getting songs", "query", query, "filters", filters)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to run get query", "error", err.Error())
		return nil, err
	}
	defer rows.Close()

	var songs []models.Song
	for rows.Next() {
		var song models.Song
		if err := rows.Scan(
			&song.Group,
			&song.Song,
			&song.ReleaseDate,
			&song.Text,
			&song.Link,
		); err != nil {
			r.logger.Error("failed to scan row", "error", err.Error())
			return nil, err
		}
		songs = append(songs, song)
	}
	r.logger.Debug("Listing songs", "count", len(songs))
	return songs, nil
}

// GetSongText Получение текста песни
func (r *SongRepository) GetSongText(ctx context.Context, id string) (string, error) {
	query := "SELECT text FROM songs WHERE id = $1"
	var text sql.NullString
	err := r.pool.QueryRow(ctx, query, id).Scan(&text)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("no songs in results set", "id", id)
			return "", models.SongNotFound
		}

		r.logger.Error("failed to get song text", "error", err.Error())
		return "", err
	}
	if text.String == "" {
		r.logger.Warn("no text found for this song", "id", id)
		return "", models.NoTextFound
	}
	r.logger.Debug("getting song text", "query", query, "id", id)
	return text.String, nil
}

// DeleteSong Удаление песни
func (r *SongRepository) DeleteSong(ctx context.Context, id string) error {
	commandTag, err := r.pool.Exec(ctx, "DELETE FROM songs WHERE id = $1", id)
	if err != nil {
		r.logger.Error("failed to delete song", "error", err.Error())
		return err
	}
	if commandTag.RowsAffected() == 0 {
		r.logger.Warn("failed to delete song", "id", id, "error", models.SongNotFound.Error())
		return models.SongNotFound
	}
	r.logger.Debug("deleted song", "query", commandTag.String(), "id", id)
	return nil
}

// AddSong Добавление новой песни
func (r *SongRepository) AddSong(ctx context.Context, req models.AddSong) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO songs (group_name, song_name)
		VALUES ($1, $2)`,
		req.Group, req.Song,
	)
	if err != nil {
		r.logger.Error("failed to add song", "error", err.Error())
	}
	r.logger.Debug("added song", "request", req)
	return err
}

// UpdateSong Обновление данных песни
func (r *SongRepository) UpdateSong(ctx context.Context, id string, song models.Song) error {
	// Мапа для хранения обновляемых полей и их значений
	fields := make(map[string]interface{})

	if song.Group != nil {
		fields["group_name"] = *song.Group
	}
	if song.Song != nil {
		fields["song_name"] = *song.Song
	}
	if song.ReleaseDate != nil {
		fields["release_date"] = *song.ReleaseDate
	}
	if song.Text != nil {
		fields["text"] = *song.Text
	}
	if song.Link != nil {
		fields["link"] = *song.Link
	}

	if len(fields) == 0 {
		r.logger.Warn("no fields to update", "id", id)
		return models.NothingToUpdate
	}

	// Динамически строим SQL-запрос
	query := "UPDATE songs SET "
	var params []interface{}
	i := 1

	for field, value := range fields {
		if i > 1 {
			query += ", "
		}
		query += field + " = $" + strconv.Itoa(i)
		params = append(params, value)
		i++
	}

	query += " WHERE id = $" + strconv.Itoa(i)
	params = append(params, id)
	r.logger.Debug("Update song", "query", query, "params", params)

	// Выполняем запрос
	commandTag, err := r.pool.Exec(ctx, query, params...)
	if err != nil {
		r.logger.Error("failed to update song", "error", err.Error())
		return err
	}
	if commandTag.RowsAffected() == 0 {
		r.logger.Warn("no to update song", "id", id)
		return models.SongNotFound
	}

	return nil
}
