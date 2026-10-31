package movieapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var ErrNotFound = errors.New("not found")

type Movie struct {
	ID int `json:"id"`
}

type MovieClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewMovieClient(url string) *MovieClient {
	return &MovieClient{baseURL: url, httpClient: &http.Client{}}
}

func (c *MovieClient) fetch(ctx context.Context, path string) ([]byte, error) {
	//full address
	fullURL := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)

	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)

	return data, err
}

func (c *MovieClient) GetMovie(ctx context.Context, id int) (*Movie, error) {
	path := fmt.Sprintf("/%d/info.0.json", id)

	data, err := c.fetch(ctx, path)
	if err != nil {
		return nil, err
	}

	var movie Movie
	if err = json.Unmarshal(data, &movie); err != nil {
		return nil, err
	}

	return &movie, nil

}

func (c *MovieClient) GetMovies(ctx context.Context, from, to int) ([]Movie, error) {
	var movies []Movie
	for id := from; id <= to; id++ {
		movie, err := c.GetMovie(ctx, id)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return movies, err
		}
		movies = append(movies, *movie)
	}

	return movies, nil
}

func (c *MovieClient) GetLastMovie(ctx context.Context) (*Movie, error) {
	low, high := 1, 1
	for {
		_, err := c.GetMovie(ctx, high)
		if err == ErrNotFound {
			break
		}
		if err != nil {
			return nil, err
		}
		low = high
		high *= 2
	}

	var lastFound *Movie
	for low <= high {
		mid := low + (high-low)/2
		movie, err := c.GetMovie(ctx, mid)

		if err == ErrNotFound {
			high = mid - 1
		} else if err != nil {
			return nil, err
		} else {
			lastFound = movie
			low = mid + 1
		}
	}
	return lastFound, nil
}
