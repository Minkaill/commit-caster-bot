package handlers

import (
	"commitcaster/config"
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

type WebhookHandler struct {
	cfg             *config.Config
	telegramService *services.TelegramService
	aiService       *services.AIService
}

func NewWebhookHandler(cfg *config.Config, telegramService *services.TelegramService, aiService *services.AIService) *WebhookHandler {
	return &WebhookHandler{
		cfg:             cfg,
		telegramService: telegramService,
		aiService:       aiService,
	}
}

// HandleGitHubWebhook обрабатывает webhook от GitHub
func (h *WebhookHandler) HandleGitHubWebhook(c *gin.Context) {
	// Читаем тело запроса
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot read body"})
		return
	}

	// Проверяем подпись GitHub
	if !h.verifySignature(body, c.GetHeader("X-Hub-Signature-256")) {
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
	go h.processCommits(payload)

	c.JSON(http.StatusOK, gin.H{"message": "Webhook received"})
}

func (h *WebhookHandler) processCommits(payload models.GitHubWebhookPayload) {
	// Собираем информацию о коммитах
	commitSummary := h.buildCommitSummary(payload)

	log.Printf("Processing commits for repo: %s", payload.Repository.Name)
	log.Printf("Commit summary: %s", commitSummary)

	// Генерируем пост с помощью AI
	post, err := h.aiService.GeneratePost(commitSummary, payload.Repository.Name)
	if err != nil {
		log.Printf("Error generating post: %v", err)
		return
	}

	// Отправляем в Telegram
	if err := h.telegramService.SendMessage(post); err != nil {
		log.Printf("Error sending to Telegram: %v", err)
		return
	}

	log.Println("Successfully posted to Telegram")
}

func (h *WebhookHandler) buildCommitSummary(payload models.GitHubWebhookPayload) string {
	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("Репозиторий: %s\n", payload.Repository.Name))
	summary.WriteString(fmt.Sprintf("Количество коммитов: %d\n\n", len(payload.Commits)))

	for i, commit := range payload.Commits {
		if i >= 5 { // Ограничиваем 5 коммитами
			summary.WriteString(fmt.Sprintf("...и ещё %d коммитов\n", len(payload.Commits)-5))
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

func (h *WebhookHandler) verifySignature(payload []byte, signature string) bool {
	if h.cfg.GitHubSecret == "" {
		log.Println("Warning: GitHub secret not configured, skipping signature verification")
		return true
	}

	if signature == "" {
		return false
	}

	// Убираем префикс "sha256="
	signature = strings.TrimPrefix(signature, "sha256=")

	// Вычисляем HMAC
	mac := hmac.New(sha256.New, []byte(h.cfg.GitHubSecret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

// HealthCheck для проверки работоспособности
func (h *WebhookHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"service": "CommitCaster",
	})
}
