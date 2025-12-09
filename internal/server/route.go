package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func CreateRoute(api fiber.Router, db *gorm.DB, redisClient *redis.Client, cfg *Cfg) {

}
