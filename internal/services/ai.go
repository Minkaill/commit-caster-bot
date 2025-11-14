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

type AIService struct {
	cfg      *config.Config
	settings *models.UserSettings
}

type GroqRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Temperature float64 `json:"temperature"`
	MaxTokens int      `json:"max_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GroqResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func NewAIService(cfg *config.Config) *AIService {
	return &AIService{cfg: cfg}
}

// NewAIServiceWithSettings создаёт AI сервис с настройками пользователя
func NewAIServiceWithSettings(settings *models.UserSettings) *AIService {
	return &AIService{settings: settings}
}

// GeneratePost генерирует пост на основе информации о коммитах
func (s *AIService) GeneratePost(commitSummary, repoName string) (string, error) {
	// Определяем промпт
	var prompt string
	if s.settings != nil && s.settings.CustomPrompt != "" {
		// Используем кастомный промпт пользователя
		prompt = fmt.Sprintf(s.settings.CustomPrompt, commitSummary, repoName)
	} else {
		// Дефолтный промпт
		prompt = fmt.Sprintf(`Ты - разработчик, который ведёт технический блог в Telegram.

Информация о коммитах:
%s

Напиши увлекательный пост для Telegram канала о проделанной работе.

Требования:
- Начни с приветствия "Всем привет!"
- Расскажи что делал сегодня в проекте %s
- Упомяни интересные моменты или сложности
- Стиль: дружелюбный, но профессиональный
- Длина: 3-5 предложений
- Используй эмодзи где уместно
- НЕ используй хештеги
- Пиши на русском языке

Пример хорошего поста:
"Всем привет! Сегодня добавил функцию редактирования ответов в чат-боте. Было интересно разбираться с async/await, но в итоге всё получилось! Теперь пользователи могут исправлять свои сообщения. Завтра планирую добавить уведомления."`, commitSummary, repoName)
	}

	// Определяем модель
	model := "llama-3.3-70b-versatile"
	if s.settings != nil && s.settings.AIModel != "" {
		model = s.settings.AIModel
	}

	// Определяем API ключ
	var apiKey string
	if s.settings != nil {
		apiKey = s.settings.GroqAPIKey
	} else if s.cfg != nil {
		apiKey = s.cfg.GroqAPIKey
	} else {
		return "", fmt.Errorf("no API key available")
	}

	reqBody := GroqRequest{
		Model: model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   500,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("groq API error (status %d): %s", resp.StatusCode, string(body))
	}

	var groqResp GroqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return groqResp.Choices[0].Message.Content, nil
}
