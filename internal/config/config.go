package config

import (
    "log"
    "os"
    "strconv"

    "github.com/joho/godotenv"
    "gopkg.in/yaml.v3"
)

type Config struct {
    Server   ServerConfig   `yaml:"server"`
    Database DatabaseConfig `yaml:"database"`
    Logging  LoggingConfig  `yaml:"logging"`
}

type ServerConfig struct {
    Port int `yaml:"port"`
}

type DatabaseConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    User     string `yaml:"user"`
    Password string `yaml:"password"`
    Name     string `yaml:"name"`
    SSLMode  string `yaml:"sslmode"`
}

type LoggingConfig struct {
    Level string `yaml:"level"`
}

func LoadConfig() (*Config, error) {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    data, err := os.ReadFile("config.yaml")
    if err != nil {
        log.Printf("Error reading config.yaml: %v", err)
        return loadConfigFromEnv(), nil
    }

    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }

    overrideFromEnv(&config)

    return &config, nil
}

func loadConfigFromEnv() *Config {
    port, _ := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
    dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))

    return &Config{
        Server: ServerConfig{
            Port: port,
        },
        Database: DatabaseConfig{
            Host:     getEnv("DB_HOST", "localhost"),
            Port:     dbPort,
            User:     getEnv("DB_USER", "postgres"),
            Password: getEnv("DB_PASSWORD", "password"),
            Name:     getEnv("DB_NAME", "subscriptions"),
            SSLMode:  getEnv("DB_SSLMODE", "disable"),
        },
        Logging: LoggingConfig{
            Level: getEnv("LOG_LEVEL", "info"),
        },
    }
}

func overrideFromEnv(config *Config) {
    if port := os.Getenv("SERVER_PORT"); port != "" {
        if p, err := strconv.Atoi(port); err == nil {
            config.Server.Port = p
        }
    }

    if host := os.Getenv("DB_HOST"); host != "" {
        config.Database.Host = host
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}