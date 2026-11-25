package main

import (
	"log"
	"os"
	"time"

	"taskflow/internal/cache"
	"taskflow/internal/config"
	"taskflow/internal/handler"
	"taskflow/internal/repository"
	"taskflow/internal/router"
	"taskflow/internal/service"
)

func main() {
	db, err := config.ConnectDatabase()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	jwtExpiresIn := 24 * time.Hour
	if rawExp := os.Getenv("JWT_EXPIRES_IN"); rawExp != "" {
		parsed, err := time.ParseDuration(rawExp)
		if err != nil {
			log.Fatal("Invalid JWT_EXPIRES_IN:", err)
		}
		jwtExpiresIn = parsed
	}

	jwtConfig := &config.JWTConfig{
		Secret:    []byte(jwtSecret),
		ExpiresIn: jwtExpiresIn,
	}

	redisClient, err := config.ConnectRedis()
	if err != nil {
		log.Fatal("Failed to connect Redis: ", err)
	}
	defer redisClient.Close()

	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	authRepository := repository.NewAuthRepository(db)

	redisCache := cache.NewRedisCache(redisClient)

	userService := service.NewUserService(userRepo)
	taskService := service.NewTaskService(taskRepo, userRepo, redisCache)
	authService := service.NewAuthService(authRepository, jwtConfig)

	userHandler := handler.NewUserHandler(userService)
	taskHandler := handler.NewTaskHandler(taskService)
	authHandler := handler.NewAuthHandler(authService)

	r := router.Setup(userHandler, taskHandler, authHandler, jwtConfig.Secret)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8081"
	}

	log.Println("Server running on port", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Server failed:", err)
	}
}
