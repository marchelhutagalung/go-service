package handlers

import (
	"encoding/json"
	"errors"
	"github.com/marchelhutagalung/go-service/internal/auth"
	"github.com/marchelhutagalung/go-service/internal/logger"
	"github.com/marchelhutagalung/go-service/internal/middleware"
	"github.com/marchelhutagalung/go-service/internal/models"
	"github.com/marchelhutagalung/go-service/internal/repository"
	"github.com/marchelhutagalung/go-service/internal/response"
	"net/http"
)

type AuthHandler struct {
	userRepo   *repository.UserRepository
	jwtService *auth.JWTService
}

func NewAuthHandler(userRepo *repository.UserRepository, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

type RegisterResponse struct {
	User  *models.UserResponse `json:"user"`
	Token string               `json:"token"`
}

type LoginResponse struct {
	User  *models.UserResponse `json:"user"`
	Token string               `json:"token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input models.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.Error("Invalid request body", logger.Field("error", err))
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	passwordHash, err := models.HashPassword(input.Password)
	if err != nil {
		logger.Error("Error hashing password", logger.Field("error", err))
		response.ErrorResponse(w, http.StatusInternalServerError, "Error processing password")
		return
	}

	user, err := h.userRepo.Create(r.Context(), &input, passwordHash)
	if err != nil {
		if errors.Is(err, repository.ErrEmailExists) {
			logger.Error("Email already exists", logger.Field("email", input.Email))
			response.ErrorResponse(w, http.StatusConflict, "Email already exists")
			return
		}
		logger.Error("Error creating user", logger.Field("error", err))
		response.ErrorResponse(w, http.StatusInternalServerError, "Error creating user")
		return
	}

	token, err := h.jwtService.GenerateToken(user.ID)
	if err != nil {
		logger.Error("Error generating token", logger.Field("error", err), logger.Field("user_id", user.ID))
		response.ErrorResponse(w, http.StatusInternalServerError, "Error generating token")
		return
	}

	responseData := RegisterResponse{
		User:  user.ToResponse(),
		Token: token,
	}

	logger.Info("User registered", logger.Field("user_id", user.ID), logger.Field("email", user.Email))
	response.SuccessResponse(w, http.StatusCreated, "User registered successfully", responseData)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input models.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.Error("Invalid request body", logger.Field("error", err))
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.userRepo.Authenticate(r.Context(), input.Email, input.Password)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidCredentials) {
			logger.Error("Login failed: invalid credentials", logger.Field("email", input.Email))
			response.ErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		logger.Error("Error authenticating user", logger.Field("error", err))
		response.ErrorResponse(w, http.StatusInternalServerError, "Error authenticating user")
		return
	}

	token, err := h.jwtService.GenerateToken(user.ID)
	if err != nil {
		logger.Error("Error generating token", logger.Field("error", err), logger.Field("user_id", user.ID))
		response.ErrorResponse(w, http.StatusInternalServerError, "Error generating token")
		return
	}

	responseData := LoginResponse{
		User:  user.ToResponse(),
		Token: token,
	}

	logger.Info("User logged in", logger.Field("user_id", user.ID), logger.Field("email", user.Email))
	response.SuccessResponse(w, http.StatusOK, "Login successful", responseData)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		logger.Error("Logout attempted without authentication")
		response.ErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	err := h.jwtService.InvalidateToken(userID)
	if err != nil {
		logger.Error("Error logging out", logger.Field("error", err), logger.Field("user_id", userID))
		response.ErrorResponse(w, http.StatusInternalServerError, "Error logging out")
		return
	}

	logger.Info("User logged out", logger.Field("user_id", userID))
	response.SuccessResponse(w, http.StatusOK, "Successfully logged out", nil)
}
