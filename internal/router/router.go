package router

import (
	"net/http"

	"taskflow/internal/handler"
	"taskflow/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(
	userHandler *handler.UserHandler,
	taskHandler *handler.TaskHandler,
	authHandler *handler.AuthHandler,
	jwtSecret []byte,
) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "TaskFlow API is running",
			"status":  "health",
		})
	})

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(jwtSecret))
		{
			users := protected.Group("/users")
			{
				users.POST("", userHandler.Create)
				users.GET("", userHandler.List)
				users.GET("/with-tasks", userHandler.ListWithTasks)
				users.GET("/:id", userHandler.GetByID)
				users.GET("/:id/tasks", taskHandler.ListByUserId)
			}

			tasks := protected.Group("/tasks")
			{
				tasks.POST("", taskHandler.Create)
				tasks.GET("", taskHandler.List)
				tasks.GET("/paginated", taskHandler.ListPaginated)
				tasks.GET("/:id", taskHandler.GetByID)
				tasks.PUT("/:id", taskHandler.Update)
				tasks.DELETE("/:id", taskHandler.Delete)
			}
		}
	}

	return r
}