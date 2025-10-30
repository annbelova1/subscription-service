//go:build integration
// +build integration

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"subscription-service/internal/models"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *models.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	Update(ctx context.Context, id uuid.UUID, req *models.UpdateSubscriptionRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*models.Subscription, error)
	GetSummary(ctx context.Context, req *models.SummaryRequest) (*models.SubscriptionSummary, error)
}

type subscriptionRepo struct {
	db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) SubscriptionRepository {
	return &subscriptionRepo{db: db}
}

func (r *subscriptionRepo) Create(ctx context.Context, sub *models.Subscription) error {
	existing, err := r.findExistingSubscription(ctx, sub.UserID, sub.ServiceName, sub.StartDate)
	if err != nil {
		return fmt.Errorf("failed to check existing subscription: %w", err)
	}
	
	if existing != nil {
		return fmt.Errorf("subscription already exists for user %s to service %s starting from %s", 
			sub.UserID, sub.ServiceName, sub.StartDate.Format("2006-01-02"))
	}

	query := `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRowContext(
		ctx,
		query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)

	if err != nil {
		if isDuplicateError(err) {
			return fmt.Errorf("subscription already exists for this user and service")
		}
		log.Printf("Error creating subscription: %v", err)
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	log.Printf("Created subscription with ID: %s", sub.ID)
	return nil
}

func (r *subscriptionRepo) findExistingSubscription(ctx context.Context, userID uuid.UUID, serviceName string, startDate time.Time) (*models.Subscription, error) {
	query, args := NewQueryBuilder("subscriptions").
		Where("user_id = $1", userID).
		Where("service_name = $2", serviceName).
		Where("start_date = $3", startDate).
		Limit(1).
		Build()

	var sub models.Subscription
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &sub, nil
}

func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	errorString := err.Error()
	return strings.Contains(errorString, "unique constraint") || 
		   strings.Contains(errorString, "duplicate key") ||
		   strings.Contains(errorString, "23505")
}

func (r *subscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	query, args := NewQueryBuilder("subscriptions").
		Where("id = $1", id).
		Build()

	var sub models.Subscription
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("subscription not found")
		}
		log.Printf("Error getting subscription by ID %s: %v", id, err)
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return &sub, nil
}

func (r *subscriptionRepo) Update(ctx context.Context, id uuid.UUID, req *models.UpdateSubscriptionRequest) error {
	query := `
		UPDATE subscriptions 
		SET service_name = COALESCE($1, service_name),
			price = COALESCE($2, price),
			end_date = COALESCE($3, end_date),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query, req.ServiceName, req.Price, req.EndDate, id)
	if err != nil {
		log.Printf("Error updating subscription %s: %v", id, err)
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("subscription not found")
	}

	log.Printf("Updated subscription with ID: %s", id)
	return nil
}

func (r *subscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		log.Printf("Error deleting subscription %s: %v", id, err)
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("subscription not found")
	}

	log.Printf("Deleted subscription with ID: %s", id)
	return nil
}

func (r *subscriptionRepo) List(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*models.Subscription, error) {
	qb := NewQueryBuilder("subscriptions").
		OrderBy("created_at", "DESC")

	if userID != nil {
		qb = qb.Where("user_id = $1", *userID)
	}

	if serviceName != nil {
		qb = qb.Where("service_name = $2", *serviceName)
	}

	query, args := qb.Build()

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("Error listing subscriptions: %v", err)
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []*models.Subscription
	for rows.Next() {
		var sub models.Subscription
		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&sub.EndDate,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subscriptions = append(subscriptions, &sub)
	}

	log.Printf("Listed %d subscriptions", len(subscriptions))
	return subscriptions, nil
}

func (r *subscriptionRepo) GetSummary(ctx context.Context, req *models.SummaryRequest) (*models.SubscriptionSummary, error) {
	qb := NewQueryBuilder("subscriptions").
		Select(`
			COALESCE(SUM(
				price * 
				CASE 
					WHEN end_date IS NOT NULL THEN 
						EXTRACT(MONTH FROM AGE(end_date, start_date)) + 1
					ELSE 
						EXTRACT(MONTH FROM AGE(COALESCE($1, CURRENT_DATE), start_date)) + 1
				END
			), 0) as total_cost
		`).
		Where("1=1")

	args := []interface{}{req.EndDate}

	if req.StartDate != nil {
		qb = qb.Where("start_date >= $2", *req.StartDate)
	}

	if req.EndDate != nil {
		qb = qb.Where("(end_date IS NULL OR end_date <= $3)", *req.EndDate)
	}

	if req.UserID != nil {
		qb = qb.Where("user_id = $4", *req.UserID)
	}

	if req.ServiceName != nil {
		qb = qb.Where("service_name = $5", *req.ServiceName)
	}

	query, finalArgs := qb.Build()

	var totalCost float64
	err := r.db.QueryRowContext(ctx, query, finalArgs...).Scan(&totalCost)
	if err != nil {
		log.Printf("Error calculating subscription summary: %v", err)
		return nil, fmt.Errorf("failed to calculate summary: %w", err)
	}

	summary := &models.SubscriptionSummary{
		TotalCost: totalCost,
	}

	log.Printf("Calculated summary: total cost = %.2f", totalCost)
	return summary, nil
}