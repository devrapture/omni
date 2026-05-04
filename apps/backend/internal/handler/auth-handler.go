package handler

import (
	"errors"
	"net/http"

	"github.com/devrapture/omni/internal/models"
	"github.com/devrapture/omni/internal/service"
	"github.com/devrapture/omni/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	userService service.UserService
}

func NewAuthHandler(userService service.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

type GoogleLoginResponse struct {
	AccessToken string      `json:"access_token"`
	User        UserPayload `json:"user"`
}

type UserPayload struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url"`
}

type LoginWithGoogleRequest struct {
	IDToken string `json:"id_token"`
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	url, err := h.userService.GetGoogleAuthURL(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to initialize google oauth")
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code not found"})
		return
	}
	if state == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "state not found")
		return
	}
	user, jwtToken, err := h.userService.HandleGoogleCallback(c, code, state)
	if err != nil {
		if errors.Is(err, service.ErrInvalidOAuthState) {
			utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "invalid oauth state")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Authentication failed")
		return
	}

	response := GoogleLoginResponse{
		AccessToken: jwtToken,
		User:        toUserPayload(user),
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", response, nil)
}

func toUserPayload(u *models.User) UserPayload {
	if u == nil {
		return UserPayload{}
	}
	return UserPayload{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		AvatarURL: u.AvatarURL,
	}
}

func (h *AuthHandler) LoginWithGoogle(c *gin.Context) {
	var req LoginWithGoogleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}
	if req.IDToken == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "id_token is required")
		return
	}
	user, jwtToken, err := h.userService.HandleLoginWithGoogle(c.Request.Context(), req.IDToken)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
		return
	}

	response := GoogleLoginResponse{
		AccessToken: jwtToken,
		User:        toUserPayload(user),
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", response, nil)
}
