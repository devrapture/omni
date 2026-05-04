package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/devrapture/omni/internal/config"
	"github.com/devrapture/omni/internal/models"
	"github.com/devrapture/omni/internal/repositories"
	"github.com/devrapture/omni/internal/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var ErrInvalidOAuthState = errors.New("invalid oauth state")

type UserService interface {
	GetGoogleAuthURL(ctx context.Context) (string, error)
	HandleGoogleCallback(ctx context.Context, code, state string) (*models.User, string, error)
	HandleLoginWithGoogle(ctx context.Context, idToken string) (*models.User, string, error)
}

type userService struct {
	repo              repositories.UserRepository
	googleOAuthConfig *oauth2.Config
	githubOAuthConfig *oauth2.Config
	cfg               *config.Config
	stateStore        *oauthStateStore
}

type oauthStateStore struct {
	mu     sync.Mutex
	tokens map[string]oauthStateEntry
	ttl    time.Duration
}

type oauthStateEntry struct {
	provider  string
	expiresAt time.Time
}

func NewUserService(cfg *config.Config, repo repositories.UserRepository) UserService {
	googleConfig := &oauth2.Config{
		ClientID:     cfg.GOOGLE_CLIENT_ID,
		ClientSecret: cfg.GOOGLE_CLIENT_SECRET,
		RedirectURL:  cfg.GOOGLE_REDIRECT_URL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	return &userService{
		repo:              repo,
		googleOAuthConfig: googleConfig,
		cfg:               cfg,
		stateStore: &oauthStateStore{
			tokens: make(map[string]oauthStateEntry),
			ttl:    10 * time.Minute,
		},
	}
}

func (s *userService) HandleGoogleCallback(ctx context.Context, code, state string) (*models.User, string, error) {
	if err := s.stateStore.ValidateAndConsume("google", state); err != nil {
		return nil, "", err
	}

	token, err := s.googleOAuthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, "", errors.New("failed to exchange token")
	}
	client := s.googleOAuthConfig.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, "", errors.New("failed to get user info")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", errors.New("failed to get user info")
	}

	var userInfo struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, "", errors.New("failed to parse user info")
	}
	user, err := s.repo.FindOrCreateUser(ctx, userInfo.ID, userInfo.Email, userInfo.Name, userInfo.Picture, "google")
	if err != nil {
		return nil, "", errors.New("failed to create user")
	}
	jwtToken, err := utils.GenerateJwt(user.ID, user.Email, s.cfg)
	if err != nil {
		return nil, "", errors.New("failed to generate jwt")
	}
	return user, jwtToken, nil
}

func (s *userService) GetGoogleAuthURL(ctx context.Context) (string, error) {
	state, err := s.stateStore.Generate("google")
	if err != nil {
		return "", err
	}

	return s.googleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

func (s *userService) GetGithubAuthUrl(ctx context.Context) (string, error) {
	state, err := s.stateStore.Generate("github")
	if err != nil {
		return "", err
	}

	return s.githubOAuthConfig.AuthCodeURL(state), nil
}

func (s *userService) HandleLoginWithGoogle(ctx context.Context, idToken string) (*models.User, string, error) {
	claims, err := utils.VerifyGoogleIDToken(ctx, idToken, s.cfg.GOOGLE_CLIENT_ID)
	if err != nil {
		return nil, "", errors.New("failed to verify google id token")
	}

	user, err := s.repo.FindOrCreateUser(ctx, claims.Subject, claims.Email, claims.Name, claims.Picture, "google")
	if err != nil {
		return nil, "", errors.New("failed to create user")
	}
	jwtToken, err := utils.GenerateJwt(user.ID, user.Email, s.cfg)
	if err != nil {
		return nil, "", errors.New("failed to generate jwt")
	}
	return user, jwtToken, nil
}

func (s *oauthStateStore) Generate(provider string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", errors.New("failed to generate oauth state")
	}

	token := base64.RawURLEncoding.EncodeToString(b)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupExpiredLocked(time.Now())
	s.tokens[token] = oauthStateEntry{
		provider:  provider,
		expiresAt: time.Now().Add(s.ttl),
	}

	return token, nil
}

func (s *oauthStateStore) ValidateAndConsume(provider, token string) error {
	if token == "" {
		return ErrInvalidOAuthState
	}

	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupExpiredLocked(now)

	entry, ok := s.tokens[token]
	if !ok || entry.provider != provider || now.After(entry.expiresAt) {
		delete(s.tokens, token)
		return ErrInvalidOAuthState
	}

	delete(s.tokens, token)
	return nil
}

func (s *oauthStateStore) cleanupExpiredLocked(now time.Time) {
	for token, entry := range s.tokens {
		if now.After(entry.expiresAt) {
			delete(s.tokens, token)
		}
	}
}
