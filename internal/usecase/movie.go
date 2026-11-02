package usecase

import (
	"context"

	"github.com/InSleepMode/VK-SPbU-BackendCourse/internal/domain"
)

type MovieProvider interface {
	GetMovie(ctx context.Context, id int) (domain.Movie, error)
}

type MovieUseCase interface {
	LoadMovies(ctx context.Context, from, to int) ([]domain.Movie, error)
	GetLastMovie(ctx context.Context) (domain.Movie, error)
}

type movieUseCase struct {
	provider MovieProvider
}

// derendency injection
func NewMovieUseCase(p MovieProvider) MovieUseCase {
	return &movieUseCase{provider: p}
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
