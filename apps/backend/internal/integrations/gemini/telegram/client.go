package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/devrapture/omni/internal/config"
)

type TelegramClient struct {
	botToken   string
	baseURL    string
	httpClient *http.Client
	cfg *config.Config
	appBaseURL string
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
		cfg: cfg,
		appBaseURL: cfg.AppBaseUrl,
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

func (t *TelegramClient) SetWebhook(ctx context.Context){
	url := fmt.Sprintf("", a ...any)
} 
