package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/sirupsen/logrus"
    "subscription-service/internal/models"
    "subscription-service/internal/service"
)

type SubscriptionHandler struct {
    service service.SubscriptionService
    logger  *logrus.Logger
}

func NewSubscriptionHandler(service service.SubscriptionService, logger *logrus.Logger) *SubscriptionHandler {
    return &SubscriptionHandler{
        service: service,
        logger:  logger,
    }
}

// CreateSubscription создает новую подписку
// @Summary Создать подписку
// @Description Создает новую запись о подписке
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param input body models.CreateSubscriptionRequest true "Данные подписки"
// @Success 201 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions [post]
func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
    var req models.CreateSubscriptionRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Warnf("Invalid request body: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

    subscription := &models.Subscription{
        ServiceName:  req.ServiceName,
        Price: req.Price,
        UserID:       req.UserID,
        StartDate:    req.StartDate,
        EndDate:      req.EndDate,
    }

    if err := h.service.CreateSubscription(c.Request.Context(), subscription); err != nil {
        h.logger.Errorf("Failed to create subscription: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription"})
        return
    }

    h.logger.Infof("Subscription created successfully: %s", subscription.ID)
    c.JSON(http.StatusCreated, subscription)
}

// GetSubscription получает подписку по ID
// @Summary Получить подписку
// @Description Возвращает подписку по её ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "ID подписки"
// @Success 200 {object} models.Subscription
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        h.logger.Warnf("Invalid subscription ID: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
        return
    }

    subscription, err := h.service.GetSubscription(c.Request.Context(), id)
    if err != nil {
        h.logger.Errorf("Failed to get subscription %s: %v", id, err)
        c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
        return
    }

    c.JSON(http.StatusOK, subscription)
}

// UpdateSubscription обновляет подписку
// @Summary Обновить подписку
// @Description Обновляет данные подписки
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Param input body models.UpdateSubscriptionRequest true "Данные для обновления"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        h.logger.Warnf("Invalid subscription ID: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
        return
    }

    var req models.UpdateSubscriptionRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Warnf("Invalid request body: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

    if err := h.service.UpdateSubscription(c.Request.Context(), id, &req); err != nil {
        h.logger.Errorf("Failed to update subscription %s: %v", id, err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
        return
    }

    h.logger.Infof("Subscription updated successfully: %s", id)
    c.JSON(http.StatusOK, gin.H{"message": "Subscription updated successfully"})
}

// DeleteSubscription удаляет подписку
// @Summary Удалить подписку
// @Description Удаляет подписку по её ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "ID подписки"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        h.logger.Warnf("Invalid subscription ID: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
        return
    }

    if err := h.service.DeleteSubscription(c.Request.Context(), id); err != nil {
        h.logger.Errorf("Failed to delete subscription %s: %v", id, err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subscription"})
        return
    }

    h.logger.Infof("Subscription deleted successfully: %s", id)
    c.JSON(http.StatusOK, gin.H{"message": "Subscription deleted successfully"})
}

// ListSubscriptions возвращает список подписок
// @Summary Список подписок
// @Description Возвращает список подписок с возможностью фильтрации
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "ID пользователя"
// @Param service_name query string false "Название сервиса"
// @Success 200 {array} models.Subscription
// @Failure 500 {object} map[string]string
// @Router /subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
    var userID *uuid.UUID
    var serviceName *string

    if userIDStr := c.Query("user_id"); userIDStr != "" {
        if id, err := uuid.Parse(userIDStr); err == nil {
            userID = &id
        }
    }

    if serviceNameStr := c.Query("service_name"); serviceNameStr != "" {
        serviceName = &serviceNameStr
    }

    subscriptions, err := h.service.ListSubscriptions(c.Request.Context(), userID, serviceName)
    if err != nil {
        h.logger.Errorf("Failed to list subscriptions: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list subscriptions"})
        return
    }

    c.JSON(http.StatusOK, subscriptions)
}

// GetSummary возвращает суммарную стоимость подписок за период
// @Summary Сумма подписок
// @Description Возвращает суммарную стоимость подписок за указанный период
// @Tags subscriptions
// @Produce json
// @Param start_date query string true "Начальная дата (YYYY-MM-DD)"
// @Param end_date query string true "Конечная дата (YYYY-MM-DD)"
// @Param user_id query string false "ID пользователя"
// @Param service_name query string false "Название сервиса"
// @Success 200 {object} models.SubscriptionSummary
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/summary [get]
func (h *SubscriptionHandler) GetSummary(c *gin.Context) {
    var req models.SummaryRequest

    if startDateStr := c.Query("start_date"); startDateStr != "" {
        startDate, err := time.Parse("2006-01-02", startDateStr)
        if err != nil {
            h.logger.Warnf("Invalid start_date: %v", err)
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD"})
            return
        }
        req.StartDate = &startDate
    }

    if endDateStr := c.Query("end_date"); endDateStr != "" {
        endDate, err := time.Parse("2006-01-02", endDateStr)
        if err != nil {
            h.logger.Warnf("Invalid end_date: %v", err)
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD"})
            return
        }
        req.EndDate = &endDate
    }

    if userIDStr := c.Query("user_id"); userIDStr != "" {
        if id, err := uuid.Parse(userIDStr); err == nil {
            req.UserID = &id
        }
    }

    if serviceNameStr := c.Query("service_name"); serviceNameStr != "" {
        req.ServiceName = &serviceNameStr
    }

    summary, err := h.service.GetSummary(c.Request.Context(), &req)
    if err != nil {
        h.logger.Errorf("Failed to get summary: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate summary"})
        return
    }

    c.JSON(http.StatusOK, summary)
}