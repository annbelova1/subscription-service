package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Subscription struct {
	ServiceName string    `json:"service_name"`
	Price       float64   `json:"price"`
	UserID      string    `json:"user_id"`
	StartDate   time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

func main() {
	baseURL := "http://localhost:8080/api/v1/subscriptions"
	userID := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

	now := time.Now()
	customEndDate1 := now.AddDate(0, 2, 0)
	customEndDate2 := now.AddDate(0, 0, 45)

	subscriptions := []Subscription{
		{"Yandex Plus", 399.00, userID, time.Now().AddDate(0, 0, -30), &customEndDate1},
		{"Netflix", 799.00, userID, time.Now().AddDate(0, 0, -15), &customEndDate2},
		{"Spotify", 299.00, userID, time.Now().AddDate(0, 0, -10), nil},
		{"YouTube Premium", 459.00, userID, time.Now().AddDate(0, 0, -5), nil},
		{"Apple Music", 199.00, userID, time.Now().AddDate(0, 0, -20), nil},
		{"Amazon Prime", 299.00, userID, time.Now().AddDate(0, 0, -25), nil},
		{"Microsoft 365", 399.00, userID, time.Now().AddDate(0, 0, -8), nil},
		{"PlayStation Plus", 599.00, userID, time.Now().AddDate(0, 0, -12), nil},
		{"Xbox Game Pass", 349.00, userID, time.Now().AddDate(0, 0, -18), nil},
		{"Telegram Premium", 299.00, userID, time.Now().AddDate(0, 0, -3), nil},
	}

	successCount := 0
	for i, sub := range subscriptions {
		fmt.Printf("Creating %d/%d: %s...\n", i+1, len(subscriptions), sub.ServiceName)
		if err := createSubscription(baseURL, sub); err != nil {
			log.Printf("Error creating subscription %s: %v", sub.ServiceName, err)
			continue
		}
		fmt.Printf("✓ Created subscription: %s (%.2f руб)\n", sub.ServiceName, sub.Price)
		successCount++
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Printf("Successfully created %d out of %d subscriptions\n", successCount, len(subscriptions))
}

func createSubscription(url string, sub Subscription) error {
	jsonData, err := json.Marshal(sub)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	// Debug: Print what we're sending
	fmt.Printf("Sending: %s\n", string(jsonData))

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("HTTP request error: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("HTTP status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}