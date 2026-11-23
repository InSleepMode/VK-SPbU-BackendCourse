package domain

import "errors"

var ErrNotFound = errors.New("not found")
var ErrInvalidRange = errors.New("invalid movie range: from must be more than 0 and less than to")

type Movie struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Year        int      `json:"year"`
	Director    string   `json:"director"`
	Description string   `json:"description"`
	Actors      []string `json:"actors"`
	Genres      []string `json:"genres"`
}
