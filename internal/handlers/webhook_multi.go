package handlers

import (
	"commitcaster/internal/database"
	"commitcaster/internal/models"
	"commitcaster/internal/services"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type MultiUserWebhookHandler struct{}

func NewMultiUserWebhookHandler() *MultiUserWebhookHandler {
	return &MultiUserWebhookHandler{}
}

// HandleGitHubWebhook обрабатывает webhook от GitHub для multi-user
func (h *MultiUserWebhookHandler) HandleGitHubWebhook(c *gin.Context) {
	webhookToken := c.Param("token")
	if webhookToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook token required"})
		return
	}

	// Находим пользователя по webhook токену
	db := database.GetDB()
	var user models.User
	if err := db.Where("webhook_token = ?", webhookToken).First(&user).Error; err != nil {
		log.Printf("User not found for token: %s", webhookToken)
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid webhook token"})
		return
	}

	// Загружаем настройки пользователя
	var settings models.UserSettings
	if err := db.Where("user_id = ?", user.ID).First(&settings).Error; err != nil {
		log.Printf("Settings not found for user: %d", user.ID)
		c.JSON(http.StatusNotFound, gin.H{"error": "User settings not configured"})
		return
	}

	// Проверяем что бот активен
	if !settings.IsActive {
		log.Printf("Bot is inactive for user: %d", user.ID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Bot is not active. Please configure your tokens first"})
		return
	}

	// Читаем тело запроса
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot read body"})
		return
	}

	// Проверяем подпись GitHub
	if !h.verifySignature(body, c.GetHeader("X-Hub-Signature-256"), settings.GitHubSecret) {
		log.Println("Invalid signature")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Парсим payload
	var payload models.GitHubWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Проверяем, что это push event с коммитами
	event := c.GetHeader("X-GitHub-Event")
	if event != "push" || len(payload.Commits) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Event ignored"})
		return
	}

	// Обрабатываем коммиты асинхронно
	go h.processCommits(payload, settings)

	c.JSON(http.StatusOK, gin.H{"message": "Webhook received"})
}

func (h *MultiUserWebhookHandler) processCommits(payload models.GitHubWebhookPayload, settings models.UserSettings) {
	// Собираем информацию о коммитах
	commitSummary := h.buildCommitSummary(payload, settings.MaxCommits)

	log.Printf("Processing commits for repo: %s (user_id: %d)", payload.Repository.Name, settings.UserID)

	// Создаём сервисы с настройками пользователя
	telegramService := services.NewTelegramServiceWithSettings(&settings)
	aiService := services.NewAIServiceWithSettings(&settings)

	// Генерируем пост с помощью AI
	post, err := aiService.GeneratePost(commitSummary, payload.Repository.Name)
	if err != nil {
		log.Printf("Error generating post: %v", err)
		return
	}

	// Отправляем в Telegram
	if err := telegramService.SendMessage(post); err != nil {
		log.Printf("Error sending to Telegram: %v", err)
		return
	}

	log.Printf("Successfully posted to Telegram for user_id: %d", settings.UserID)
}

func (h *MultiUserWebhookHandler) buildCommitSummary(payload models.GitHubWebhookPayload, maxCommits int) string {
	var summary strings.Builder

	if maxCommits == 0 {
		maxCommits = 5
	}

	summary.WriteString(fmt.Sprintf("Репозиторий: %s\n", payload.Repository.Name))
	summary.WriteString(fmt.Sprintf("Количество коммитов: %d\n\n", len(payload.Commits)))

	for i, commit := range payload.Commits {
		if i >= maxCommits {
			summary.WriteString(fmt.Sprintf("...и ещё %d коммитов\n", len(payload.Commits)-maxCommits))
			break
		}

		summary.WriteString(fmt.Sprintf("Коммит %d:\n", i+1))
		summary.WriteString(fmt.Sprintf("Сообщение: %s\n", commit.Message))

		if len(commit.Added) > 0 {
			summary.WriteString(fmt.Sprintf("Добавлено файлов: %d\n", len(commit.Added)))
		}
		if len(commit.Modified) > 0 {
			summary.WriteString(fmt.Sprintf("Изменено файлов: %d\n", len(commit.Modified)))
		}
		if len(commit.Removed) > 0 {
			summary.WriteString(fmt.Sprintf("Удалено файлов: %d\n", len(commit.Removed)))
		}
		summary.WriteString("\n")
	}

	return summary.String()
}

func (h *MultiUserWebhookHandler) verifySignature(payload []byte, signature string, secret string) bool {
	if secret == "" {
		log.Println("Warning: GitHub secret not configured, skipping signature verification")
		return true
	}

	if signature == "" {
		return false
	}

	// Убираем префикс "sha256="
	signature = strings.TrimPrefix(signature, "sha256=")

	// Вычисляем HMAC
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}
