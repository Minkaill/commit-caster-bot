package main

import (
	"commitcaster/config"
	"commitcaster/internal/database"
	"commitcaster/internal/handlers"
	"commitcaster/internal/middleware"
	"commitcaster/internal/services"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "commitcaster/docs" // Swagger docs
)

// @title CommitCaster API
// @version 1.0
// @description API для автоматической публикации GitHub коммитов в Telegram с помощью AI

// @contact.name API Support
// @contact.email support@commitcaster.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	// Проверяем режим работы
	databaseURL := os.Getenv("DATABASE_URL")
	isSaaSMode := databaseURL != ""

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Swagger documentation (доступен в обоих режимах)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Добавляем CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	if isSaaSMode {
		log.Println("🌐 Starting in SaaS mode (multi-user)")

		// Подключаемся к БД
		if err := database.Connect(); err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		// Проверяем JWT_SECRET
		if os.Getenv("JWT_SECRET") == "" {
			log.Fatal("JWT_SECRET not set (required for SaaS mode)")
		}

		// API handlers
		apiHandler := handlers.NewAPIHandler()
		multiWebhookHandler := handlers.NewMultiUserWebhookHandler()

		// Public routes
		r.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "ok",
				"mode":   "saas",
			})
		})
		r.POST("/api/auth/register", apiHandler.Register)
		r.POST("/api/auth/login", apiHandler.Login)

		// Protected routes (require JWT)
		protected := r.Group("/api")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/settings", apiHandler.GetSettings)
			protected.PUT("/settings", apiHandler.UpdateSettings)
			protected.GET("/webhook", apiHandler.GetWebhookInfo)
		}

		// GitHub webhook endpoint (по токену пользователя)
		r.POST("/webhook/github/:token", multiWebhookHandler.HandleGitHubWebhook)

		log.Println("📋 API Endpoints:")
		log.Println("  POST /api/auth/register - Register new user")
		log.Println("  POST /api/auth/login - Login")
		log.Println("  GET  /api/settings - Get user settings (protected)")
		log.Println("  PUT  /api/settings - Update settings (protected)")
		log.Println("  GET  /api/webhook - Get webhook URL (protected)")
		log.Println("  POST /webhook/github/:token - GitHub webhook")
		log.Println("")
		log.Printf("📖 Swagger UI: http://localhost:%s/swagger/index.html", cfg.Port)

	} else {
		log.Println("👤 Starting in single-user mode")

		// Проверяем обязательные переменные для single-user режима
		if cfg.TelegramBotToken == "" {
			log.Fatal("TELEGRAM_BOT_TOKEN не установлен")
		}
		if cfg.TelegramChannelID == "" {
			log.Fatal("TELEGRAM_CHANNEL_ID не установлен")
		}
		if cfg.GroqAPIKey == "" {
			log.Fatal("GROQ_API_KEY не установлен")
		}

		// Инициализируем сервисы
		telegramService := services.NewTelegramService(cfg)
		aiService := services.NewAIService(cfg)
		webhookHandler := handlers.NewWebhookHandler(cfg, telegramService, aiService)

		// Роуты для single-user режима
		r.GET("/health", webhookHandler.HealthCheck)
		r.POST("/webhook/github", webhookHandler.HandleGitHubWebhook)

		log.Printf("📡 Webhook URL: http://localhost:%s/webhook/github", cfg.Port)
		log.Printf("📖 Swagger UI: http://localhost:%s/swagger/index.html", cfg.Port)
	}

	// Запускаем сервер
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 CommitCaster запущен на порту %s", cfg.Port)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
