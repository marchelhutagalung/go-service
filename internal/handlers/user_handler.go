package handlers

import (
	"errors"
	"github.com/marchelhutagalung/go-service/internal/logger"
	"github.com/marchelhutagalung/go-service/internal/middleware"
	"github.com/marchelhutagalung/go-service/internal/repository"
	"github.com/marchelhutagalung/go-service/internal/response"
	"net/http"
)

type UserHandler struct {
	userRepo *repository.UserRepository
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logger.Error("Get current user attempted without authentication")
		response.ErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			logger.Error("User not found", logger.Field("user_id", userID))
			response.ErrorResponse(w, http.StatusNotFound, "User not found")
			return
		}
		logger.Error("Error fetching user", logger.Field("error", err), logger.Field("user_id", userID))
		response.ErrorResponse(w, http.StatusInternalServerError, "Error fetching user")
		return
	}

	logger.Info("User fetched", logger.Field("user_id", userID))
	response.SuccessResponse(w, http.StatusOK, "User retrieved successfully", user.ToResponse())
}
