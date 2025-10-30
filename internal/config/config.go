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
    Log  LogConfig  `yaml:"Log"`
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

type LogConfig struct {
    Level    string `yaml:"level"`
    Format   string `yaml:"format"`
    Output   string `yaml:"output"`
    FilePath string `yaml:"filepath"`
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
        Log: LogConfig{
            Level:    getEnv("LOG_LEVEL", "info"),
            Format:   getEnv("LOG_FORMAT", "text"),
            Output:   getEnv("LOG_OUTPUT", "stdout"),
            FilePath: getEnv("LOG_FILE_PATH", "/var/log/subscription-service.log"),
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

    if level := os.Getenv("LOG_LEVEL"); level != "" {
        config.Log.Level = level
    }

    if format := os.Getenv("LOG_FORMAT"); format != "" {
        config.Log.Format = format
    }

    if output := os.Getenv("LOG_OUTPUT"); output != "" {
        config.Log.Output = output
    }

    if filePath := os.Getenv("LOG_FILE_PATH"); filePath != "" {
        config.Log.FilePath = filePath
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
