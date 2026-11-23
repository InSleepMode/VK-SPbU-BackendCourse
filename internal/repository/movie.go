package repository

import (
	"context"
	"errors"

	"github.com/InSleepMode/VK-SPbU-BackendCourse/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MovieRepository struct {
	pool *pgxpool.Pool
}

func NewMovieRepository(pool *pgxpool.Pool) *MovieRepository {
	return &MovieRepository{pool: pool}
}

func (r *MovieRepository) GetMovie(ctx context.Context, id int) (domain.Movie, error) {
	var m domain.Movie
	query := `
		SELECT id, title, release_year, director
		FROM movies 
		WHERE id = $1;
	`

	err := r.pool.QueryRow(ctx, query, id).Scan(&m.ID, &m.Title, &m.Year, &m.Director)
	if errors.Is(err, pgx.ErrNoRows) {
		return m, domain.ErrNotFound
	}
	if err != nil {
		return domain.Movie{}, err
	}
	return m, nil
}

func (r *MovieRepository) GetMaxID(ctx context.Context) (int, error) {
	var movieID int

	query := `
		SELECT COALESCE(MAX(id), 0)
		FROM movies
	`

	err := r.pool.QueryRow(ctx, query).Scan(&movieID)
	if err != nil {
		return 0, err
	}
	return movieID, nil

}

func (r *MovieRepository) SaveMovie(ctx context.Context, m domain.Movie) error {
	query := `
		INSERT INTO movies (id, title, release_year, director)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, m.ID, m.Title, m.Year, m.Director)
	if err != nil {
		return err
	}
	return nil
}

func (r *MovieRepository) GetMoviesByIDs(ctx context.Context, ids []int) ([]domain.Movie, error) {
	if len(ids) == 0 {
		return []domain.Movie{}, nil
	}

	query := `
		SELECT id, title, release_year, director
		FROM movies
		WHERE id = ANY($1);
	`

	rows, err := r.pool.Query(ctx, query, ids)
	if err != nil {
		return []domain.Movie{}, err
	}
	defer rows.Close()

	movies := make([]domain.Movie, 0, len(ids))
	for rows.Next() {
		var m domain.Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Year, &m.Director); err != nil {
			return nil, err
		}
		movies = append(movies, m)
	}

	return movies, rows.Err()
}

func (r *MovieRepository) GetAllMovies(ctx context.Context) ([]domain.Movie, error) {

	query := `
		SELECT id, title, release_year, director
		FROM movies;
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return []domain.Movie{}, err
	}
	defer rows.Close()

	movies := []domain.Movie{}
	for rows.Next() {
		var m domain.Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Year, &m.Director); err != nil {
			return nil, err
		}
		movies = append(movies, m)
	}

	return movies, rows.Err()
}
