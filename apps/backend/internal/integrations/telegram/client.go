package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/devrapture/omni/internal/config"
	apperrors "github.com/devrapture/omni/internal/errors"
)

type TelegramClient struct {
	botToken   string
	baseURL    string
	httpClient *http.Client
}

type SendMessageRequest struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func NewTelegramClient(cfg *config.Config) *TelegramClient {
	return &TelegramClient{
		botToken: cfg.TelegramBotToken,
		baseURL:  fmt.Sprintf("https://api.telegram.org/bot%s", cfg.TelegramBotToken),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *TelegramClient) SendMessage(ctx context.Context, chatID, text string) error {
	url := fmt.Sprintf("%s/sendMessage", t.baseURL)
	payload := SendMessageRequest{
		ChatID: chatID,
		Text:   text,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create telegram request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram api returned status code %d", resp.StatusCode)
	}

	return nil
}

func (t *TelegramClient) ValidateBotToken(ctx context.Context, token string) (username string, err error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", token)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create telegram request: %w", err)
	}

	res, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to validate token: %w", err)
	}
	defer res.Body.Close()

	var body struct {
		Ok     bool `json:"ok"`
		Result struct {
			Username string `json:"username"`
		} `json:"result"`
	}

	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusNotFound {
		return "", apperrors.ErrInvalidTelegramBotToken
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("telegram api returned status code %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if !body.Ok {
		return "", apperrors.ErrInvalidTelegramBotToken
	}
	return body.Result.Username, nil
}
