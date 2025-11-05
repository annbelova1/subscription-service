//go:build integration
// +build integration

package main

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "os"
    "testing"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "subscription-service/internal/config"
    "subscription-service/internal/database"
    "subscription-service/internal/handlers"
    "subscription-service/internal/models"
    "subscription-service/internal/repository"
    "subscription-service/internal/service"
)

func getTestDBConfig() *config.DatabaseConfig {
    dbHost := os.Getenv("DB_HOST")
    if dbHost == "" {
        if _, err := os.Stat("/.dockerenv"); err == nil {
            dbHost = "postgres-test"
        } else {
            dbHost = "localhost"
        }
    }

    return &config.DatabaseConfig{
        Host:     dbHost,
        Port:     5432,
        User:     "postgres",
        Password: "password",
        Name:     "subscriptions_test",
        SSLMode:  "disable",
    }
}

func setupIntegrationTest(t *testing.T) (*httptest.Server, func()) {
    cfg := getTestDBConfig()

    // Подключаемся к тестовой базе
    db, err := database.NewDB(cfg)
    if err != nil {
        t.Skipf("Skipping integration test - cannot connect to test database: %v", err)
        return nil, func() {}
    }

    // Проверяем подключение с таймаутом
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := db.PingContext(ctx); err != nil {
        t.Skipf("Skipping integration test - cannot ping test database: %v", err)
        db.Close()
        return nil, func() {}
    }

    // Очищаем таблицу перед тестом
    _, err = db.Exec("DELETE FROM subscriptions")
    if err != nil {
        t.Skipf("Skipping integration test - cannot clean test database: %v", err)
        db.Close()
        return nil, func() {}
    }

    // Создаем цепочку зависимостей
    repo := repository.NewSubscriptionRepository(db)
    svc := service.NewSubscriptionService(repo)
    logger := logrus.New()
    logger.SetLevel(logrus.ErrorLevel)
    handler := handlers.NewSubscriptionHandler(svc, logger)
    
    // Настраиваем роутер
    router := setupTestRouter(handler)
    server := httptest.NewServer(router)

    return server, func() {
        server.Close()
        db.Close()
    }
}

func setupTestRouter(handler *handlers.SubscriptionHandler) *gin.Engine {
    gin.SetMode(gin.TestMode)
    router := gin.Default()
    
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
    
    return router
}

func TestIntegration_CreateAndGetSubscription(t *testing.T) {
    server, cleanup := setupIntegrationTest(t)
    defer cleanup()

    userID := uuid.New()
    startDate := time.Now()

    // Создаем подписку
    createReq := models.CreateSubscriptionRequest{
        ServiceName:  "Integration Test Netflix",
        Price: 599.00,
        UserID:       userID,
        StartDate:    startDate,
    }

    body, _ := json.Marshal(createReq)
    resp, err := http.Post(server.URL+"/api/v1/subscriptions", "application/json", bytes.NewBuffer(body))
    assert.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    // Получаем созданную подписку
    var createdSub models.Subscription
    err = json.NewDecoder(resp.Body).Decode(&createdSub)
    assert.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, createdSub.ID)
    assert.Equal(t, "Integration Test Netflix", createdSub.ServiceName)

    // Получаем подписку по ID
    resp, err = http.Get(server.URL + "/api/v1/subscriptions/" + createdSub.ID.String())
    assert.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusOK, resp.StatusCode)

    var retrievedSub models.Subscription
    err = json.NewDecoder(resp.Body).Decode(&retrievedSub)
    assert.NoError(t, err)

    assert.Equal(t, createdSub.ID, retrievedSub.ID)
    assert.Equal(t, "Integration Test Netflix", retrievedSub.ServiceName)
    assert.Equal(t, 599.00, retrievedSub.Price)
}

func TestIntegration_CreateDuplicateSubscription(t *testing.T) {
    server, cleanup := setupIntegrationTest(t)
    defer cleanup()

    userID := uuid.New()
    startDate := time.Now()

    createReq := models.CreateSubscriptionRequest{
        ServiceName:  "Duplicate Test Service",
        Price: 299.00,
        UserID:       userID,
        StartDate:    startDate,
    }

    // Первое создание - успех
    body, _ := json.Marshal(createReq)
    resp, err := http.Post(server.URL+"/api/v1/subscriptions", "application/json", bytes.NewBuffer(body))
    assert.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)
    resp.Body.Close()

    // Второе создание с теми же данными - ошибка
    body, _ = json.Marshal(createReq)
    resp, err = http.Post(server.URL+"/api/v1/subscriptions", "application/json", bytes.NewBuffer(body))
    assert.NoError(t, err)
    assert.Equal(t, http.StatusConflict, resp.StatusCode)
    resp.Body.Close()
}

func TestIntegration_GetSummary(t *testing.T) {
    server, cleanup := setupIntegrationTest(t)
    defer cleanup()

    userID := uuid.New()
    startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

    // Создаем тестовые подписки
    subscriptions := []models.CreateSubscriptionRequest{
        {
            ServiceName:  "Integration Test Service 1",
            Price: 100.00,
            UserID:       userID,
            StartDate:    startDate,
        },
        {
            ServiceName:  "Integration Test Service 2", 
            Price: 200.00,
            UserID:       userID,
            StartDate:    startDate.AddDate(0, 1, 0),
        },
    }

    // Создаем подписки
    for _, sub := range subscriptions {
        body, _ := json.Marshal(sub)
        resp, err := http.Post(server.URL+"/api/v1/subscriptions", "application/json", bytes.NewBuffer(body))
        assert.NoError(t, err)
        assert.Equal(t, http.StatusCreated, resp.StatusCode)
        resp.Body.Close()
    }

    // Тестируем summary
    resp, err := http.Get(server.URL + "/api/v1/subscriptions/summary?start_date=2024-01-01&end_date=2024-12-31&user_id=" + userID.String())
    assert.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusOK, resp.StatusCode)

    var summary models.SubscriptionSummary
    err = json.NewDecoder(resp.Body).Decode(&summary)
    assert.NoError(t, err)

    assert.Equal(t, 300.00, summary.TotalCost)
}

func TestIntegration_UpdateSubscription(t *testing.T) {
    server, cleanup := setupIntegrationTest(t)
    defer cleanup()

    userID := uuid.New()
    startDate := time.Now()

    // Создаем подписку
    createReq := models.CreateSubscriptionRequest{
        ServiceName:  "Update Test Service",
        Price: 199.00,
        UserID:       userID,
        StartDate:    startDate,
    }

    body, _ := json.Marshal(createReq)
    resp, err := http.Post(server.URL+"/api/v1/subscriptions", "application/json", bytes.NewBuffer(body))
    assert.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    var createdSub models.Subscription
    err = json.NewDecoder(resp.Body).Decode(&createdSub)
    assert.NoError(t, err)
    resp.Body.Close()

    // Обновляем подписку
    updateReq := models.UpdateSubscriptionRequest{
        Price: func(f float64) *float64 { return &f }(299.00),
    }

    body, _ = json.Marshal(updateReq)
    req, err := http.NewRequest("PUT", server.URL+"/api/v1/subscriptions/"+createdSub.ID.String(), bytes.NewBuffer(body))
    assert.NoError(t, err)
    req.Header.Set("Content-Type", "application/json")

    resp, err = http.DefaultClient.Do(req)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    resp.Body.Close()

    // Проверяем обновление
    resp, err = http.Get(server.URL + "/api/v1/subscriptions/" + createdSub.ID.String())
    assert.NoError(t, err)
    defer resp.Body.Close()

    var updatedSub models.Subscription
    err = json.NewDecoder(resp.Body).Decode(&updatedSub)
    assert.NoError(t, err)

    assert.Equal(t, 299.00, updatedSub.Price)
}

func TestIntegration_DeleteSubscription(t *testing.T) {
    server, cleanup := setupIntegrationTest(t)
    defer cleanup()

    userID := uuid.New()
    startDate := time.Now()

    // Создаем подписку
    createReq := models.CreateSubscriptionRequest{
        ServiceName:  "Delete Test Service",
        Price: 99.00,
        UserID:       userID,
        StartDate:    startDate,
    }

    body, _ := json.Marshal(createReq)
    resp, err := http.Post(server.URL+"/api/v1/subscriptions", "application/json", bytes.NewBuffer(body))
    assert.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    var createdSub models.Subscription
    err = json.NewDecoder(resp.Body).Decode(&createdSub)
    assert.NoError(t, err)
    resp.Body.Close()

    // Удаляем подписку
    req, err := http.NewRequest("DELETE", server.URL+"/api/v1/subscriptions/"+createdSub.ID.String(), nil)
    assert.NoError(t, err)

    resp, err = http.DefaultClient.Do(req)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    resp.Body.Close()

    // Проверяем что подписка удалена
    resp, err = http.Get(server.URL + "/api/v1/subscriptions/" + createdSub.ID.String())
    assert.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}