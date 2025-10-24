# subscription-service

# Сборка и запуск
docker-compose up --build

# Только запуск
docker-compose up

# Запуск в фоне
docker-compose up -d

# Создание подписки
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 399.00,
    "user_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "start_date": "2024-01-01T00:00:00Z"
  }'


# Получение подписок
curl http://localhost:8080/api/v1/subscriptions"

# Просмотр Swagger документации

http://localhost:8080/swagger/index.html


# Подсчет суммарной стоимости всех подписок за выбранный период с фильтрацией по id пользователя и названию подписки
curl "http://localhost:8080/api/v1/subscriptions/summary?user_id=123e4567-e89b-12d3-a456-426614174000&service_name=Spotify&start_date=2024-01-01&end_date=2024-12-31"
