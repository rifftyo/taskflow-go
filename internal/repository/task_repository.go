package repository

import (
	"context"
	"errors"
	"taskflow/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepository struct {
	DB *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{DB: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *model.Task) error {
	query := `
		INSERT INTO tasks (user_id, title, description, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	return r.DB.QueryRow(
		ctx,
		query,
		task.UserID,
		task.Title,
		task.Description,
		task.Status,
	).Scan(
		&task.ID,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
}

func (r *TaskRepository) GetByID(ctx context.Context, id int64) (*model.Task, error) {
	query := `
		SELECT id, user_id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	var task model.Task

	err := r.DB.QueryRow(ctx, query, id).Scan(
		&task.ID,
		&task.UserID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, pgx.ErrNoRows
	}

	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *model.Task) error {
	query := `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`

	err := r.DB.QueryRow(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.ID,
	).Scan(&task.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return pgx.ErrNoRows
	}

	return err
}

func (r *TaskRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.DB.Exec(
		ctx,
		"DELETE FROM tasks WHERE id = $1",
		id,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *TaskRepository) List(ctx context.Context) ([]model.Task, error) {
	query := `
		SELECT id, user_id, title, description, status, created_at, updated_at
		FROM tasks
		ORDER BY id DESC
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]model.Task, 0)

	for rows.Next() {
		var task model.Task

		err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *TaskRepository) ListByUserID(ctx context.Context, userID int64) ([]model.Task, error) {
	query := `
		SELECT id, user_id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE user_id = $1
		ORDER BY id DESC
	`

	rows, err := r.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]model.Task, 0)

	for rows.Next() {
		var task model.Task

		err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return tasks, nil
}

func (r *TaskRepository) ListPaginated(
	ctx context.Context,
	page int,
	limit int,
	status string,
	userID int64,
) ([]model.Task, int64, error) {
	where := `WHERE ($1 = '' OR status = $1)
		AND ($2 = 0 OR user_id = $2)`

	countQuery := `
		SELECT COUNT(*)
		FROM tasks
	` + where

	var total int64

	err := r.DB.QueryRow(
		ctx,
		countQuery,
		status,
		userID,
	).Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, title, description,
			status, created_at, updated_at
		FROM tasks
		` + where + `
		ORDER BY id DESC
		LIMIT $3 OFFSET $4
	`

	offset := (page - 1) * limit

	rows, err := r.DB.Query(
		ctx,
		query,
		status,
		userID,
		limit,
		offset,
	)

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tasks := make([]model.Task, 0)

	for rows.Next() {
		var task model.Task

		err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}