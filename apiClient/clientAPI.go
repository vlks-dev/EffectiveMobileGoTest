package apiClient

import (
	"encoding/json"
	"fmt"
	"github.com/vlks-dev/EffectiveMobileGoTest/internal/models"
	"net/http"
	"net/url"
)

type APIClient interface {
	GetSongDetails(group, song string) (*models.Song, error)
}

type ExternalAPIClient struct {
	baseURL string
}

func (c *ExternalAPIClient) GetSongDetails(group, song *string) (*models.Song, error) {

	resp, err := http.Get(fmt.Sprintf("%s/info?group=%s&song=%s", c.baseURL, url.QueryEscape(*group), url.QueryEscape(*song)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var detail models.SongDetail
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		return nil, err
	}

	result := &models.Song{
		Group:      group,
		Song:       song,
		SongDetail: detail,
	}

	return result, nil
}
