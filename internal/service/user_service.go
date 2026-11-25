package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"taskflow/internal/model"
	"taskflow/internal/repository"

	"github.com/jackc/pgx/v5"
)

var ErrNotFound = errors.New("resource not found")
var ErrInvalidInput = errors.New("invalid input")

type UserService struct {
	Repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{Repo: repo}
}

func (s *UserService) Create(ctx context.Context, user *model.User) error{
	user.Name = strings.TrimSpace(user.Name)
	user.Email = strings.TrimSpace(user.Email)

	if user.Name == "" || len (user.Name) > 100 {
		return ErrInvalidInput
	}

	if len(user.Email) > 150 {
		return ErrInvalidInput
	}

	if _, err := mail.ParseAddress(user.Email); err != nil {
		return ErrInvalidInput
	}

	return s.Repo.Create(ctx, user)
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*model.User, error) {
	user, err := s.Repo.GetByID(ctx, id)

	if errors.Is(err, pgx.ErrNoRows) {
        return nil, ErrNotFound
    }

    if err != nil {
        return nil, err
    }

	return user, nil
}

func (s *UserService) List(ctx context.Context) ([]model.User, error) {
	return s.Repo.List(ctx)
}

func (s *UserService) ListWithTask(
	ctx context.Context,
) ([]model.UserWithTasks, error) {
	return s.Repo.ListWithTasks(ctx)
}