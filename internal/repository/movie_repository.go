package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/marchelhutagalung/go-service/internal/database"
	"github.com/marchelhutagalung/go-service/internal/models"
	"strings"
	"time"
)

var (
	ErrMovieNotFound = errors.New("movie not found")
)

type MovieRepository struct {
	db *database.PostgresDB
}

func NewMovieRepository(db *database.PostgresDB) *MovieRepository {
	return &MovieRepository{
		db: db,
	}
}

func (r *MovieRepository) Create(ctx context.Context, input *models.CreateMovieInput) (*models.Movie, error) {
	query := `
		INSERT INTO movies (title, description, release_date, rating, duration, genre, director, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		RETURNING id, title, description, release_date, rating, duration, genre, director, created_at, updated_at
	`

	now := time.Now()
	movie := &models.Movie{}

	err := r.db.QueryRowContext(
		ctx, query,
		input.Title, input.Description, input.ReleaseDate, input.Rating,
		input.Duration, input.Genre, input.Director, now,
	).Scan(
		&movie.ID, &movie.Title, &movie.Description, &movie.ReleaseDate,
		&movie.Rating, &movie.Duration, &movie.Genre, &movie.Director,
		&movie.CreatedAt, &movie.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return movie, nil
}

func (r *MovieRepository) GetByID(ctx context.Context, id int64) (*models.Movie, error) {
	query := `
		SELECT id, title, description, release_date, rating, duration, genre, director, created_at, updated_at
		FROM movies
		WHERE id = $1
	`

	movie := &models.Movie{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&movie.ID, &movie.Title, &movie.Description, &movie.ReleaseDate,
		&movie.Rating, &movie.Duration, &movie.Genre, &movie.Director,
		&movie.CreatedAt, &movie.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMovieNotFound
		}
		return nil, err
	}

	return movie, nil
}

func (r *MovieRepository) Update(ctx context.Context, id int64, input *models.UpdateMovieInput) (*models.Movie, error) {
	// Get current movie data
	movie, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Build update query dynamically
	updates := []string{}
	args := []interface{}{}
	argPosition := 1

	if input.Title != nil {
		updates = append(updates, fmt.Sprintf("title = $%d", argPosition))
		args = append(args, *input.Title)
		argPosition++
	}

	if input.Description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argPosition))
		args = append(args, *input.Description)
		argPosition++
	}

	if input.ReleaseDate != nil {
		updates = append(updates, fmt.Sprintf("release_date = $%d", argPosition))
		args = append(args, *input.ReleaseDate)
		argPosition++
	}

	if input.Rating != nil {
		updates = append(updates, fmt.Sprintf("rating = $%d", argPosition))
		args = append(args, *input.Rating)
		argPosition++
	}

	if input.Duration != nil {
		updates = append(updates, fmt.Sprintf("duration = $%d", argPosition))
		args = append(args, *input.Duration)
		argPosition++
	}

	if input.Genre != nil {
		updates = append(updates, fmt.Sprintf("genre = $%d", argPosition))
		args = append(args, *input.Genre)
		argPosition++
	}

	if input.Director != nil {
		updates = append(updates, fmt.Sprintf("director = $%d", argPosition))
		args = append(args, *input.Director)
		argPosition++
	}

	// Add updated_at
	updates = append(updates, fmt.Sprintf("updated_at = $%d", argPosition))
	args = append(args, time.Now())
	argPosition++

	// Add ID to args
	args = append(args, id)

	// If no updates, return the movie
	if len(updates) == 1 { // Only updated_at
		return movie, nil
	}

	// Build and execute query
	query := fmt.Sprintf(`
		UPDATE movies
		SET %s
		WHERE id = $%d
		RETURNING id, title, description, release_date, rating, duration, genre, director, created_at, updated_at
	`, strings.Join(updates, ", "), argPosition)

	updatedMovie := &models.Movie{}
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&updatedMovie.ID, &updatedMovie.Title, &updatedMovie.Description, &updatedMovie.ReleaseDate,
		&updatedMovie.Rating, &updatedMovie.Duration, &updatedMovie.Genre, &updatedMovie.Director,
		&updatedMovie.CreatedAt, &updatedMovie.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return updatedMovie, nil
}

func (r *MovieRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM movies WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrMovieNotFound
	}

	return nil
}

func (r *MovieRepository) List(ctx context.Context, query *models.MovieQuery) ([]*models.Movie, int, error) {
	// Build the query
	countQuery := `SELECT COUNT(*) FROM movies WHERE 1=1`
	selectQuery := `
		SELECT id, title, description, release_date, rating, duration, genre, director, created_at, updated_at
		FROM movies
		WHERE 1=1
	`

	// Add filters
	args := []interface{}{}
	argPosition := 1
	whereClause := ""

	if query.Title != "" {
		whereClause += fmt.Sprintf(" AND title ILIKE $%d", argPosition)
		args = append(args, "%"+query.Title+"%")
		argPosition++
	}

	if query.Genre != "" {
		whereClause += fmt.Sprintf(" AND genre ILIKE $%d", argPosition)
		args = append(args, "%"+query.Genre+"%")
		argPosition++
	}

	if query.Director != "" {
		whereClause += fmt.Sprintf(" AND director ILIKE $%d", argPosition)
		args = append(args, "%"+query.Director+"%")
		argPosition++
	}

	countQuery += whereClause
	selectQuery += whereClause

	// Add sorting
	if query.SortBy != "" {
		orderDir := "ASC"
		if strings.ToUpper(query.Order) == "DESC" {
			orderDir = "DESC"
		}

		// Validate sort column to prevent SQL injection
		allowedColumns := map[string]bool{
			"id":           true,
			"title":        true,
			"release_date": true,
			"rating":       true,
			"duration":     true,
			"genre":        true,
			"director":     true,
			"created_at":   true,
		}

		if allowedColumns[query.SortBy] {
			selectQuery += fmt.Sprintf(" ORDER BY %s %s", query.SortBy, orderDir)
		} else {
			selectQuery += " ORDER BY created_at DESC"
		}
	} else {
		selectQuery += " ORDER BY created_at DESC"
	}

	// Add pagination
	if query.Page <= 0 {
		query.Page = 1
	}

	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	offset := (query.Page - 1) * query.PageSize
	selectQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPosition, argPosition+1)
	args = append(args, query.PageSize, offset)

	// Get total count
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args[:argPosition-1]...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Get movies
	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	movies := []*models.Movie{}
	for rows.Next() {
		movie := &models.Movie{}
		err := rows.Scan(
			&movie.ID, &movie.Title, &movie.Description, &movie.ReleaseDate,
			&movie.Rating, &movie.Duration, &movie.Genre, &movie.Director,
			&movie.CreatedAt, &movie.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		movies = append(movies, movie)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return movies, totalCount, nil
}
