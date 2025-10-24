package main

import (
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"

    "subscription-service/internal/config"
    "subscription-service/internal/database"
    "subscription-service/internal/handlers"
    "subscription-service/internal/repository"
    "subscription-service/internal/service"

    _ "subscription-service/docs"
)

// @title Subscription Service API
// @version 1.0
// @description API для управления подписками пользователей
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io


// @host localhost:8080
func main() {
    // Загрузка конфигурации
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    logger := logrus.New()
    logger.SetLevel(logrus.InfoLevel)
    logger.SetFormatter(&logrus.JSONFormatter{})

    db, err := database.NewDB(&cfg.Database)
    if err != nil {
        logger.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()

    repo := repository.NewSubscriptionRepository(db)
    svc := service.NewSubscriptionService(repo)
    handler := handlers.NewSubscriptionHandler(svc, logger)

    router := gin.Default()

    router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    api := router.Group("/api/v1")
    {
        subscriptions := api.Group("/subscriptions")
        {
            subscriptions.POST("", handler.CreateSubscription)
            subscriptions.GET("", handler.ListSubscriptions)
            subscriptions.GET("/summary", handler.GetSummary)
            subscriptions.GET("/:id", handler.GetSubscription)
            subscriptions.PUT("/:id", handler.UpdateSubscription)
            subscriptions.DELETE("/:id", handler.DeleteSubscription)
        }
    }

    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    logger.Infof("Starting server on port %d", cfg.Server.Port)
    if err := router.Run(":8080"); err != nil {
        logger.Fatalf("Failed to start server: %v", err)
    }
}