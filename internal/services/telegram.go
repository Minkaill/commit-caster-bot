package services

import (
	"bytes"
	"commitcaster/config"
	"commitcaster/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type TelegramService struct {
	cfg      *config.Config
	settings *models.UserSettings
}

type TelegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

func NewTelegramService(cfg *config.Config) *TelegramService {
	return &TelegramService{cfg: cfg}
}

// NewTelegramServiceWithSettings создаёт сервис с настройками пользователя
func NewTelegramServiceWithSettings(settings *models.UserSettings) *TelegramService {
	return &TelegramService{settings: settings}
}

// SendMessage отправляет сообщение в Telegram канал
func (s *TelegramService) SendMessage(text string) error {
	var botToken, channelID string

	if s.settings != nil {
		botToken = s.settings.TelegramBotToken
		channelID = s.settings.TelegramChannelID
	} else if s.cfg != nil {
		botToken = s.cfg.TelegramBotToken
		channelID = s.cfg.TelegramChannelID
	} else {
		return fmt.Errorf("no configuration available")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	message := TelegramMessage{
		ChatID:    channelID,
		Text:      text,
		ParseMode: "Markdown",
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
