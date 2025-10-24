FROM golang:1.21-alpine

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g cmd/server/main.go

RUN go build -o main ./cmd/server

EXPOSE 8080

CMD ["./main"]