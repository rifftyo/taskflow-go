package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"taskflow/internal/model"
	"taskflow/internal/service"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	Service *service.TaskService
}

func NewTaskHandler(s *service.TaskService) *TaskHandler {
	return &TaskHandler{Service: s}
}

func (h *TaskHandler) Create(c *gin.Context) {
	var task model.Task

	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body",
		})
		return
	}

	err := h.Service.Create(c.Request.Context(), &task)

	if errors.Is(err, service.ErrInvalidInput) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid task data",
		})
		return
	}

	if errors.Is(err, service.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "User not found",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to create task",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Task created successfully",
		"data": task,
	})
}

func (h *TaskHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid task ID",
		})
		return
	}

	task, err := h.Service.GetById(c.Request.Context(), id)

	if errors.Is(err, service.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Task not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": task,
	})
}

func (h *TaskHandler) List(c *gin.Context) {
	tasks, err := h.Service.List(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to retrieve tasks",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tasks,
	})
}

func (h *TaskHandler) ListByUserId(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid user ID",
		})
		return
	}

	tasks, err := h.Service.ListByUserID(c.Request.Context(), id)

	if errors.Is(err, service.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "User not found",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to retrieve tasks",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tasks,
	})
}

func (h *TaskHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid task ID",
		})
		return
	}

	var task model.Task

	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body",
		})
		return
	}

	task.ID = id

	err = h.Service.Update(c.Request.Context(), &task)

	if errors.Is(err, service.ErrInvalidInput) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invallid task data",
		})
		return
	}

	if errors.Is(err, service.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Task not found",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update task",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task update successfully",
		"data": task,
	})
}

func (h *TaskHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid task ID",
		})
		return
	}

	err = h.Service.Delete(c.Request.Context(), id)

	if errors.Is(err, service.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Task not found",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to delete task",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task deleted successfully",
	})
}

func (h *TaskHandler) ListPaginated(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid page",
		})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid limit",
		})
		return
	}

	status := c.Query("status")

	userID := int64(0)

	if value := c.Query("user_id"); value != "" {
		userID, err = strconv.ParseInt(value, 10, 64)

		if err != nil || userID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid user_id",
			})
			return
		}
	}

	tasks, total, err := h.Service.ListPaginated(
		c.Request.Context(),
		page,
		limit,
		status,
		userID,
	)

	if errors.Is(err, service.ErrInvalidInput) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid filter",
		})
		return
	}

	if err != nil {
		// ==========================================
		// TAMBAHAN DEBUGGING DI SINI
		// ==========================================
		log.Printf("[ERROR] Failed to retrieve paginated tasks: %v", err)
		// ==========================================

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to retrieve tasks",
			"error":   err.Error(), // (Opsional) Membantu melihat detail error langsung di Postman/Client
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tasks,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}