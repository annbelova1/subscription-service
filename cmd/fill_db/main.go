package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Subscription struct {
	ServiceName string    `json:"service_name"`
	Price       float64   `json:"price"`
	UserID      string    `json:"user_id"`
	StartDate   time.Time `json:"start_date"`
}

func main() {
	baseURL := "http://localhost:8080/api/v1/subscriptions"
	userID := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

	subscriptions := []Subscription{
		{"Yandex Plus", 399.00, userID, time.Now().AddDate(0, 0, -30)},
		{"Netflix", 799.00, userID, time.Now().AddDate(0, 0, -15)},
		{"Spotify", 299.00, userID, time.Now().AddDate(0, 0, -10)},
		{"YouTube Premium", 459.00, userID, time.Now().AddDate(0, 0, -5)},
		{"Apple Music", 199.00, userID, time.Now().AddDate(0, 0, -20)},
		{"Amazon Prime", 299.00, userID, time.Now().AddDate(0, 0, -25)},
		{"Microsoft 365", 399.00, userID, time.Now().AddDate(0, 0, -8)},
		{"PlayStation Plus", 599.00, userID, time.Now().AddDate(0, 0, -12)},
		{"Xbox Game Pass", 349.00, userID, time.Now().AddDate(0, 0, -18)},
		{"Telegram Premium", 299.00, userID, time.Now().AddDate(0, 0, -3)},
	}

	for _, sub := range subscriptions {
		if err := createSubscription(baseURL, sub); err != nil {
			log.Printf("Error creating subscription %s: %v", sub.ServiceName, err)
			continue
		}
		fmt.Printf("✓ Created subscription: %s (%.2f руб)\n", sub.ServiceName, sub.Price)
		time.Sleep(100 * time.Millisecond) // небольшая пауза
	}

	fmt.Println("All subscriptions created successfully!")
}

func createSubscription(url string, sub Subscription) error {
	jsonData, err := json.Marshal(sub)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("HTTP request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	return nil
}