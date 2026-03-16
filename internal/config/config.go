package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	MongoURI    string
	MongoDB     string
	RedisURI    string
	RabbitMQURI string
	JWTSecret   string
	WSPort      string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	return &Config{
		Port:        os.Getenv("PORT"),
		MongoURI:    os.Getenv("MONGO_URI"),
		MongoDB:     os.Getenv("MONGO_DB"),
		RedisURI:    os.Getenv("REDIS_URI"),
		RabbitMQURI: os.Getenv("RABBITMQ_URI"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		WSPort:      os.Getenv("WS_PORT"),
	}
}
