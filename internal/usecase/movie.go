package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/InSleepMode/VK-SPbU-BackendCourse/internal/domain"
)

type MovieProvider interface {
	GetMovie(ctx context.Context, id int) (domain.Movie, error)
	GetAllMovies(ctx context.Context) ([]domain.Movie, error)
	GetMoviesByIDs(ctx context.Context, ids []int) ([]domain.Movie, error)
}

type MovieUseCase interface {
	LoadMovies(ctx context.Context, from, to int) ([]domain.Movie, error)
	GetLastMovie(ctx context.Context) (domain.Movie, error)
	WarmUpIndex(ctx context.Context) error
	SearchMovies(ctx context.Context, query string) ([]domain.Movie, error)
}

type SearchProvider interface {
	BuildIndex(ctx context.Context, movies []domain.Movie) error
	SearchIDs(ctx context.Context, query string) ([]int, error)
	SetCachedSearch(ctx context.Context, query string, movies []domain.Movie) error
	GetCachedSearch(ctx context.Context, query string) ([]domain.Movie, error)
}

type movieUseCase struct {
	provider MovieProvider
	search   SearchProvider
}

// derendency injection
func NewMovieUseCase(p MovieProvider, s SearchProvider) MovieUseCase {
	return &movieUseCase{
		provider: p,
		search:   s,
	}
}

func (u *movieUseCase) LoadMovies(ctx context.Context, from, to int) ([]domain.Movie, error) {
	if from <= 0 || from > to {
		return nil, domain.ErrInvalidRange
	}

	var movies []domain.Movie
	for id := from; id <= to; id++ {
		movie, err := u.provider.GetMovie(ctx, id)
		if err == domain.ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}

	return movies, nil
}

func (u *movieUseCase) GetLastMovie(ctx context.Context) (domain.Movie, error) {
	low, high := 1, 1
	for {
		_, err := u.provider.GetMovie(ctx, high)
		if err == domain.ErrNotFound {
			break
		}
		if err != nil {
			return domain.Movie{}, err
		}
		low = high
		high *= 2
	}

	var lastFound domain.Movie
	for low <= high {
		mid := low + (high-low)/2
		movie, err := u.provider.GetMovie(ctx, mid)

		if err == domain.ErrNotFound {
			high = mid - 1
		} else if err != nil {
			return domain.Movie{}, err
		} else {
			lastFound = movie
			low = mid + 1
		}
	}
	return lastFound, nil
}

func (u *movieUseCase) WarmUpIndex(ctx context.Context) error {
	movies, err := u.provider.GetAllMovies(ctx)
	if err != nil {
		return err
	}

	return u.search.BuildIndex(ctx, movies)
}

func (u *movieUseCase) SearchMovies(ctx context.Context, query string) ([]domain.Movie, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("empty search query")
	}
	cached, err := u.search.GetCachedSearch(ctx, query)
	if err != nil {
		return nil, err
	}
	if cached != nil {
		return cached, nil
	}

	movies := []domain.Movie{}
	ids, err := u.search.SearchIDs(ctx, query)
	if err != nil {
		return nil, err
	}
	//если фильмов нет
	if len(ids) == 0 {
		_ = u.search.SetCachedSearch(ctx, query, movies)
		return movies, nil
	}
	movies, err = u.provider.GetMoviesByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	_ = u.search.SetCachedSearch(ctx, query, movies)
	return movies, nil
}
