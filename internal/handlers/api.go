package handlers

import (
	"commitcaster/internal/auth"
	"commitcaster/internal/database"
	"commitcaster/internal/models"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type APIHandler struct{}

func NewAPIHandler() *APIHandler {
	return &APIHandler{}
}

// Структуры для запросов/ответов

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token        string `json:"token"`
	WebhookToken string `json:"webhook_token"`
	WebhookURL   string `json:"webhook_url"`
}

type SettingsRequest struct {
	TelegramBotToken  string `json:"telegram_bot_token"`
	TelegramChannelID string `json:"telegram_channel_id"`
	GroqAPIKey        string `json:"groq_api_key"`
	GitHubSecret      string `json:"github_secret"`
	AIModel           string `json:"ai_model"`
	PostLanguage      string `json:"post_language"`
	MaxCommits        int    `json:"max_commits"`
	CustomPrompt      string `json:"custom_prompt"`
}

// Register регистрирует нового пользователя
// @Summary Регистрация нового пользователя
// @Description Создаёт нового пользователя и возвращает JWT токен
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "User registration data"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /auth/register [post]
func (h *APIHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Проверяем существование пользователя
	var existingUser models.User
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// Создаём пользователя
	user := models.User{
		Email:        req.Email,
		Name:         req.Name,
		WebhookToken: generateToken(),
	}

	if err := user.SetPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Создаём с настройками по умолчанию
	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Создаём пустые настройки
	settings := models.UserSettings{
		UserID:   user.ID,
		IsActive: false, // Неактивен пока не заполнены токены
	}
	if err := db.Create(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create settings"})
		return
	}

	// Генерируем JWT
	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Token:        token,
		WebhookToken: user.WebhookToken,
		WebhookURL:   fmt.Sprintf("%s/webhook/github/%s", baseURL, user.WebhookToken),
	})
}

// Login авторизует пользователя
// @Summary Авторизация пользователя
// @Description Авторизует пользователя и возвращает JWT токен
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *APIHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	var user models.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token:        token,
		WebhookToken: user.WebhookToken,
		WebhookURL:   fmt.Sprintf("%s/webhook/github/%s", baseURL, user.WebhookToken),
	})
}

// GetSettings получает настройки текущего пользователя
// @Summary Получить настройки пользователя
// @Description Возвращает настройки текущего пользователя
// @Tags settings
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.UserSettings
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /settings [get]
func (h *APIHandler) GetSettings(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	db := database.GetDB()

	var settings models.UserSettings
	if err := db.Where("user_id = ?", userID).First(&settings).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Settings not found"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateSettings обновляет настройки текущего пользователя
// @Summary Обновить настройки пользователя
// @Description Обновляет настройки текущего пользователя (все поля опциональные)
// @Tags settings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body SettingsRequest true "Settings to update"
// @Success 200 {object} models.UserSettings
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /settings [put]
func (h *APIHandler) UpdateSettings(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req SettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	var settings models.UserSettings
	if err := db.Where("user_id = ?", userID).First(&settings).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Settings not found"})
		return
	}

	// Обновляем поля
	if req.TelegramBotToken != "" {
		settings.TelegramBotToken = req.TelegramBotToken
	}
	if req.TelegramChannelID != "" {
		settings.TelegramChannelID = req.TelegramChannelID
	}
	if req.GroqAPIKey != "" {
		settings.GroqAPIKey = req.GroqAPIKey
	}
	if req.GitHubSecret != "" {
		settings.GitHubSecret = req.GitHubSecret
	}
	if req.AIModel != "" {
		settings.AIModel = req.AIModel
	}
	if req.PostLanguage != "" {
		settings.PostLanguage = req.PostLanguage
	}
	if req.MaxCommits > 0 {
		settings.MaxCommits = req.MaxCommits
	}
	if req.CustomPrompt != "" {
		settings.CustomPrompt = req.CustomPrompt
	}

	// Проверяем что все необходимые токены заполнены
	if settings.TelegramBotToken != "" && settings.TelegramChannelID != "" && settings.GroqAPIKey != "" {
		settings.IsActive = true
	}

	if err := db.Save(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// GetWebhookInfo возвращает информацию о webhook URL
// @Summary Получить webhook URL
// @Description Возвращает уникальный webhook URL пользователя для настройки GitHub
// @Tags webhook
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /webhook [get]
func (h *APIHandler) GetWebhookInfo(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	db := database.GetDB()

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	c.JSON(http.StatusOK, gin.H{
		"webhook_token": user.WebhookToken,
		"webhook_url":   fmt.Sprintf("%s/webhook/github/%s", baseURL, user.WebhookToken),
	})
}

// generateToken генерирует случайный токен для webhook
func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
