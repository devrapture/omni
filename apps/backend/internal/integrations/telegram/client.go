package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/devrapture/omni/internal/config"
	apperrors "github.com/devrapture/omni/internal/errors"
)

type TelegramClient struct {
	botToken   string
	baseURL    string
	httpClient *http.Client
	appBaseURL string
	cfg        *config.Config
}

type SendMessageRequest struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

type SetWebhookResponse struct {
	Ok          bool   `json:"ok"`
	Result      bool   `json:"result"`
	Description string `json:"description"`
}
type WebhookInfo struct {
	URL                  string `json:"url"`
	HasCustomCertificate bool   `json:"has_custom_certificate"`
	PendingUpdateCount   int    `json:"pending_update_count"`
	LastErrorDate        int64  `json:"last_error_date"`
	LastErrorMessage     string `json:"last_error_message"`
	MaxConnections       int    `json:"max_connections"`
}

type GetMeResponse struct {
	OK     bool `json:"ok"`
	Result struct {
		ID        int64  `json:"id"`
		IsBot     bool   `json:"is_bot"`
		FirstName string `json:"first_name"`
		Username  string `json:"username"`
	} `json:"result"`
	Description string `json:"description"`
}

func NewTelegramClient(cfg *config.Config) *TelegramClient {
	return &TelegramClient{
		botToken: cfg.TelegramBotToken,
		baseURL:  fmt.Sprintf("https://api.telegram.org/bot%s", cfg.TelegramBotToken),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		appBaseURL: cfg.AppBaseUrl,
		cfg:        cfg,
	}
}

func (t *TelegramClient) SendMessage(ctx context.Context, chatID, text string) error {
	url := fmt.Sprintf("%s/sendMessage", t.baseURL)
	payload := SendMessageRequest{
		ChatID: chatID,
		Text:   text,
	}

	return t.post(ctx, url, payload, nil)

}

func (t *TelegramClient) ValidateBotToken(ctx context.Context, token string) (username string, err error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", token)
	var body struct {
		Ok     bool `json:"ok"`
		Result struct {
			Username string `json:"username"`
		} `json:"result"`
	}

	if err := t.post(ctx, url, nil, &body); err != nil {
		return "", err
	}

	if !body.Ok {
		return "", apperrors.ErrInvalidTelegramBotToken
	}

	return body.Result.Username, nil
}

func (t *TelegramClient) SetWebHook(ctx context.Context, telegramBotToken string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook", telegramBotToken)
	payload := map[string]string{
		"url":          t.appBaseURL,
		"secret_token": t.cfg.TelegramWebhookSecret,
	}
	var result SetWebhookResponse

	if err := t.post(ctx, url, payload, &result); err != nil {
		return err
	}

	if !result.Ok {
		return fmt.Errorf("Telegram rejected webhook registration: %s", result.Description)
	}
	return nil
}

func (t *TelegramClient) DeleteWebHook(ctx context.Context, telegramBotToken string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/deleteWebhook", telegramBotToken)
	payload := map[string]string{
		"url":          t.appBaseURL,
		"secret_token": t.cfg.TelegramWebhookSecret,
	}
	var result SetWebhookResponse
	if err := t.post(ctx, url, payload, &result); err != nil {
		return err
	}

	if !result.Ok {
		return fmt.Errorf("Telegram rejected webhook deletion: %s", result.Description)
	}
	return nil
}

func (t *TelegramClient) GetWebhookInfo(ctx context.Context, telegramBotToken string) (*WebhookInfo, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getWebhookInfo", telegramBotToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook info:%w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get telegram webhook:%w", err)
	}
	defer res.Body.Close()

	var result struct {
		Ok     bool        `json:"ok"`
		Result WebhookInfo `json:"result"`
	}
	if res.StatusCode == http.StatusUnauthorized {
		return nil, apperrors.ErrInvalidTelegramBotToken
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode webhook info response:%w", err)
	}

	return &result.Result, nil
}

func (t *TelegramClient) post(ctx context.Context, url string, payload, result interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram API request failed: %w", err)
	}

	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read telegram response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram api returned status code %d", res.StatusCode)
	}

	if res.StatusCode == http.StatusUnauthorized {
		return apperrors.ErrInvalidTelegramBotToken
	}

	if result != nil {
		if err := json.Unmarshal(resBody, result); err != nil {
			return fmt.Errorf("failed to parse telegram response: %w", err)
		}
	}

	return nil
}
