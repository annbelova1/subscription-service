//go:build integration
// +build integration

package repository

import (
    "context"
    "database/sql"
    "fmt"
    "os"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "subscription-service/internal/models"
    
    _ "github.com/lib/pq"
)

type SubscriptionRepositoryIntegrationTestSuite struct {
    suite.Suite
    db   *sql.DB
    repo SubscriptionRepository
    ctx  context.Context
}

func (suite *SubscriptionRepositoryIntegrationTestSuite) SetupSuite() {
    suite.ctx = context.Background()
    
    dbHost := getEnvDefault("DB_HOST", "localhost")
	dbPort := getEnvDefault("DB_PORT", "5432") 
	dbUser := getEnvDefault("DB_USER", "postgres")
	dbPassword := getEnvDefault("DB_PASSWORD", "password")
	dbName := getEnvDefault("DB_NAME", "subscriptions_test")	

    connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        dbHost, dbPort, dbUser, dbPassword, dbName)

    suite.T().Logf("Connecting to database: %s", connStr)

    var err error
    suite.db, err = sql.Open("postgres", connStr)
    if err != nil {
        suite.T().Fatalf("Failed to open database: %v", err)
    }

    suite.db.SetMaxOpenConns(10)
    suite.db.SetMaxIdleConns(5)
    suite.db.SetConnMaxLifetime(time.Hour)

    err = suite.waitForDB()
    if err != nil {
        suite.T().Fatalf("Database not ready: %v", err)
    }

    suite.repo = NewSubscriptionRepository(suite.db)
    suite.T().Log("Database connection established")
}

func (suite *SubscriptionRepositoryIntegrationTestSuite) waitForDB() error {
    ctx, cancel := context.WithTimeout(suite.ctx, 30*time.Second)
    defer cancel()

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            err := suite.db.PingContext(ctx)
            if err == nil {
                return nil
            }
            time.Sleep(1 * time.Second)
        }
    }
}

func (suite *SubscriptionRepositoryIntegrationTestSuite) TearDownSuite() {
    if suite.db != nil {
        suite.db.Close()
    }
}

func (suite *SubscriptionRepositoryIntegrationTestSuite) SetupTest() {
    if suite.db == nil {
        suite.T().Skip("Database not available")
        return
    }

    // Очищаем таблицу перед каждым тестом
    _, err := suite.db.ExecContext(suite.ctx, "DELETE FROM subscriptions")
    if err != nil {
        suite.T().Fatalf("Failed to clean database: %v", err)
    }
}

func (suite *SubscriptionRepositoryIntegrationTestSuite) TestCreateSubscription() {
    userID := uuid.New()
    startDate := time.Now()

    subscription := &models.Subscription{
        ServiceName: "Netflix",
        Price:       599.00,
        UserID:      userID,
        StartDate:   startDate,
    }

    err := suite.repo.Create(suite.ctx, subscription)
    assert.NoError(suite.T(), err)
    assert.NotEqual(suite.T(), uuid.Nil, subscription.ID)

    // Проверяем, что подписка сохранилась в базе
    var count int
    err = suite.db.QueryRowContext(suite.ctx, 
        "SELECT COUNT(*) FROM subscriptions WHERE id = $1", subscription.ID).Scan(&count)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), 1, count)
}

func (suite *SubscriptionRepositoryIntegrationTestSuite) TestCreateDuplicateSubscription() {
    userID := uuid.New()
    startDate := time.Now()

    subscription := &models.Subscription{
        ServiceName: "Netflix",
        Price:       599.00,
        UserID:      userID,
        StartDate:   startDate,
    }

    err := suite.repo.Create(suite.ctx, subscription)
    assert.NoError(suite.T(), err)

    duplicateSub := &models.Subscription{
        ServiceName: "Netflix",
        Price:       699.00, // другая цена
        UserID:      userID,
        StartDate:   startDate,
    }

    err = suite.repo.Create(suite.ctx, duplicateSub)
    assert.Error(suite.T(), err)
    assert.Contains(suite.T(), err.Error(), "already exists")
}

func (suite *SubscriptionRepositoryIntegrationTestSuite) TestGetSubscription() {
    userID := uuid.New()
    startDate := time.Now()

    subscription := &models.Subscription{
        ServiceName: "Yandex Plus",
        Price:       399.00,
        UserID:      userID,
        StartDate:   startDate,
    }

    err := suite.repo.Create(suite.ctx, subscription)
    assert.NoError(suite.T(), err)

    found, err := suite.repo.GetByID(suite.ctx, subscription.ID)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), subscription.ID, found.ID)
    assert.Equal(suite.T(), "Yandex Plus", found.ServiceName)
    assert.Equal(suite.T(), 399.00, found.Price)
    assert.Equal(suite.T(), userID, found.UserID)
}

func (suite *SubscriptionRepositoryIntegrationTestSuite) TestGetNonExistentSubscription() {
    nonExistentID := uuid.New()
    _, err := suite.repo.GetByID(suite.ctx, nonExistentID)
    assert.Error(suite.T(), err)
    assert.Contains(suite.T(), err.Error(), "not found")
}

func (suite *SubscriptionRepositoryIntegrationTestSuite) TestUpdateSubscription() {
    userID := uuid.New()
    startDate := time.Now()

    subscription := &models.Subscription{
        ServiceName: "Spotify",
        Price:       299.00,
        UserID:      userID,
        StartDate:   startDate,
    }

    err := suite.repo.Create(suite.ctx, subscription)
    assert.NoError(suite.T(), err)

    newPrice := 349.00
    updateReq := &models.UpdateSubscriptionRequest{
        Price: &newPrice,
    }

    err = suite.repo.Update(suite.ctx, subscription.ID, updateReq)
    assert.NoError(suite.T(), err)

    updated, err := suite.repo.GetByID(suite.ctx, subscription.ID)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), 349.00, updated.Price)
    assert.Equal(suite.T(), "Spotify", updated.ServiceName)
}

func (suite *SubscriptionRepositoryIntegrationTestSuite) TestDeleteSubscription() {
    userID := uuid.New()
    startDate := time.Now()

    subscription := &models.Subscription{
        ServiceName: "Apple Music",
        Price:       199.00,
        UserID:      userID,
        StartDate:   startDate,
    }

    err := suite.repo.Create(suite.ctx, subscription)
    assert.NoError(suite.T(), err)

    err = suite.repo.Delete(suite.ctx, subscription.ID)
    assert.NoError(suite.T(), err)

    _, err = suite.repo.GetByID(suite.ctx, subscription.ID)
    assert.Error(suite.T(), err)
    assert.Contains(suite.T(), err.Error(), "not found")
}

func (suite *SubscriptionRepositoryIntegrationTestSuite) TestListSubscriptions() {
    userID := uuid.New()

    subscriptions := []*models.Subscription{
        {
            ServiceName: "Netflix",
            Price:       599.00,
            UserID:      userID,
            StartDate:   time.Now(),
        },
        {
            ServiceName: "Yandex Plus",
            Price:       399.00,
            UserID:      userID,
            StartDate:   time.Now().AddDate(0, 1, 0),
        },
    }

    for _, sub := range subscriptions {
        err := suite.repo.Create(suite.ctx, sub)
        assert.NoError(suite.T(), err)
    }

    // Тестируем List с user_id
    result, err := suite.repo.List(suite.ctx, &userID, nil)
    assert.NoError(suite.T(), err)
    assert.Len(suite.T(), result, 2)

    // Тестируем List без фильтров
    allResults, err := suite.repo.List(suite.ctx, nil, nil)
    assert.NoError(suite.T(), err)
    assert.Len(suite.T(), allResults, 2)

    // Тестируем List с service_name фильтром
    serviceName := "Netflix"
    filteredResults, err := suite.repo.List(suite.ctx, &userID, &serviceName)
    assert.NoError(suite.T(), err)
    assert.Len(suite.T(), filteredResults, 1)
    assert.Equal(suite.T(), "Netflix", filteredResults[0].ServiceName)
}

func (suite *SubscriptionRepositoryIntegrationTestSuite) TestGetSummary() {
    userID := uuid.New()
    startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    endDate := startDate.AddDate(1, 0, 0)

    subscriptions := []*models.Subscription{
        {
            ServiceName: "Netflix",
            Price:       599.00,
            UserID:      userID,
            StartDate:   startDate,
        },
        {
            ServiceName: "Yandex Plus",
            Price:       399.00,
            UserID:      userID,
            StartDate:   startDate.AddDate(0, 1, 0),
        },
    }

    for _, sub := range subscriptions {
        err := suite.repo.Create(suite.ctx, sub)
        assert.NoError(suite.T(), err)
    }

    req := &models.SummaryRequest{
        StartDate: &startDate,
        EndDate:   &endDate,
        UserID:    &userID,
    }

    summary, err := suite.repo.GetSummary(suite.ctx, req)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), 998.00, summary.TotalCost)
}

func (suite *SubscriptionRepositoryIntegrationTestSuite) TestGetSummary_NoSubscriptions() {
    userID := uuid.New()
    startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    endDate := startDate.AddDate(1, 0, 0)

    req := &models.SummaryRequest{
        StartDate: &startDate,
        EndDate:   &endDate,
        UserID:    &userID,
    }

    summary, err := suite.repo.GetSummary(suite.ctx, req)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), 0.00, summary.TotalCost)
}

func getEnvDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}