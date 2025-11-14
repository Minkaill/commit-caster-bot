package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User представляет пользователя сервиса
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Email        string `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string `gorm:"not null" json:"-"`
	Name         string `json:"name"`

	// Webhook URL уникален для каждого пользователя
	WebhookToken string `gorm:"uniqueIndex;not null" json:"webhook_token"`

	// Связь с настройками
	Settings UserSettings `gorm:"constraint:OnDelete:CASCADE;" json:"settings"`
}

// UserSettings хранит токены и настройки пользователя
type UserSettings struct {
	ID     uint `gorm:"primarykey" json:"id"`
	UserID uint `gorm:"uniqueIndex;not null" json:"user_id"`

	// Telegram настройки
	TelegramBotToken  string `json:"telegram_bot_token"`
	TelegramChannelID string `json:"telegram_channel_id"`

	// Groq API для AI
	GroqAPIKey string `json:"groq_api_key"`

	// GitHub webhook secret
	GitHubSecret string `json:"github_secret"`

	// Дополнительные настройки
	IsActive      bool   `gorm:"default:true" json:"is_active"`
	AIModel       string `gorm:"default:llama-3.3-70b-versatile" json:"ai_model"`
	PostLanguage  string `gorm:"default:ru" json:"post_language"`
	MaxCommits    int    `gorm:"default:5" json:"max_commits"`
	CustomPrompt  string `gorm:"type:text" json:"custom_prompt,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SetPassword хеширует и устанавливает пароль
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword проверяет пароль
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}
