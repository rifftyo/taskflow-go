package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"taskflow/internal/model"
	"taskflow/internal/repository"
	"time"

	"github.com/jackc/pgx/v5"
)

type TaskCache interface {
	Get(ctx context.Context, key string, dest any) (bool, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error 
}

type TaskService struct {
	Repo *repository.TaskRepository
	UserRepo *repository.UserRepository
	Cache TaskCache
}

func NewTaskService(
	repo *repository.TaskRepository,
	userRepo *repository.UserRepository,
	taskCache TaskCache,
) *TaskService {
	return &TaskService{
		Repo: repo,
		UserRepo: userRepo,
		Cache: taskCache,
	}
}

func (s *TaskService) Create(ctx context.Context, task *model.Task) error {
	task.Title = strings.TrimSpace(task.Title)
	task.Description = strings.TrimSpace(task.Description)

	if task.Title == "" || len(task.Title) > 200 {
		return ErrInvalidInput
	}

	if task.Status == "" {
		task.Status = "pending"
	}

	if !isValidStatus(task.Status) {
		return ErrInvalidInput
	}

	_, err := s.UserRepo.GetByID(ctx, task.UserID)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	if err != nil {
		return err
	}

	if s.Cache != nil {
		cacheKey := fmt.Sprintf("taskflow:user:%d:tasks", task.UserID)
		_ = s.Cache.Delete(ctx, cacheKey)
	}

	return s.Repo.Create(ctx, task)
}

func (s *TaskService) GetById(ctx context.Context, id int64) (*model.Task, error) {
	task, err := s.Repo.GetByID(ctx, id)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil , ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskService) List(ctx context.Context) ([]model.Task, error) {
	return s.Repo.List(ctx)
}

func (s *TaskService) ListByUserID(ctx context.Context, userID int64) ([]model.Task, error) {
	cacheKey := fmt.Sprintf("taskflow:user:%d:tasks", userID)

	var tasks []model.Task

	if s.Cache != nil {
		found, err := s.Cache.Get(ctx, cacheKey, &tasks)

		if err == nil && found {
			return tasks, nil
		}
	}

	_, err := s.UserRepo.GetByID(ctx, userID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	tasks, err = s.Repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if tasks == nil {
		tasks = []model.Task{}
	}

	if s.Cache != nil {
		_ = s.Cache.Set(ctx, cacheKey, tasks, 60*time.Second)
	}

	return tasks, nil
}

func (s *TaskService) Update(ctx context.Context, task *model.Task) error {
	task.Title = strings.TrimSpace(task.Title)
	task.Description = strings.TrimSpace(task.Description)

	if task.Title == "" || len(task.Title) > 200 {
		return ErrInvalidInput
	}

	if !isValidStatus(task.Status) {
		return ErrInvalidInput
	}

	return s.Repo.Update(ctx, task)
}

func (s *TaskService) Delete(ctx context.Context, id int64) error {
	err := s.Repo.Delete(ctx, id)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	return err
}

func isValidStatus(status string) bool {
	return status == "pending" || status == "in_progress" || status == "completed"
}

func (s *TaskService) ListPaginated(
	ctx context.Context,
	page int,
	limit int,
	status string,
	userID int64,
) ([]model.Task, int64, error) {
	if (page < 1) {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	if status != "" && !isValidStatus(status) {
		return nil, 0, ErrInvalidInput
	}

	if userID < 0 {
		return nil, 0, ErrInvalidInput
	}

	return s.Repo.ListPaginated(
		ctx,
		page,
		limit,
		status,
		userID,
	)
}