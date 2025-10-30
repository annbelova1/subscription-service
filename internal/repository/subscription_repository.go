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
    List(ctx context.Context, userID *uuid.UUID, serviceName *string, page, pageSize int) ([]*models.Subscription, *models.Pagination, error)
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
    query := `
        SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
        FROM subscriptions 
        WHERE user_id = $1 AND service_name = $2 AND start_date = $3
        LIMIT 1
    `

    var sub models.Subscription
    err := r.db.QueryRowContext(ctx, query, userID, serviceName, startDate).Scan(
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
    query := `
        SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
        FROM subscriptions 
        WHERE id = $1
    `

    var sub models.Subscription
    err := r.db.QueryRowContext(ctx, query, id).Scan(
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

func (r *subscriptionRepo) List(ctx context.Context, userID *uuid.UUID, serviceName *string, page int, pageSize int) ([]*models.Subscription, *models.Pagination, error) {
    query := `
        SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
        FROM subscriptions 
        WHERE 1=1
    `
    countQuery := `SELECT COUNT(*) FROM subscriptions WHERE 1=1`

    args := []interface{}{}
    argPos := 1

    if userID != nil {
        query += fmt.Sprintf(" AND user_id = $%d", argPos)
        args = append(args, *userID)
        argPos++
    }

    if serviceName != nil {
        query += fmt.Sprintf(" AND service_name = $%d", argPos)
        args = append(args, *serviceName)
        argPos++
    }

    query += " ORDER BY created_at DESC"

    if pageSize > 0 {
        offset := (page - 1) * pageSize
        query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
        args = append(args, pageSize, offset)
    }

    var total int
    err := r.db.QueryRowContext(ctx, countQuery, args[:len(args)-2]...).Scan(&total)
    if err != nil {
        log.Printf("Error counting subscriptions: %v", err)
        return nil, nil, fmt.Errorf("failed to count subscriptions: %w", err)
    }

    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        log.Printf("Error listing subscriptions: %v", err)
        return nil, nil, fmt.Errorf("failed to list subscriptions: %w", err)
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
            return nil, nil, fmt.Errorf("failed to scan subscription: %w", err)
        }
        subscriptions = append(subscriptions, &sub)
    }

    pagination := &models.Pagination{
        Page:       page,
        PageSize:   pageSize,
        Total:      total,
        TotalPages: (total + pageSize - 1) / pageSize,
    }

    log.Printf("Listed %d subscriptions (page %d, pageSize %d, total %d)", 
        len(subscriptions), page, pageSize, total)
    return subscriptions, pagination, nil
}

func (r *subscriptionRepo) GetSummary(ctx context.Context, req *models.SummaryRequest) (*models.SubscriptionSummary, error) {
    query := `
        SELECT COALESCE(SUM(
            price * 
            CASE 
                WHEN end_date IS NOT NULL THEN 
                    EXTRACT(MONTH FROM AGE(end_date, start_date)) + 1
                ELSE 
                    EXTRACT(MONTH FROM AGE(COALESCE($1, CURRENT_DATE), start_date)) + 1
            END
        ), 0) as total_cost 
        FROM subscriptions 
        WHERE 1=1
    `
    
    args := []interface{}{req.EndDate}
    argPos := 2

    var conditions []string

    if req.StartDate != nil {
        conditions = append(conditions, fmt.Sprintf("start_date >= $%d", argPos))
        args = append(args, *req.StartDate)
        argPos++
    }

    if req.EndDate != nil {
        conditions = append(conditions, fmt.Sprintf("(end_date IS NULL OR end_date <= $%d)", argPos))
        args = append(args, *req.EndDate)
        argPos++
    }

    if req.UserID != nil {
        conditions = append(conditions, fmt.Sprintf("user_id = $%d", argPos))
        args = append(args, *req.UserID)
        argPos++
    }

    if req.ServiceName != nil {
        conditions = append(conditions, fmt.Sprintf("service_name = $%d", argPos))
        args = append(args, *req.ServiceName)
        argPos++
    }

    if len(conditions) > 0 {
        query += " AND " + strings.Join(conditions, " AND ")
    }

    var totalCost float64
    err := r.db.QueryRowContext(ctx, query, args...).Scan(&totalCost)
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