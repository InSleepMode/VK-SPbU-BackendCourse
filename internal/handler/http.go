package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/InSleepMode/VK-SPbU-BackendCourse/internal/usecase"
)

type HTTPServer struct {
	uc usecase.MovieUseCase
}

func NewHTTPServer(uc usecase.MovieUseCase) *HTTPServer {
	return &HTTPServer{uc: uc}
}

// @Summary Получить последний фильм
// @Description Ищет и возвращает последний добавленный фильм на стенде
// @Tags movies
// @Produce json
// @Success 200 "Успешный ответ"
// @Router /last-movie [get]
func (s *HTTPServer) handleGetLastMovie(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	movie, err := s.uc.GetLastMovie(ctx)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(movie)
}

// @Summary Поиск фильмов
// @Description Ищет фильмы через инвертированный индекс в Redis
// @Tags movies
// @Produce json
// @Param q query string true "Поисковый запрос"
// @Success 200 "Успешный ответ"
// @Router /search [get]
func (s *HTTPServer) handleSearchMovies(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query().Get("q")

	if strings.TrimSpace(query) == "" {
		http.Error(w, "Empty search query", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	movies, err := s.uc.SearchMovies(ctx, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}

// @Summary Загрузить диапазон фильмов
// @Description Загружает фильмы с указанными ID (от from до to)
// @Tags movies
// @Produce json
// @Param from query int true "Начальный ID фильма"
// @Param to query int true "Конечный ID фильма"
// @Success 200 "Успешный ответ"
// @Router /load-movies [get]
func (s *HTTPServer) handleLoadMovies(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	from, err := strconv.Atoi(fromStr)
	if err != nil {
		http.Error(w, "invalid from", http.StatusBadRequest)
		return
	}
	to, err := strconv.Atoi(toStr)

	if err != nil {
		http.Error(w, "invalid to", http.StatusBadRequest)
		return
	}

	if from <= 0 || from > to {
		http.Error(w, "invalid range", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	movies, err := s.uc.LoadMovies(ctx, from, to)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(movies)
}

func (s *HTTPServer) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/last-movie", s.handleGetLastMovie)
	mux.HandleFunc("/load-movies", s.handleLoadMovies)
	mux.HandleFunc("/search", s.handleSearchMovies)
}
