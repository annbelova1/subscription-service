package service

import (
    "context"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "subscription-service/internal/models"
)

// Mock репозитория
type MockSubscriptionRepository struct {
    mock.Mock
}

func (m *MockSubscriptionRepository) Create(ctx context.Context, sub *models.Subscription) error {
    args := m.Called(ctx, sub)
    return args.Error(0)
}

func (m *MockSubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateSubscriptionRequest) error {
    args := m.Called(ctx, id, req)
    return args.Error(0)
}

func (m *MockSubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}

func (m *MockSubscriptionRepository) List(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*models.Subscription, error) {
    args := m.Called(ctx, userID, serviceName)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) GetSummary(ctx context.Context, req *models.SummaryRequest) (*models.SubscriptionSummary, error) {
    args := m.Called(ctx, req)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.SubscriptionSummary), args.Error(1)
}

func TestSubscriptionService_CreateSubscription(t *testing.T) {
    mockRepo := new(MockSubscriptionRepository)
    service := NewSubscriptionService(mockRepo)

    ctx := context.Background()
    subscription := &models.Subscription{
        ServiceName:  "Netflix",
        Price: 599.00,
        UserID:       uuid.New(),
        StartDate:    time.Now(),
    }

    mockRepo.On("Create", anyContext(), subscription).Return(nil)

    err := service.CreateSubscription(ctx, subscription)
    assert.NoError(t, err)
    mockRepo.AssertCalled(t, "Create", anyContext(), subscription)
}

func TestSubscriptionService_GetSubscription(t *testing.T) {
    mockRepo := new(MockSubscriptionRepository)
    service := NewSubscriptionService(mockRepo)

    ctx := context.Background()
    subscriptionID := uuid.New()
    expectedSubscription := &models.Subscription{
        ID:           subscriptionID,
        ServiceName:  "Netflix",
        Price: 599.00,
        UserID:       uuid.New(),
        StartDate:    time.Now(),
    }

    mockRepo.On("GetByID", anyContext(), subscriptionID).Return(expectedSubscription, nil)

    result, err := service.GetSubscription(ctx, subscriptionID)
    assert.NoError(t, err)
    assert.Equal(t, expectedSubscription, result)
    mockRepo.AssertCalled(t, "GetByID", anyContext(), subscriptionID)
}

func TestSubscriptionService_UpdateSubscription(t *testing.T) {
    mockRepo := new(MockSubscriptionRepository)
    service := NewSubscriptionService(mockRepo)

    ctx := context.Background()
    subscriptionID := uuid.New()
    newPrice := 699.00
    updateReq := &models.UpdateSubscriptionRequest{
        Price: &newPrice,
    }

    mockRepo.On("Update", anyContext(), subscriptionID, updateReq).Return(nil)

    err := service.UpdateSubscription(ctx, subscriptionID, updateReq)
    assert.NoError(t, err)
    mockRepo.AssertCalled(t, "Update", anyContext(), subscriptionID, updateReq)
}

func TestSubscriptionService_DeleteSubscription(t *testing.T) {
    mockRepo := new(MockSubscriptionRepository)
    service := NewSubscriptionService(mockRepo)

    ctx := context.Background()
    subscriptionID := uuid.New()

    mockRepo.On("Delete", anyContext(), subscriptionID).Return(nil)

    err := service.DeleteSubscription(ctx, subscriptionID)
    assert.NoError(t, err)
    mockRepo.AssertCalled(t, "Delete", anyContext(), subscriptionID)
}

func TestSubscriptionService_ListSubscriptions(t *testing.T) {
    mockRepo := new(MockSubscriptionRepository)
    service := NewSubscriptionService(mockRepo)

    ctx := context.Background()
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

    mockRepo.On("List", anyContext(), uuidPtrMatcher(userID), nilStringPtr()).Return(expectedSubscriptions, nil)

    result, err := service.ListSubscriptions(ctx, &userID, nil)
    assert.NoError(t, err)
    assert.Equal(t, expectedSubscriptions, result)
    mockRepo.AssertCalled(t, "List", anyContext(), uuidPtrMatcher(userID), nilStringPtr())
}

func TestSubscriptionService_ListSubscriptions_WithServiceName(t *testing.T) {
    mockRepo := new(MockSubscriptionRepository)
    service := NewSubscriptionService(mockRepo)

    ctx := context.Background()
    userID := uuid.New()
    serviceName := "Netflix"
    expectedSubscriptions := []*models.Subscription{
        {
            ID:           uuid.New(),
            ServiceName:  "Netflix",
            Price: 599.00,
            UserID:       userID,
            StartDate:    time.Now(),
        },
    }

    mockRepo.On("List", anyContext(), uuidPtrMatcher(userID), stringPtrMatcher(serviceName)).Return(expectedSubscriptions, nil)

    result, err := service.ListSubscriptions(ctx, &userID, &serviceName)
    assert.NoError(t, err)
    assert.Equal(t, expectedSubscriptions, result)
    mockRepo.AssertCalled(t, "List", anyContext(), uuidPtrMatcher(userID), stringPtrMatcher(serviceName))
}

func TestSubscriptionService_GetSummary(t *testing.T) {
    mockRepo := new(MockSubscriptionRepository)
    service := NewSubscriptionService(mockRepo)

    ctx := context.Background()
    userID := uuid.New()
    startDate := time.Now()
    endDate := startDate.AddDate(1, 0, 0)
    
    req := &models.SummaryRequest{
        StartDate: &startDate,
        EndDate:   &endDate,
        UserID:    &userID,
    }
    expectedSummary := &models.SubscriptionSummary{
        TotalCost: 998.00,
    }

    mockRepo.On("GetSummary", anyContext(), req).Return(expectedSummary, nil)

    result, err := service.GetSummary(ctx, req)
    assert.NoError(t, err)
    assert.Equal(t, expectedSummary, result)
    mockRepo.AssertCalled(t, "GetSummary", anyContext(), req)
}

func TestSubscriptionService_ListSubscriptions_NoFilters(t *testing.T) {
    mockRepo := new(MockSubscriptionRepository)
    service := NewSubscriptionService(mockRepo)

    ctx := context.Background()
    expectedSubscriptions := []*models.Subscription{
        {
            ID:           uuid.New(),
            ServiceName:  "Netflix",
            Price: 599.00,
            UserID:       uuid.New(),
            StartDate:    time.Now(),
        },
    }

    // Тестируем без фильтров (оба параметра nil) - используем правильные матчеры
    mockRepo.On("List", anyContext(), nilUUIDPtr(), nilStringPtr()).Return(expectedSubscriptions, nil)

    result, err := service.ListSubscriptions(ctx, nil, nil)
    assert.NoError(t, err)
    assert.Equal(t, expectedSubscriptions, result)
    mockRepo.AssertCalled(t, "List", anyContext(), nilUUIDPtr(), nilStringPtr())
}

func TestSubscriptionService_ListSubscriptions_OnlyUserID(t *testing.T) {
    mockRepo := new(MockSubscriptionRepository)
    service := NewSubscriptionService(mockRepo)

    ctx := context.Background()
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

    // Только userID, serviceName = nil
    mockRepo.On("List", anyContext(), uuidPtrMatcher(userID), nilStringPtr()).Return(expectedSubscriptions, nil)

    result, err := service.ListSubscriptions(ctx, &userID, nil)
    assert.NoError(t, err)
    assert.Equal(t, expectedSubscriptions, result)
    mockRepo.AssertCalled(t, "List", anyContext(), uuidPtrMatcher(userID), nilStringPtr())
}

func TestSubscriptionService_ListSubscriptions_OnlyServiceName(t *testing.T) {
    mockRepo := new(MockSubscriptionRepository)
    service := NewSubscriptionService(mockRepo)

    ctx := context.Background()
    serviceName := "Netflix"
    expectedSubscriptions := []*models.Subscription{
        {
            ID:           uuid.New(),
            ServiceName:  "Netflix",
            Price: 599.00,
            UserID:       uuid.New(),
            StartDate:    time.Now(),
        },
    }

    // Только serviceName, userID = nil
    mockRepo.On("List", anyContext(), nilUUIDPtr(), stringPtrMatcher(serviceName)).Return(expectedSubscriptions, nil)

    result, err := service.ListSubscriptions(ctx, nil, &serviceName)
    assert.NoError(t, err)
    assert.Equal(t, expectedSubscriptions, result)
    mockRepo.AssertCalled(t, "List", anyContext(), nilUUIDPtr(), stringPtrMatcher(serviceName))
}

func TestSubscriptionService_GetSubscription_NotFound(t *testing.T) {
    mockRepo := new(MockSubscriptionRepository)
    service := NewSubscriptionService(mockRepo)

    ctx := context.Background()
    subscriptionID := uuid.New()

    mockRepo.On("GetByID", anyContext(), subscriptionID).Return(nil, assert.AnError)

    result, err := service.GetSubscription(ctx, subscriptionID)
    assert.Error(t, err)
    assert.Nil(t, result)
    mockRepo.AssertCalled(t, "GetByID", anyContext(), subscriptionID)
}

func TestSubscriptionService_CreateSubscription_Error(t *testing.T) {
    mockRepo := new(MockSubscriptionRepository)
    service := NewSubscriptionService(mockRepo)

    ctx := context.Background()
    subscription := &models.Subscription{
        ServiceName:  "Netflix",
        Price: 599.00,
        UserID:       uuid.New(),
        StartDate:    time.Now(),
    }

    mockRepo.On("Create", anyContext(), subscription).Return(assert.AnError)

    err := service.CreateSubscription(ctx, subscription)
    assert.Error(t, err)
    mockRepo.AssertCalled(t, "Create", anyContext(), subscription)
}

// Вспомогательные функции для мокинга
func anyContext() interface{} {
    return mock.MatchedBy(func(ctx context.Context) bool {
        return ctx != nil
    })
}

func anyUUID() interface{} {
    return mock.MatchedBy(func(id uuid.UUID) bool {
        return id != uuid.Nil
    })
}

// Функции для UUID указателей
func uuidPtrMatcher(expected uuid.UUID) interface{} {
    return mock.MatchedBy(func(id *uuid.UUID) bool {
        return id != nil && *id == expected
    })
}

func nilUUIDPtr() interface{} {
    return mock.MatchedBy(func(id *uuid.UUID) bool {
        return id == nil
    })
}

func anyUUIDPtr() interface{} {
    return mock.MatchedBy(func(id *uuid.UUID) bool {
        return id != nil
    })
}

// Функции для string указателей
func nilStringPtr() interface{} {
    return mock.MatchedBy(func(s *string) bool {
        return s == nil
    })
}

func stringPtrMatcher(expected string) interface{} {
    return mock.MatchedBy(func(s *string) bool {
        return s != nil && *s == expected
    })
}

func anyStringPtr() interface{} {
    return mock.MatchedBy(func(s *string) bool {
        return s != nil
    })
}