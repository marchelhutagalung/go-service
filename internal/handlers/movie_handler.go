package handlers

import (
	"encoding/json"
	"errors"
	"github.com/marchelhutagalung/go-service/internal/logger"
	"github.com/marchelhutagalung/go-service/internal/models"
	"github.com/marchelhutagalung/go-service/internal/repository"
	"github.com/marchelhutagalung/go-service/internal/response"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type MovieHandler struct {
	movieRepo *repository.MovieRepository
}

func NewMovieHandler(movieRepo *repository.MovieRepository) *MovieHandler {
	return &MovieHandler{
		movieRepo: movieRepo,
	}
}

type PaginatedMovieResponse struct {
	Movies     []*models.Movie `json:"movies"`
	TotalCount int             `json:"total_count"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

func (h *MovieHandler) CreateMovie(w http.ResponseWriter, r *http.Request) {
	var input models.CreateMovieInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.Error("Invalid request body", logger.Field("error", err))
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	movie, err := h.movieRepo.Create(r.Context(), &input)
	if err != nil {
		logger.Error("Error creating movie", logger.Field("error", err))
		response.ErrorResponse(w, http.StatusInternalServerError, "Error creating movie")
		return
	}

	logger.Info("Movie created", logger.Field("movie_id", movie.ID), logger.Field("title", movie.Title))
	response.SuccessResponse(w, http.StatusCreated, "Movie created successfully", movie)
}

func (h *MovieHandler) GetMovie(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("Invalid movie ID", logger.Field("id", idStr), logger.Field("error", err))
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid movie ID")
		return
	}

	movie, err := h.movieRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrMovieNotFound) {
			logger.Error("Movie not found", logger.Field("movie_id", id))
			response.ErrorResponse(w, http.StatusNotFound, "Movie not found")
			return
		}
		logger.Error("Error getting movie", logger.Field("error", err), logger.Field("movie_id", id))
		response.ErrorResponse(w, http.StatusInternalServerError, "Error getting movie")
		return
	}

	logger.Info("Movie retrieved", logger.Field("movie_id", movie.ID), logger.Field("title", movie.Title))
	response.SuccessResponse(w, http.StatusOK, "Movie retrieved successfully", movie)
}

func (h *MovieHandler) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("Invalid movie ID", logger.Field("id", idStr), logger.Field("error", err))
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid movie ID")
		return
	}

	var input models.UpdateMovieInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.Error("Invalid request body", logger.Field("error", err))
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	movie, err := h.movieRepo.Update(r.Context(), id, &input)
	if err != nil {
		if errors.Is(err, repository.ErrMovieNotFound) {
			logger.Error("Movie not found", logger.Field("movie_id", id))
			response.ErrorResponse(w, http.StatusNotFound, "Movie not found")
			return
		}
		logger.Error("Error updating movie", logger.Field("error", err), logger.Field("movie_id", id))
		response.ErrorResponse(w, http.StatusInternalServerError, "Error updating movie")
		return
	}

	logger.Info("Movie updated", logger.Field("movie_id", movie.ID), logger.Field("title", movie.Title))
	response.SuccessResponse(w, http.StatusOK, "Movie updated successfully", movie)
}

func (h *MovieHandler) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("Invalid movie ID", logger.Field("id", idStr), logger.Field("error", err))
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid movie ID")
		return
	}

	err = h.movieRepo.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrMovieNotFound) {
			logger.Error("Movie not found", logger.Field("movie_id", id))
			response.ErrorResponse(w, http.StatusNotFound, "Movie not found")
			return
		}
		logger.Error("Error deleting movie", logger.Field("error", err), logger.Field("movie_id", id))
		response.ErrorResponse(w, http.StatusInternalServerError, "Error deleting movie")
		return
	}

	logger.Info("Movie deleted", logger.Field("movie_id", id))
	response.SuccessResponse(w, http.StatusOK, "Movie deleted successfully", nil)
}

func (h *MovieHandler) ListMovies(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := &models.MovieQuery{
		Title:    r.URL.Query().Get("title"),
		Genre:    r.URL.Query().Get("genre"),
		Director: r.URL.Query().Get("director"),
		SortBy:   r.URL.Query().Get("sort_by"),
		Order:    r.URL.Query().Get("order"),
	}

	// Parse pagination parameters
	if page := r.URL.Query().Get("page"); page != "" {
		pageInt, err := strconv.Atoi(page)
		if err == nil && pageInt > 0 {
			query.Page = pageInt
		}
	}

	if pageSize := r.URL.Query().Get("page_size"); pageSize != "" {
		pageSizeInt, err := strconv.Atoi(pageSize)
		if err == nil && pageSizeInt > 0 {
			query.PageSize = pageSizeInt
		}
	}

	// Default values
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	movies, totalCount, err := h.movieRepo.List(r.Context(), query)
	if err != nil {
		logger.Error("Error listing movies", logger.Field("error", err))
		response.ErrorResponse(w, http.StatusInternalServerError, "Error listing movies")
		return
	}

	// Calculate total pages
	totalPages := totalCount / query.PageSize
	if totalCount%query.PageSize != 0 {
		totalPages++
	}

	responseData := PaginatedMovieResponse{
		Movies:     movies,
		TotalCount: totalCount,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}

	logger.Info("Movies listed",
		logger.Field("count", len(movies)),
		logger.Field("total", totalCount),
		logger.Field("page", query.Page),
	)
	response.SuccessResponse(w, http.StatusOK, "Movies retrieved successfully", responseData)
}
