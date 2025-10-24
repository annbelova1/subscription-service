package service

import (
    "context"

    "github.com/google/uuid"
    "subscription-service/internal/models"
    "subscription-service/internal/repository"
)

type SubscriptionService interface {
    CreateSubscription(ctx context.Context, sub *models.Subscription) error
    GetSubscription(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
    UpdateSubscription(ctx context.Context, id uuid.UUID, req *models.UpdateSubscriptionRequest) error
    DeleteSubscription(ctx context.Context, id uuid.UUID) error
    ListSubscriptions(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*models.Subscription, error)
    GetSummary(ctx context.Context, req *models.SummaryRequest) (*models.SubscriptionSummary, error)
}

type subscriptionService struct {
    repo repository.SubscriptionRepository
}

func NewSubscriptionService(repo repository.SubscriptionRepository) SubscriptionService {
    return &subscriptionService{repo: repo}
}

func (s *subscriptionService) CreateSubscription(ctx context.Context, sub *models.Subscription) error {
    return s.repo.Create(ctx, sub)
}

func (s *subscriptionService) GetSubscription(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
    return s.repo.GetByID(ctx, id)
}

func (s *subscriptionService) UpdateSubscription(ctx context.Context, id uuid.UUID, req *models.UpdateSubscriptionRequest) error {
    return s.repo.Update(ctx, id, req)
}

func (s *subscriptionService) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
    return s.repo.Delete(ctx, id)
}

func (s *subscriptionService) ListSubscriptions(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*models.Subscription, error) {
    return s.repo.List(ctx, userID, serviceName)
}

func (s *subscriptionService) GetSummary(ctx context.Context, req *models.SummaryRequest) (*models.SubscriptionSummary, error) {
    return s.repo.GetSummary(ctx, req)
}