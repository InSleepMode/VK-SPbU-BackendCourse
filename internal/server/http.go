package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/InSleepMode/VK-SPbU-BackendCourse/internal/movieapi"
)

type Server struct {
	client *movieapi.MovieClient
}

func NewServer(client *movieapi.MovieClient) *Server {
	return &Server{client: client}
}

// @Summary Получить последний фильм
// @Description Ищет и возвращает последний добавленный фильм на стенде
// @Tags movies
// @Produce json
// @Success 200 "Успешный ответ"
// @Router /last-movie [get]
func (s *Server) handleGetLastMovie(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	movie, err := s.client.GetLastMovie(ctx)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(movie)
}

// @Summary Загрузить диапазон фильмов
// @Description Загружает фильмы с указанными ID (от from до to)
// @Tags movies
// @Produce json
// @Param from query int true "Начальный ID фильма"
// @Param to query int true "Конечный ID фильма"
// @Success 200 "Успешный ответ"
// @Router /load-movies [get]
func (s *Server) handleLoadMovies(w http.ResponseWriter, r *http.Request) {
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

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	movies, err := s.client.GetMovies(ctx, from, to)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(movies)
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/last-movie", s.handleGetLastMovie)
	mux.HandleFunc("/load-movies", s.handleLoadMovies)
}
