# subscription-service

# Сборка и запуск
make up

make run-fill

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

curl -s "http://localhost:8080/api/v1/subscriptions" | jq '.'

# c фильтрацией

curl -s "http://localhost:8080/api/v1/subscriptions" | jq '.[] | {id, service_name, price}'

# Получить только общую стоимость

curl -s "http://localhost:8080/api/v1/subscriptions/summary?start_date=2025-01-01&end_date=2025-12-31" | jq '.total_cost'

# Неправильный UUID
curl -X GET "http://localhost:8080/api/v1/subscriptions/invalid-uuid"

output: {"error":"Invalid subscription ID"}

# Обновление подписки (PUT /api/v1/subscriptions/:id)

# Обновить цену подписки
curl -X PUT http://localhost:8080/api/v1/subscriptions/3d49eb54-b321-4426-95f7-be49268154cf \
  -H "Content-Type: application/json" \
  -d '{ "price": 699.00 }'

# Обновить несколько полей
curl -X PUT http://localhost:8080/api/v1/subscriptions/3d49eb54-b321-4426-95f7-be49268154cf \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Netflix Premium",
    "price": 799.00,
    "end_date": "2024-06-30T23:59:59Z"
  }'

# Удалить подписку
curl -X DELETE "http://localhost:8080/api/v1/subscriptions/123e4567-e89b-12d3-a456-426614174000

# Просмотр Swagger документации

http://localhost:8080/swagger/index.html


# Подсчет суммарной стоимости всех подписок за выбранный период с фильтрацией по id пользователя и названию подписки
curl "http://localhost:8080/api/v1/subscriptions/summary?user_id=123e4567-e89b-12d3-a456-426614174000&service_name=Spotify&start_date=2024-01-01&end_date=2024-12-31"

# Test

make test-all
