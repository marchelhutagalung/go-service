package models

import "time"

type Movie struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ReleaseDate time.Time `json:"release_date"`
	Rating      float64   `json:"rating"`
	Duration    int       `json:"duration"` // in minutes
	Genre       string    `json:"genre"`
	Director    string    `json:"director"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateMovieInput struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ReleaseDate time.Time `json:"release_date"`
	Rating      float64   `json:"rating"`
	Duration    int       `json:"duration"`
	Genre       string    `json:"genre"`
	Director    string    `json:"director"`
}

type UpdateMovieInput struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	ReleaseDate *time.Time `json:"release_date"`
	Rating      *float64   `json:"rating"`
	Duration    *int       `json:"duration"`
	Genre       *string    `json:"genre"`
	Director    *string    `json:"director"`
}

type MovieQuery struct {
	Title    string `json:"title"`
	Genre    string `json:"genre"`
	Director string `json:"director"`
	SortBy   string `json:"sort_by"`
	Order    string `json:"order"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}
