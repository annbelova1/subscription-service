package models

import (
    "time"

    "github.com/google/uuid"
)

type Subscription struct {
    ID           uuid.UUID `json:"id" db:"id"`
    ServiceName  string    `json:"service_name" db:"service_name"`
    Price float64   `json:"price" db:"price"`
    UserID       uuid.UUID `json:"user_id" db:"user_id"`
    StartDate    time.Time `json:"start_date" db:"start_date"`
    EndDate      *time.Time `json:"end_date,omitempty" db:"end_date"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type CreateSubscriptionRequest struct {
    ServiceName  string    `json:"service_name" binding:"required"`
    Price float64   `json:"price" binding:"required,gt=0"`
    UserID       uuid.UUID `json:"user_id" binding:"required"`
    StartDate    time.Time `json:"start_date" binding:"required"`
    EndDate      *time.Time `json:"end_date,omitempty"`
}

type UpdateSubscriptionRequest struct {
    ServiceName  *string    `json:"service_name,omitempty"`
    Price *float64   `json:"price,omitempty"`
    EndDate      *time.Time `json:"end_date,omitempty"`
}

type SubscriptionSummary struct {
    TotalCost   float64 `json:"total_cost"`
    Subscriptions []Subscription `json:"subscriptions,omitempty"`
}

type SummaryRequest struct {
    StartDate   *time.Time `form:"start_date,omitempty"`
    EndDate     *time.Time `form:"end_date,omitempty"`
    UserID     *uuid.UUID `form:"user_id,omitempty"`
    ServiceName *string    `form:"service_name,omitempty"`
}