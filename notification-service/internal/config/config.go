package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// Config holds the application configuration for the Notification service
type Config struct {
	Port                string
	JobServiceGRPCAddr  string
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
	DBSSLMode           string
	RedisAddr           string
	JWTSecret           string
	SMTPHost            string
	SMTPPort            string
	SMTPUsername        string
	SMTPPassword        string
	SMTPFromEmail       string
}

// Load loads the configuration from .env file, environment variables, or defaults
func Load() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Set configuration defaults
	viper.SetDefault("PORT", "8081")
	viper.SetDefault("JOB_SERVICE_GRPC_ADDR", "localhost:50051")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5434")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "notifications")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("JWT_SECRET", "secret")

	// SMTP Defaults
	viper.SetDefault("SMTP_HOST", "localhost")
	viper.SetDefault("SMTP_PORT", "587")
	viper.SetDefault("SMTP_USERNAME", "")
	viper.SetDefault("SMTP_PASSWORD", "")
	viper.SetDefault("SMTP_FROM_EMAIL", "no-reply@jobboard.com")

	// Read config file if exists
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: No .env file found or read failed. Using defaults and system env vars: %v", err)
	}

	return &Config{
		Port:                viper.GetString("PORT"),
		JobServiceGRPCAddr:  viper.GetString("JOB_SERVICE_GRPC_ADDR"),
		DBHost:              viper.GetString("DB_HOST"),
		DBPort:              viper.GetString("DB_PORT"),
		DBUser:              viper.GetString("DB_USER"),
		DBPassword:          viper.GetString("DB_PASSWORD"),
		DBName:              viper.GetString("DB_NAME"),
		DBSSLMode:           viper.GetString("DB_SSLMODE"),
		RedisAddr:           viper.GetString("REDIS_ADDR"),
		JWTSecret:           viper.GetString("JWT_SECRET"),
		SMTPHost:            viper.GetString("SMTP_HOST"),
		SMTPPort:            viper.GetString("SMTP_PORT"),
		SMTPUsername:        viper.GetString("SMTP_USERNAME"),
		SMTPPassword:        viper.GetString("SMTP_PASSWORD"),
		SMTPFromEmail:       viper.GetString("SMTP_FROM_EMAIL"),
	}
}

// GetDBConnString formats and returns the PostgreSQL DSN connection string
func (c *Config) GetDBConnString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}
