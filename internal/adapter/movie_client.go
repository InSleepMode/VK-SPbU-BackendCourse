package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/InSleepMode/VK-SPbU-BackendCourse/internal/domain"
)

type HTTPMovieClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPMovieClient(url string) *HTTPMovieClient {
	return &HTTPMovieClient{baseURL: url, httpClient: &http.Client{}}
}

func (c *HTTPMovieClient) GetMovie(ctx context.Context, id int) (domain.Movie, error) {
	path := fmt.Sprintf("%s/%d/info.0.json", c.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return domain.Movie{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return domain.Movie{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return domain.Movie{}, domain.ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return domain.Movie{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var movie domain.Movie
	if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
		return domain.Movie{}, err
	}

	return movie, nil
}
