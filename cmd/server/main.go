package main

import (
    "context"
    "database/sql"
    "log"
    "net/http"
    "os"
    "os/signal"
    "strconv"
    "syscall"
    "time"
    "path/filepath"

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

// MultiOutputHook для записи в несколько выходов
type MultiOutputHook struct {
    file *os.File
    formatter logrus.Formatter
}

func (h *MultiOutputHook) Levels() []logrus.Level {
    return logrus.AllLevels
}

func (h *MultiOutputHook) Fire(entry *logrus.Entry) error {
    // Запись в файл если он открыт
    if h.file != nil {
        line, err := h.formatter.Format(entry)
        if err != nil {
            return err
        }
        if _, err := h.file.Write(line); err != nil {
            return err
        }
    }
    return nil
}

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

    logger := setupLogger(cfg)

    db, err := database.NewDB(&cfg.Database)
    if err != nil {
        logger.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()

    repo := repository.NewSubscriptionRepository(db)
    svc := service.NewSubscriptionService(repo)
    handler := handlers.NewSubscriptionHandler(svc, logger)

    router := gin.Default()

    router.Use(ginLogger(logger))

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


    // router.Run() уже внутри себя запускает HTTP сервер, и он блокирующий. Если обернуть его в goroutine, то main функция продолжит выполнение и сразу завершит программу.
    // logger.Infof("Starting server on port %d", cfg.Server.Port)
    // if err := router.Run(":8080"); err != nil {
    //     logger.Fatalf("Failed to start server: %v", err)
    // }

    srv := &http.Server{
        Addr:    ":" + strconv.Itoa(cfg.Server.Port),
        Handler: router,
    }

    go func() {
        logger.Infof("Starting server on port %d", cfg.Server.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatalf("Failed to start server: %v", err)
        }
    }()

    waitForShutdown(srv, db, logger, cfg.Server.ShutdownTimeout)
}

func waitForShutdown(srv *http.Server, db *sql.DB, logger *logrus.Logger, timeout time.Duration) {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    
    // Блокируемся до получения сигнала
    sig := <-quit
    logger.Infof("Received signal: %s. Starting graceful shutdown...", sig)

    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    logger.Info("Shutting down HTTP server...")
    if err := srv.Shutdown(ctx); err != nil {
        logger.Errorf("HTTP server shutdown error: %v", err)
    } else {
        logger.Info("HTTP server stopped gracefully")
    }

    logger.Info("Closing database connections...")
    if err := db.Close(); err != nil {
        logger.Errorf("Database connection close error: %v", err)
    } else {
        logger.Info("Database connections closed")
    }

    logger.Info("Closing log files...")
    for _, hook := range logger.Hooks {
        for _, h := range hook {
            if multiHook, ok := h.(*MultiOutputHook); ok && multiHook.file != nil {
                multiHook.file.Close()
            }
        }
    }

    logger.Info("Application shutdown completed")
}

func setupLogger(cfg *config.Config) *logrus.Logger {
    logger := logrus.New()
    
    level, err := logrus.ParseLevel(cfg.Log.Level)
    if err != nil {
        logger.SetLevel(logrus.InfoLevel)
        logger.Warnf("Invalid log level '%s', using 'info' instead", cfg.Log.Level)
    } else {
        logger.SetLevel(level)
    }

    if cfg.Log.Format == "json" {
        logger.SetFormatter(&logrus.JSONFormatter{
            TimestampFormat: "2006-01-02 15:04:05",
        })
    } else {
        logger.SetFormatter(&logrus.TextFormatter{
            FullTimestamp:   true,
            TimestampFormat: "2006-01-02 15:04:05",
            ForceColors:     true,
        })
    }

    logger.SetOutput(os.Stdout)

    // Добавляем файловый вывод если нужно
    if cfg.Log.Output == "file" && cfg.Log.FilePath != "" {
        // Создаем директорию если её нет
        dir := filepath.Dir(cfg.Log.FilePath)
        if err := os.MkdirAll(dir, 0755); err != nil {
            logger.Warnf("Failed to create log directory %s: %v", dir, err)
        } else {
            file, err := os.OpenFile(cfg.Log.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
            if err != nil {
                logger.Warnf("Failed to open log file %s: %v", cfg.Log.FilePath, err)
            } else {
                // Добавляем hook для записи в файл
                logger.AddHook(&MultiOutputHook{
                    file: file,
                    formatter: logger.Formatter,
                })
                logger.Infof("Logging to file: %s", cfg.Log.FilePath)
            }
        }
    }

    logger.Infof("Logger initialized with level: %s", logger.GetLevel())
    return logger
}

func ginLogger(logger *logrus.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start)
        
        entry := logger.WithFields(logrus.Fields{
            "method":     c.Request.Method,
            "path":       c.Request.URL.Path,
            "status":     c.Writer.Status(),
            "duration":   duration,
            "client_ip":  c.ClientIP(),
            "user_agent": c.Request.UserAgent(),
        })
        
        if c.Writer.Status() >= 500 {
            entry.Error("Request completed with server error")
        } else if c.Writer.Status() >= 400 {
            entry.Warn("Request completed with client error")
        } else {
            entry.Info("Request completed successfully")
        }
    }
}