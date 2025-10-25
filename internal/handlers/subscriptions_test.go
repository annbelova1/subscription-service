package handlers

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "subscription-service/internal/models"
)

// Mock сервиса
type MockSubscriptionService struct {
    mock.Mock
}

func (m *MockSubscriptionService) CreateSubscription(ctx context.Context, sub *models.Subscription) error {
    args := m.Called(ctx, sub)
    return args.Error(0)
}

func (m *MockSubscriptionService) GetSubscription(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) UpdateSubscription(ctx context.Context, id uuid.UUID, req *models.UpdateSubscriptionRequest) error {
    args := m.Called(ctx, id, req)
    return args.Error(0)
}

func (m *MockSubscriptionService) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}

func (m *MockSubscriptionService) ListSubscriptions(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*models.Subscription, error) {
    args := m.Called(ctx, userID, serviceName)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) GetSummary(ctx context.Context, req *models.SummaryRequest) (*models.SubscriptionSummary, error) {
    args := m.Called(ctx, req)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.SubscriptionSummary), args.Error(1)
}

func setupTestRouter(handler *SubscriptionHandler) *gin.Engine {
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

func TestSubscriptionHandler_CreateSubscription(t *testing.T) {
    mockService := new(MockSubscriptionService)
    logger := logrus.New()
    handler := NewSubscriptionHandler(mockService, logger)
    router := setupTestRouter(handler)

    userID := uuid.New()
    startDate := time.Now()

    createReq := models.CreateSubscriptionRequest{
        ServiceName:  "Netflix",
        Price: 599.00,
        UserID:       userID,
        StartDate:    startDate,
    }

    // Используем anyContext() для любого типа контекста
    mockService.On("CreateSubscription", anyContext(), anySubscription()).Return(nil)

    body, _ := json.Marshal(createReq)
    req, _ := http.NewRequest("POST", "/api/v1/subscriptions", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)
    mockService.AssertCalled(t, "CreateSubscription", anyContext(), anySubscription())
}

func TestSubscriptionHandler_GetSubscription(t *testing.T) {
    mockService := new(MockSubscriptionService)
    logger := logrus.New()
    handler := NewSubscriptionHandler(mockService, logger)
    router := setupTestRouter(handler)

    subscriptionID := uuid.New()
    expectedSubscription := &models.Subscription{
        ID:           subscriptionID,
        ServiceName:  "Netflix",
        Price: 599.00,
        UserID:       uuid.New(),
        StartDate:    time.Now(),
    }

    mockService.On("GetSubscription", anyContext(), uuidMatcher(subscriptionID)).Return(expectedSubscription, nil)

    req, _ := http.NewRequest("GET", "/api/v1/subscriptions/"+subscriptionID.String(), nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    
    var response models.Subscription
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, subscriptionID, response.ID)
}

func TestSubscriptionHandler_ListSubscriptions(t *testing.T) {
    mockService := new(MockSubscriptionService)
    logger := logrus.New()
    handler := NewSubscriptionHandler(mockService, logger)
    router := setupTestRouter(handler)

    userID := uuid.New()
    expectedSubscriptions := []*models.Subscription{
        {
            ID:           uuid.New(),
            ServiceName:  "Netflix",
            Price: 599.00,
            UserID:       userID,
            StartDate:    time.Now(),
        },
    }

    mockService.On("ListSubscriptions", 
        anyContext(), 
        uuidPtrMatcher(userID),
        (*string)(nil),
    ).Return(expectedSubscriptions, nil)

    req, _ := http.NewRequest("GET", "/api/v1/subscriptions?user_id="+userID.String(), nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    mockService.AssertCalled(t, "ListSubscriptions", 
        anyContext(), 
        uuidPtrMatcher(userID),
        (*string)(nil),
    )
}

func TestSubscriptionHandler_GetSummary(t *testing.T) {
    mockService := new(MockSubscriptionService)
    logger := logrus.New()
    handler := NewSubscriptionHandler(mockService, logger)
    router := setupTestRouter(handler)

    expectedSummary := &models.SubscriptionSummary{
        TotalCost: 998.00,
    }

    mockService.On("GetSummary", anyContext(), mock.AnythingOfType("*models.SummaryRequest")).Return(expectedSummary, nil)

    req, _ := http.NewRequest("GET", "/api/v1/subscriptions/summary?start_date=2024-01-01&end_date=2024-12-31", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    
    var response models.SubscriptionSummary
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, 998.00, response.TotalCost)
}

func TestSubscriptionHandler_UpdateSubscription(t *testing.T) {
    mockService := new(MockSubscriptionService)
    logger := logrus.New()
    handler := NewSubscriptionHandler(mockService, logger)
    router := setupTestRouter(handler)

    subscriptionID := uuid.New()
    newPrice := 699.00
    updateReq := models.UpdateSubscriptionRequest{
        Price: &newPrice,
    }

    mockService.On("UpdateSubscription", anyContext(), uuidMatcher(subscriptionID), &updateReq).Return(nil)

    body, _ := json.Marshal(updateReq)
    req, _ := http.NewRequest("PUT", "/api/v1/subscriptions/"+subscriptionID.String(), bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
}

func TestSubscriptionHandler_DeleteSubscription(t *testing.T) {
    mockService := new(MockSubscriptionService)
    logger := logrus.New()
    handler := NewSubscriptionHandler(mockService, logger)
    router := setupTestRouter(handler)

    subscriptionID := uuid.New()

    mockService.On("DeleteSubscription", anyContext(), uuidMatcher(subscriptionID)).Return(nil)

    req, _ := http.NewRequest("DELETE", "/api/v1/subscriptions/"+subscriptionID.String(), nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
}

func TestSubscriptionHandler_CreateSubscription_InvalidJSON(t *testing.T) {
    mockService := new(MockSubscriptionService)
    logger := logrus.New()
    handler := NewSubscriptionHandler(mockService, logger)
    router := setupTestRouter(handler)

    body := bytes.NewBuffer([]byte(`{"invalid": json`))
    req, _ := http.NewRequest("POST", "/api/v1/subscriptions", body)
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubscriptionHandler_GetSubscription_InvalidUUID(t *testing.T) {
    mockService := new(MockSubscriptionService)
    logger := logrus.New()
    handler := NewSubscriptionHandler(mockService, logger)
    router := setupTestRouter(handler)

    req, _ := http.NewRequest("GET", "/api/v1/subscriptions/invalid-uuid", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubscriptionHandler_GetSummary_InvalidDate(t *testing.T) {
    mockService := new(MockSubscriptionService)
    logger := logrus.New()
    handler := NewSubscriptionHandler(mockService, logger)
    router := setupTestRouter(handler)

    req, _ := http.NewRequest("GET", "/api/v1/subscriptions/summary?start_date=invalid-date", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Вспомогательная функция для проверки любого контекста
func anyContext() interface{} {
    return mock.MatchedBy(func(ctx context.Context) bool {
        return ctx != nil
    })
}

// Вспомогательная функция для проверки подписки
func anySubscription() interface{} {
    return mock.MatchedBy(func(sub *models.Subscription) bool {
        return sub != nil
    })
}

// Вспомогательная функция для проверки UUID
func uuidMatcher(expected uuid.UUID) interface{} {
    return mock.MatchedBy(func(id uuid.UUID) bool {
        return id == expected
    })
}

// Вспомогательная функция для проверки указателя на UUID
func uuidPtrMatcher(expected uuid.UUID) interface{} {
    return mock.MatchedBy(func(id *uuid.UUID) bool {
        return id != nil && *id == expected
    })
}