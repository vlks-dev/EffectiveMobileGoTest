package handlers

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	_ "github.com/swaggo/swag/example/celler/model"
	"github.com/vlks-dev/EffectiveMobileGoTest/internal/models"
	"github.com/vlks-dev/EffectiveMobileGoTest/internal/services"
	"net/http"
	"time"
)

type SongHandler struct {
	service *services.SongService
}

func NewSongHandler(service *services.SongService) *SongHandler {
	return &SongHandler{service: service}
}

// @Description Retrieve a list of songs with optional filters
// @Tags Songs
// @Accept json
// @Produce json
// @Param group query string false "Filter by group name"
// @Param song query string false "Filter by song name"
// @Param release_date query string false "Filter by release date"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Success 200 {array} models.Song "List of songs"
// @Failure 400
// @Failure 500
// @Router /songs [get]
func (h *SongHandler) GetSongs(c *gin.Context) {
	group := c.Query("group")
	song := c.Query("song")
	releaseDate := c.Query("release_date")
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	timeout, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	filters := make(map[string]interface{})
	filters["group_name"] = group
	filters["song_name"] = song
	filters["release_date"] = releaseDate

	songs, err := h.service.GetSongs(timeout, filters, page, limit)
	if songs == nil {
		c.JSON(http.StatusOK, gin.H{"no songs on such filters": filters})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, songs)
}

// @Summary Get song text
// @Description Retrieve the text of a song by its ID, with optional pagination
// @Tags Songs
// @Accept json
// @Produce json
// @Param id path string true "Song ID"
// @Param page query int false "Page number for pagination" default(1)
// @Success 200
// @Failure 400
// @Failure 500
// @Router /songs/{id}/text [get]
func (h *SongHandler) GetSongText(c *gin.Context) {
	id := c.Param("id")
	page := c.DefaultQuery("page", "1")

	timeout, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	verses, err := h.service.GetSongText(timeout, id, page)
	if err != nil {
		if errors.Is(err, models.NoTextFound) {
			c.JSON(http.StatusOK, gin.H{"no text found for song with id": id})
			return
		}
		if errors.Is(err, models.SongNotFound) {
			c.JSON(http.StatusOK, gin.H{"song not found with id": id})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, verses)
}

// @Summary Delete a song
// @Description Delete a song by its ID
// @Tags Songs
// @Accept json
// @Produce json
// @Param id path string true "Song ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400
// @Failure 500
// @Router /songs/{id} [delete]
func (h *SongHandler) DeleteSong(c *gin.Context) {
	id := c.Param("id")

	timeout, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.service.DeleteSong(timeout, id); err != nil {
		if errors.Is(err, models.SongNotFound) {
			c.JSON(http.StatusOK, gin.H{"message": "song not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Song deleted"})
}

// @Summary Add a song
// @Description Add a new song with group and title
// @Tags Songs
// @Accept json
// @Produce json
// @Param song body models.AddSong true "Song details"
// @Success 201 {object} map[string]string "Success message"
// @Failure 400
// @Failure 500
// @Router /songs [post]
func (h *SongHandler) AddSong(c *gin.Context) {
	timeout, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var req models.AddSong
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddSong(timeout, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Song added"})
}

// @Summary Update a song
// @Description Update the details of a song by its ID
// @Tags Songs
// @Accept json
// @Produce json
// @Param id path string true "Song ID"
// @Param song body models.Song true "Updated song details"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400
// @Failure 500
// @Router /songs/{id} [patch]
func (h *SongHandler) UpdateSong(c *gin.Context) {
	id := c.Param("id")

	timeout, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var updatedSong models.Song
	if err := c.ShouldBindJSON(&updatedSong); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateSong(timeout, id, updatedSong); err != nil {
		if errors.Is(err, models.NothingToUpdate) {
			c.JSON(http.StatusOK, gin.H{"message": "Nothing to update"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Song updated"})
}
